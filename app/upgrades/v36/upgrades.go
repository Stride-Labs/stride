package v36

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/Stride-Labs/stride/v35/utils"
)

// CreateUpgradeHandler returns the v36 upgrade handler, which moves the three governance-earmarked treasury accounts
// (Community Pool Growth, Incentive Program, Community Pool Security) to the team multisig. Like the v35 community pool
// transfer, a coin that fails to send is logged and skipped rather than failing the upgrade.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (treasury accounts to the multisig)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		TransferTreasuryToAuthority(ctx, bankKeeper)

		return vm, nil
	}
}

// TransferTreasuryToAuthority sends the spendable balance of each treasury account, in every denom, to the multisig.
// Each coin is sent on its own in a cache context, so one that fails stays behind while the rest still transfer.
func TransferTreasuryToAuthority(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	recipient := sdk.MustAccAddressFromBech32(UpgradeAuthority)
	for _, treasuryAddress := range TreasuryAddresses {
		sender := sdk.MustAccAddressFromBech32(treasuryAddress)

		// Coins are sorted by denom, so every node sends in the same sequence
		transferred := sdk.NewCoins()
		for _, coin := range bankKeeper.SpendableCoins(ctx, sender) {
			err := utils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
				return bankKeeper.SendCoins(cacheCtx, sender, recipient, sdk.NewCoins(coin))
			})
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("v36: skipping transfer of %s from %s: %s", coin, treasuryAddress, err))
				continue
			}
			transferred = transferred.Add(coin)
		}

		ctx.Logger().Info(fmt.Sprintf("v36: transferred %s from %s to %s", transferred, treasuryAddress, UpgradeAuthority))
	}
}
