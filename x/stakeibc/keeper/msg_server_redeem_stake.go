package keeper

import (
	"context"
	"fmt"

	"github.com/spf13/cast"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/utils"
)

func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}
	// then make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone is invalid: %s", msg.HostZone)
	}

	// ensure the recipient address is a valid bech32 address on the hostZone
	// TODO(TEST-112) do we need to check the hostZone before this check? Would need access to keeper
	_, err = utils.AccAddressFromBech32(msg.Receiver, hostZone.Bech32Prefix)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	if msg.Amount > hostZone.StakedBal {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %d", msg.Amount)
	}

	// construct desired unstaking amount from host zone
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	nativeAmount := sdk.NewDec(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()

	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %d", msg.Amount)
	}
	// 	- Creator owns at least "amount" stAssets
	senderStAssetBalance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer stAsset balance: %v%s", senderStAssetBalance.Amount, senderStAssetBalance.Denom))
	k.Logger(ctx).Info(fmt.Sprintf("Redemption requested redemption amount: %v%s", msg.Amount, hostZone.HostDenom))
	if senderStAssetBalance.Amount.LT(sdk.NewInt(msg.Amount)) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins,
			"sender balance is less than redemption amount. redemption amount: %d, balance %d: ", msg.Amount, senderStAssetBalance.Amount)
	}
	// UNBONDING RECORD KEEPING
	// first construct a user redemption record
	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "epoch tracker not found: %s", "day")
	}
	epochNumberInt, err := cast.ToInt64E(epochTracker.EpochNumber)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrIntCast, "epoch number in redeem stake")
	}
	senderAddr := sender.String()
	redemptionId := recordstypes.UserRedemptionRecordKeyFormatter(hostZone.ChainId, epochTracker.EpochNumber, senderAddr)
	userRedemptionRecord := recordstypes.UserRedemptionRecord{
		Id:          redemptionId,
		Sender:      senderAddr,
		Receiver:    msg.Receiver,
		Amount:      nativeAmount.Uint64(),
		Denom:       hostZone.HostDenom,
		HostZoneId:  hostZone.ChainId,
		EpochNumber: epochNumberInt,
		IsClaimable: false,
	}
	_, found = k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if found {
		return nil, sdkerrors.Wrapf(recordstypes.ErrRedemptionAlreadyExists, "user already redeemed this epoch: %s", redemptionId)
	}
	// then add undelegation amount to epoch unbonding records
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochTracker.EpochNumber)
	if !found {
		k.Logger(ctx).Error("latest epoch unbonding record not found")
		return nil, sdkerrors.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, "latest epoch unbonding record not found")
	}
	// get relevant host zone on this epoch unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone not found in unbondings: %s", hostZone.ChainId)
	}
	hostZoneUnbonding.NativeTokenAmount += nativeAmount.Uint64()
	hostZoneUnbonding.UserRedemptionRecords = append(hostZoneUnbonding.UserRedemptionRecords, userRedemptionRecord.Id)

	// Escrow user's balance
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, sdk.NewInt(msg.Amount)))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %d %s tokens to module account. err: %s", msg.Amount, stDenom, err.Error())
	}

	// record the number of stAssets that should be burned after unbonding
	stTokenAmount, err := cast.ToUint64E(msg.Amount)
	if err != nil {
		errMsg := fmt.Sprintf("Could not convert redemption amount to int64 in redeem stake | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}
	hostZoneUnbonding.StTokenAmount += stTokenAmount

	// Actually set the records, we wait until now to prevent any errors
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Set the UserUnbondingRecords on the proper HostZoneUnbondingRecord
	hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
	if len(hostZoneUnbondings) == 0 {
		hostZoneUnbondings = []*recordstypes.HostZoneUnbonding{}
		epochUnbondingRecord.HostZoneUnbondings = hostZoneUnbondings
	}
	updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding)
	if !success {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v", epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding))
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	return &types.MsgRedeemStakeResponse{}, nil
}
