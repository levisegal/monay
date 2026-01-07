package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func accountsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts",
	}

	cmd.AddCommand(accountsListCommand())
	cmd.AddCommand(accountsRenameCommand())
	cmd.AddCommand(accountsDeleteCommand())

	return cmd
}

func accountsDeleteCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			conn, err := database.Open(ctx, cfg.Database.ConnString())
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)

			account, err := queries.GetAccountByName(ctx, name)
			if err != nil {
				return fmt.Errorf("account %q not found: %w", name, err)
			}

			if err := queries.DeleteAccount(ctx, account.ID); err != nil {
				return fmt.Errorf("failed to delete account: %w", err)
			}

			fmt.Printf("Deleted account %q\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Account name to delete")
	cmd.MarkFlagRequired("name")

	return cmd
}

func accountsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			conn, err := database.Open(ctx, cfg.Database.ConnString())
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)
			accounts, err := queries.ListAccounts(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("\n%-40s %-20s %-15s %s\n", "ID", "Name", "Institution", "External #")
			fmt.Printf("%-40s %-20s %-15s %s\n", "----", "----", "-----------", "----------")
			for _, a := range accounts {
				fmt.Printf("%-40s %-20s %-15s %s\n",
					a.ID,
					a.Name,
					a.InstitutionName,
					a.ExternalAccountNumber.String,
				)
			}
			return nil
		},
	}
}

func accountsRenameCommand() *cobra.Command {
	var oldName, newName string

	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			conn, err := database.Open(ctx, cfg.Database.ConnString())
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)

			account, err := queries.GetAccountByName(ctx, oldName)
			if err != nil {
				return fmt.Errorf("account %q not found: %w", oldName, err)
			}

			updated, err := queries.UpdateAccount(ctx, db.UpdateAccountParams{
				ID:   account.ID,
				Name: newName,
			})
			if err != nil {
				return fmt.Errorf("failed to rename account: %w", err)
			}

			fmt.Printf("Renamed %q -> %q\n", oldName, updated.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&oldName, "old", "", "Current account name")
	cmd.Flags().StringVar(&newName, "new", "", "New account name")
	cmd.MarkFlagRequired("old")
	cmd.MarkFlagRequired("new")

	return cmd
}

