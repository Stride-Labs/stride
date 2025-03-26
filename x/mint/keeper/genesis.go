package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/mint/types"
)

// InitGenesis new mint genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if genState.Minter.EpochProvisions.IsZero() {
		genState.Minter.EpochProvisions = genState.Params.GenesisEpochProvisions
	}
	k.SetMinter(ctx, genState.Minter)
	k.SetParams(ctx, genState.Params)

	if !k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName)) {
		k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	}

	// set up new community module accounts
	k.SetupNewModuleAccount(ctx, types.CommunityGrowthSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	k.SetupNewModuleAccount(ctx, types.CommunitySecurityBudgetSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	k.SetupNewModuleAccount(ctx, types.CommunityUsageSubmoduleName, types.SubmoduleCommunityNamespaceKey)

	k.SetLastReductionEpochNum(ctx, genState.ReductionStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)
	lastReductionEpoch := k.GetLastReductionEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastReductionEpoch)
}
