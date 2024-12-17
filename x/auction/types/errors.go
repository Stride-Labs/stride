package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/auction module sentinel errors
var (
	ErrAuctionAlreadyExists = sdkerrors.Register(ModuleName, 7001, "auction already exists")
	ErrAuctionDoesntExist   = sdkerrors.Register(ModuleName, 7002, "auction doesn't exists")
)
