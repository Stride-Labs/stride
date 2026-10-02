// x/stakeibc/keeper/msg_server_wind_down.go
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// Wind-down admin txs (spec §7). Each handler is a thin delegate to the keeper function in the
// matching wind_down_*.go file. The sweep is delivered by the next PR.

// UndelegateFromValidators: the wind-down drain (Task 3 replaces the body).
// Delegates to Keeper.UndelegateFromValidators in wind_down_undelegate.go.
func (k msgServer) UndelegateFromValidators(goCtx context.Context, msg *types.MsgUndelegateFromValidators) (*types.MsgUndelegateFromValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numBatches, err := k.Keeper.UndelegateFromValidators(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUndelegateFromValidatorsResponse{NumBatchesSubmitted: numBatches}, nil
}

// TransferFromIca: ICA balance to the Osmosis vault (Task 4 replaces the body).
// Delegates to Keeper.TransferFromIca in wind_down_transfer_from_ica.go.
func (k msgServer) TransferFromIca(goCtx context.Context, msg *types.MsgTransferFromIca) (*types.MsgTransferFromIcaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.TransferFromIca(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgTransferFromIcaResponse{}, nil
}

// TransferStaketiaClaimBalance: claim-address TIA to the celestia delegation ICA (Task 5
// replaces the body). Delegates to Keeper.TransferStaketiaClaimBalance in wind_down_staketia_claim.go.
func (k msgServer) TransferStaketiaClaimBalance(goCtx context.Context, msg *types.MsgTransferStaketiaClaimBalance) (*types.MsgTransferStaketiaClaimBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferred, err := k.Keeper.TransferStaketiaClaimBalance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgTransferStaketiaClaimBalanceResponse{Transferred: transferred}, nil
}

// SweepTokensOffStride is the batched token sweep, gated on the sweep operator in ValidateBasic
func (k msgServer) SweepTokensOffStride(goCtx context.Context, msg *types.MsgSweepTokensOffStride) (*types.MsgSweepTokensOffStrideResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numTransfers, numSkipped, err := k.Keeper.SweepTokensOffStride(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSweepTokensOffStrideResponse{NumTransfers: numTransfers, NumSkipped: numSkipped}, nil
}
