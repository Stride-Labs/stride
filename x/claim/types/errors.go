package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/claim module sentinel errors
var (
	ErrIncorrectModuleAccountBalance = sdkerrors.Register(ModuleName, 1100,
		"claim module account balance != sum of all claim record InitialClaimableAmounts")
	ErrTotalWeightNotSet = sdkerrors.Register(ModuleName, 1101,
		"total weight not set!")
	ErrTotalWeightParse = sdkerrors.Register(ModuleName, 1102,
		"total weight parse error!")
	ErrFailedToGetTotalWeight = sdkerrors.Register(ModuleName, 1104,
		"failed to get total weight!")
)
