package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cast"

	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// WithdrawalBalanceCallback is a callback handler for WithdrawalBalance queries.
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback executing, QueryId: %vs, Host: %s, QueryType: %s, Connection: %s",
		query.Id, query.ChainId, query.QueryType, query.ConnectionId))

	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Unmarshal the CB args into a coin type
	withdrawalBalanceCoin := sdk.Coin{}
	err := k.cdc.Unmarshal(args, &withdrawalBalanceCoin)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal balance in callback args for zone: %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}

	// Check if the coin is nil (which would indicate the account never had a balance)
	if withdrawalBalanceCoin.IsNil() || withdrawalBalanceCoin.Amount.IsNil() {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: balance query returned a nil coin for address %s on %s, meaning the account has never had a balance on the host",
			hostZone.WithdrawalAccount.GetAddress(), hostZone.ChainId))
		return nil
	}

	// Confirm the balance is greater than zero
	if withdrawalBalanceCoin.Amount.LTE(sdk.ZeroInt()) {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: no balance to transfer for zone: %s, accAddr: %v, coin: %v",
			hostZone.ChainId, hostZone.WithdrawalAccount.GetAddress(), withdrawalBalanceCoin.String()))
		return nil
	}

	// Sweep the withdrawal account balance, to the commission and the delegation accounts
	k.Logger(ctx).Info(fmt.Sprintf("ICA Bank Sending %v%s from withdrawalAddr to delegationAddr.",
		withdrawalBalanceCoin.Amount, withdrawalBalanceCoin.Denom))

	withdrawalAccount := hostZone.GetWithdrawalAccount()
	if withdrawalAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no withdrawal account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}
	delegationAccount := hostZone.GetDelegationAccount()
	if delegationAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no delegation account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}
	feeAccount := hostZone.GetFeeAccount()
	if feeAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no fee account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}

	params := k.GetParams(ctx)
	strideCommissionInt, err := cast.ToInt64E(params.GetStrideCommission())
	if err != nil {
		return err
	}

	// check that stride commission is between 0 and 1
	strideCommission := sdk.NewDec(strideCommissionInt).Quo(sdk.NewDec(100))
	if strideCommission.LT(sdk.ZeroDec()) || strideCommission.GT(sdk.OneDec()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Aborting reinvestment callback -- Stride commission must be between 0 and 1!")
	}

	withdrawalBalanceAmount := withdrawalBalanceCoin.Amount
	strideClaim := strideCommission.Mul(withdrawalBalanceAmount.ToDec())
	strideClaimFloored := strideClaim.TruncateInt()

	// back the reinvestment amount out of the total less the commission
	reinvestAmountCeil := sdk.NewInt(withdrawalBalanceAmount.Int64()).Sub(strideClaimFloored)

	// TODO(TEST-112) safety check, balances should add to original amount
	if (strideClaimFloored.Int64() + reinvestAmountCeil.Int64()) != withdrawalBalanceAmount.Int64() {
		ctx.Logger().Error(fmt.Sprintf("Error with withdraw logic: %v, Fee portion: %v, reinvestPortion %v", withdrawalBalanceAmount, strideClaimFloored, reinvestAmountCeil))
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Failed to subdivide rewards to feeAccount and delegationAccount")
	}
	strideCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, strideClaimFloored)
	reinvestCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, reinvestAmountCeil)

	var msgs []sdk.Msg
	if strideCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.GetAddress(),
			ToAddress:   feeAccount.GetAddress(),
			Amount:      sdk.NewCoins(strideCoin),
		})
	}
	if reinvestCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.GetAddress(),
			ToAddress:   delegationAccount.GetAddress(),
			Amount:      sdk.NewCoins(reinvestCoin),
		})
	}
	ctx.Logger().Info(fmt.Sprintf("Submitting withdrawal sweep messages for: %v", msgs))

	// add callback data
	reinvestCallback := types.ReinvestCallback{
		ReinvestAmount: reinvestCoin,
		HostZoneId:     hostZone.ChainId,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling ReinvestCallback args: %v", reinvestCallback))
	marshalledCallbackArgs, err := k.MarshalReinvestCallbackArgs(ctx, reinvestCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, *withdrawalAccount, ICACallbackID_Reinvest, marshalledCallbackArgs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to SubmitTxs for %s - %s, Messages: %v | err: %s", hostZone.ChainId, hostZone.ConnectionId, msgs, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICATxFailed, errMsg)
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
