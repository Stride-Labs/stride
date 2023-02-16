package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/utils"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
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
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	withdrawalBalanceAmount, err := UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
		"Query response - Withdrawal Balance: %v %s", withdrawalBalanceAmount, hostZone.HostDenom))

	// Confirm the balance is greater than zero
	if withdrawalBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"No balance to transfer for address: %v, balance: %v", hostZone.WithdrawalAccount.GetAddress(), withdrawalBalanceAmount))
		return nil
	}

	// Get the host zone's ICA accounts
	withdrawalAccount := hostZone.WithdrawalAccount
	if withdrawalAccount == nil || withdrawalAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", chainId)
	}
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}
	feeAccount := hostZone.FeeAccount
	if feeAccount == nil || feeAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no fee account found for %s", chainId)
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
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Aborting withdrawal balance callback -- Stride commission must be between 0 and 1!")
	}

	// Split out the reinvestment amount from the fee amount
	feeAmount := strideCommission.Mul(sdk.NewDecFromInt(withdrawalBalanceAmount)).TruncateInt()
	reinvestAmount := withdrawalBalanceAmount.Sub(feeAmount)

	// Safety check, balances should add to original amount
	if !feeAmount.Add(reinvestAmount).Equal(withdrawalBalanceAmount) {
		k.Logger(ctx).Error(fmt.Sprintf("Error with withdraw logic: %v, Fee Portion: %v, Reinvest Portion %v", withdrawalBalanceAmount, feeAmount, reinvestAmount))
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Failed to subdivide rewards to feeAccount and delegationAccount")
	}

	// Prepare MsgSends from the withdrawal account
	feeCoin := sdk.NewCoin(hostZone.HostDenom, feeAmount)
	reinvestCoin := sdk.NewCoin(hostZone.HostDenom, reinvestAmount)

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
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", withdrawalBalanceAmount.String()),
		),
	)

	return nil
}

// Helper function to unmarshal a Balance query response across SDK versions
// Before SDK v46, the query response returned a sdk.Coin type. SDK v46 returns an int type
// https://github.com/cosmos/cosmos-sdk/pull/9832
func UnmarshalAmountFromBalanceQuery(cdc codec.BinaryCodec, queryResponseBz []byte) (amount sdkmath.Int, err error) {
	// An nil should not be possible, exit immediately if it occurs
	if queryResponseBz == nil {
		return sdkmath.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "query response is nil")
	}

	// If the query response is empty, that means the account was never registed (and thus has a 0 balance)
	if len(queryResponseBz) == 0 {
		return sdkmath.ZeroInt(), nil
	}

	// First attempt to unmarshal as an Int (for SDK v46+)
	// If the result was serialized as a `Coin` type, it should contain a string (representing the denom)
	// which will cause the unmarshalling to throw an error
	intError := amount.Unmarshal(queryResponseBz)
	if intError == nil {
		return amount, nil
	}

	// If the Int unmarshaling was unsuccessful, attempt again using a Coin type (for SDK v45 and below)
	// If successful, return the amount field from the coin (if the coin is not nil)
	var coin sdk.Coin
	coinError := cdc.Unmarshal(queryResponseBz, &coin)
	if coinError == nil {
		return coin.Amount, nil
	}

	// If it failed unmarshaling with either data structure, return an error with the failure messages combined
	return sdkmath.Int{}, errorsmod.Wrapf(types.ErrUnmarshalFailure,
		"unable to unmarshal balance query response %v as sdkmath.Int (err: %s) or sdk.Coin (err: %s)", queryResponseBz, intError.Error(), coinError.Error())
}
