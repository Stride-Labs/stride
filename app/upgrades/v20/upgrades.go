package v20

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v4/x/ccv/consumer/keeper"
	ccvtypes "github.com/cosmos/interchain-security/v4/x/ccv/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v20/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v20/x/stakeibc/types"
)

const (
	UpgradeName                      = "v20"
	DydxCommunityPoolTreasuryAddress = "dydx15ztc7xy42tn2ukkc0qjthkucw9ac63pgp70urn"
	DydxChainId                      = "dydx-mainnet-1"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper ccvconsumerkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v20...")

		ctx.Logger().Info("Running module migrations...")
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to run module migrations")
		}

		ctx.Logger().Info("Migrating ICS outstanding downtime...")
		if err := MigrateICSOutstandingDowntime(ctx, consumerKeeper); err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to migrate ICS downtime for v4")
		}

		ctx.Logger().Info("Migrating ICS params...")
		MigrateICSParams(ctx, consumerKeeper)

		ctx.Logger().Info("Adding DYDX Community Pool Treasury Address...")
		if err := SetDydxCommunityPoolTreasuryAddress(ctx, stakeibcKeeper); err != nil {
			return newVm, err
		}

		return newVm, nil
	}
}

// Write the Community Pool Treasury Address to the DYDX host_zone struct
func SetDydxCommunityPoolTreasuryAddress(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	// Get the dydx host_zone
	hostZone, found := k.GetHostZone(ctx, DydxChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound.Wrapf("dydx host zone not found")
	}

	// Set the treasury address
	hostZone.CommunityPoolTreasuryAddress = DydxCommunityPoolTreasuryAddress

	// Save the dydx host_zone
	k.SetHostZone(ctx, hostZone)

	return nil
}

// Migrates Outstanding Downtime for upgrade to ICS v4
// https://github.com/cosmos/interchain-security/blob/release/v4.0.x/UPGRADING.md#v40x
func MigrateICSOutstandingDowntime(ctx sdk.Context, ck ccvconsumerkeeper.Keeper) error {
	downtimes := ck.GetAllOutstandingDowntimes(ctx)
	for _, od := range downtimes {
		consAddr, err := sdk.ConsAddressFromBech32(od.ValidatorConsensusAddress)
		if err != nil {
			return err
		}
		ck.DeleteOutstandingDowntime(ctx, consAddr)
	}
	return nil
}

// Migrates ICS Params to add the new RetryDelayParam
// Because RetryDelayPeriod isn't in the store yet, we can't call
// ck.GetConsumerParams(ctx) and modify the RetryDelayPeriod,
// as calling ck.GetConsumerParams(ctx) will panic if the param isn't in the store.
// Instead, we get the DefaultParams and modify each value to
// their value directly before the v20 upgrade
func MigrateICSParams(ctx sdk.Context, ck ccvconsumerkeeper.Keeper) {
	// Get default consumer params and set the new RetryDelayPeriod
	params := ccvtypes.DefaultParams()
	params.RetryDelayPeriod = ccvtypes.DefaultRetryDelayPeriod
	// Modify to the correct params values
	params.Enabled = true
	params.BlocksPerDistributionTransmission = 1000
	params.DistributionTransmissionChannel = "channel-0"
	params.ProviderFeePoolAddrStr = "cosmos1ap0mh6xzfn8943urr84q6ae7zfnar48am2erhd"
	params.CcvTimeoutPeriod = time.Second * 2419200
	params.TransferTimeoutPeriod = time.Second * 3600
	params.ConsumerRedistributionFraction = "0.85"
	params.HistoricalEntries = 10000
	params.UnbondingPeriod = time.Second * 1209600
	params.SoftOptOutThreshold = "0.05"
	params.RewardDenoms = []string{
		"stucmdx",
		"stuatom",
		"staevmos",
		"stinj",
		"stujuno",
		"stuosmo",
		"stuluna",
		"stusomm",
		"stustars",
		"stuumee",
		"ustrd",
		"stadydx",
		"stutia",
		"stadym",
	}
	params.ProviderRewardDenoms = []string{}
	// Set the new params
	ck.SetParams(ctx, params)
}
