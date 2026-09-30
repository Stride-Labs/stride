package v35

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GovModuleAddress is the gov module account, the only address that may upload code or
// administer the deploy key's contracts after the upgrade.
func GovModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(govtypes.ModuleName)
}

// SetWasmUploadAccessToGov restricts code upload to the gov module. This is the one handler
// step that fails the upgrade on error: it can only fail on invalid params, and leaving
// upload open to the two deploy keys with nothing but a log line is the worse outcome (spec §5).
func SetWasmUploadAccessToGov(ctx sdk.Context, k wasmkeeper.Keeper) error {
	params := k.GetParams(ctx)
	params.CodeUploadAccess = wasmtypes.AccessTypeAnyOfAddresses.With(GovModuleAddress())
	if err := k.SetParams(ctx, params); err != nil {
		return errorsmod.Wrap(err, "v35: unable to set wasm code upload access to the gov module")
	}
	ctx.Logger().Info("v35: wasm code upload access restricted to the gov module")
	return nil
}

// MoveDeployKeyContractAdminsToGov sets the admin of every contract currently administered by
// WasmDeployKey to the gov module. The gov permission keeper's authorization policy allows
// the change without the current admin's signature. A failed update is logged and skipped.
func MoveDeployKeyContractAdminsToGov(ctx sdk.Context, k wasmkeeper.Keeper) {
	govKeeper := wasmkeeper.NewGovPermissionKeeper(k)
	gov := GovModuleAddress()

	// Collect first: IterateContractInfo must not see the store change under it
	var toMove []sdk.AccAddress
	k.IterateContractInfo(ctx, func(contractAddr sdk.AccAddress, info wasmtypes.ContractInfo) bool {
		if info.Admin == WasmDeployKey {
			toMove = append(toMove, contractAddr)
		}
		return false
	})

	numMoved := 0
	for _, contractAddr := range toMove {
		if err := govKeeper.UpdateContractAdmin(ctx, contractAddr, gov, gov); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: unable to move admin of %s to gov, skipping: %s", contractAddr, err))
			continue
		}
		numMoved++
		ctx.Logger().Info(fmt.Sprintf("v35: admin of %s moved from the deploy key to gov", contractAddr))
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d contract admin(s) moved to gov", numMoved))
}
