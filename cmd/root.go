// Package cmd implements the CLI subcommands.
package cmd

import "github.com/spf13/cobra"

// SubscriptionID is set globally via the --subscription persistent flag.
var SubscriptionID string

// AttachGlobalFlags registers persistent flags shared across all subcommands.
func AttachGlobalFlags(root *cobra.Command) {
	root.PersistentFlags().StringVar(&SubscriptionID, "subscription", "", "Azure subscription ID (defaults to AZURE_SUBSCRIPTION_ID or first listed)")
}
