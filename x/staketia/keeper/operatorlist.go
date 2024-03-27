package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v21/x/staketia/types"
)

// CheckIsOperatorAddress checks if the given address is the operator address
func isOperatorAddress(hostZone types.HostZone, address string) bool {
	return address == hostZone.OperatorAddressOnStride
}

// CheckIsSafeAddress checks if the given address is the safe address
func isSafeAddress(hostZone types.HostZone, address string) bool {
	return address == hostZone.SafeAddressOnStride
}

// CheckIsOperatorAddress checks if the given address is EITHER the safe admin OR operator admin address
func (k Keeper) CheckIsOperatorAddress(ctx sdk.Context, address string) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	if !isOperatorAddress(hostZone, address) {
		return types.ErrInvalidAdmin.Wrapf("invalid operator address %s", address)
	}

	return nil
}

// CheckIsSafeAddress checks if the given address is the safe admin address
func (k Keeper) CheckIsSafeAddress(ctx sdk.Context, address string) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	if !isSafeAddress(hostZone, address) {
		return types.ErrInvalidAdmin.Wrapf("invalid safe address %s", address)
	}

	return nil
}

// CheckIsSafeOrOperatorAddress checks if the given address is EITHER the safe admin OR operator admin address
func (k Keeper) CheckIsSafeOrOperatorAddress(ctx sdk.Context, address string) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	if !isSafeAddress(hostZone, address) && !isOperatorAddress(hostZone, address) {
		return types.ErrInvalidAdmin.Wrapf("invalid admin address %s", address)
	}

	return nil
}
