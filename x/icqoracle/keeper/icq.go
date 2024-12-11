package keeper

import (
	"fmt"
	"regexp"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/utils"
	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
	icqkeeper "github.com/Stride-Labs/stride/v24/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/v24/x/interchainquery/types"
)

// Submits an ICQ to get a validator's shares to tokens rate
func (k Keeper) SubmitSpotPriceV2CallbackICQ(
	ctx sdk.Context,
	tokenPrice types.TokenPrice,
) error {
	k.Logger(ctx).Info(utils.LogWithPriceToken(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, "Submitting SpotPriceV2 ICQ"))

	params := k.GetParams(ctx)

	// Submit validator sharesToTokens rate ICQ
	// Considering this query is executed manually, we can be conservative with the timeout
	query := icqtypes.Query{
		Id:              fmt.Sprintf("%s|%s-%d", tokenPrice.BaseDenom, tokenPrice.QuoteDenom, ctx.BlockHeight()), // TODO fix?
		ChainId:         params.OsmosisChainId,
		ConnectionId:    params.OsmosisConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF, // TODO fix
		RequestData:     []byte{},                                // TODO fix
		CallbackModule:  types.ModuleName,
		CallbackId:      "banana",                                   // TODO fix
		CallbackData:    []byte{},                                   // TODO fix
		TimeoutDuration: 10 * time.Minute,                           // TODO fix
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST, // TODO fix
	}
	if err := k.icqKeeper.SubmitICQRequest(ctx, query, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, "Error submitting SpotPriceV2 ICQ, error '%s'", err.Error()))
		return err
	}

	if err := k.SetTokenPriceQueryInProgress(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, "Error updating queryInProgress=true, error '%s'", err.Error()))
		return err
	}

	return nil
}

var queryIdRegex = regexp.MustCompile(`^(.+)\|(.+)-(\d+)$`)

// Helper function to parse the ID string
func ParseQueryID(id string) (baseDenom, quoteDenom, blockHeight string, ok bool) {
	matches := queryIdRegex.FindStringSubmatch(id)
	if len(matches) != 4 {
		return "", "", "", false
	}
	return matches[1], matches[2], matches[3], true
}

func SpotPriceV2Callback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	baseDenom, quoteDenom, _, ok := ParseQueryID(query.Id)
	if !ok {
		return fmt.Errorf("unable to parse baseDenom and quoteDenom from queryId '%s'", query.Id)
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithPriceToken(baseDenom, quoteDenom, "SpotPriceV2",
		"Starting SpotPriceV2 ICQ callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Unmarshal the query response args to determine the balance
	spotPrice, err := icqkeeper.UnmarshalSpotPriceFromSpotPriceV2Query(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine spot price from query response")
	}

	tokenPrice := types.TokenPrice{
		BaseDenom:       baseDenom,
		QuoteDenom:      quoteDenom,
		Price:           spotPrice,
		UpdatedAt:       ctx.BlockTime(),
		QueryInProgress: false,
	}

	if err := k.SetTokenPrice(ctx, tokenPrice); err != nil {
		return errorsmod.Wrap(err, "unable to update spot price from query response")
	}

	return nil

	// k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
	// 	"Query response - Fee Balance: %v %s", feeBalanceAmount, hostZone.HostDenom))

	// // Confirm the balance is greater than zero
	// if feeBalanceAmount.LTE(sdkmath.ZeroInt()) {
	// 	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
	// 		"No balance to transfer for address: %s, balance: %v", hostZone.FeeIcaAddress, feeBalanceAmount))
	// 	return nil
	// }

	// // Confirm the fee account has been initiated
	// if hostZone.FeeIcaAddress == "" {
	// 	return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no fee account found for %s", chainId)
	// }

	// // The ICA and transfer should both timeout before the end of the epoch
	// timeout, err := k.GetICATimeoutNanos(ctx, epochtypes.STRIDE_EPOCH)
	// if err != nil {
	// 	return errorsmod.Wrapf(err, "Failed to get ICATimeout from %s epoch", epochtypes.STRIDE_EPOCH)
	// }

	// // get counterparty chain's transfer channel
	// transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, transfertypes.PortID, hostZone.TransferChannelId)
	// if !found {
	// 	return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", hostZone.TransferChannelId)
	// }
	// counterpartyChannelId := transferChannel.Counterparty.ChannelId

	// // Prepare a MsgTransfer from the fee account to the rewards collector account
	// rewardsCoin := sdk.NewCoin(hostZone.HostDenom, feeBalanceAmount)
	// rewardsCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
	// transferMsg := ibctypes.NewMsgTransfer(
	// 	transfertypes.PortID,
	// 	counterpartyChannelId,
	// 	rewardsCoin,
	// 	hostZone.FeeIcaAddress,
	// 	rewardsCollectorAddress.String(),
	// 	clienttypes.Height{},
	// 	timeout,
	// 	"",
	// )

	// msgs := []proto.Message{transferMsg}
	// k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_FeeBalance,
	// 	"Preparing MsgTransfer of %v from the fee account to the rewards collector module account (for commission)", rewardsCoin.String()))

	// // Send the transaction through SubmitTx
	// if _, err := k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_FEE, ICACallbackID_Reinvest, nil); err != nil {
	// 	return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	// }

	// return nil
}
