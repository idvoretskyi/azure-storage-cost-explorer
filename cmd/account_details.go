package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var (
	accountDetailsOutputCSV string
)

// AccountDetailsCmd is the `account-details` subcommand.
var AccountDetailsCmd = &cobra.Command{
	Use:   "account-details <storage-account>",
	Short: "Get detailed capacity breakdown for a specific storage account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}
		var match *explorer.StorageAccount
		for i, a := range accounts {
			if strings.EqualFold(a.Name, accountName) {
				match = &accounts[i]
				break
			}
		}
		if match == nil {
			return fmt.Errorf("storage account %q not found in subscription %s", accountName, e.SubscriptionID)
		}

		fmt.Printf("Retrieving detailed information for account: %s\n", match.Name)
		cap, err := e.GetAccountCapacity(ctx, match.ResourceID)
		if err != nil {
			return err
		}

		headers := []string{"Service / Tier", "Size"}
		var rows [][]string
		var total float64

		// Blob with per-tier breakdown when available
		rows = append(rows, []string{"Blob (total)", explorer.FormatBytes(cap.BlobCapacityBytes)})
		total += cap.BlobCapacityBytes
		if len(cap.BlobByTier) > 0 {
			// Sort tiers by canonical AccessTiers order
			for _, tier := range explorer.AccessTiers {
				if v, ok := cap.BlobByTier[tier]; ok {
					rows = append(rows, []string{"  Blob: " + tier, explorer.FormatBytes(v)})
				}
			}
			// Anything else (Unknown etc.) sorted alphabetically
			var extras []string
			canonical := map[string]bool{}
			for _, t := range explorer.AccessTiers {
				canonical[t] = true
			}
			for k := range cap.BlobByTier {
				if !canonical[k] {
					extras = append(extras, k)
				}
			}
			sort.Strings(extras)
			for _, k := range extras {
				rows = append(rows, []string{"  Blob: " + k, explorer.FormatBytes(cap.BlobByTier[k])})
			}
		}
		rows = append(rows, []string{"File", explorer.FormatBytes(cap.FileCapacityBytes)})
		total += cap.FileCapacityBytes
		rows = append(rows, []string{"Queue", explorer.FormatBytes(cap.QueueCapacityBytes)})
		total += cap.QueueCapacityBytes
		rows = append(rows, []string{"Table", explorer.FormatBytes(cap.TableCapacityBytes)})
		total += cap.TableCapacityBytes

		if accountDetailsOutputCSV != "" {
			csvRows := append(rows, []string{"Total", explorer.FormatBytes(total)})
			if err := output.WriteCSV(accountDetailsOutputCSV, headers, csvRows); err != nil {
				return err
			}
			fmt.Printf("Account details exported to %s\n", accountDetailsOutputCSV)
		} else {
			fmt.Printf("\nCapacity Breakdown for %s:\n", match.Name)
			output.PrintTable(headers, rows)
			fmt.Printf("\nTotal Size: %s\n", explorer.FormatBytes(total))
		}
		return nil
	},
}

func init() {
	AccountDetailsCmd.Flags().StringVar(&accountDetailsOutputCSV, "csv", "", "Export to CSV file")
}
