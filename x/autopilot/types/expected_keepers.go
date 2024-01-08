package types

import (
	context "context"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type IbcTransferKeeper interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}
