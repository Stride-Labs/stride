package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/airdrop module sentinel errors
var (
	ErrAirdropAlreadyExists        = sdkerrors.Register(ModuleName, 2001, "airdrop already exists")
	ErrAirdropNotFound             = sdkerrors.Register(ModuleName, 2002, "airdrop not found")
	ErrUserAllocationAlreadyExists = sdkerrors.Register(ModuleName, 2003, "user allocation already exists")
	ErrUserAllocationNotFound      = sdkerrors.Register(ModuleName, 2004, "user allocation not found")
)
