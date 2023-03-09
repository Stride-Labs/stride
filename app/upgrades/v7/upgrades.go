package v7

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
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
	mintkeeper "github.com/Stride-Labs/stride/v6/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v6/x/mint/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v6/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName             = "v7"
	IncentiveProgramAddress = "stride1tlxk4as9sgpqkh42cfaxqja0mdj6qculqshy0gg3glazmrnx3y8s8gsvqk"
	StrideFoundationAddress = "stride1yz3mp7c2m739nftfrv5r3h6j64aqp95f3degpf"
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
	epochsKeeper epochskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	icahostKeeper icahostkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v7...")

		// TODO:
		//  	add min/max redemption rate
		//      autopilot store key
		//  	Create RewardCollector module account

		// Add an hourly epoch which will be used by the rate limit store
		AddHourEpoch(ctx, epochsKeeper)

		// Increase stride inflation to 1078 STRD
		IncreaseStrideInflation(ctx, mintKeeper)

		// Add stride messages to the ICA host allow messages
		AddICAHostAllowMessages(ctx, icahostKeeper)

		// Add min/max redemption rate threshold and a `Halted`` boolean to each host zone
		AddRedemptionRateSafetyChecks(ctx, stakeibcKeeper)

		// Change the juno unbonding frequency to 5
		if err := ModifyJunoUnbondingFrequency(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to modify juno unbonding frequency")
		}

		// Incentive diversification
		if err := IncentiveDiversification(ctx, bankKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to send ustrd tokens for prop #153 (incentive diversification)")
		}

		ctx.Logger().Info("Running module mogrations...")
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

// Increase stride inflation from 285 STRD to 1078 STRD
func IncreaseStrideInflation(ctx sdk.Context, k mintkeeper.Keeper) {
	ctx.Logger().Info("Increasing STRD inflation")

	epochProvisions := sdk.NewDec(1_078_767_123)
	minter := minttypes.NewMinter(epochProvisions)
	k.SetMinter(ctx, minter)
}

// Add stride messages (LiquidStake, RedeemStake, Claim) to the ICAHost allow messages
// to allow other protocols to liquid stake via ICA
func AddICAHostAllowMessages(ctx sdk.Context, k icahostkeeper.Keeper) {
	ctx.Logger().Info("Adding ICA Host allow messages")

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

// Add the min/max redemption rate to each host zone as safety bounds, using the default for each
// Also set the Halted boolean to false
func AddRedemptionRateSafetyChecks(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Setting min/max redemption rate safety bounds on each host zone")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		// hostZone.Halted = false
		// hostZone.MinRedemptionRate =
		// hostZone.MaxRedemptionRate =

		k.SetHostZone(ctx, hostZone)
	}
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

// Incentive diversification (Prop #153) - Send 3M STRD from Incentive Program to Stride Foundation
func IncentiveDiversification(ctx sdk.Context, k bankkeeper.Keeper) error {
	ctx.Logger().Info("Sending funds for Prop #153")

	incentiveProgramAddress, err := sdk.AccAddressFromBech32(IncentiveProgramAddress)
	if err != nil {
		return err
	}
	strideFoundationAddress, err := sdk.AccAddressFromBech32(StrideFoundationAddress)
	if err != nil {
		return err
	}
	amount := sdk.NewCoin("ustrd", sdk.NewInt(3_000_000_000_000))
	k.SendCoins(ctx, incentiveProgramAddress, strideFoundationAddress, sdk.NewCoins(amount))

	return nil
}
