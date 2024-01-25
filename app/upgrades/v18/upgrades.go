package v18

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v17/x/records/keeper"
	recordtypes "github.com/Stride-Labs/stride/v17/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v17/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v18
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v18...")

		ctx.Logger().Info("Updating redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Updating unbonding records...")
		err := UpdateUnbondingRecords(ctx, stakeibcKeeper, recordsKeeper, RedemptionRatesBeforeProp, RedemptionRatesAtTimeOfProp)
		if err != nil {
			return vm, err
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

// Modify HostZoneUnbonding and UserRedemptionRecords NativeTokenAmount to reflect new data structs
func UpdateUnbondingRecords(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
	rk recordskeeper.Keeper,
	redemptionRatesBeforeProp map[string]map[uint64]sdk.Dec,
	redemptionRatesDuringProp map[string]sdk.Dec,
) error {
	// loop over host zone unbonding records
	for _, epochUnbondingRecord := range rk.GetAllEpochUnbondingRecord(ctx) {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {

			// we can ignore any record that's not currently unbonding
			if hostZoneUnbonding.Status != recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE {
				continue
			}

			// Grab the redemption rates from before the prop was posted, for a given chain
			// across all the epochs that unbonded
			hostZoneRRBeforeProp, ok := redemptionRatesBeforeProp[hostZoneUnbonding.HostZoneId]
			if !ok {
				ctx.Logger().Error("Host zone from unbonding record not included in redemption rate mapping")
				continue
			}

			// Grab the redemption rate for this specific epoch
			// If it's not found, that means the unbonding for this epoch occurred after the prop was live
			recordRedemptionRate, recordUnbondedBeforeProp := hostZoneRRBeforeProp[epochUnbondingRecord.EpochNumber]

			// If we don't have the redemption rate, estimate it
			if !recordUnbondedBeforeProp {
				hostZone, found := sk.GetHostZone(ctx, hostZoneUnbonding.HostZoneId)
				if !found {
					return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound, "unable to find host zone with chain-id %s", hostZoneUnbonding.HostZoneId)
				}

				redemptionRateDuringProp := redemptionRatesDuringProp[hostZoneUnbonding.HostZoneId]
				redemptionRateDuringUpgrade := hostZone.RedemptionRate
				recordRedemptionRate = redemptionRateDuringProp.Add(redemptionRateDuringUpgrade).Quo(sdk.NewDec(2))
			}

			// now update all userRedemptionRecords
			for _, userRedemptionRecordId := range hostZoneUnbonding.UserRedemptionRecords {
				userRedemptionRecord, found := rk.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
				if !found {
					return errorsmod.Wrapf(recordtypes.ErrHostUnbondingRecordNotFound, "unable to find user redemption record with id %s", userRedemptionRecordId)
				}
				userRedemptionRecord.NativeTokenAmount = sdk.NewDecFromInt(userRedemptionRecord.StTokenAmount).Mul(recordRedemptionRate).TruncateInt()
				rk.SetUserRedemptionRecord(ctx, userRedemptionRecord)
			}

			// finally, update the hostZoneUnbonding record
			return rk.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZoneUnbonding.HostZoneId, *hostZoneUnbonding)
		}
	}
	return nil
}
