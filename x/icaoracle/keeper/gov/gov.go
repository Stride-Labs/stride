package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Toggles whether an oracle is currently active (meaning it's a destination for metric pushes)
func ToggleOracle(ctx sdk.Context, k keeper.Keeper, proposal *types.ToggleOracleProposal) error {
	return k.ToggleOracle(ctx, proposal.OracleChainId, proposal.Active)
}

// Removes an oracle from the store
func RemoveOracle(ctx sdk.Context, k keeper.Keeper, proposal *types.RemoveOracleProposal) error {
	_, found := k.GetOracle(ctx, proposal.OracleChainId)
	if !found {
		return types.ErrOracleNotFound
	}

	k.RemoveOracle(ctx, proposal.OracleChainId)
	return nil
}

// Updates the cosmwasm contract address for an oracle
func UpdateOracleContract(ctx sdk.Context, k keeper.Keeper, proposal *types.UpdateOracleContractProposal) error {
	oracle, found := k.GetOracle(ctx, proposal.OracleChainId)
	if !found {
		return types.ErrOracleNotFound
	}

	oracle.ContractAddress = proposal.ContractAddress
	k.SetOracle(ctx, oracle)

	return nil
}
