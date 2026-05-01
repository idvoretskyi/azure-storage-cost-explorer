package explorer

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

// ListContainerBlobsByTier enumerates the blobs in a container and aggregates
// total size per access tier (Hot/Cool/Cold/Archive). This is the analog of the
// AWS ListObjectsV2 fallback because Azure Monitor only exposes account-level
// (not container-level) capacity metrics.
//
// For large containers this may be slow.
func (e *StorageCostExplorer) ListContainerBlobsByTier(ctx context.Context, accountName, containerName string) (map[string]float64, error) {
	svcURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	svcClient, err := service.NewClient(svcURL, e.Cred, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating blob service client for %s: %w", accountName, err)
	}
	containerClient := svcClient.NewContainerClient(containerName)

	tierData := make(map[string]float64)
	pager := containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing blobs in %s/%s: %w", accountName, containerName, err)
		}
		if page.Segment == nil {
			continue
		}
		for _, b := range page.Segment.BlobItems {
			if b == nil || b.Properties == nil {
				continue
			}
			tier := "Unknown"
			if b.Properties.AccessTier != nil {
				tier = string(*b.Properties.AccessTier)
			}
			var size float64
			if b.Properties.ContentLength != nil {
				size = float64(*b.Properties.ContentLength)
			}
			tierData[tier] += size
		}
	}
	return tierData, nil
}
