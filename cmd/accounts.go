package cmd

import (
	"context"
	"fmt"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var (
	accountsDetailed  bool
	accountsOutputCSV string
)

// AccountsCmd is the `accounts` subcommand.
var AccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List all storage accounts with capacity across Blob, File, Queue, and Table services",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving storage accounts (subscription %s)...\n", e.SubscriptionID)
		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}
		if len(accounts) == 0 {
			fmt.Println("No storage accounts found in the subscription.")
			return nil
		}

		var headers []string
		if accountsDetailed {
			headers = []string{"Account", "Resource Group", "Location", "Service", "Size"}
		} else {
			headers = []string{"Account", "Resource Group", "Location", "SKU", "Total Size"}
		}

		var rows [][]string
		for _, acc := range accounts {
			fmt.Printf("Analyzing account: %s\n", acc.Name)
			cap, err := e.GetAccountCapacity(ctx, acc.ResourceID)
			if err != nil {
				fmt.Println(err)
			}

			if accountsDetailed {
				rows = append(rows,
					[]string{acc.Name, acc.ResourceGroup, acc.Location, "Blob", explorer.FormatBytes(cap.BlobCapacityBytes)},
					[]string{acc.Name, acc.ResourceGroup, acc.Location, "File", explorer.FormatBytes(cap.FileCapacityBytes)},
					[]string{acc.Name, acc.ResourceGroup, acc.Location, "Queue", explorer.FormatBytes(cap.QueueCapacityBytes)},
					[]string{acc.Name, acc.ResourceGroup, acc.Location, "Table", explorer.FormatBytes(cap.TableCapacityBytes)},
				)
			} else {
				total := cap.BlobCapacityBytes + cap.FileCapacityBytes + cap.QueueCapacityBytes + cap.TableCapacityBytes
				rows = append(rows, []string{acc.Name, acc.ResourceGroup, acc.Location, acc.SKU, explorer.FormatBytes(total)})
			}
		}

		if len(rows) == 0 {
			return nil
		}

		if accountsOutputCSV != "" {
			if err := output.WriteCSV(accountsOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Account data exported to %s\n", accountsOutputCSV)
		} else {
			if accountsDetailed {
				fmt.Println("\nStorage Accounts (Detailed):")
			} else {
				fmt.Println("\nStorage Account Summary:")
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	AccountsCmd.Flags().BoolVar(&accountsDetailed, "detailed", false, "Show per-service capacity breakdown")
	AccountsCmd.Flags().StringVar(&accountsOutputCSV, "csv", "", "Export to CSV file")
}
