package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/utils"
)

func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("redeem stake: %s", msg.String()))

	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("creator address is invalid: %s. err: %s: %s", msg.Creator, err.Error(), types.ErrInvalidAddress.Error())
	}
	// then make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, fmt.Errorf("host zone is invalid: %s: %s", msg.HostZone, types.ErrInvalidHostZone.Error())
	}
	// first construct a user redemption record
	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return nil, fmt.Errorf("epoch tracker found: %s: %s", "day", types.ErrEpochNotFound.Error())
	}
	senderAddr := sender.String()
	redemptionId := recordstypes.UserRedemptionRecordKeyFormatter(hostZone.ChainId, epochTracker.EpochNumber, senderAddr)
	_, found = k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if found {
		return nil, fmt.Errorf("user already redeemed this epoch: %s: %s", redemptionId, recordstypes.ErrRedemptionAlreadyExists.Error())
	}

	// ensure the recipient address is a valid bech32 address on the hostZone
	// TODO(TEST-112) do we need to check the hostZone before this check? Would need access to keeper
	_, err = utils.AccAddressFromBech32(msg.Receiver, hostZone.Bech32Prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid receiver address (%s): %s", err, types.ErrInvalidAddress.Error())
	}

	// construct desired unstaking amount from host zone
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()

	if nativeAmount.GT(hostZone.StakedBal) {
		return nil, fmt.Errorf("cannot unstake an amount g.t. staked balance on host zone: %v: %s", msg.Amount, types.ErrInvalidAmount.Error())
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, fmt.Errorf("%s: %s", errMsg, types.ErrRedemptionRateOutsideSafetyBounds.Error())
	}

	// TODO(TEST-112) bigint safety
	coinString := nativeAmount.String() + stDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, fmt.Errorf("could not parse inCoin: %s. err: %s: %s", coinString, err.Error(), types.ErrInvalidCoins.Error())
	}
	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, fmt.Errorf("amount must be greater than 0. found: %v: %s", msg.Amount, types.ErrInvalidCoins.Error())
	}
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer IBCDenom balance: %v%s", balance.Amount, balance.Denom))
	k.Logger(ctx).Info(fmt.Sprintf("Redemption requested redemotion amount: %v%s", inCoin.Amount, inCoin.Denom))
	if balance.Amount.LT(msg.Amount) {
		return nil, fmt.Errorf("balance is lower than redemption amount. redemption amount: %v, balance %v: %s", msg.Amount, balance.Amount, types.ErrInvalidCoins.Error())
	}
	// UNBONDING RECORD KEEPING
	userRedemptionRecord := recordstypes.UserRedemptionRecord{
		Id:          redemptionId,
		Sender:      senderAddr,
		Receiver:    msg.Receiver,
		Amount:      nativeAmount,
		Denom:       hostZone.HostDenom,
		HostZoneId:  hostZone.ChainId,
		EpochNumber: epochTracker.EpochNumber,
		// claimIsPending represents whether a redemption is currently being claimed,
		// contingent on the host zone unbonding having status CLAIMABLE
		ClaimIsPending: false,
	}
	// then add undelegation amount to epoch unbonding records
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochTracker.EpochNumber)
	if !found {
		k.Logger(ctx).Error("latest epoch unbonding record not found")
		return nil, fmt.Errorf("latest epoch unbonding record not found: %s ", recordstypes.ErrEpochUnbondingRecordNotFound.Error())
	}
	// get relevant host zone on this epoch unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, fmt.Errorf("host zone not found in unbondings: %s: %s", hostZone.ChainId, types.ErrInvalidHostZone.Error())
	}
	hostZoneUnbonding.NativeTokenAmount = hostZoneUnbonding.NativeTokenAmount.Add(nativeAmount)
	hostZoneUnbonding.UserRedemptionRecords = append(hostZoneUnbonding.UserRedemptionRecords, userRedemptionRecord.Id)

	// Escrow user's balance
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, msg.Amount))
	bech32ZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.Address, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoins(ctx, sender, bech32ZoneAddress, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, fmt.Errorf("couldn't send %v derivative %s tokens to module account. err: %s: %s", msg.Amount, hostZone.HostDenom, err.Error(), types.ErrInsufficientFunds.Error())
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
		return nil, fmt.Errorf("couldn't set host zone epoch unbonding record. err: %s: %s", err.Error(), types.ErrEpochNotFound.Error())
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	k.Logger(ctx).Info(fmt.Sprintf("executed redeem stake: %s", msg.String()))
	return &types.MsgRedeemStakeResponse{}, nil
}
