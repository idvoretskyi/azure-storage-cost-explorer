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
	containerDetailsOutputCSV string
)

// ContainerDetailsCmd is the `container-details` subcommand.
// Argument format: <storage-account>/<container>
var ContainerDetailsCmd = &cobra.Command{
	Use:   "container-details <account>/<container>",
	Short: "Get per-tier size breakdown for a specific blob container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("argument must be in the form <storage-account>/<container>")
		}
		accountName, containerName := parts[0], parts[1]

		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving detailed information for container: %s/%s\n", accountName, containerName)
		tierData, err := e.ListContainerBlobsByTier(ctx, accountName, containerName)
		if err != nil {
			return err
		}
		if len(tierData) == 0 {
			fmt.Printf("No data found for container: %s/%s\n", accountName, containerName)
			return nil
		}

		headers := []string{"Access Tier", "Size"}
		var rows [][]string
		var total float64

		seen := map[string]bool{}
		for _, tier := range explorer.AccessTiers {
			if v, ok := tierData[tier]; ok {
				rows = append(rows, []string{tier, explorer.FormatBytes(v)})
				total += v
				seen[tier] = true
			}
		}
		var extras []string
		for k := range tierData {
			if !seen[k] {
				extras = append(extras, k)
			}
		}
		sort.Strings(extras)
		for _, k := range extras {
			rows = append(rows, []string{k, explorer.FormatBytes(tierData[k])})
			total += tierData[k]
		}

		if containerDetailsOutputCSV != "" {
			csvRows := append(rows, []string{"Total", explorer.FormatBytes(total)})
			if err := output.WriteCSV(containerDetailsOutputCSV, headers, csvRows); err != nil {
				return err
			}
			fmt.Printf("Container details exported to %s\n", containerDetailsOutputCSV)
		} else {
			fmt.Printf("\nAccess Tier Breakdown for %s/%s:\n", accountName, containerName)
			output.PrintTable(headers, rows)
			fmt.Printf("\nTotal Size: %s\n", explorer.FormatBytes(total))
		}
		return nil
	},
}

func init() {
	ContainerDetailsCmd.Flags().StringVar(&containerDetailsOutputCSV, "csv", "", "Export to CSV file")
}
