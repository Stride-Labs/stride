package v28

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
)

var (
	UpgradeName = "v28"

	// Redemption rate bounds updated to give slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.50")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("1.00")

	DeliveryAccount = "stride198f9skhtnpzpsxtmlkg3ry8yglwqn9pm9ugl28"

	VestingEndTime    = int64(1785988800)        // Thu Aug 06 2026 04:00:00 GMT+0000
	LockedTokenAmount = int64(4_000_000_000_000) // 4 million STRD
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v28...")

		ak := accountKeeper

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		// Deliver locked tokens
		if err := DeliverLockedTokens(ctx, ak); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to deliver tokens to account %s", DeliveryAccount)
		}

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

func DeliverLockedTokens(ctx sdk.Context, ak authkeeper.AccountKeeper) error {
	// Get account
	account := ak.GetAccount(ctx, sdk.MustAccAddressFromBech32(DeliveryAccount))
	if account == nil {
		return nil
	}
	baseAccount, isBaseAccount := account.(*authtypes.BaseAccount)
	if !isBaseAccount {
		// Maybe return an error here?
		ctx.Logger().Error("Account at DeliveryAccount is not a BaseAccount, cannot create DelayedVestingAccount")
		return nil
	}

	originalVesting := sdk.NewCoins(sdk.NewCoin("ustrd", sdk.NewInt(LockedTokenAmount)))
	dva := vesting.NewDelayedVestingAccount(baseAccount, originalVesting, VestingEndTime)

	// No tokens on the account are staked, so we don't need to set delegated_free / delegated_vesting

	ak.SetAccount(ctx, dva)
	return nil
}
