package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/mint/keeper"
	"github.com/Stride-Labs/stride/v4/x/mint/types"
)

// InitGenesis new mint genesis.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, data *types.GenesisState) {
	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)

	if !ak.HasAccount(ctx, ak.GetModuleAddress(types.ModuleName)) {
		ak.GetModuleAccount(ctx, types.ModuleName)
	}

	// set up new community module accounts
	keeper.SetupNewModuleAccount(ctx, types.CommunityGrowthSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	keeper.SetupNewModuleAccount(ctx, types.CommunitySecurityBudgetSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	keeper.SetupNewModuleAccount(ctx, types.CommunityUsageSubmoduleName, types.SubmoduleCommunityNamespaceKey)

	keeper.SetLastReductionEpochNum(ctx, data.ReductionStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	lastReductionEpoch := keeper.GetLastReductionEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastReductionEpoch)
}
