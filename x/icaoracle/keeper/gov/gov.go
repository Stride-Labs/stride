package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Toggles whether an oracle is currently active (meaning it's a destination for metric pushes)
func ToggleOracle(ctx sdk.Context, k keeper.Keeper, p *types.ToggleOracleProposal) error {
	// TODO
	return nil
}

// Removes an oracle from the store
func RemoveOracle(ctx sdk.Context, k keeper.Keeper, p *types.RemoveOracleProposal) error {
	// TODO
	return nil
}

// Updates the cosmwasm contract address for an oracle
func UpdateOracleContract(ctx sdk.Context, k keeper.Keeper, p *types.UpdateOracleContractProposal) error {
	// TODO
	return nil
}
