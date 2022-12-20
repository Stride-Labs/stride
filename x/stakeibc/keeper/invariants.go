package keeper

// DONTCOVER

import (
	"fmt"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "balance-stake-hostzone-invariant", BalanceStakeHostZoneInvariant(k))
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		// return "", false
		msg, broke := BalanceStakeHostZoneInvariant(k)(ctx)
		return msg, broke

	}
}

// func RedemptionRateInvariant(k Keeper) sdk.Invariant {
// 	return func(ctx sdk.Context) (string, bool) {
// 		listHostZone := k.GetAllHostZone(ctx)
// 	}
// }

// BalanceStakeHostZoneInvariant ensure that balance stake of all host zone are equal to of validator's delegation
func BalanceStakeHostZoneInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)
		for _, host := range listHostZone {
			balanceStake := host.StakedBal
			totalDelegateOfVals := k.GetTotalValidatorDelegations(host)
			if !balanceStake.Equal(totalDelegateOfVals) {
				return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant",
					fmt.Sprintf("\tBalance stake of hostzone %s is not equal to total of validator's delegations \n\tBalance stake actually: %d\n\t Total of validator's delegations: %d\n",
						host.ChainId, host.StakedBal, totalDelegateOfVals,
					)), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant", "All host zones have balances stake is equal to total of validator's delegations"), false
	}
}

// BalanceUnbondedTokensOfRedemptionAccountInvariant ensure that total unbonded tokens from the delegation account is equal to redemption account
// func BalanceUnbondedTokensOfRedemptionAccountInvariant(k Keeper) sdk.Invariant {
// 	return func(ctx sdk.Context) (string, bool) {
// 		// listHostZone := k.GetAllHostZone(ctx)
// 		// for _, host := range listHostZone {

// 		// }
// 	}
// }

func getTotalUnbondedTokens(ctx sdk.Context, k Keeper, hostZone types.HostZone) (sdk.Int, bool) {

	totalAmtTransferToRedemptionAcct := sdk.ZeroInt()

	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return sdk.ZeroInt(), false
	}
	epochNumber := strideEpochTracker.GetEpochNumber()
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
	if !found {
		return sdk.ZeroInt(), false
	}
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
	if !found {
		return sdk.ZeroInt(), false
	}
	blockTime, err := k.GetLightClientTimeSafely(ctx, hostZone.ConnectionId)
	if err != nil {
		return sdk.ZeroInt(), false
	}
	inTransferQueue := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
	validUnbondingTime := hostZoneUnbonding.UnbondingTime > 0 && hostZoneUnbonding.UnbondingTime < blockTime
	if inTransferQueue && validUnbondingTime {
		totalAmtTransferToRedemptionAcct = hostZoneUnbonding.NativeTokenAmount
	}
	return totalAmtTransferToRedemptionAcct, true
}
