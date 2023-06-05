package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqkeeper "github.com/Stride-Labs/stride/v9/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// FeeBalanceCallback is a callback handler for FeeBalnce queries.
// The query response will return the fee account balance
// If the balance is non-zero, an ICA MsgTransfer is initated to the RewardsCollector account
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func FeeBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_FeeBalance,
		"Starting fee balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	feeBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
		"Query response - Fee Balance: %v %s", feeBalanceAmount, hostZone.HostDenom))

	// Confirm the balance is greater than zero
	if feeBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
			"No balance to transfer for address: %v, balance: %v", hostZone.FeeAccount.GetAddress(), feeBalanceAmount))
		return nil
	}

	// Confirm the fee account has been initiated
	feeAccount := hostZone.FeeAccount
	if feeAccount == nil || feeAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no fee account found for %s", chainId)
	}

	// The ICA and transfer should both timeout before the end of the epoch
	timeout, err := k.GetICATimeoutNanos(ctx, epochtypes.STRIDE_EPOCH)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get ICATimeout from %s epoch", epochtypes.STRIDE_EPOCH)
	}

	// get counterparty chain's transfer channel
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, transfertypes.PortID, hostZone.TransferChannelId)
	if !found {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", hostZone.TransferChannelId)
	}
	counterpartyChannelId := transferChannel.Counterparty.ChannelId

	// Prepare a MsgTransfer from the fee account to the rewards collector account
	rewardsCoin := sdk.NewCoin(hostZone.HostDenom, feeBalanceAmount)
	rewardsCollectorAddress := k.accountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
	transferMsg := ibctypes.NewMsgTransfer(
		transfertypes.PortID,
		counterpartyChannelId,
		rewardsCoin,
		feeAccount.Address,
		rewardsCollectorAddress.String(),
		clienttypes.Height{},
		timeout,
		"",
	)

	msgs := []proto.Message{transferMsg}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
		"Preparing MsgTransfer of %v from the fee account to the rewards collector module account (for commission)", rewardsCoin.String()))

	// Send the transaction through SubmitTx
	if _, err := k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, *feeAccount, ICACallbackID_Reinvest, nil); err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}

	return nil
}
