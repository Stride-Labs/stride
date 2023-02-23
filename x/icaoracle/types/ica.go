package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

const (
	ICAAccountType_Oracle = "ORACLE"
)

type ICATx struct {
	ConnectionId string
	ChannelId    string
	PortId       string
	Messages     []sdk.Msg
	Timeout      uint64
	CallbackArgs proto.Message
	CallbackId   string
}

func (i ICATx) ValidateICATx(ctx sdk.Context) error {
	blockTime := ctx.BlockTime().UnixNano()
	if i.ConnectionId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "connection-id is empty")
	}
	if i.ChannelId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "channel-id is empty")
	}
	if i.PortId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "port-id is empty")
	}
	if len(i.Messages) < 1 {
		return errorsmod.Wrapf(ErrInvalidICARequest, "messages are empty")
	}
	if i.Timeout < uint64(ctx.BlockTime().UnixNano()) {
		return errorsmod.Wrapf(ErrInvalidICARequest,
			"timeout is not in the future, timeout: %d, block time: %d", i.Timeout, blockTime)
	}
	if i.CallbackId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "callback-id is empty")
	}
	return nil
}

func FormatICAAccountOwner(chainId string, accountType string) string {
	return fmt.Sprintf("%s.%s", chainId, accountType)
}
