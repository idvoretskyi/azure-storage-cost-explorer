package explorer

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// AccessTiers lists Azure Blob access tiers in a stable order.
var AccessTiers = []string{"Hot", "Cool", "Cold", "Archive"}

// AccountCapacity holds capacity metrics for a storage account, broken down per service.
type AccountCapacity struct {
	BlobCapacityBytes  float64
	BlobByTier         map[string]float64 // populated when available; otherwise empty
	FileCapacityBytes  float64
	QueueCapacityBytes float64
	TableCapacityBytes float64
}

// GetAccountCapacity queries Azure Monitor for capacity metrics on the given storage account.
// Returns zero values for any service whose metric is unavailable.
//
// Resource IDs:
//   - account-level: <account.ResourceID>
//   - blob service:  <account.ResourceID>/blobServices/default
//   - file service:  <account.ResourceID>/fileServices/default
//   - queue service: <account.ResourceID>/queueServices/default
//   - table service: <account.ResourceID>/tableServices/default
func (e *StorageCostExplorer) GetAccountCapacity(ctx context.Context, accountResourceID string) (AccountCapacity, error) {
	cap := AccountCapacity{
		BlobByTier: make(map[string]float64),
	}

	end := time.Now().UTC()
	start := end.Add(-48 * time.Hour)
	timespan := start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
	interval := "PT1H"
	aggregation := "Average"

	type metricSpec struct {
		resourceID string
		metricName string
		assignTo   *float64
	}

	specs := []metricSpec{
		{
			resourceID: accountResourceID + "/blobServices/default",
			metricName: "BlobCapacity",
			assignTo:   &cap.BlobCapacityBytes,
		},
		{
			resourceID: accountResourceID + "/fileServices/default",
			metricName: "FileCapacity",
			assignTo:   &cap.FileCapacityBytes,
		},
		{
			resourceID: accountResourceID + "/queueServices/default",
			metricName: "QueueCapacity",
			assignTo:   &cap.QueueCapacityBytes,
		},
		{
			resourceID: accountResourceID + "/tableServices/default",
			metricName: "TableCapacity",
			assignTo:   &cap.TableCapacityBytes,
		},
	}

	resultType := armmonitor.ResultTypeData
	for _, s := range specs {
		ns := metricNamespaceFor(s.metricName)
		opts := &armmonitor.MetricsClientListOptions{
			Timespan:        &timespan,
			Interval:        &interval,
			Metricnames:     &s.metricName,
			Aggregation:     &aggregation,
			ResultType:      &resultType,
			Metricnamespace: &ns,
		}
		resp, err := e.MetricsClient.List(ctx, s.resourceID, opts)
		if err != nil {
			// Not all metric namespaces exist for every account kind; skip silently.
			continue
		}
		val := latestMetricValue(resp.Value)
		if s.assignTo != nil {
			*s.assignTo = val
		}
	}

	// Query BlobCapacity per Tier dimension separately.
	blobResource := accountResourceID + "/blobServices/default"
	blobNS := "Microsoft.Storage/storageAccounts/blobServices"
	for _, tier := range AccessTiers {
		filter := fmt.Sprintf("Tier eq '%s'", tier)
		metricName := "BlobCapacity"
		opts := &armmonitor.MetricsClientListOptions{
			Timespan:        &timespan,
			Interval:        &interval,
			Metricnames:     &metricName,
			Aggregation:     &aggregation,
			Filter:          &filter,
			ResultType:      &resultType,
			Metricnamespace: &blobNS,
		}
		resp, err := e.MetricsClient.List(ctx, blobResource, opts)
		if err != nil {
			continue
		}
		val := latestMetricValue(resp.Value)
		if val > 0 {
			cap.BlobByTier[tier] = val
		}
	}

	return cap, nil
}

func metricNamespaceFor(metric string) string {
	switch metric {
	case "BlobCapacity":
		return "Microsoft.Storage/storageAccounts/blobServices"
	case "FileCapacity":
		return "Microsoft.Storage/storageAccounts/fileServices"
	case "QueueCapacity":
		return "Microsoft.Storage/storageAccounts/queueServices"
	case "TableCapacity":
		return "Microsoft.Storage/storageAccounts/tableServices"
	default:
		return "Microsoft.Storage/storageAccounts"
	}
}

// latestMetricValue scans the metric data and returns the most recent
// average value across all timeseries.
func latestMetricValue(metrics []*armmonitor.Metric) float64 {
	var latestT time.Time
	var latestVal float64
	for _, m := range metrics {
		if m == nil {
			continue
		}
		for _, ts := range m.Timeseries {
			if ts == nil {
				continue
			}
			for _, dp := range ts.Data {
				if dp == nil || dp.Average == nil || dp.TimeStamp == nil {
					continue
				}
				if dp.TimeStamp.After(latestT) {
					latestT = *dp.TimeStamp
					latestVal = *dp.Average
				}
			}
		}
	}
	return latestVal
}
