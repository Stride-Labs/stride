package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker of icaoracle module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// TODO: Add metric count to store so this lookup is faster
	// For each queued metric, submit an ICA to each oracle
	// and then remove the metric from the queue
	for _, metric := range k.GetAllMetricsFromQueue(ctx) {
		for _, oracle := range k.GetAllOracles(ctx) {
			if oracle.Active {
				k.Logger(ctx).Error(fmt.Sprintf("Submitting oracle metric update - Metric: %+v, Oracle: %+v", metric, oracle))
				if err := k.SubmitMetricUpdate(ctx, oracle, metric); err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Failed to submit a metric update ICA - Metric: %+v, Oracle: %+v, %s", metric, oracle, err.Error()))
				}
			}
		}
		k.RemoveMetricFromQueue(ctx, metric.Key)
	}
}
