package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/autopilot module sentinel errors
var (
	ErrInvalidPacketMetadata       = errorsmod.Register(ModuleName, 1501, "invalid packet metadata")
	ErrUnsupportedStakeibcAction   = errorsmod.Register(ModuleName, 1502, "unsupported stakeibc action")
	ErrInvalidClaimAirdropId       = errorsmod.Register(ModuleName, 1503, "invalid claim airdrop ID (cannot be empty)")
	ErrMulitpleAutopilotRoutesInTx = errorsmod.Register(ModuleName, 1504, "multiple autopilot routes in the same transaction")
	ErrUnsupportedAutopilotRoute   = errorsmod.Register(ModuleName, 1505, "unsupported autpilot route")
)
