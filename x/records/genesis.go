package records

import (
	"github.com/Stride-Labs/stride/x/records/keeper"
	"github.com/Stride-Labs/stride/x/records/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the userRedemptionRecord
	for _, elem := range genState.UserRedemptionRecordList {
		k.SetUserRedemptionRecord(ctx, elem)
	}

	// Set userRedemptionRecord count
	k.SetUserRedemptionRecordCount(ctx, genState.UserRedemptionRecordCount)
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.UserRedemptionRecordList = k.GetAllUserRedemptionRecord(ctx)
	genesis.UserRedemptionRecordCount = k.GetUserRedemptionRecordCount(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
