package explorer

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
)

// GetStorageCosts returns the total storage cost for the subscription over the
// last `days` days (using ActualCost).
func (e *StorageCostExplorer) GetStorageCosts(ctx context.Context, days int) (float64, error) {
	scope := fmt.Sprintf("/subscriptions/%s", e.SubscriptionID)
	from := time.Now().UTC().AddDate(0, 0, -days)
	to := time.Now().UTC()

	def := buildStorageCostDefinition(from, to, "")
	resp, err := e.CostQueryClient.Usage(ctx, scope, def, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting storage costs: %w", err)
	}
	rows := costRowsFromResult(resp.QueryResult)

	var total float64
	for _, r := range rows {
		total += r.Cost
	}
	return total, nil
}

// CostBreakdownItem represents a single line in the detailed breakdown.
type CostBreakdownItem struct {
	Key  string
	Cost float64
}

// GetDetailedStorageCosts returns storage costs grouped by MeterSubCategory
// (Blob, File, Queue, Table, ...) over the last `days` days.
func (e *StorageCostExplorer) GetDetailedStorageCosts(ctx context.Context, days int) ([]CostBreakdownItem, error) {
	scope := fmt.Sprintf("/subscriptions/%s", e.SubscriptionID)
	from := time.Now().UTC().AddDate(0, 0, -days)
	to := time.Now().UTC()

	def := buildStorageCostDefinition(from, to, "MeterSubCategory")
	resp, err := e.CostQueryClient.Usage(ctx, scope, def, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting detailed storage costs: %w", err)
	}
	rows := costRowsFromResult(resp.QueryResult)

	totals := make(map[string]float64)
	for _, r := range rows {
		key := r.GroupKey
		if key == "" {
			key = "(unspecified)"
		}
		totals[key] += r.Cost
	}
	var items []CostBreakdownItem
	for k, v := range totals {
		if v > 0 {
			items = append(items, CostBreakdownItem{Key: k, Cost: v})
		}
	}
	return items, nil
}

// buildStorageCostDefinition constructs a Cost Management query definition
// scoped to the Storage service. If groupBy is non-empty, results are grouped
// by that dimension.
func buildStorageCostDefinition(from, toTime time.Time, groupBy string) armcostmanagement.QueryDefinition {
	def := armcostmanagement.QueryDefinition{
		Type:      to.Ptr(armcostmanagement.ExportTypeActualCost),
		Timeframe: to.Ptr(armcostmanagement.TimeframeTypeCustom),
		TimePeriod: &armcostmanagement.QueryTimePeriod{
			From: to.Ptr(from),
			To:   to.Ptr(toTime),
		},
		Dataset: &armcostmanagement.QueryDataset{
			Granularity: to.Ptr(armcostmanagement.GranularityTypeDaily),
			Aggregation: map[string]*armcostmanagement.QueryAggregation{
				"totalCost": {
					Name:     to.Ptr("Cost"),
					Function: to.Ptr(armcostmanagement.FunctionTypeSum),
				},
			},
			Filter: &armcostmanagement.QueryFilter{
				Dimensions: &armcostmanagement.QueryComparisonExpression{
					Name:     to.Ptr("ServiceName"),
					Operator: to.Ptr(armcostmanagement.QueryOperatorTypeIn),
					Values: []*string{
						to.Ptr("Storage"),
					},
				},
			},
		},
	}
	if groupBy != "" {
		def.Dataset.Grouping = []*armcostmanagement.QueryGrouping{
			{
				Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
				Name: to.Ptr(groupBy),
			},
		}
	}
	return def
}

type costRow struct {
	Cost     float64
	GroupKey string
}

// costRowsFromResult parses a QueryResult into normalized rows.
// The Cost Management API returns columns dynamically, so we look up by name.
func costRowsFromResult(qr armcostmanagement.QueryResult) []costRow {
	if qr.Properties == nil {
		return nil
	}
	props := qr.Properties

	costIdx, groupIdx := -1, -1
	for i, c := range props.Columns {
		if c == nil || c.Name == nil {
			continue
		}
		switch *c.Name {
		case "PreTaxCost", "Cost", "CostUSD", "PreTaxCostUSD":
			if costIdx == -1 {
				costIdx = i
			}
		default:
			// Track the first non-cost, non-currency, non-date column as the group dimension.
			n := *c.Name
			if n != "UsageDate" && n != "Currency" && groupIdx == -1 {
				groupIdx = i
			}
		}
	}

	var rows []costRow
	for _, raw := range props.Rows {
		if costIdx < 0 || costIdx >= len(raw) {
			continue
		}
		var r costRow
		switch v := raw[costIdx].(type) {
		case float64:
			r.Cost = v
		case float32:
			r.Cost = float64(v)
		case int:
			r.Cost = float64(v)
		case int64:
			r.Cost = float64(v)
		}
		if groupIdx >= 0 && groupIdx < len(raw) {
			if s, ok := raw[groupIdx].(string); ok {
				r.GroupKey = s
			}
		}
		rows = append(rows, r)
	}
	return rows
}
