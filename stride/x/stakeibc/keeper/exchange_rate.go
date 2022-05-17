package keeper

import (
	"fmt"

	icqtypes "github.com/Stride-labs/stride/x/interchainquery/types"
	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// // is this the right keeper?
// func (k Keeper) getStTokenExchRate(goCtx context.Context, hostZone types.HostZone, inclOutstandingRewards bool) (sdk.Dec, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// TODO() enforce order for ICQ calls;
// 	// with these ICQ calls running async and writing to HostZone store, it's possible one doesn't finish in time
// 	// update delegated balance
// 	k.ICQGetHostStakedBalance(ctx, hostZone)

// 	// ICQ accrued rewards
// 	if inclOutstandingRewards {
// 		k.GetHostAccruedRewardsBalance(ctx, hostZone)
// 		outstandingRewardsOfVirtualPoolOnHost := k.Interchain.Distribution.OutstandingRewards()
// 		balOfVirtualPoolOnHost := delegatedBalOfVirtualPoolOnHost + outstandingRewardsOfVirtualPoolOnHost
// 	} else {
// 		balOfVirtualPoolOnHost := delegatedBalOfVirtualPoolOnHost
// 	}

// 	// Read stAsset supply
// 	// supplyOfStTokens := k.bankKeeper.Supply(stDenom(inCoin.Denom))
// 	// k.Logger(ctx).Info("stAsset outstanding supply:", supplyOfStTokens)

// 	exchRate := balOfVirtualPoolOnHost.toDec() / supplyOfStTokens.toDec()

// 	return exchRate
// }

// set store var to balance on host (either delegated balance, accumulated rewards or totalbalance)
func (k Keeper) ICQGetHostBalance(ctx sdk.Context, hostZone types.HostZone, query_type string) error {

	// does this func need to be a "Callback" type?
	// seems like callback functions can't return anything (why?) so we'll need to write the result to state
	var cbTotalBalance = func(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
		zone, found := k.GetHostZoneInfo(ctx, query.GetChainId())
		if !found {
			return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
		}

		queryRes := banktypes.QueryAllBalancesResponse{}
		err := k.cdc.UnmarshalJSON(args, &queryRes)
		if err != nil {
			k.Logger(ctx).Error("Unable to unmarshal validators info for zone", "zone", zone.ChainId, "err", err)
			return err
		}
		hostZone.TotalAllBalances = sdk.Dec(queryRes.Balances.AmountOf(hostZone.BaseDenom))
		return nil
	}

	var cbDelegations = func(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
		_, found := k.GetHostZoneInfo(ctx, query.GetChainId())
		if !found {
			return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
		}

		var response stakingtypes.QueryDelegatorDelegationsResponse
		err := k.cdc.UnmarshalJSON(args, &response)
		if err != nil {
			return err
		}

		var delegatorSum sdk.Dec = sdk.ZeroDec()
		for _, delegation := range response.DelegationResponses {
			//TODO make sure delegation.Balance is type sdk.Dec or this will error
			delegatorSum = delegatorSum.Add(delegation.Balance.Amount.ToDec())
			if err != nil {
				return err
			}
		}
		// TODO make sure this is writing to state rather than in memory (I suspect it's in memory bc passed by value)
		hostZone.TotalDelegatorDelegations = delegatorSum

		return nil
	}

	var cbAccrRewards = func(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
		_, found := k.GetHostZoneInfo(ctx, query.GetChainId())
		if !found {
			return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
		}
		//TODO(TEST-46) Figure out how to query outstanding rewards for a (validator, delegator) pair using ICQ
		// TODO make sure this is writing to state rather than in memory (I suspect it's in memory bc passed by value)
		hostZone.TotalOutstandingRewards = sdk.ZeroDec()
		return nil
	}

	switch query_type {
	case "totalBalance":
		for _, da := range hostZone.DelegationAccounts {
			k.ICQKeeper.MakeRequest(
				ctx,
				hostZone.ConnectionID,
				hostZone.ChainId,
				"cosmos.bank.v1beta1.Query/AllBalances",
				map[string]string{"address": da.Address},
				sdk.NewInt(-1),
				types.ModuleName,
				cbTotalBalance,
			)
		}
	case "accrRewards":
		for _, da := range hostZone.DelegationAccounts {
			k.ICQKeeper.MakeRequest(
				ctx,
				hostZone.ConnectionID,
				hostZone.ChainId,
				// Figure out how to query outstanding rewards for a (validator, delegator) pair using ICQ.
				// Currently querying for the entire validator, so this is an overestimate
				"cosmos.distribution.v1beta1.ValidatorOutstandingRewards",
				map[string]string{"address": da.Address},
				sdk.NewInt(-1),
				types.ModuleName,
				cbAccrRewards,
			)
		}
	case "delegations":
		for _, da := range hostZone.DelegationAccounts {
			k.ICQKeeper.MakeRequest(
				ctx,
				hostZone.ConnectionID,
				hostZone.ChainId,
				"cosmos.staking.v1beta1.Query/DelegatorDelegations",
				map[string]string{"address": da.Address},
				sdk.NewInt(-1),
				types.ModuleName,
				cbDelegations,
			)
		}
	}

	return nil
}
