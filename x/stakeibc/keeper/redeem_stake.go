package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v28/utils"
	epochtypes "github.com/Stride-Labs/stride/v28/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v28/x/records/types"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

// TODO [cleanup]: Cleanup this function (errors, logs, comments, whitespace, operation ordering)
// Exchanges a user's stTokens for native tokens using the current redemption rate
func (k Keeper) RedeemStake(ctx sdk.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	k.Logger(ctx).Info(fmt.Sprintf("redeem stake: %s", msg.String()))

	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}

	// confirm the host zone is not halted and has redemptions enabled
	hostZone, err := k.GetActiveHostZone(ctx, msg.HostZone)
	if err != nil {
		return nil, err
	}
	if !hostZone.RedemptionsEnabled {
		return nil, errorsmod.Wrapf(types.ErrRedemptionsDisabled, "redemptions disabled for %s", msg.HostZone)
	}

	// ensure the recipient address is a valid bech32 address on the hostZone
	_, err = utils.AccAddressFromBech32(msg.Receiver, hostZone.Bech32Prefix)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to check if redemption rate is within safety bounds")
	}
	if !rateIsSafe {
		return nil, types.ErrRedemptionRateOutsideSafetyBounds
	}

	// construct desired unstaking amount from host zone
	nativeAmount := sdkmath.LegacyNewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).TruncateInt()
	if nativeAmount.LTE(sdkmath.ZeroInt()) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}
	if nativeAmount.GT(hostZone.TotalDelegations) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %v", msg.Amount)
	}

	// ----------------- UNBONDING RECORD KEEPING -----------------
	// Fetch the record
	epochTracker, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker found: %s", epochtypes.DAY_EPOCH)
	}

	redemptionId := recordstypes.UserRedemptionRecordKeyFormatter(hostZone.ChainId, epochTracker.EpochNumber, msg.Receiver)
	userRedemptionRecord, userHasRedeemedThisEpoch := k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if userHasRedeemedThisEpoch {
		k.Logger(ctx).Info(fmt.Sprintf("UserRedemptionRecord found for %s", redemptionId))
		// Add the unbonded amount to the UserRedemptionRecord
		// The record is set below
		userRedemptionRecord.StTokenAmount = userRedemptionRecord.StTokenAmount.Add(msg.Amount)
		userRedemptionRecord.NativeTokenAmount = userRedemptionRecord.NativeTokenAmount.Add(nativeAmount)
	} else {
		// First time a user is redeeming this epoch
		userRedemptionRecord = recordstypes.UserRedemptionRecord{
			Id:                redemptionId,
			Receiver:          msg.Receiver,
			NativeTokenAmount: nativeAmount,
			Denom:             hostZone.HostDenom,
			HostZoneId:        hostZone.ChainId,
			EpochNumber:       epochTracker.EpochNumber,
			StTokenAmount:     msg.Amount,
			// claimIsPending represents whether a redemption is currently being claimed,
			// contingent on the host zone unbonding having status CLAIMABLE
			ClaimIsPending: false,
		}
		k.Logger(ctx).Info(fmt.Sprintf("UserRedemptionRecord not found - creating for %s", redemptionId))
	}

	// then add undelegation amount to epoch unbonding records
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochTracker.EpochNumber)
	if !found {
		k.Logger(ctx).Error("latest epoch unbonding record not found")
		return nil, errorsmod.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, "latest epoch unbonding record not found")
	}
	// get relevant host zone on this epoch unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "host zone not found in unbondings: %s", hostZone.ChainId)
	}
	hostZoneUnbonding.NativeTokenAmount = hostZoneUnbonding.NativeTokenAmount.Add(nativeAmount)
	if !userHasRedeemedThisEpoch {
		// Only append a UserRedemptionRecord to the HZU if it wasn't previously appended
		hostZoneUnbonding.UserRedemptionRecords = append(hostZoneUnbonding.UserRedemptionRecords, userRedemptionRecord.Id)
	}

	// Escrow user's balance
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, msg.Amount))
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.DepositAddress, hostZone.ChainId)
	}
	// Note: checkBlockedAddr=false because depositAddress is a module
	err = utils.SafeSendCoins(false, k.bankKeeper, ctx, sender, depositAddress, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, errorsmod.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens to module account. err: %s", msg.Amount, hostZone.HostDenom, err.Error())
	}

	// record the number of stAssets that should be burned after unbonding
	hostZoneUnbonding.StTokenAmount = hostZoneUnbonding.StTokenAmount.Add(msg.Amount)

	// Actually set the records, we wait until now to prevent any errors
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Set the UserUnbondingRecords on the proper HostZoneUnbondingRecord
	hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
	if hostZoneUnbondings == nil {
		hostZoneUnbondings = []*recordstypes.HostZoneUnbonding{}
		epochUnbondingRecord.HostZoneUnbondings = hostZoneUnbondings
	}
	if err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, *hostZoneUnbonding); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info(fmt.Sprintf("executed redeem stake: %s", msg.String()))
	EmitSuccessfulRedeemStakeEvent(ctx, msg, hostZone, nativeAmount, msg.Amount)

	return &types.MsgRedeemStakeResponse{}, nil
}
