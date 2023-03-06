package keeper

import (
	"encoding/json"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Submits an ICA to update the metric in the CW contract
func (k Keeper) SubmitMetricUpdate(ctx sdk.Context, oracle types.Oracle, metric types.Metric) error {
	// Validate ICA is setup properly, contract has been instantiated, and oracle is active
	if err := oracle.ValidateICASetup(); err != nil {
		return err
	}
	if err := oracle.ValidateContractInstantiated(); err != nil {
		return err
	}
	if !oracle.Active {
		return errorsmod.Wrapf(types.ErrOracleInactive, "oracle (%s) is inactive", oracle.ChainId)
	}

	// Build contract message with metric update
	contractMsg := types.MsgExecuteContractPostMetric{
		PostMetric: &metric,
	}
	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to marshal execute contract update metric: %s", err.Error())
	}

	// Build ICA message to execute the CW contract
	msgs := []sdk.Msg{&types.MsgExecuteContract{
		Sender:   oracle.IcaAddress,
		Contract: oracle.ContractAddress,
		Msg:      contractMsgBz,
	}}

	// QUESTION/TODO: Not sure what makes the most sense for the timeout
	// I think we can be more conservative than our epochly logic
	// The oracle querier can enforce filters to ensure the data is recent, so I think from the Stride
	// perspective, we should lean more conservative and do our best to avoid timeout's and channel closure's
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour * 24).Nanoseconds())

	// Submit the ICA to execute the contract
	callbackArgs := types.UpdateOracleCallback{
		OracleChainId: oracle.ChainId,
		Metric:        &metric,
	}
	icaTx := types.ICATx{
		ConnectionId: oracle.ConnectionId,
		ChannelId:    oracle.ChannelId,
		PortId:       oracle.PortId,
		Messages:     msgs,
		Timeout:      timeout,
		CallbackArgs: &callbackArgs,
		CallbackId:   ICACallbackID_UpdateOracle,
	}
	if err := k.SubmitICATx(ctx, icaTx); err != nil {
		return errorsmod.Wrapf(err, "unable to submit update oracle contract ICA")
	}

	// Add the metric to the pending store
	pendingMetricUpdate := types.PendingMetricUpdate{
		OracleChainId: oracle.ChainId,
		Metric:        &metric,
	}
	k.SetMetricUpdateInProgress(ctx, pendingMetricUpdate)

	return nil
}
