package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v18/utils"
)

func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("redeem stake: %s", msg.String()))

	// ----------------- PRELIMINARY CHECKS -----------------
	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}
	// then make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "host zone is invalid: %s", msg.HostZone)
	}

	if hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone halted for zone (%s)", msg.HostZone))
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for zone (%s)", msg.HostZone)
	}

	// first construct a user redemption record
	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker found: %s", "day")
	}

	// ensure the recipient address is a valid bech32 address on the hostZone
	_, err = utils.AccAddressFromBech32(msg.Receiver, hostZone.Bech32Prefix)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	// construct desired unstaking amount from host zone
	// TODO [cleanup]: Consider changing to truncate int
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()

	if nativeAmount.GT(hostZone.TotalDelegations) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %v", msg.Amount)
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	if balance.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %v, balance %v: ", msg.Amount, balance.Amount)
	}

	// ----------------- UNBONDING RECORD KEEPING -----------------
	// Fetch the record
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
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, msg.Amount))
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.DepositAddress, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoins(ctx, sender, depositAddress, redeemCoin)
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
	updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding)
	if !success {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v", epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding))
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	k.Logger(ctx).Info(fmt.Sprintf("executed redeem stake: %s", msg.String()))
	return &types.MsgRedeemStakeResponse{}, nil
}
