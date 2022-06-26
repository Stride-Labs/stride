package records

import (
	"github.com/Stride-Labs/stride/x/records/keeper"
	"github.com/Stride-Labs/stride/x/records/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// Set all the depositRecord
	for _, elem := range genState.DepositRecordList {
		k.SetDepositRecord(ctx, elem)
	}

	// Set depositRecord count
	k.SetDepositRecordCount(ctx, genState.DepositRecordCount)

}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.DepositRecordList = k.GetAllDepositRecord(ctx)
	genesis.DepositRecordCount = k.GetDepositRecordCount(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
