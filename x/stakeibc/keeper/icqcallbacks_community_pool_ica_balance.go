package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v29/utils"
	icqkeeper "github.com/Stride-Labs/stride/v29/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/v29/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

// CommunityPoolBalanceCallback is a callback handler for CommunityPoolBalance queries.
// The query response will return the balance for a specific denom in the deposit or return ica

// If the address queried was the deposit ICA address, call TransferCommunityPoolDepositToHolding
// If the address queried was the return ICA address, call FundCommunityPool

// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func CommunityPoolIcaBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_CommunityPoolIcaBalance,
		"Starting community pool balance callback, QueryId: %vs, QueryType: %s, Connection: %s",
		query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance, denom, and icaType
	//  get amount from the query response, get denom and icaType from marshalled callback data
	amount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine amount from query response")
	}

	// Unmarshal the callback data containing the denom being queried
	var callbackData types.CommunityPoolBalanceQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal community pool balance query callback data")
	}
	icaType := callbackData.IcaType
	denom := callbackData.Denom

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolIcaBalance,
		"Query response - Community Pool Balance: %+v %s %s", amount, icaType.String(), denom))

	// Confirm the balance is greater than zero for now...
	// ...perhaps use a positive threshold in the future to avoid work when transfer would be small
	if amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolIcaBalance,
			"No need to transfer tokens -- not enough found %v %s", amount, denom))
		return nil
	}

	token := sdk.NewCoin(denom, amount)

	// Based on the account type, we kick off the relevant ICA (transfer or fund)
	// If either of the ICAs fails midway through it's invocation, we swallow the
	// error and revert any partial state so that the query response submission can finish
	if icaType == types.ICAAccountType_COMMUNITY_POOL_DEPOSIT {
		// Send ICA msg to kick off transfer from deposit ICA to stake holding address
		err := utils.ApplyFuncIfNoError(ctx, func(c sdk.Context) error {
			return k.TransferCommunityPoolDepositToHolding(ctx, hostZone, token)
		})
		if err != nil {
			k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolIcaBalance,
				"Initiating transfer to holding address failed: %s", err.Error()))
		}
	} else if icaType == types.ICAAccountType_COMMUNITY_POOL_RETURN {
		// Send ICA msg to FundCommunityPool with token found in return ICA
		err := utils.ApplyFuncIfNoError(ctx, func(c sdk.Context) error {
			return k.FundCommunityPool(ctx, hostZone, token, icaType)
		})
		if err != nil {
			k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolIcaBalance,
				"Initiating community pool fund failed: %s", err.Error()))
		}
	}

	return nil
}
