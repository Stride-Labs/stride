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
	ErrClaimTypeUnavailable        = sdkerrors.Register(ModuleName, 2005, "claim type is unavailable due to a previous selection")
	ErrDistributionNotStarted      = sdkerrors.Register(ModuleName, 2006, "airdrop distribution has not started")
	ErrDistributionEnded           = sdkerrors.Register(ModuleName, 2007, "airdrop distribution has ended")
	ErrNoUnclaimedRewards          = sdkerrors.Register(ModuleName, 2008, "no unclaimed rewards")
	ErrAfterDecisionDeadline       = sdkerrors.Register(ModuleName, 2009, "claim type decision deadline passed")
	ErrUserLinkAlreadyExists       = sdkerrors.Register(ModuleName, 2010, "user link already exists")
	ErrAddUserLink                 = sdkerrors.Register(ModuleName, 2011, "cannot add user link")
	ErrUserLinkNotFound            = sdkerrors.Register(ModuleName, 2012, "user links not found")
)
