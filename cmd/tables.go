package cmd

import (
	"context"
	"fmt"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var tablesOutputCSV string

// TablesCmd is the `tables` subcommand.
var TablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "List all storage tables across storage accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving tables (subscription %s)...\n", e.SubscriptionID)
		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}

		headers := []string{"Account", "Table"}
		var rows [][]string
		for _, acc := range accounts {
			tables, err := e.ListTables(ctx, acc.ResourceGroup, acc.Name)
			if err != nil {
				continue
			}
			for _, t := range tables {
				rows = append(rows, []string{acc.Name, t})
			}
		}

		if len(rows) == 0 {
			fmt.Println("No tables found.")
			return nil
		}

		if tablesOutputCSV != "" {
			if err := output.WriteCSV(tablesOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Table data exported to %s\n", tablesOutputCSV)
		} else {
			fmt.Println("\nStorage Tables:")
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	TablesCmd.Flags().StringVar(&tablesOutputCSV, "csv", "", "Export to CSV file")
}
