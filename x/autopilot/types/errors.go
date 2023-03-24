package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/autopilot module sentinel errors
var (
	ErrInvalidReceiverData       = errorsmod.Register(ModuleName, 1501, "invalid receiver data")
	ErrUnsupportedStakeibcAction = errorsmod.Register(ModuleName, 1502, "unsupported stakeibc action")
	ErrInvalidClaimAirdropId     = errorsmod.Register(ModuleName, 1503, "invalid claim airdrop ID (cannot be empty)")
)
