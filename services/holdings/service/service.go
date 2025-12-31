package service

import (
	"context"
	"os"
	"sync"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/levisegal/monay/services/holdings/config"
	monayv1beta1 "github.com/levisegal/monay/services/holdings/gen/api/monay/v1beta1"
	"github.com/levisegal/monay/services/holdings/plaid"
	"github.com/levisegal/monay/services/holdings/version"
)

type HoldingsService struct {
	config      *config.Config
	plaidClient *plaid.Client

	mu          sync.RWMutex
	accessToken string
}

func New(cfg *config.Config, plaidClient *plaid.Client) *HoldingsService {
	svc := &HoldingsService{
		config:      cfg,
		plaidClient: plaidClient,
	}
	svc.loadAccessToken()
	return svc
}

func (s *HoldingsService) GetVersion(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
) (*connect.Response[monayv1beta1.GetVersionResponse], error) {
	v, revision := version.GetReleaseInfo()
	return connect.NewResponse(&monayv1beta1.GetVersionResponse{
		Version:  v,
		Revision: revision,
	}), nil
}

// loadAccessToken reads the Plaid access token from disk.
// The access token is returned by Plaid after a user completes Link and is
// used to fetch that user's data. In production, store this in the database
// tied to a user record. File storage is a dev-only workaround.
func (s *HoldingsService) loadAccessToken() {
	data, err := os.ReadFile(".access_token")
	if err == nil {
		s.accessToken = string(data)
	}
}

func (s *HoldingsService) saveAccessToken(token string) {
	os.WriteFile(".access_token", []byte(token), 0600)
}
