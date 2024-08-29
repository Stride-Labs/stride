package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/utils"

	"github.com/Stride-Labs/stride/v24/x/staketia/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// User transaction to liquid stake native tokens into stTokens
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	stToken, err := k.Keeper.LiquidStake(ctx, msg.Staker, msg.NativeAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgLiquidStakeResponse{StToken: stToken}, nil
}

// User transaction to redeem stake stTokens into native tokens
func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	nativeToken, err := k.Keeper.RedeemStake(ctx, msg.Redeemer, msg.StTokenAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgRedeemStakeResponse{NativeToken: nativeToken}, nil
}

// Operator transaction to confirm a delegation was submitted on the host chain
func (k msgServer) ConfirmDelegation(goCtx context.Context, msg *types.MsgConfirmDelegation) (*types.MsgConfirmDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to either admin (SAFE or OPERATOR) address
	if err := k.CheckIsSafeOrOperatorAddress(ctx, msg.Operator); err != nil {
		return nil, err
	}

	err := k.Keeper.ConfirmDelegation(ctx, msg.RecordId, msg.TxHash, msg.Operator)
	if err != nil {
		return nil, err
	}

	return &types.MsgConfirmDelegationResponse{}, nil
}

// Operator transaction to confirm an undelegation was submitted on the host chain
func (k msgServer) ConfirmUndelegation(goCtx context.Context, msg *types.MsgConfirmUndelegation) (*types.MsgConfirmUndelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to either admin (SAFE or OPERATOR) address
	if err := k.CheckIsSafeOrOperatorAddress(ctx, msg.Operator); err != nil {
		return nil, err
	}

	err := k.Keeper.ConfirmUndelegation(ctx, msg.RecordId, msg.TxHash, msg.Operator)
	if err != nil {
		return nil, err
	}

	return &types.MsgConfirmUndelegationResponse{}, err
}

// Operator transaction to confirm unbonded tokens were transferred back to stride
func (k msgServer) ConfirmUnbondedTokenSweep(goCtx context.Context, msg *types.MsgConfirmUnbondedTokenSweep) (*types.MsgConfirmUnbondedTokenSweepResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to either admin (SAFE or OPERATOR) address
	if err := k.CheckIsSafeOrOperatorAddress(ctx, msg.Operator); err != nil {
		return nil, err
	}

	err := k.Keeper.ConfirmUnbondedTokenSweep(ctx, msg.RecordId, msg.TxHash, msg.Operator)
	if err != nil {
		return nil, err
	}

	return &types.MsgConfirmUnbondedTokenSweepResponse{}, nil
}

// SAFE transaction to adjust the delegated balance after a validator was slashed
// - creates a slash record as a log
// - allow negative amounts in case we want to fix our record keeping
func (k msgServer) AdjustDelegatedBalance(goCtx context.Context, msg *types.MsgAdjustDelegatedBalance) (*types.MsgAdjustDelegatedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msg.Operator); err != nil {
		return nil, err
	}

	// add offset to the delegated balance and write to host zone
	// Note: we're intentionally not checking the zone is halted
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return nil, err
	}
	hostZone.DelegatedBalance = hostZone.DelegatedBalance.Add(msg.DelegationOffset)

	// safety check that this will not cause the delegated balance to be negative
	if hostZone.DelegatedBalance.IsNegative() {
		return nil, types.ErrNegativeNotAllowed.Wrapf("offset would cause the delegated balance to be negative")
	}
	k.SetHostZone(ctx, hostZone)

	// create a corresponding slash record
	latestSlashRecordId := k.IncrementSlashRecordId(ctx)
	slashRecord := types.SlashRecord{
		Id:               latestSlashRecordId,
		Time:             uint64(ctx.BlockTime().Unix()),
		NativeAmount:     msg.DelegationOffset,
		ValidatorAddress: msg.ValidatorAddress,
	}
	k.SetSlashRecord(ctx, slashRecord)

	return &types.MsgAdjustDelegatedBalanceResponse{}, nil
}

