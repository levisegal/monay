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

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/gen/api/monay/v1beta1/monayv1beta1connect"
	"github.com/levisegal/monay/services/holdings/plaid"
	"github.com/levisegal/monay/services/holdings/service"
	"github.com/levisegal/monay/services/holdings/version"
)

func Start(ctx context.Context, cfg *config.Config) error {
	slog.Info("Starting Monay Holdings API server...")
	ver, rev := version.GetReleaseInfo()
	slog.Info("  Version", "version", ver)
	slog.Info("  Revision", "revision", rev)

	slog.Info("Plaid config",
		"client_id_set", cfg.Plaid.ClientID != "",
		"secret_set", cfg.Plaid.Secret != "",
		"env", cfg.Plaid.Env,
	)
	plaidClient := plaid.NewClient(cfg.Plaid)
	svc := service.New(cfg, plaidClient)

	interceptors := connect.WithInterceptors(
		NewLoggingInterceptor(),
	)

	apiMux := http.NewServeMux()

	apiMux.Handle(monayv1beta1connect.NewPlaidServiceHandler(svc, interceptors))

	// Health check
	checker := grpchealth.NewStaticChecker(
		monayv1beta1connect.PlaidServiceName,
	)
	apiMux.Handle(grpchealth.NewHandler(checker))

	// Reflection
	reflector := grpcreflect.NewStaticReflector(
		monayv1beta1connect.PlaidServiceName,
	)
	apiMux.Handle(grpcreflect.NewHandlerV1(reflector))
	apiMux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Serve test page at root - temporary dev UI, will be replaced by Next.js frontend
	apiMux.HandleFunc("/", serveTestPage)

	webListener, err := net.Listen("tcp", cfg.ListenAddr)
	slog.Info("Plaid environment", "env", cfg.Plaid.Env)
	if err != nil {
		return err
	}

	webServer := &http.Server{
		Handler: h2c.NewHandler(apiMux, &http2.Server{}),
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

func serveTestPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(testPageHTML))
}

// TODO: Delete this HTML and serveTestPage when Next.js frontend is ready.
// This is a temporary dev-only UI for testing the Plaid integration.
const testPageHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Monay - Plaid Test</title>
  <script src="https://cdn.plaid.com/link/v2/stable/link-initialize.js"></script>
  <style>
    * { box-sizing: border-box; }
    body {
      font-family: system-ui, -apple-system, sans-serif;
      max-width: 800px;
      margin: 0 auto;
      padding: 40px 20px;
      background: #0a0a0a;
      color: #e5e5e5;
    }
    h1 { color: #fff; margin-bottom: 8px; }
    .subtitle { color: #888; margin-bottom: 32px; }
    button {
      background: #2563eb;
      color: white;
      border: none;
      padding: 12px 24px;
      font-size: 16px;
      border-radius: 8px;
      cursor: pointer;
      margin-right: 12px;
      margin-bottom: 12px;
    }
    button:hover { background: #1d4ed8; }
    button:disabled { background: #374151; cursor: not-allowed; }
    #status {
      margin-top: 24px;
      padding: 16px;
      background: #1a1a1a;
      border-radius: 8px;
      white-space: pre-wrap;
      font-family: monospace;
      font-size: 14px;
    }
    .success { color: #22c55e; }
    .error { color: #ef4444; }
  </style>
</head>
<body>
  <h1>Monay</h1>
  <p class="subtitle">Plaid Integration Test</p>
  
  <button id="connect-btn">Connect Account</button>
  <button id="holdings-btn" disabled>Fetch Holdings</button>
  
  <div id="status">Ready. Click "Connect Account" to start.</div>

  <script>
    const statusEl = document.getElementById('status');
    const connectBtn = document.getElementById('connect-btn');
    const holdingsBtn = document.getElementById('holdings-btn');
    
    function log(msg, type) {
      const line = document.createElement('div');
      line.textContent = msg;
      if (type) line.className = type;
      statusEl.appendChild(line);
      statusEl.scrollTop = statusEl.scrollHeight;
    }
    
    async function connectAccount() {
      statusEl.innerHTML = '';
      log('Creating link token...');
      
      try {
        const resp = await fetch('/monay.v1beta1.PlaidService/CreateLinkToken', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ userId: 'test-user-' + Date.now() })
        });
        
        if (!resp.ok) throw new Error(await resp.text());
        const data = await resp.json();
        log('Got link token: ' + data.linkToken.substring(0, 30) + '...');
        
        const handler = Plaid.create({
          token: data.linkToken,
          onSuccess: async (publicToken, metadata) => {
            log('Connected! Institution: ' + metadata.institution.name, 'success');
            log('Exchanging public token...');
            
            const exchangeResp = await fetch('/monay.v1beta1.PlaidService/ExchangePublicToken', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ publicToken: publicToken })
            });
            
            if (!exchangeResp.ok) throw new Error(await exchangeResp.text());
            log('Access token stored!', 'success');
            holdingsBtn.disabled = false;
          },
          onExit: (err, metadata) => {
            if (err) log('Link error: ' + err.display_message, 'error');
          },
        });
        
        handler.open();
      } catch (err) {
        log('Error: ' + err.message, 'error');
      }
    }
    
    async function fetchHoldings() {
      log('Fetching holdings...');
      
      try {
        const resp = await fetch('/monay.v1beta1.PlaidService/GetHoldings', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({})
        });
        
        if (!resp.ok) throw new Error(await resp.text());
        const data = await resp.json();
        
        log('\nAccounts:', 'success');
        for (const acct of data.accounts || []) {
          log('  ' + acct.name + ' (' + acct.type + '): $' + acct.balance.toFixed(2));
        }
        
        log('\nHoldings:', 'success');
        for (const h of data.holdings || []) {
          const qty = (h.quantity || 0).toFixed(4);
          const price = (h.price || 0).toFixed(2);
          const value = (h.value || 0).toFixed(2);
          log('  ' + (h.symbol || h.name || 'Unknown') + ': ' + qty + ' @ $' + price + ' = $' + value);
        }
      } catch (err) {
        log('Error: ' + err.message, 'error');
      }
    }
    
    connectBtn.addEventListener('click', connectAccount);
    holdingsBtn.addEventListener('click', fetchHoldings);
  </script>
</body>
</html>`
