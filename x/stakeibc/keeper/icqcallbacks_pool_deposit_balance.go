package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v14/utils"
	icqkeeper "github.com/Stride-Labs/stride/v14/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// PoolDepositBalanceCallback is a callback handler for CommunityPoolDepositBalance queries.
// The query response will return the deposit account balance for a specific denom
// If the balance is non-zero, call IBCTransferCommunityPoolTokens to transfer tokens to Stride
//  autopilot memo causes (hostDenom -> stake), (stHostDenom -> redeem), or (otherDenom -> do nothing)

// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func CommunityPoolDepositBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_CommunityPoolDepositBalance,
		"Starting community pool deposit balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	communityPoolHostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance *and* the denom
	//  get amount from the query response, get denom from the marshalled callback data
	depositBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine deposit balance from query response")
	}

	// Unmarshal the callback data containing the denom being queried
	var callbackData types.CommunityPoolDepositQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal community pool deposit query callback data")
	}
	depositDenom := callbackData.Denom

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolDepositBalance,
		"Query response - Community Pool Deposit Balance: %v %s", depositBalanceAmount, depositDenom))

	// Confirm the balance is greater than zero for now...
	// ...perhaps use a positive threshold in the future to avoid work when transfer would be small
	if depositBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_CommunityPoolDepositBalance,
			"No need to transfer tokens -- not enough found %v %s", depositBalanceAmount, depositDenom))
		return nil
	}

	// Transfer the tokens and potentially use autopilot to operate on them (stake/redeem) on landing
	autoPilotAction := NoAction
	if depositDenom == communityPoolHostZone.HostDenom {
		autoPilotAction = LiquidStake
	}
	ibcStDenom := k.GetStakedHostTokenDenomOnHostZone(communityPoolHostZone)
	if depositDenom == ibcStDenom {
		autoPilotAction = RedeemStake
	}
	transferCoin := sdk.NewCoin(depositDenom, depositBalanceAmount)

	return k.TransferCommunityPoolDepositToHolding(ctx, communityPoolHostZone, transferCoin, autoPilotAction)
}
