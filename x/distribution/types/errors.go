package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/distribution module sentinel errors
var (
	ErrEmptyDelegatorAddr      = sdkerrors.Register(ModuleName, 1518, "delegator address is empty")
	ErrEmptyWithdrawAddr       = sdkerrors.Register(ModuleName, 1519, "withdraw address is empty")
	ErrEmptyValidatorAddr      = sdkerrors.Register(ModuleName, 1520, "validator address is empty")
	ErrEmptyDelegationDistInfo = sdkerrors.Register(ModuleName, 1521, "no delegation distribution info")
	ErrNoValidatorDistInfo     = sdkerrors.Register(ModuleName, 1522, "no validator distribution info")
	ErrNoValidatorCommission   = sdkerrors.Register(ModuleName, 1523, "no validator commission to withdraw")
	ErrSetWithdrawAddrDisabled = sdkerrors.Register(ModuleName, 1524, "set withdraw address disabled")
	ErrBadDistribution         = sdkerrors.Register(ModuleName, 1525, "community pool does not have sufficient coins to distribute")
	ErrInvalidProposalAmount   = sdkerrors.Register(ModuleName, 1526, "invalid community pool spend proposal amount")
	ErrEmptyProposalRecipient  = sdkerrors.Register(ModuleName, 1527, "invalid community pool spend proposal recipient")
	ErrNoValidatorExists       = sdkerrors.Register(ModuleName, 1528, "validator does not exist")
	ErrNoDelegationExists      = sdkerrors.Register(ModuleName, 1529, "delegation does not exist")
)
