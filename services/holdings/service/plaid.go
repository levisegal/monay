package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/plaid/plaid-go/v29/plaid"

	monayv1beta1 "github.com/levisegal/monay/services/holdings/gen/api/monay/v1beta1"
)

func (s *HoldingsService) CreateLinkToken(
	ctx context.Context,
	req *connect.Request[monayv1beta1.CreateLinkTokenRequest],
) (*connect.Response[monayv1beta1.CreateLinkTokenResponse], error) {
	userID := req.Msg.UserId
	if userID == "" {
		userID = "default-user"
	}

	token, err := s.plaidClient.CreateLinkToken(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&monayv1beta1.CreateLinkTokenResponse{
		LinkToken: token,
	}), nil
}

func (s *HoldingsService) ExchangePublicToken(
	ctx context.Context,
	req *connect.Request[monayv1beta1.ExchangePublicTokenRequest],
) (*connect.Response[monayv1beta1.ExchangePublicTokenResponse], error) {
	accessToken, err := s.plaidClient.ExchangePublicToken(ctx, req.Msg.PublicToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	s.mu.Lock()
	s.accessToken = accessToken
	s.mu.Unlock()

	s.saveAccessToken(accessToken)

	return connect.NewResponse(&monayv1beta1.ExchangePublicTokenResponse{}), nil
}

func (s *HoldingsService) GetHoldings(
	ctx context.Context,
	req *connect.Request[monayv1beta1.GetHoldingsRequest],
) (*connect.Response[monayv1beta1.GetHoldingsResponse], error) {
	s.mu.RLock()
	token := s.accessToken
	s.mu.RUnlock()

	if token == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("no access token, connect an account first"))
	}

	resp, err := s.plaidClient.GetHoldings(ctx, token)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	securities := buildSecurityMaps(resp.Securities)

	return connect.NewResponse(&monayv1beta1.GetHoldingsResponse{
		Accounts: buildAccounts(resp.Accounts),
		Holdings: buildHoldings(resp.Holdings, securities),
	}), nil
}

type securityMaps struct {
	symbols map[string]string
	names   map[string]string
}

func buildSecurityMaps(securities []plaid.Security) securityMaps {
	m := securityMaps{
		symbols: make(map[string]string),
		names:   make(map[string]string),
	}
	for _, sec := range securities {
		m.symbols[sec.GetSecurityId()] = sec.GetTickerSymbol()
		m.names[sec.GetSecurityId()] = sec.GetName()
	}
	return m
}

func buildAccounts(accounts []plaid.AccountBase) []*monayv1beta1.Account {
	result := make([]*monayv1beta1.Account, 0, len(accounts))
	for _, acct := range accounts {
		balance := 0.0
		if b := acct.GetBalances().Current.Get(); b != nil {
			balance = *b
		}
		result = append(result, &monayv1beta1.Account{
			AccountId: acct.GetAccountId(),
			Name:      acct.GetName(),
			Type:      string(acct.GetType()),
			Subtype:   string(acct.GetSubtype()),
			Balance:   balance,
		})
	}
	return result
}

func buildHoldings(holdings []plaid.Holding, sec securityMaps) []*monayv1beta1.Holding {
	result := make([]*monayv1beta1.Holding, 0, len(holdings))
	for _, h := range holdings {
		symbol := sec.symbols[h.GetSecurityId()]
		name := sec.names[h.GetSecurityId()]
		if symbol == "" {
			symbol = name
		}
		result = append(result, &monayv1beta1.Holding{
			AccountId:  h.GetAccountId(),
			SecurityId: h.GetSecurityId(),
			Symbol:     symbol,
			Name:       name,
			Quantity:   h.GetQuantity(),
			Price:      h.GetInstitutionPrice(),
			Value:      h.GetInstitutionValue(),
		})
	}
	return result
}
