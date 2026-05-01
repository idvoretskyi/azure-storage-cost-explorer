package cmd

import (
	"context"
	"fmt"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var queuesOutputCSV string

// QueuesCmd is the `queues` subcommand.
var QueuesCmd = &cobra.Command{
	Use:   "queues",
	Short: "List all storage queues across storage accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving queues (subscription %s)...\n", e.SubscriptionID)
		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}

		headers := []string{"Account", "Queue"}
		var rows [][]string
		for _, acc := range accounts {
			queues, err := e.ListQueues(ctx, acc.ResourceGroup, acc.Name)
			if err != nil {
				// Some account kinds (e.g., BlobStorage) don't support queues; ignore.
				continue
			}
			for _, q := range queues {
				rows = append(rows, []string{acc.Name, q})
			}
		}

		if len(rows) == 0 {
			fmt.Println("No queues found.")
			return nil
		}

		if queuesOutputCSV != "" {
			if err := output.WriteCSV(queuesOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Queue data exported to %s\n", queuesOutputCSV)
		} else {
			fmt.Println("\nStorage Queues:")
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	QueuesCmd.Flags().StringVar(&queuesOutputCSV, "csv", "", "Export to CSV file")
}
