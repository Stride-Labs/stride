package keeper

import (

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

type StakeIbcAnteDecorator struct {
	StakeIbcKeeper Keeper
}

func NewStakeIbcAnteDecorator(stakeIbcAnteDecorator Keeper) StakeIbcAnteDecorator {
	return StakeIbcAnteDecorator{
		StakeIbcKeeper: stakeIbcAnteDecorator,
	}
}

// This posthandler will save the stTokenSupply & moduleAccount balance before tx to store
func (stakeIbcAnteDec StakeIbcAnteDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	moduleAddr := stakeIbcAnteDec.StakeIbcKeeper.AccountKeeper.GetModuleAccount(ctx, stakeibctypes.ModuleName).GetAddress()
	hostzones := stakeIbcAnteDec.StakeIbcKeeper.GetAllHostZone(ctx)
	for _, hz := range hostzones {
		stSupplyBeforeTx := stakeIbcAnteDec.StakeIbcKeeper.bankKeeper.GetSupply(ctx, hz.HostDenom)
		ibcDenomModuleAccountBalance := stakeIbcAnteDec.StakeIbcKeeper.bankKeeper.GetBalance(ctx, moduleAddr, hz.IbcDenom)
		stakeIbcAnteDec.StakeIbcKeeper.SetStSupply(ctx, hz, stSupplyBeforeTx)
		stakeIbcAnteDec.StakeIbcKeeper.SetModuleAccountIbcBalance(ctx, hz, ibcDenomModuleAccountBalance)
	}
	return next(ctx, tx, simulate)
}