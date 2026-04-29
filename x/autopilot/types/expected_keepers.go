package types

import (
	context "context"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	BlockedAddr(addr sdk.AccAddress) bool
	SendCoins(ctx context.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type IbcTransferKeeper interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}
