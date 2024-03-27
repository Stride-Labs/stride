package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v21/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v21/utils"
	icqtypes "github.com/Stride-Labs/stride/v21/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

// WithdrawalHostBalanceCallback is a callback handler for WithdrawalBalance queries.
// The query response will return the withdrawal account balance for the native denom (i.e. "host denom")
// If the balance is non-zero, ICA MsgSends are submitted to transfer from the withdrawal account
// to the delegation account (for reinvestment) and fee account (for commission)
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalHostBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_WithdrawalHostBalance,
		"Starting withdrawal balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	withdrawalBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalHostBalance,
		"Query response - Withdrawal Balance: %v %s", withdrawalBalanceAmount, hostZone.HostDenom))

	// Confirm the balance is greater than zero
	if withdrawalBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalHostBalance,
			"No balance to transfer for address: %s, balance: %v", hostZone.WithdrawalIcaAddress, withdrawalBalanceAmount))
		return nil
	}

	// Get the host zone's ICA accounts
	if hostZone.WithdrawalIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", chainId)
	}
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}
	if hostZone.FeeIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no fee account found for %s", chainId)
	}

	// Split the withdrawal amount into the stride fee and reinvest portion
	rewardsSplit, err := k.CalculateRewardsSplit(ctx, hostZone, withdrawalBalanceAmount)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to split reward amount into fee and reinvest amounts")
	}

	// Prepare MsgSends from the withdrawal account
	feeCoin := sdk.NewCoin(hostZone.HostDenom, rewardsSplit.StrideFeeAmount)
	reinvestCoin := sdk.NewCoin(hostZone.HostDenom, rewardsSplit.ReinvestAmount)
	rebateCoin := sdk.NewCoin(hostZone.HostDenom, rewardsSplit.RebateAmount)

	var msgs []proto.Message
	if feeCoin.Amount.GT(sdkmath.ZeroInt()) {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: hostZone.WithdrawalIcaAddress,
			ToAddress:   hostZone.FeeIcaAddress,
			Amount:      sdk.NewCoins(feeCoin),
		})
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalHostBalance,
			"Preparing MsgSends of %v from the withdrawal account to the fee account (for commission)", feeCoin.String()))
	}
	if reinvestCoin.Amount.GT(sdkmath.ZeroInt()) {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: hostZone.WithdrawalIcaAddress,
			ToAddress:   hostZone.DelegationIcaAddress,
			Amount:      sdk.NewCoins(reinvestCoin),
		})
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalHostBalance,
			"Preparing MsgSends of %v from the withdrawal account to the delegation account (for reinvestment)", reinvestCoin.String()))
	}
	if rebateCoin.Amount.GT(sdkmath.ZeroInt()) {
		fundMsg, err := k.BuildFundCommunityPoolMsg(ctx, hostZone, sdk.NewCoins(rebateCoin), types.ICAAccountType_WITHDRAWAL)
		if err != nil {
			return err
		}
		msgs = append(msgs, fundMsg...)
	}

	// add callback data before calling reinvestment ICA
	reinvestCallback := types.ReinvestCallback{
		ReinvestAmount: reinvestCoin,
		HostZoneId:     hostZone.ChainId,
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalHostBalance, "Marshalling ReinvestCallback args: %v", reinvestCallback))
	marshalledCallbackArgs, err := k.MarshalReinvestCallbackArgs(ctx, reinvestCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_WITHDRAWAL, ICACallbackID_Reinvest, marshalledCallbackArgs)
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
