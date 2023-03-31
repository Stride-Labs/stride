package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) SetLSMTokenDeposit(ctx sdk.Context, deposit types.LSMTokenDeposit) {
	// TODO [LSM]
}

func (k Keeper) RemoveLSMTokenDeposit(ctx sdk.Context, denom string) {
	// TODO [LSM]
}

func (k Keeper) GetLSMTokenDeposit(ctx sdk.Context, denom string) (deposit types.LSMTokenDeposit, found bool) {
	// TODO [LSM]
	return
}

func (k Keeper) GetAllLSMTokenDeposit(ctx sdk.Context) []types.LSMTokenDeposit {
	// TODO [LSM]
	return []types.LSMTokenDeposit{}
}

func (k Keeper) AddLSMTokenDeposit(ctx sdk.Context, deposit types.LSMDepositStatus) {
	// TODO [LSM]
	// See if a deposit already exists for this denom
	// If so, increment the amount
	// otherwise, create a new deposit
}

func (k Keeper) UpdateLSMTokenDepositStatus(ctx sdk.Context, deposit types.LSMDepositStatus, status types.LSMDepositStatus) {
	// TODO [LSM]
}

func (k Keeper) GetLSMDepositsForHostZone(ctx sdk.Context, deposit types.LSMDepositStatus, chainId string) []types.LSMDepositStatus {
	// TODO [LSM]
	return []types.LSMDepositStatus{}
}

func (k Keeper) GetLSMDepositsForHostZoneWithStatus(
	ctx sdk.Context,
	deposit types.LSMDepositStatus,
	chainId string,
	status types.LSMDepositStatus,
) []types.LSMDepositStatus {
	// TODO [LSM]
	return []types.LSMDepositStatus{}
}
