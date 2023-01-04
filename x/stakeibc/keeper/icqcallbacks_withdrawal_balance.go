package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// WithdrawalBalanceCallback is a callback handler for WithdrawalBalance queries.
// The query response will return the withdrawal account balance
// If the balance is non-zero, ICA MsgSends are submitted to transfer from the withdrawal account
//  to the delegation account (for reinvestment) and fee account (for commission)
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_WithdrawalBalance,
		"Starting withdrawal balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args into a coin type
	withdrawalBalanceCoin := sdk.Coin{}
	err := k.cdc.Unmarshal(args, &withdrawalBalanceCoin)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal query response into Coin type: %s, err: %s", err.Error())
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance, "Query response - Coin: %+v", withdrawalBalanceCoin))

	// Check if the coin is nil (which would indicate the account never had a balance)
	if withdrawalBalanceCoin.IsNil() || withdrawalBalanceCoin.Amount.IsNil() {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"Balance query returned a nil coin for address %v, meaning the account has never had a balance on the host",
			hostZone.WithdrawalAccount.GetAddress()))
		return nil
	}

	// Confirm the balance is greater than zero
	if withdrawalBalanceCoin.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"No balance to transfer for address: %v, coin: %v",
			hostZone.ChainId, hostZone.WithdrawalAccount.GetAddress(), withdrawalBalanceCoin.String()))
		return nil
	}

	// Get the host zone's ICA accounts
	withdrawalAccount := hostZone.WithdrawalAccount
	if withdrawalAccount == nil || withdrawalAccount.Address == "" {
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", chainId)
	}
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}
	feeAccount := hostZone.FeeAccount
	if feeAccount == nil || feeAccount.Address == "" {
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, "no fee account found found for %s", chainId)
	}

	// Determine the stride commission rate to the relevant portion can be sent to the fee account
	params := k.GetParams(ctx)
	strideCommissionInt, err := cast.ToInt64E(params.StrideCommission)
	if err != nil {
		return err
	}

	// check that stride commission is between 0 and 1
	strideCommission := sdk.NewDec(strideCommissionInt).Quo(sdk.NewDec(100))
	if strideCommission.LT(sdk.ZeroDec()) || strideCommission.GT(sdk.OneDec()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Aborting withdrawal balance callback -- Stride commission must be between 0 and 1!")
	}

	// Split out the reinvestment amount from the fee amount
	withdrawalBalanceAmount := withdrawalBalanceCoin.Amount
	feeAmount := strideCommission.Mul(sdk.NewDecFromInt(withdrawalBalanceAmount)).TruncateInt()
	reinvestAmount := withdrawalBalanceAmount.Sub(feeAmount)

	// Safety check, balances should add to original amount
	if !feeAmount.Add(reinvestAmount).Equal(withdrawalBalanceAmount) {
		k.Logger(ctx).Error(fmt.Sprintf("Error with withdraw logic: %v, Fee Portion: %v, Reinvest Portion %v", withdrawalBalanceAmount, feeAmount, reinvestAmount))
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Failed to subdivide rewards to feeAccount and delegationAccount")
	}

	// Prepare MsgSends from the withdrawal account
	feeCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, feeAmount)
	reinvestCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, reinvestAmount)

	var msgs []sdk.Msg
	if feeCoin.Amount.GT(sdk.ZeroInt()) {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.Address,
			ToAddress:   feeAccount.Address,
			Amount:      sdk.NewCoins(feeCoin),
		})
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"Preparing MsgSends of %v from the withdrawal account to the fee account (for commission)", feeCoin.String()))
	}
	if reinvestCoin.Amount.GT(sdk.ZeroInt()) {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.Address,
			ToAddress:   delegationAccount.Address,
			Amount:      sdk.NewCoins(reinvestCoin),
		})
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"Preparing MsgSends of %v from the withdrawal account to the delegation account (for reinvestment)", reinvestCoin.String()))
	}

	// add callback data before calling reinvestment ICA
	reinvestCallback := types.ReinvestCallback{
		ReinvestAmount: reinvestCoin,
		HostZoneId:     hostZone.ChainId,
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance, "Marshalling ReinvestCallback args: %v", reinvestCallback))
	marshalledCallbackArgs, err := k.MarshalReinvestCallbackArgs(ctx, reinvestCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, *withdrawalAccount, ICACallbackID_Reinvest, marshalledCallbackArgs)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", withdrawalBalanceCoin.Amount.String()),
		),
	)

	return nil
}
