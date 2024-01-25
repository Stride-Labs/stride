package v18

import (
	"fmt"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v17/x/records/keeper"
	recordtypes "github.com/Stride-Labs/stride/v17/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v17/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

func NewIntFromString(val string) sdkmath.Int {
	out_int, ok := sdkmath.NewIntFromString(val)
	if !ok {
		panic(fmt.Sprintf("invalid string %s", val))
	}
	return out_int
}

var (
	UpgradeName = "v18"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")

	// Get Initial Redemption Rates for Unbonding Records Migration
	InitialRedemptionRates = map[string]sdk.Dec{
		"comdex-1":     sdk.MustNewDecFromStr("1.199107579864670291"),
		"cosmoshub-4":  sdk.MustNewDecFromStr("1.295759028526771630"),
		"evmos_9001-2": sdk.MustNewDecFromStr("1.489961582552824897"),
		"injective-1":  sdk.MustNewDecFromStr("1.211514556746897031"),
		"juno-1":       sdk.MustNewDecFromStr("1.414898221278138305"),
		"osmosis-1":    sdk.MustNewDecFromStr("1.199042531372783387"),
		"phoenix-1":    sdk.MustNewDecFromStr("1.175092793361528752"),
		"sommelier-3":  sdk.MustNewDecFromStr("1.025191976926387110"),
		"stargaze-1":   sdk.MustNewDecFromStr("1.426572776580499529"),
		"umee-1":       sdk.MustNewDecFromStr("1.125555216159213353"),
	}

	// Get Amount Unbonded for each HostZone for Unbonding Records Migration
	TotalAmountUnbonded = map[string]map[string]sdkmath.Int{
		"comdex-1": {
			"1705863627": NewIntFromString("1640812802"),
			"1706209223": NewIntFromString("164530181"),
			"1706554815": NewIntFromString("966328519777"),
			"1706900422": NewIntFromString("334284598402"),
			"1707246012": NewIntFromString("78876482507"),
		},
		"cosmoshub-4": {
			"1706209219": NewIntFromString("495520981916"),
			"1707246028": NewIntFromString("153692215135"),
			"1706900433": NewIntFromString("101590127858"),
			"1706554814": NewIntFromString("221083458468"),
			"1705863634": NewIntFromString("142964081509"),
		},
		"evmos_9001-2": {
			"1705690816": NewIntFromString("420455301467019532445082"),
			"1705950016": NewIntFromString("1056499615224518202135947"),
			"1706209220": NewIntFromString("149337601388876187205598"),
			"1706468419": NewIntFromString("97606478355885814008907"),
			"1706727614": NewIntFromString("1214076305566194153808"),
		},
		"injective-1": {
			"1707246015": NewIntFromString("51607643878504534441739"),
		},
		"juno-1": {
			"1706554816": NewIntFromString("157918257"),
			"1706986817": NewIntFromString("176556452"),
			"1707850813": NewIntFromString("447906130"),
			"1706122818": NewIntFromString("4784225328"),
			"1707418816": NewIntFromString("11456373"),
			"1705690818": NewIntFromString("32566654"),
		},
		"osmosis-1": {
			"1705690839": NewIntFromString("69534834841"),
			"1705950052": NewIntFromString("34185116581"),
			"1706209219": NewIntFromString("31400303772"),
			"1706468412": NewIntFromString("58456099905"),
			"1706727627": NewIntFromString("29541728198"),
		},
		"phoenix-1": {
			"1706554818": NewIntFromString("14560879404"),
			"1706900423": NewIntFromString("1576394459"),
			"1707246013": NewIntFromString("1060350649"),
		},
		"sommelier-3": {
			"1705690818": NewIntFromString("14786162726"),
			"1706122819": NewIntFromString("3634"),
			"1706554818": NewIntFromString("19330705153"),
			"1706986821": NewIntFromString("21799461066"),
			"1707418817": NewIntFromString("140098739"),
			"1707850817": NewIntFromString("7879600093"),
		},
		"stargaze-1": {
			"1705690816": NewIntFromString("327679537357"),
			"1705950009": NewIntFromString("437989561516"),
			"1706209215": NewIntFromString("437233672994"),
			"1706468412": NewIntFromString("508229473185"),
			"1706727610": NewIntFromString("1033532361858"),
		},
		"umee-1": {
			"1705690815": NewIntFromString("510084698639"),
		},
	}
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
		err := UpdateUnbondingRecords(ctx, stakeibcKeeper, recordsKeeper)
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
func UpdateUnbondingRecords(ctx sdk.Context, sk stakeibckeeper.Keeper, rk recordskeeper.Keeper) error {
	// loop over host zone unbonding records
	for _, epochUnbondingRecord := range rk.GetAllEpochUnbondingRecord(ctx) {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {

			// we can ignore any record that's not currently unbonding
			if hostZoneUnbonding.Status != recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE {
				continue
			}
			amountUnbonded := sdkmath.ZeroInt()

			// check if the unbond amount is stored in cache above
			allAmountsUnbondedForHost, found := TotalAmountUnbonded[hostZoneUnbonding.HostZoneId]
			if !found {
				// check if this unbond time is in the cache above
				unbondTime := strconv.FormatUint(hostZoneUnbonding.UnbondingTime, 10)
				for unbondTimePrefixInCache, amountUnbondedInCache := range allAmountsUnbondedForHost {
					if strings.HasPrefix(unbondTime, unbondTimePrefixInCache) {
						amountUnbonded = amountUnbondedInCache
						break
					}
				}
			}

			// if we can't find unbonding amount, then we use an estimated redemption rate
			if amountUnbonded.IsZero() {
				initialRedemptionRate := InitialRedemptionRates[hostZoneUnbonding.HostZoneId]
				hostZone, found := sk.GetHostZone(ctx, hostZoneUnbonding.HostZoneId)
				if !found {
					return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound, "unable to find host zone with chain-id %s", hostZoneUnbonding.HostZoneId)
				}
				currentRedemptionRate := hostZone.RedemptionRate
				blendedRedemptionRate := initialRedemptionRate.Add(currentRedemptionRate).Quo(sdk.NewDecFromInt(sdk.NewInt(2)))
				amountUnbonded = sdk.NewDecFromInt(hostZoneUnbonding.StTokenAmount).Mul(blendedRedemptionRate).TruncateInt()
			}
			hostZoneUnbonding.NativeTokenAmount = amountUnbonded

			// get implied redemption rate, based on the amount unbonded and stTokens burned
			impliedRedemptionRate := sdk.NewDecFromInt(amountUnbonded).Quo(sdk.NewDecFromInt(hostZoneUnbonding.StTokenAmount))

			// now update all userRedemptionRecords
			for _, userRedemptionRecordId := range hostZoneUnbonding.UserRedemptionRecords {
				userRedemptionRecord, found := rk.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
				if !found {
					return errorsmod.Wrapf(recordtypes.ErrHostUnbondingRecordNotFound, "unable to find user redemption record with id %s", userRedemptionRecordId)
				}
				userRedemptionRecord.NativeTokenAmount = sdk.NewDecFromInt(userRedemptionRecord.StTokenAmount).Mul(impliedRedemptionRate).TruncateInt()
				rk.SetUserRedemptionRecord(ctx, userRedemptionRecord)
			}

			// finally, update the hostZoneUnbonding record
			return rk.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZoneUnbonding.HostZoneId, *hostZoneUnbonding)
		}
	}
	return nil
}
