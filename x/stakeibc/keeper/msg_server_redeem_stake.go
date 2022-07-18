package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/Stride-Labs/stride/utils"
)

func (k Keeper) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "address is invalid: %s", msg.Creator)
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
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Amount > hostZone.StakedBal {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %d", msg.Amount)
	}

	// construct desired unstaking amount from host zone
	coinDenom := "st" + hostZone.HostDenom
	nativeAmount := sdk.NewDec(msg.Amount).Mul(hostZone.RedemptionRate)
	// TODO(TEST-112) bigint safety
	coinString := nativeAmount.RoundInt().String() + coinDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse inCoin: %s", coinString)
	}
	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !inCoin.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %d", msg.Amount)
	}
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, coinDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer IBCDenom balance: %d%s", balance.Amount, balance.Denom))
	k.Logger(ctx).Info(fmt.Sprintf("Redemption requested redemotion amount: %v%s", inCoin.Amount, inCoin.Denom))
	if balance.Amount.LT(sdk.NewInt(msg.Amount)) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %d, balance %d: ", msg.Amount, balance.Amount)
	}
	// UNBONDING RECORD KEEPING
	// first construct a user redemption record
	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "epoch tracker found: %s", "day")
	}
	senderAddr := sender.String()
	redemptionId := recordstypes.UserRedemptionRecordKeyFormatter(hostZone.ChainId, epochTracker.EpochNumber, senderAddr)
	userRedemptionRecord := recordstypes.UserRedemptionRecord{
		Id:          redemptionId,
		Sender:      senderAddr,
		Receiver:    msg.Receiver,
		Amount:      inCoin.Amount.Uint64(),
		Denom:       hostZone.HostDenom,
		HostZoneId:  hostZone.ChainId,
		EpochNumber: int64(epochTracker.EpochNumber),
		IsClaimable: false,
	}
	_, found = k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if found {
		return nil, sdkerrors.Wrapf(recordstypes.ErrRedemptionAlreadyExists, "user already redeemed this epoch: %s", redemptionId)
	}
	// then add undelegation amount to epoch unbonding records
	epochUnbondingRecord, found := k.RecordsKeeper.GetLatestEpochUnbondingRecord(ctx)
	if !found {
		k.Logger(ctx).Error("latest epoch unbonding record not found")
		return nil, sdkerrors.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, "latest epoch unbonding record not found")
	}
	// get relevant host zone on this epoch unbonding record
	hostZoneUnbonding, found := epochUnbondingRecord.HostZoneUnbondings[hostZone.ChainId]
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone not found in unbondings: %s", hostZone.ChainId)
	}
	hostZoneUnbonding.Amount += inCoin.Amount.Uint64()
	hostZoneUnbonding.UserRedemptionRecords = append(hostZoneUnbonding.UserRedemptionRecords, userRedemptionRecord.Id)

	// Escrow user's balance
	redeemCoin := sdk.NewCoins(sdk.NewCoin(coinDenom, sdk.NewInt(msg.Amount)))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, redeemCoin)
	if err != nil {
		k.Logger(ctx).Info("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %d %s tokens to module account", msg.Amount, coinDenom)
	}

	// burn stAssets upon successful unbonding
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %v", err))
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't burn %d %s tokens in module account", msg.Amount, coinDenom)
	}

	// Actually set the records, we wait until now to prevent any errors
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Set the UserUnbondingRecords on the proper HostZoneUnbondingRecord
	hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
	if len(hostZoneUnbondings) == 0 {
		hostZoneUnbondings = make(map[string]*recordstypes.HostZoneUnbonding)
		epochUnbondingRecord.HostZoneUnbondings = hostZoneUnbondings
	}
	epochUnbondingRecord.HostZoneUnbondings[hostZone.ChainId] = hostZoneUnbonding
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)

	return &types.MsgRedeemStakeResponse{}, nil
}
