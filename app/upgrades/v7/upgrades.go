package v7

import (
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochskeeper "github.com/Stride-Labs/stride/v9/x/epochs/keeper"
	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	mintkeeper "github.com/Stride-Labs/stride/v9/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v9/x/mint/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	epochsKeeper epochskeeper.Keeper,
	icahostKeeper icahostkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v7...")

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
		if err := ExecuteProp153(ctx, bankKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to send ustrd tokens for prop #153 (incentive diversification)")
		}

		// Create reward collector module account
		if err := CreateRewardCollectorModuleAccount(ctx, accountKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable create reward collector module account")
		}

		ctx.Logger().Info("Running module mogrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Add a new hourly epoch that will be used by the rate limit module
func AddHourEpoch(ctx sdk.Context, k epochskeeper.Keeper) {
	ctx.Logger().Info("Adding hour epoch")

	startTime := ctx.BlockTime().Truncate(time.Hour)
	hourEpoch := epochstypes.EpochInfo{
		Identifier:              epochstypes.HOUR_EPOCH,
		StartTime:               startTime,
		Duration:                time.Hour,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: ctx.BlockHeight(),
		CurrentEpochStartTime:   startTime,
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

	// Set new stakeibc params
	// In this case, we're using `DefaultParams` because all of our current params are the defaults
	// You can verify this by hand by running `strided q stakeibc params`, and comparing the output to the values defined in params.go.
	// This is a safer way of adding a new parameter that avoids having to unmarshal using the old types
	// In future upgrades, if we are simply modifying parameter values (instead of adding a new parameter),
	// we should read in params using GetParams, modify them, and then set them using SetParams
	params := stakeibctypes.DefaultParams()
	k.SetParams(ctx, params)

	// Get default min/max redemption rate
	defaultMinRedemptionRate := sdk.NewDecWithPrec(int64(params.DefaultMinRedemptionRateThreshold), 2)
	defaultMaxRedemptionRate := sdk.NewDecWithPrec(int64(params.DefaultMaxRedemptionRateThreshold), 2)

	for _, hostZone := range k.GetAllHostZone(ctx) {

		hostZone.MinRedemptionRate = defaultMinRedemptionRate
		hostZone.MaxRedemptionRate = defaultMaxRedemptionRate
		hostZone.Halted = false

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
// so that it can be swapped for USDC and deployed as incentives
// The 3M STRD will be swapped to USDC.
func ExecuteProp153(ctx sdk.Context, k bankkeeper.Keeper) error {
	ctx.Logger().Info("Sending 3M STRD from the incentive pool to Stride Foundation for Prop #153")

	incentiveProgramAddress, err := sdk.AccAddressFromBech32(IncentiveProgramAddress)
	if err != nil {
		return err
	}
	strideFoundationAddress, err := sdk.AccAddressFromBech32(StrideFoundationAddress_F4)
	if err != nil {
		return err
	}
	amount := sdk.NewCoin(Ustrd, sdk.NewInt(STRDProp153SendAmount))
	if err := k.SendCoins(ctx, incentiveProgramAddress, strideFoundationAddress, sdk.NewCoins(amount)); err != nil {
		return err
	}

	return nil
}

// Create reward collector module account for Prop #8
func CreateRewardCollectorModuleAccount(ctx sdk.Context, k authkeeper.AccountKeeper) error {
	ctx.Logger().Info("Creating reward collector module account")

	rewardCollectorAddress := address.Module(stakeibctypes.RewardCollectorName, []byte(stakeibctypes.RewardCollectorName))
	return utils.CreateModuleAccount(ctx, k, rewardCollectorAddress)
}
