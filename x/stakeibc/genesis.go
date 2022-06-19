package stakeibc

import (
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// TODO: TEST-10
	// Set if defined
	if genState.ICAAccount != nil {
		k.SetICAAccount(ctx, *genState.ICAAccount)
	}
	// Set all the hostZone
	for _, elem := range genState.HostZoneList {
		k.SetHostZone(ctx, elem)
	}

	// Set hostZone count
	k.SetHostZoneCount(ctx, genState.HostZoneCount)
	// Set all the depositRecord
	for _, elem := range genState.DepositRecordList {
		k.SetDepositRecord(ctx, elem)
	}

	// Set depositRecord count
	k.SetDepositRecordCount(ctx, genState.DepositRecordCount)
	// Set all the controllerBalances
for _, elem := range genState.ControllerBalancesList {
	k.SetControllerBalances(ctx, elem)
}
// this line is used by starport scaffolding # genesis/module/init
	// TODO(TEST-22): Set ports
	// k.SetPort(ctx, genState.PortId)
	// // Only try to bind to port if it is not already bound, since we may already own
	// // port capability from capability InitGenesis
	// if !k.IsBound(ctx, genState.PortId) {
	// 	// module binds to the port on InitChain
	// 	// and claims the returned capability
	// 	err := k.BindPort(ctx, genState.PortId)
	// 	if err != nil {
	// 		panic("could not claim port capability: " + err.Error())
	// 	}
	// }
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// ibc v2 scaffolded code
	// genesis.PortId = k.GetPort(ctx)
	// Get all iCAAccount
	iCAAccount, found := k.GetICAAccount(ctx)
	if found {
		genesis.ICAAccount = &iCAAccount
	}
	genesis.HostZoneList = k.GetAllHostZone(ctx)
	genesis.HostZoneCount = k.GetHostZoneCount(ctx)
	genesis.DepositRecordList = k.GetAllDepositRecord(ctx)
	genesis.DepositRecordCount = k.GetDepositRecordCount(ctx)
	genesis.ControllerBalancesList = k.GetAllControllerBalances(ctx)
// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
