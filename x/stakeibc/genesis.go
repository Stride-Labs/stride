package stakeibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// TODO: TEST-10
	// Set if defined
	if genState.IcaAccount != nil {
		k.SetICAAccount(ctx, *genState.IcaAccount)
	}
	// Set all the hostZone
	for _, elem := range genState.HostZoneList {
		k.SetHostZone(ctx, elem)
	}

	// Set hostZone count
	k.SetHostZoneCount(ctx, genState.HostZoneCount)
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
		genesis.IcaAccount = &iCAAccount
	}
	genesis.EpochTrackerList = k.GetAllEpochTracker(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
