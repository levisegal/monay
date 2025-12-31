package plaid

import (
	"context"
	"fmt"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/plaid/plaid-go/v29/plaid"
)

type Client struct {
	api         *plaid.APIClient
	redirectURI string
}

func NewClient(cfg config.Plaid) *Client {
	plaidCfg := plaid.NewConfiguration()
	plaidCfg.AddDefaultHeader("PLAID-CLIENT-ID", cfg.ClientID)
	plaidCfg.AddDefaultHeader("PLAID-SECRET", cfg.Secret)

	switch cfg.Env {
	case "production":
		plaidCfg.UseEnvironment(plaid.Production)
	default:
		plaidCfg.UseEnvironment(plaid.Sandbox)
	}

	return &Client{
		api:         plaid.NewAPIClient(plaidCfg),
		redirectURI: cfg.RedirectURI,
	}
}

func (c *Client) GetHoldings(ctx context.Context, accessToken string) (*plaid.InvestmentsHoldingsGetResponse, error) {
	req := plaid.NewInvestmentsHoldingsGetRequest(accessToken)
	resp, _, err := c.api.PlaidApi.InvestmentsHoldingsGet(ctx).InvestmentsHoldingsGetRequest(*req).Execute()
	if err != nil {
		return nil, plaidError(err)
	}
	return &resp, nil
}

func (c *Client) ExchangePublicToken(ctx context.Context, publicToken string) (string, error) {
	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := c.api.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return "", plaidError(err)
	}
	return resp.GetAccessToken(), nil
}

func (c *Client) API() *plaid.APIClient {
	return c.api
}

func plaidError(err error) error {
	if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
		return fmt.Errorf("%s: %s", plaidErr.Error(), string(plaidErr.Body()))
	}
	return err
}
