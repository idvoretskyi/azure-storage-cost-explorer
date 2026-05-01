// Package explorer provides Azure clients and the core StorageCostExplorer type.
package explorer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// StorageCostExplorer holds Azure SDK clients used by all commands.
type StorageCostExplorer struct {
	Cred           azcore.TokenCredential
	SubscriptionID string

	AccountsClient   *armstorage.AccountsClient
	ContainersClient *armstorage.BlobContainersClient
	SharesClient     *armstorage.FileSharesClient
	BlobSvcClient    *armstorage.BlobServicesClient
	FileSvcClient    *armstorage.FileServicesClient
	QueueSvcClient   *armstorage.QueueServicesClient
	TableSvcClient   *armstorage.TableServicesClient
	QueuesClient     *armstorage.QueueClient
	TablesClient     *armstorage.TableClient
	MetricsClient    *armmonitor.MetricsClient
	CostQueryClient  *armcostmanagement.QueryClient
}

// New creates a new StorageCostExplorer using DefaultAzureCredential.
// Subscription resolution order:
//  1. explicit subscriptionID parameter (e.g. from --subscription flag)
//  2. AZURE_SUBSCRIPTION_ID env var
//  3. first subscription returned by the management API
func New(ctx context.Context, subscriptionID string) (*StorageCostExplorer, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain Azure credentials: %w", err)
	}

	subID := strings.TrimSpace(subscriptionID)
	if subID == "" {
		subID = strings.TrimSpace(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	}
	if subID == "" {
		subClient, err := armsubscription.NewSubscriptionsClient(cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscriptions client: %w", err)
		}
		pager := subClient.NewListPager(nil)
		found := false
		for pager.More() && !found {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list subscriptions: %w", err)
			}
			for _, s := range page.Value {
				if s != nil && s.SubscriptionID != nil {
					subID = *s.SubscriptionID
					found = true
					break
				}
			}
		}
		if subID == "" {
			return nil, fmt.Errorf("no Azure subscription found; set AZURE_SUBSCRIPTION_ID or use --subscription")
		}
	}

	accountsClient, err := armstorage.NewAccountsClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}
	containersClient, err := armstorage.NewBlobContainersClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob containers client: %w", err)
	}
	sharesClient, err := armstorage.NewFileSharesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file shares client: %w", err)
	}
	blobSvc, err := armstorage.NewBlobServicesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob services client: %w", err)
	}
	fileSvc, err := armstorage.NewFileServicesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file services client: %w", err)
	}
	queueSvc, err := armstorage.NewQueueServicesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue services client: %w", err)
	}
	tableSvc, err := armstorage.NewTableServicesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create table services client: %w", err)
	}
	queuesClient, err := armstorage.NewQueueClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue client: %w", err)
	}
	tablesClient, err := armstorage.NewTableClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create table client: %w", err)
	}
	metricsClient, err := armmonitor.NewMetricsClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}
	costClient, err := armcostmanagement.NewQueryClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cost management client: %w", err)
	}

	return &StorageCostExplorer{
		Cred:             cred,
		SubscriptionID:   subID,
		AccountsClient:   accountsClient,
		ContainersClient: containersClient,
		SharesClient:     sharesClient,
		BlobSvcClient:    blobSvc,
		FileSvcClient:    fileSvc,
		QueueSvcClient:   queueSvc,
		TableSvcClient:   tableSvc,
		QueuesClient:     queuesClient,
		TablesClient:     tablesClient,
		MetricsClient:    metricsClient,
		CostQueryClient:  costClient,
	}, nil
}

// ParseResourceID extracts the resource group from an ARM resource ID like:
//   /subscriptions/<sub>/resourceGroups/<rg>/providers/...
// Returns "" if not found.
func ParseResourceGroup(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, p := range parts {
		if strings.EqualFold(p, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
