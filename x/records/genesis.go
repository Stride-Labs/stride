package records

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/records/keeper"
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the userRedemptionRecord
	for _, elem := range genState.UserRedemptionRecordList {
		k.SetUserRedemptionRecord(ctx, elem)
	}

	// Set all the epochUnbondingRecord
	for _, elem := range genState.EpochUnbondingRecordList {
		k.SetEpochUnbondingRecord(ctx, elem)
	}

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

	genesis.UserRedemptionRecordList = k.GetAllUserRedemptionRecord(ctx)
	genesis.EpochUnbondingRecordList = k.GetAllEpochUnbondingRecord(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
