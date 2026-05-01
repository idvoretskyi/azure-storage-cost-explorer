// Azure Storage Cost Explorer CLI Tool
// Retrieve costs and usage across Blob, File, Queue, and Table services.
package main

import (
	"fmt"
	"os"

	"github.com/idvoretskyi/azure-storage-cost-explorer/cmd"
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags="-X main.version=<tag>".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "azure-storage-cost-explorer",
	Short:   "Azure Storage Cost Explorer CLI Tool",
	Long:    "Retrieve costs and usage across Blob, File, Queue, and Table services in your Azure subscription.",
	Version: version,
}

func main() {
	cmd.AttachGlobalFlags(rootCmd)
	rootCmd.AddCommand(cmd.CostsCmd)
	rootCmd.AddCommand(cmd.AccountsCmd)
	rootCmd.AddCommand(cmd.AccountDetailsCmd)
	rootCmd.AddCommand(cmd.ContainersCmd)
	rootCmd.AddCommand(cmd.ContainerDetailsCmd)
	rootCmd.AddCommand(cmd.SharesCmd)
	rootCmd.AddCommand(cmd.QueuesCmd)
	rootCmd.AddCommand(cmd.TablesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
