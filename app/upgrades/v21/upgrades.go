package v21

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordstypes "github.com/Stride-Labs/stride/v21/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v21/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

const (
	UpgradeName = "v21"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v21
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v21...")

		// Remove dydx host zone
		if err := RemoveDydxHostZone(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to remove dydx host zone")
		}

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func RemoveDydxHostZone(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	chainId := "dydx-testnet-4"

	// Remove module accounts
	depositAddress := stakeibctypes.NewHostZoneDepositAddress(chainId)
	communityPoolStakeAddress := stakeibctypes.NewHostZoneModuleAddress(chainId, stakeibckeeper.CommunityPoolStakeHoldingAddressKey)
	communityPoolRedeemAddress := stakeibctypes.NewHostZoneModuleAddress(chainId, stakeibckeeper.CommunityPoolRedeemHoldingAddressKey)

	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, depositAddress))
	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, communityPoolStakeAddress))
	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, communityPoolRedeemAddress))

	// Remove all deposit records for the host zone
	for _, depositRecord := range k.RecordsKeeper.GetAllDepositRecord(ctx) {
		if depositRecord.HostZoneId == chainId {
			k.RecordsKeeper.RemoveDepositRecord(ctx, depositRecord.Id)
		}
	}

	// Remove all epoch unbonding records for the host zone
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		updatedHostZoneUnbondings := []*recordstypes.HostZoneUnbonding{}
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			if hostZoneUnbonding.HostZoneId != chainId {
				updatedHostZoneUnbondings = append(updatedHostZoneUnbondings, hostZoneUnbonding)
			}
		}
		epochUnbondingRecord.HostZoneUnbondings = updatedHostZoneUnbondings
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}

	return nil
}
