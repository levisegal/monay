package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/plaid"
	plaidgo "github.com/plaid/plaid-go/v29/plaid"
	"github.com/spf13/cobra"
)

func plaidCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plaid",
		Short: "Plaid integration commands",
	}

	cmd.AddCommand(plaidLinkCommand())
	cmd.AddCommand(plaidHoldingsCommand())

	return cmd
}

func plaidLinkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Run Plaid Link flow to connect an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client := plaid.NewClient(cfg.Plaid)

			fmt.Println("[1] Creating link token...")
			userID := fmt.Sprintf("user-%d", time.Now().Unix())
			linkToken, err := client.CreateLinkToken(ctx, userID)
			if err != nil {
				return fmt.Errorf("create link token: %w", err)
			}

			fmt.Println("[2] Open http://localhost:8080 in your browser")
			fmt.Println("    For sandbox, use test credentials: user_good / pass_good")
			fmt.Println()
			fmt.Println("    Waiting for connection...")

			publicToken, err := plaid.ServeLinkAndWait(linkToken, 5*time.Minute)
			if err != nil {
				return fmt.Errorf("link flow: %w", err)
			}

			fmt.Println("[3] Exchanging public token for access token...")
			accessToken, err := client.ExchangePublicToken(ctx, publicToken)
			if err != nil {
				return fmt.Errorf("exchange token: %w", err)
			}

			if err := saveAccessToken(accessToken); err != nil {
				fmt.Printf("Warning: could not save access token: %v\n", err)
			}

			fmt.Println("[4] Success! Access token saved to .access_token")
			return nil
		},
	}
}

func plaidHoldingsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "holdings",
		Short: "Fetch investment holdings from Plaid",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client := plaid.NewClient(cfg.Plaid)

			accessToken, err := loadAccessToken()
			if err != nil {
				return fmt.Errorf("no access token found, run 'holdings plaid link' first: %w", err)
			}

			fmt.Println("Fetching investment holdings...")
			resp, err := client.GetHoldings(ctx, accessToken)
			if err != nil {
				return fmt.Errorf("get holdings: %w", err)
			}

			printHoldings(resp)
			return nil
		},
	}
}

func printHoldings(resp *plaidgo.InvestmentsHoldingsGetResponse) {
	fmt.Printf("\nAccounts (%d):\n", len(resp.Accounts))
	for _, acct := range resp.Accounts {
		fmt.Printf("  - %s (%s): $%.2f\n", acct.GetName(), acct.GetType(), acct.GetBalances().Current.Get())
	}

	fmt.Printf("\nHoldings (%d):\n", len(resp.Holdings))

	securities := make(map[string]plaidgo.Security)
	for _, sec := range resp.Securities {
		securities[sec.GetSecurityId()] = sec
	}

	for _, h := range resp.Holdings {
		sec := securities[h.GetSecurityId()]
		symbol := sec.GetTickerSymbol()
		if symbol == "" {
			symbol = sec.GetName()
		}
		fmt.Printf("  %s: %.4f units @ $%.2f = $%.2f\n",
			symbol,
			h.GetQuantity(),
			h.GetInstitutionPrice(),
			h.GetInstitutionValue(),
		)
	}

	fmt.Println("\n--- Raw JSON ---")
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func loadAccessToken() (string, error) {
	data, err := os.ReadFile(".access_token")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func saveAccessToken(token string) error {
	return os.WriteFile(".access_token", []byte(token), 0600)
}
