package v2

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v3/modules/core/02-client/keeper"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	clientKeeper ibcclientkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Upgrade inspired by Evmos v4 light client upgrade
		// https://github.com/evmos/evmos/blob/main/app/upgrades/v4/upgrades.go
		logger := ctx.Logger()

		if err := UpdateLightClient(ctx, clientKeeper); err != nil {
			logger.Error(fmt.Sprintf("Failed to update light client in upgrade handler | %s", err.Error()))
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func UpdateLightClient(ctx sdk.Context, k ibcclientkeeper.Keeper) error {
	proposal := &ibcclienttypes.ClientUpdateProposal{
		Title:              "Update expired GAIA light client on STRIDE",
		Description:        fmt.Sprintf("Update the existing expired GAIA light client (%s).", ExpiredGaiaClient),
		SubjectClientId:    ExpiredGaiaClient,
		SubstituteClientId: ActiveGaiaClient,
	}

	if err := k.ClientUpdateProposal(ctx, proposal); err != nil {
		return sdkerrors.Wrap(err, "Failed to update GAIA light client")
	}

	return nil
}
