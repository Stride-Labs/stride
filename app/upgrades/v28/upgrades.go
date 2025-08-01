package v28

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqkeeper "github.com/Stride-Labs/stride/v27/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
)

var (
	UpgradeName = "v28"

	EvmosChainId          = "evmos_9001-2"
	QueryId               = "2c39af4c3d2ecb96d8bbf7f3386468c5909e51fe3364b8d1f9d6fce173dd1f7a"
	QueryValidatorAddress = "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vx6z88"
	EvmosDelegationIca    = "evmos1d67tx0zekagfhw6chhgza6qmhyad5qprru0nwazpx5s85ld0wh2sdhhznd"

	// Redemption rate bounds updated to give slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.50")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("1.00")

	DeliveryAccount = "stride198f9skhtnpzpsxtmlkg3ry8yglwqn9pm9ugl28"
	FromAccount     = "stride1alnn79kh0xka0r5h4h82uuaqfhpdmph6rvpf5f"

	VestingEndTime    = int64(1785988800)        // Thu Aug 06 2026 04:00:00 GMT+0000
	LockedTokenAmount = int64(4_000_000_000_000) // 4 million STRD
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v28...")

		ak := accountKeeper
		bk := bankKeeper

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		// Deliver locked tokens
		if err := DeliverLockedTokens(ctx, ak, bk); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to deliver tokens to account %s", DeliveryAccount)
		}

		ctx.Logger().Info("Processing stale ICQ...")
		ClearStuckEvmosQuery(ctx, stakeibcKeeper, icqKeeper)

		return versionMap, nil
	}
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate outer bounds...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMaxAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

func DeliverLockedTokens(ctx sdk.Context, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper) error {
	// Get account
	from := sdk.MustAccAddressFromBech32(FromAccount)
	to := sdk.MustAccAddressFromBech32(DeliveryAccount)
	amt := sdk.NewCoins(sdk.NewInt64Coin("ustrd", LockedTokenAmount))

	// Destination must exist and be a plain BaseAccount.
	base, isBaseAccount := ak.GetAccount(ctx, to).(*authtypes.BaseAccount)
	if !isBaseAccount {
		// Maybe return an error here?
		ctx.Logger().Error("Account at DeliveryAccount is not a BaseAccount, cannot create DelayedVestingAccount")
		return nil
	}

	// Send the 4 000 000 STRD
	if err := bk.SendCoins(ctx, from, to, amt); err != nil {
		return err
	}

	// Convert the credited account to DelayedVesting.
	dva := vestingtypes.NewDelayedVestingAccount(base, amt, VestingEndTime)
	ak.SetAccount(ctx, dva)

	return nil
}

// Cleans up the stale ICQ
func ClearStuckEvmosQuery(ctx sdk.Context, k stakeibckeeper.Keeper, icqKeeper icqkeeper.Keeper) {
	ctx.Logger().Info("Deleting stale ICQ...")
	icqKeeper.DeleteQuery(ctx, QueryId)

	ctx.Logger().Info("Setting validator slash_query_in_progress to false...")
	hostZone, found := k.GetHostZone(ctx, EvmosChainId)
	if !found {
		ctx.Logger().Error("host zone not found")
		return
	}

	// find the right validator and set slash_query_in_progress to false
	for i, validator := range hostZone.Validators {
		if validator.Address == QueryValidatorAddress {
			validator.SlashQueryInProgress = false
			hostZone.Validators[i] = validator
			k.SetHostZone(ctx, hostZone)

			ctx.Logger().Info("Set validator slash_query_in_progress to false")
			return
		}
	}
}
