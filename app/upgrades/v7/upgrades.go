package v7

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahostkeeper "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	epochskeeper "github.com/Stride-Labs/stride/v6/x/epochs/keeper"
	epochstypes "github.com/Stride-Labs/stride/v6/x/epochs/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v6/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v7"
)

// Helper function to log the migrated modules consensus versions
func logModuleMigration(ctx sdk.Context, versionMap module.VersionMap, moduleName string) {
	currentVersion := versionMap[moduleName]
	ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, currentVersion-1, currentVersion))
}

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	epochskeeper epochskeeper.Keeper,
	stakeibckeeper stakeibckeeper.Keeper,
	icahostkeeper icahostkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// TODO:
		//  	add min/max redemption rate
		// 		incentive diversification
		// 		inflation
		//  	BaseAccount issue

		// Add an hourly epoch which will be used by the rate limit store
		AddHourEpoch(ctx, epochskeeper)

		// Add stride messages to the ICA host allow messages
		AddICAHostAllowMessages(ctx, icahostkeeper)

		// Change the juno unbonding frequency to 5
		if err := ModifyJunoUnbondingFrequency(ctx, stakeibckeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to modify juno unbonding frequency")
		}

		// Add min/max redemption rate threshold for each host zone
		if err := AddMinMaxRedemptionRate(ctx, stakeibckeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set min/max redemption rate on host zones")
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Add a new hourly epoch that will be used by the rate limit module
func AddHourEpoch(ctx sdk.Context, k epochskeeper.Keeper) {
	ctx.Logger().Info("Adding hour epoch")

	hourEpoch := epochstypes.EpochInfo{
		Identifier:              epochstypes.HOUR_EPOCH,
		StartTime:               time.Time{},
		Duration:                time.Hour,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}

	k.SetEpochInfo(ctx, hourEpoch)
}

// Add stride messages (LiquidStake, RedeemStake, Claim) to the ICAHost allow messages
// to allow other protocols to liquid stake via ICA
func AddICAHostAllowMessages(ctx sdk.Context, k icahostkeeper.Keeper) {
	params := icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&transfertypes.MsgTransfer{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
		},
	}
	k.SetParams(ctx, params)
}

// Update the unbonding frequency of juno to 5 to align with the 28 day unbonding period
func ModifyJunoUnbondingFrequency(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	ctx.Logger().Info("Updating juno unbonding frequency")

	junoChainId := "juno-1"
	unbondingFrequency := uint64(5)

	junoHostZone, found := k.GetHostZone(ctx, junoChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound
	}
	junoHostZone.UnbondingFrequency = unbondingFrequency
	k.SetHostZone(ctx, junoHostZone)

	return nil
}

// Add the min/max redemption rate to each host zone as safety bounds
// Use the default min/max for each
func AddMinMaxRedemptionRate(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	// TODO
	return nil
}