// Adjusts the inner redemption rate bounds on the host zone
func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to the BOUNDS address
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return nil, types.ErrInvalidAdmin
	}

	// Fetch the zone
	zone, err := k.GetHostZone(ctx)
	if err != nil {
		return nil, err
	}

	// Get the outer bounds
	maxOuterBound := zone.MaxRedemptionRate
	minOuterBound := zone.MinRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	maxInnerBound := msg.MaxInnerRedemptionRate
	minInnerBound := msg.MinInnerRedemptionRate
	if maxInnerBound.GT(maxOuterBound) {
		return nil, types.ErrInvalidRedemptionRateBounds
	}
	if minInnerBound.LT(minOuterBound) {
		return nil, types.ErrInvalidRedemptionRateBounds
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = minInnerBound
	zone.MaxInnerRedemptionRate = maxInnerBound

	// Update the host zone
	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

// Unhalts the host zone if redemption rates were exceeded
// BOUNDS: verified in ValidateBasic
func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to the BOUNDS address
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return nil, types.ErrInvalidAdmin
	}

	// Note: of course we don't want to fail this if the zone is halted!
	zone, err := k.GetHostZone(ctx)
	if err != nil {
		return nil, err
	}

	// Check the zone is halted
	if !zone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotHalted, "zone is not halted")
	}

	stDenom := utils.StAssetDenomFromHostZoneDenom(zone.NativeTokenDenom)
	k.ratelimitKeeper.RemoveDenomFromBlacklist(ctx, stDenom)

	// Resume zone
	zone.Halted = false
	k.SetHostZone(ctx, zone)

	return &types.MsgResumeHostZoneResponse{}, nil
}

// trigger updating the redemption rate
func (k msgServer) RefreshRedemptionRate(goCtx context.Context, msgTriggerRedemptionRate *types.MsgRefreshRedemptionRate) (*types.MsgRefreshRedemptionRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msgTriggerRedemptionRate.Creator); err != nil {
		return nil, err
	}

	err := k.UpdateRedemptionRate(ctx)

	return &types.MsgRefreshRedemptionRateResponse{}, err
}

// overwrite a delegation record
func (k msgServer) OverwriteDelegationRecord(goCtx context.Context, msgOverwriteDelegationRecord *types.MsgOverwriteDelegationRecord) (*types.MsgOverwriteDelegationRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msgOverwriteDelegationRecord.Creator); err != nil {
		return nil, err
	}

	k.Keeper.SetDelegationRecord(ctx, *msgOverwriteDelegationRecord.DelegationRecord)

	return &types.MsgOverwriteDelegationRecordResponse{}, nil
}

// overwrite a unbonding record
func (k msgServer) OverwriteUnbondingRecord(goCtx context.Context, msgOverwriteUnbondingRecord *types.MsgOverwriteUnbondingRecord) (*types.MsgOverwriteUnbondingRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msgOverwriteUnbondingRecord.Creator); err != nil {
		return nil, err
	}

	k.Keeper.SetUnbondingRecord(ctx, *msgOverwriteUnbondingRecord.UnbondingRecord)

	return &types.MsgOverwriteUnbondingRecordResponse{}, nil
}

// overwrite a redemption record
func (k msgServer) OverwriteRedemptionRecord(goCtx context.Context, msgOverwriteRedemptionRecord *types.MsgOverwriteRedemptionRecord) (*types.MsgOverwriteRedemptionRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msgOverwriteRedemptionRecord.Creator); err != nil {
		return nil, err
	}

	k.Keeper.SetRedemptionRecord(ctx, *msgOverwriteRedemptionRecord.RedemptionRecord)

	return &types.MsgOverwriteRedemptionRecordResponse{}, nil
}

// Sets the OPERATOR address for the host zone
// - only SAFE can execute this message
func (k msgServer) SetOperatorAddress(goCtx context.Context, msg *types.MsgSetOperatorAddress) (*types.MsgSetOperatorAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// gate this transaction to only the SAFE address
	if err := k.CheckIsSafeAddress(ctx, msg.Signer); err != nil {
		return nil, err
	}

	// Fetch the zone
	// Note: we're intentionally not checking the zone is halted
	zone, err := k.GetHostZone(ctx)
	if err != nil {
		return nil, err
	}

	// set the operator field
	zone.OperatorAddressOnStride = msg.Operator

	// Update the host zone
	k.SetHostZone(ctx, zone)

	return &types.MsgSetOperatorAddressResponse{}, nil
}
