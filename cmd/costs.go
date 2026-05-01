package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/explorer"
	"github.com/idvoretskyi/azure-storage-cost-explorer/internal/output"
	"github.com/spf13/cobra"
)

var (
	costsDays      int
	costsOutputCSV string
)

// CostsCmd is the `costs` subcommand.
var CostsCmd = &cobra.Command{
	Use:   "costs",
	Short: "Get Azure Storage costs for the specified period",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving Azure Storage costs for the last %d days (subscription %s)...\n", costsDays, e.SubscriptionID)
		total, err := e.GetStorageCosts(ctx, costsDays)
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal Storage Cost (last %d days): $%.2f\n", costsDays, total)

		detailed, err := e.GetDetailedStorageCosts(ctx, costsDays)
		if err != nil {
			return err
		}
		if len(detailed) == 0 {
			return nil
		}

		// Sort by cost descending
		sort.Slice(detailed, func(i, j int) bool { return detailed[i].Cost > detailed[j].Cost })

		headers := []string{"Meter Sub-Category", "Cost"}
		var rows [][]string
		for _, item := range detailed {
			rows = append(rows, []string{item.Key, fmt.Sprintf("$%.2f", item.Cost)})
		}

		if costsOutputCSV != "" {
			if err := output.WriteCSV(costsOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Cost data exported to %s\n", costsOutputCSV)
		} else {
			fmt.Println("\nDetailed Cost Breakdown:")
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	CostsCmd.Flags().IntVar(&costsDays, "days", 30, "Number of days to analyze (default: 30)")
	CostsCmd.Flags().StringVar(&costsOutputCSV, "csv", "", "Export to CSV file")
}
