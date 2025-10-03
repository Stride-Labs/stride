package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Initialize module account in account keeper if not already initialized
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	// Set Total STRD Burned
	k.SetProtocolStrdBurned(ctx, genState.ProtocolUstrdBurned)
	k.SetTotalUserStrdBurned(ctx, genState.TotalUserUstrdBurned)

	// Set STRD burned by address
	allAddresses := map[string]bool{}
	for _, accountBurned := range genState.BurnedByAccount {
		if allAddresses[accountBurned.Address] {
			panic(fmt.Sprintf("Duplicate address found: %s", accountBurned.Address))
		}

		address, err := sdk.AccAddressFromBech32(accountBurned.Address)
		if err != nil {
			panic(fmt.Sprintf("Invalid burner address: %s", accountBurned.Address))
		}
		k.SetStrdBurnedByAddress(ctx, address, accountBurned.Amount)

		allAddresses[accountBurned.Address] = true
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.ProtocolUstrdBurned = k.GetProtocolStrdBurned(ctx)
	genesis.TotalUserUstrdBurned = k.GetTotalUserStrdBurned(ctx)
	genesis.TotalUstrdBurned = k.GetTotalStrdBurned(ctx)
	genesis.BurnedByAccount = k.GetAllStrdBurnedAcrossAddresses(ctx)
	return genesis
}
