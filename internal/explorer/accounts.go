package explorer

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// StorageAccount represents a discovered Azure storage account.
type StorageAccount struct {
	Name          string
	ResourceGroup string
	Location      string
	Kind          string
	SKU           string
	ResourceID    string
}

// ListStorageAccounts returns all storage accounts in the subscription.
func (e *StorageCostExplorer) ListStorageAccounts(ctx context.Context) ([]StorageAccount, error) {
	pager := e.AccountsClient.NewListPager(nil)
	var accounts []StorageAccount
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing storage accounts: %w", err)
		}
		for _, a := range page.Value {
			if a == nil {
				continue
			}
			acc := StorageAccount{
				Name:       safeStr(a.Name),
				Location:   safeStr(a.Location),
				ResourceID: safeStr(a.ID),
			}
			acc.ResourceGroup = ParseResourceGroup(acc.ResourceID)
			if a.Kind != nil {
				acc.Kind = string(*a.Kind)
			}
			if a.SKU != nil && a.SKU.Name != nil {
				acc.SKU = string(*a.SKU.Name)
			}
			accounts = append(accounts, acc)
		}
	}
	return accounts, nil
}

// ListContainers returns all blob containers within the given storage account.
func (e *StorageCostExplorer) ListContainers(ctx context.Context, rg, account string) ([]string, error) {
	pager := e.ContainersClient.NewListPager(rg, account, nil)
	var names []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing containers in %s: %w", account, err)
		}
		for _, c := range page.Value {
			if c != nil && c.Name != nil {
				names = append(names, *c.Name)
			}
		}
	}
	return names, nil
}

// FileShare represents an Azure file share.
type FileShare struct {
	Name        string
	QuotaGB     int32
	AccessTier  string
	UsedBytes   int64
	EnabledTier string
}

// ListFileShares returns file shares within the given storage account.
func (e *StorageCostExplorer) ListFileShares(ctx context.Context, rg, account string) ([]FileShare, error) {
	pager := e.SharesClient.NewListPager(rg, account, &armstorage.FileSharesClientListOptions{
		Expand: stringPtr("stats"),
	})
	var shares []FileShare
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing file shares in %s: %w", account, err)
		}
		for _, s := range page.Value {
			if s == nil {
				continue
			}
			share := FileShare{
				Name: safeStr(s.Name),
			}
			if s.Properties != nil {
				if s.Properties.ShareQuota != nil {
					share.QuotaGB = *s.Properties.ShareQuota
				}
				if s.Properties.AccessTier != nil {
					share.AccessTier = string(*s.Properties.AccessTier)
				}
				if s.Properties.ShareUsageBytes != nil {
					share.UsedBytes = *s.Properties.ShareUsageBytes
				}
				if s.Properties.EnabledProtocols != nil {
					share.EnabledTier = string(*s.Properties.EnabledProtocols)
				}
			}
			shares = append(shares, share)
		}
	}
	return shares, nil
}

// ListQueues returns queues within the given storage account (control-plane).
func (e *StorageCostExplorer) ListQueues(ctx context.Context, rg, account string) ([]string, error) {
	pager := e.QueuesClient.NewListPager(rg, account, nil)
	var names []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing queues in %s: %w", account, err)
		}
		for _, q := range page.Value {
			if q != nil && q.Name != nil {
				names = append(names, *q.Name)
			}
		}
	}
	return names, nil
}

// ListTables returns tables within the given storage account (control-plane).
func (e *StorageCostExplorer) ListTables(ctx context.Context, rg, account string) ([]string, error) {
	pager := e.TablesClient.NewListPager(rg, account, nil)
	var names []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing tables in %s: %w", account, err)
		}
		for _, t := range page.Value {
			if t != nil && t.Name != nil {
				names = append(names, *t.Name)
			}
		}
	}
	return names, nil
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtr(s string) *string { return &s }
