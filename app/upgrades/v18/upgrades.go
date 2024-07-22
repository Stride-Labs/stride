package v18

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v23/x/records/keeper"
	recordtypes "github.com/Stride-Labs/stride/v23/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v23/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v23/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v23/x/stakeibc/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v18
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v18...")

		ctx.Logger().Info("Updating redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Resetting delegation changes in progress...")
		if err := DecrementTerraDelegationChangesInProgress(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to reset delegation changes in progress")
		}

		ctx.Logger().Info("Updating unbonding records...")
		err := UpdateUnbondingRecords(
			ctx,
			stakeibcKeeper,
			recordsKeeper,
			StartingEstimateEpoch,
			RedemptionRatesBeforeProp,
			RedemptionRatesAtTimeOfProp,
		)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "unable to update unbonding records")
		}

		ctx.Logger().Info(fmt.Sprintf("Checking on prop %d status...", Prop228ProposalId))
		if err := ExecuteProp228IfPassed(ctx, bankKeeper, govKeeper); err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to check on or execute prop %d: %s",
				Prop228ProposalId, err.Error()))
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		// Give osmosis a bit more slack since OSMO stakers collect real yield
		outerAdjustment := RedemptionRateOuterMaxAdjustment
		if hostZone.ChainId == OsmosisChainId {
			outerAdjustment = outerAdjustment.Add(OsmosisRedemptionRateBuffer)
		}

		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(outerAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

// Decrement DelegationChangesInProgress on Terra vals by 3
// - Fetches terra host zone
// - Loops validators
// - Decrements each validator's DelegationChangeInProgress by 3
func DecrementTerraDelegationChangesInProgress(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
) error {

	// grab the terra host zone
	hostZone, found := sk.GetHostZone(ctx, TerraChainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrapf("failed to fetch %s", TerraChainId)
	}

	// iterate the validators
	for _, val := range hostZone.Validators {

		// subtract 3, flooring at 0
		if val.DelegationChangesInProgress < 3 {
			val.DelegationChangesInProgress = 0
		} else {
			val.DelegationChangesInProgress = val.DelegationChangesInProgress - 3
		}
	}

	// set the host zone
	sk.SetHostZone(ctx, hostZone)

	return nil
}

// Modify HostZoneUnbonding and UserRedemptionRecords NativeTokenAmount to reflect new data structs
func UpdateUnbondingRecords(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
	rk recordskeeper.Keeper,
	startingEstimateEpoch uint64,
	redemptionRatesBeforeProp map[string]map[uint64]sdk.Dec,
	redemptionRatesAtTimeOfProp map[string]sdk.Dec,
) error {
	// loop over host zone unbonding records
	for _, epochUnbondingRecord := range rk.GetAllEpochUnbondingRecord(ctx) {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			epochNumber := epochUnbondingRecord.EpochNumber
			chainId := hostZoneUnbonding.HostZoneId

			// we can ignore any record that's not currently unbonding
			if hostZoneUnbonding.Status != recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE {
				continue
			}

			// Grab the redemption rates from before the prop was posted, for a given chain
			// across all the epochs that unbonded
			hostZoneRRBeforeProp, ok := redemptionRatesBeforeProp[hostZoneUnbonding.HostZoneId]
			if !ok {
				ctx.Logger().Error(fmt.Sprintf("Host zone %s not included in redemption rate mapping", chainId))
				continue
			}

			// Grab the redemption rate for this specific epoch
			// If it's not found, that means the unbonding for this epoch occurred after the prop was live
			recordRedemptionRate, recordUnbondedBeforeProp := hostZoneRRBeforeProp[epochUnbondingRecord.EpochNumber]

			if !recordUnbondedBeforeProp && (epochNumber < startingEstimateEpoch) {
				ctx.Logger().Info(fmt.Sprintf("Skipping unbonding record adjustment for chain %s epoch %d",
					chainId, epochNumber))
				continue
			}

			// If we don't have the redemption rate, estimate it
			if !recordUnbondedBeforeProp {
				hostZone, found := sk.GetHostZone(ctx, hostZoneUnbonding.HostZoneId)
				if !found {
					return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound,
						"unable to find host zone with chain-id %s", hostZoneUnbonding.HostZoneId)
				}

				redemptionRateAtTimeOfProp := redemptionRatesAtTimeOfProp[hostZoneUnbonding.HostZoneId]
				redemptionRateDuringUpgrade := hostZone.RedemptionRate
				recordRedemptionRate = redemptionRateAtTimeOfProp.Add(redemptionRateDuringUpgrade).Quo(sdk.NewDec(2))
			}

			// now update all userRedemptionRecords by using the redemption rate to set the native token amount
			totalNativeAmount := sdkmath.ZeroInt()
			for _, userRedemptionRecordId := range hostZoneUnbonding.UserRedemptionRecords {
				userRedemptionRecord, found := rk.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
				if !found {
					return errorsmod.Wrapf(recordtypes.ErrHostUnbondingRecordNotFound,
						"unable to find user redemption record with id %s", userRedemptionRecordId)
				}

				userNativeAmount := sdk.NewDecFromInt(userRedemptionRecord.StTokenAmount).Mul(recordRedemptionRate).TruncateInt()
				totalNativeAmount = totalNativeAmount.Add(userNativeAmount)

				userRedemptionRecord.NativeTokenAmount = userNativeAmount
				rk.SetUserRedemptionRecord(ctx, userRedemptionRecord)
			}

			// finally, update the hostZoneUnbonding record
			hostZoneUnbonding.NativeTokenAmount = totalNativeAmount
			if err := rk.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, *hostZoneUnbonding); err != nil {
				return errorsmod.Wrapf(err, "unable to set host zone unbonding for %s and epoch %d",
					hostZoneUnbonding.HostZoneId, epochUnbondingRecord.EpochNumber)
			}
		}
	}
	return nil
}

// Executes the bank send for prop 228 if it passed
func ExecuteProp228IfPassed(ctx sdk.Context, bk bankkeeper.Keeper, gk govkeeper.Keeper) error {
	// Grab proposal from gov store
	proposal, found := gk.GetProposal(ctx, Prop228ProposalId)
	if !found {
		return fmt.Errorf("Prop %d not found", Prop228ProposalId)
	}

	// Check if it passed - if it didn't do nothing
	if proposal.Status != govtypes.ProposalStatus_PROPOSAL_STATUS_PASSED {
		ctx.Logger().Info(fmt.Sprintf("Prop %d did not pass", Prop228ProposalId))
		return nil
	}
	ctx.Logger().Info(fmt.Sprintf("Prop %d passed - executing corresponding bank send", Prop228ProposalId))

	// Transfer from incentive program address to F4
	fromAddress, err := sdk.AccAddressFromBech32(IncentiveProgramAddress)
	if err != nil {
		return errorsmod.Wrap(err, "invalid prop sender address")
	}

	toAddress, err := sdk.AccAddressFromBech32(StrideFoundationAddress_F4)
	if err != nil {
		return errorsmod.Wrap(err, "invalid prop recipient address")
	}

	return bk.SendCoins(ctx, fromAddress, toAddress, sdk.NewCoins(sdk.NewCoin(Strd, Prop228SendAmount)))
}
