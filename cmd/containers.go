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
	containersDetailed  bool
	containersOutputCSV string
)

// ContainersCmd is the `containers` subcommand.
var ContainersCmd = &cobra.Command{
	Use:   "containers",
	Short: "List all blob containers across storage accounts (with per-tier breakdown via blob enumeration)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		e, err := explorer.New(ctx, SubscriptionID)
		if err != nil {
			return err
		}

		fmt.Printf("Retrieving blob containers (subscription %s)...\n", e.SubscriptionID)
		accounts, err := e.ListStorageAccounts(ctx)
		if err != nil {
			return err
		}
		if len(accounts) == 0 {
			fmt.Println("No storage accounts found in the subscription.")
			return nil
		}

		var headers []string
		if containersDetailed {
			headers = []string{"Account", "Container", "Access Tier", "Size"}
		} else {
			headers = []string{"Account", "Container", "Total Size", "Tiers"}
		}

		var rows [][]string
		for _, acc := range accounts {
			containers, err := e.ListContainers(ctx, acc.ResourceGroup, acc.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, c := range containers {
				fmt.Printf("Analyzing container: %s/%s\n", acc.Name, c)
				tierData, err := e.ListContainerBlobsByTier(ctx, acc.Name, c)
				if err != nil {
					fmt.Println(err)
				}
				if len(tierData) == 0 {
					rows = append(rows, []string{acc.Name, c, "No data", "N/A"})
					continue
				}
				if containersDetailed {
					// Print canonical tiers first, then any extras
					seen := map[string]bool{}
					for _, tier := range explorer.AccessTiers {
						if v, ok := tierData[tier]; ok {
							rows = append(rows, []string{acc.Name, c, tier, explorer.FormatBytes(v)})
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
						rows = append(rows, []string{acc.Name, c, k, explorer.FormatBytes(tierData[k])})
					}
				} else {
					var totalSize float64
					var tiers []string
					for _, tier := range explorer.AccessTiers {
						if v, ok := tierData[tier]; ok {
							totalSize += v
							tiers = append(tiers, tier)
						}
					}
					var extras []string
					seen := map[string]bool{}
					for _, t := range tiers {
						seen[t] = true
					}
					for k, v := range tierData {
						if !seen[k] {
							totalSize += v
							extras = append(extras, k)
						}
					}
					sort.Strings(extras)
					tiers = append(tiers, extras...)
					rows = append(rows, []string{acc.Name, c, explorer.FormatBytes(totalSize), joinComma(tiers)})
				}
			}
		}

		if len(rows) == 0 {
			fmt.Println("No containers found.")
			return nil
		}

		if containersOutputCSV != "" {
			if err := output.WriteCSV(containersOutputCSV, headers, rows); err != nil {
				return err
			}
			fmt.Printf("Container data exported to %s\n", containersOutputCSV)
		} else {
			if containersDetailed {
				fmt.Println("\nBlob Containers (Detailed):")
			} else {
				fmt.Println("\nBlob Container Summary:")
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func init() {
	ContainersCmd.Flags().BoolVar(&containersDetailed, "detailed", false, "Show per-tier breakdown")
	ContainersCmd.Flags().StringVar(&containersOutputCSV, "csv", "", "Export to CSV file")
}

func joinComma(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
