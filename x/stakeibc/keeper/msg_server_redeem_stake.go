package keeper

import (
	"context"
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// vars
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "address is invalid: %s", msg.Creator)
	}
	coinString := strconv.Itoa(int(msg.Amount)) + msg.Denom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse inCoin: %s", coinString)
	}
	hostZone, err := k.GetHostZoneFromIBCDenom(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}
	delegationAccount := hostZone.GetDelegationAccount()
	withdrawAccount := hostZone.GetWithdrawalAccount()
	connectionId := hostZone.GetConnectionId()
	
	// Safety checks
	// Redemption amount must be positive
	if !inCoin.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %s", msg.Amount)
	}
	// Denom is valid
	// Should we register stAssets somewhere and add an additional check here?
	if types.IsStAsset(msg.Denom) != true {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "denom is not a valid stAsset. found: %s", msg.Denom)
	}
	// Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, msg.Denom)
	if balance.IsLT(inCoin) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %s, balance %s: ", msg.Amount, balance.Amount)
	}

	// Escrow user's balance
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(inCoin))
	if err != nil {
		k.Logger(ctx).Info("Failed to send sdk.NewCoins(inCoins) from account to module")
		panic(err)
	}

	// calculate the redemption rate
	// when redeeming tokens, multiply stAssets by the exchange rate (allStakedAssets / allStAssets)
	// TODO(TEST-7): Update redemption_rate via ICQ
	var rate sdk.Dec
	rate = hostZone.LastRedemptionRate
	if hostZone.RedemptionRate.LT(rate) {
		rate = hostZone.RedemptionRate
	}
	native_tokens := inCoin.Amount.ToDec().Mul(rate).TruncateInt()
	outCoin := sdk.NewCoin(hostZone.HostDenom, native_tokens)

	// Select validators for unbonding
	// TODO(TEST-39): Implement validator selection
	validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e"  // gval2

	// Construct the transaction. Note, this transaction must be atomically executed.
	var msgs []sdk.Msg
	// 1. MsgSetWithdrawalAddress
	setWithdrawAddressUser := &distributionTypes.MsgSetWithdrawAddress{DelegatorAddress: delegationAccount.GetAddress(), WithdrawAddress: sender.String()}
	msgs = append(msgs, setWithdrawAddressUser)
	// 2. MsgUndelegate
	undelegateToUser := &stakingTypes.MsgUndelegate{DelegatorAddress: delegationAccount.GetAddress(), ValidatorAddress: validator_address, Amount: outCoin}
	msgs = append(msgs, undelegateToUser)
	// 3. MsgSetWithdrawalAddress
	setWithdrawAddressIca := &distributionTypes.MsgSetWithdrawAddress{DelegatorAddress: delegationAccount.GetAddress(), WithdrawAddress: withdrawAccount.GetAddress()}
	msgs = append(msgs, setWithdrawAddressIca)
	// Send the ICA transaction
	k.SubmitTxs(ctx, connectionId, msgs, *delegationAccount)

	return &types.MsgRedeemStakeResponse{}, nil
}
