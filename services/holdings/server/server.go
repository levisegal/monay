package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/version"
)

func Start(ctx context.Context, cfg *config.Config) error {
	slog.Info("Starting Monay Holdings API server...")
	ver, rev := version.GetReleaseInfo()
	slog.Info("  Version", "version", ver)
	slog.Info("  Revision", "revision", rev)

	conn, err := database.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	queries := db.New(conn)
	router := NewRouter(queries)

	webListener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}

	webServer := &http.Server{
		Handler: h2c.NewHandler(router, &http2.Server{}),
		ErrorLog: log.New(
			slogErrorWriter{logger: slog.With(slog.String("component", "http"))},
			"",
			0,
		),
	}

	g, gctx := errgroup.WithContext(ctx)
	go func() {
		<-gctx.Done()
		slog.Info("Shutting down the server...")
		shutdownHttpServer(webServer)
	}()

	g.Go(func() error {
		slog.Info("API server is running", "listen_addr", cfg.ListenAddr)
		return serveHttp(webServer, webListener)
	})

	return g.Wait()
}

func serveHttp(s *http.Server, l net.Listener) error {
	if l == nil || s == nil {
		return nil
	}
	if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func shutdownHttpServer(s *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}

type slogErrorWriter struct {
	logger *slog.Logger
}

func (w slogErrorWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	w.logger.Error(msg)
	return len(p), nil
}
