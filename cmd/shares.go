package cmd

import (
	"context"
	"fmt"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var sharesOutputCSV string

// SharesCmd is the `shares` subcommand.
var SharesCmd = &cobra.Command{
	Use:   "shares",
	Short: "List all Azure file shares across storage accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving file shares (subscription %s)...\n", e.SubscriptionID)
		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}

		headers := []string{"Account", "Share", "Quota (GB)", "Used", "Access Tier"}
		var rows [][]string
		for _, acc := range accounts {
			shares, err := e.ListFileShares(ctx, acc.ResourceGroup, acc.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, s := range shares {
				rows = append(rows, []string{
					acc.Name,
					s.Name,
					fmt.Sprintf("%d", s.QuotaGB),
					explorer.FormatBytes(float64(s.UsedBytes)),
					s.AccessTier,
				})
			}
		}

		if len(rows) == 0 {
			fmt.Println("No file shares found.")
			return nil
		}

		if sharesOutputCSV != "" {
			if err := output.WriteCSV(sharesOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Share data exported to %s\n", sharesOutputCSV)
		} else {
			fmt.Println("\nFile Shares:")
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	SharesCmd.Flags().StringVar(&sharesOutputCSV, "csv", "", "Export to CSV file")
}
