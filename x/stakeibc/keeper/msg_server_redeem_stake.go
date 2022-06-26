package keeper

import (
	"context"
	"fmt"
	"strconv"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// vars
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "address is invalid: %s", msg.Creator)
	}
	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "receiver address is invalid: %s", msg.Receiver)
	}
	// we get stAsset from the host zone
	// first make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone is invalid: %s", msg.HostZone)
	}
	coinString := strconv.Itoa(int(msg.Amount)) + "st" + hostZone.IBCDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse inCoin: %s", coinString)
	}
	// remove st prefix to get the base denom
	delegationAccount := hostZone.GetDelegationAccount()
	connectionId := hostZone.GetConnectionId()

	// Safety checks
	// Redemption amount must be positive
	if !inCoin.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %s", msg.Amount)
	}
	// Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, hostZone.IBCDenom)
	if balance.IsLT(inCoin) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %s, balance %s: ", msg.Amount, balance.Amount)
	}

	// calculate the redemption rate
	// when redeeming tokens, multiply stAssets by the exchange rate (allStakedAssets / allStAssets)
	// TODO(TEST-7): Update redemption_rate via ICQ
	var rate sdk.Dec
	rate = hostZone.LastRedemptionRate
	// QUESTION: should we give the lower of the two rates here?
	if hostZone.RedemptionRate.LT(rate) {
		rate = hostZone.RedemptionRate
	}
	native_tokens := inCoin.Amount.ToDec().Mul(rate).TruncateInt()
	outCoin := sdk.NewCoin(hostZone.HostDenom, native_tokens)
	_ = outCoin
	// Select validators for unbonding
	// TODO(TEST-39): Implement validator selection
	validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
	_ = validator_address

	// Implement record keeping logic!
	recordsKeeper := k.recordsKeeper
	// TODO I thought we had parameterized stride_epoch? if so, change this to parameter
	epochInfo, found := k.epochsKeeper.GetEpochInfo(ctx, "stride_epoch")
	currentEpoch := epochInfo.CurrentEpoch
	senderAddr := sender.String()
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "epoch not found: %s", "stride_epoch")
	}
	redemptionId := fmt.Sprintf("%s.%d.%s", hostZone.ChainId, currentEpoch, senderAddr) // {chain_id}.{epoch}.{sender}
	userRedemptionRecord := recordstypes.UserRedemptionRecord{
		Id:          redemptionId,
		Sender:      senderAddr,
		Receiver:    receiver.String(),
		Amount:      inCoin.Amount.Uint64(),
		Denom:       hostZone.HostDenom,
		HostZoneId:  hostZone.ChainId,
		EpochNumber: currentEpoch,
		IsClaimable: false,
	}
	_, found = recordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if found {
		return nil, sdkerrors.Wrapf(recordstypes.ErrRedemptionAlreadyExists, "user already redeemed this epoch: %s", redemptionId)
	}
	recordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Escrow user's balance
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(inCoin))
	if err != nil {
		k.Logger(ctx).Info("Failed to send sdk.NewCoins(inCoins) from account to module")
		panic(err)
	}

	// Construct the transaction. Note, this transaction must be atomically executed.
	var msgs []sdk.Msg
	//msgs = append(msgs, types.NewMsgRedeemStake(sender, connectionId, outCoin))
	// Send the ICA transaction
	k.SubmitTxs(ctx, connectionId, msgs, *delegationAccount)

	return &types.MsgRedeemStakeResponse{}, nil
}
