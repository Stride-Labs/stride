package types

import (
	fmt "fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	proto "github.com/cosmos/gogoproto/proto"
)

const (
	ICAAccountType_Oracle = "ORACLE"
)

type ICATx struct {
	ConnectionId    string
	ChannelId       string
	PortId          string
	Owner           string
	Messages        []proto.Message
	RelativeTimeout time.Duration
	CallbackArgs    proto.Message
	CallbackId      string
}

func (i ICATx) ValidateICATx() error {
	if i.ConnectionId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "connection-id is empty")
	}
	if i.ChannelId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "channel-id is empty")
	}
	if i.PortId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "port-id is empty")
	}
	if i.Owner == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "owner is empty")
	}
	if len(i.Messages) < 1 {
		return errorsmod.Wrapf(ErrInvalidICARequest, "messages are empty")
	}
	if i.RelativeTimeout <= 0 {
		return errorsmod.Wrapf(ErrInvalidICARequest,
			"relative timeout must be greater than 0, timeout: %d", i.RelativeTimeout)
	}
	if i.CallbackId == "" {
		return errorsmod.Wrapf(ErrInvalidICARequest, "callback-id is empty")
	}
	return nil
}

func (i ICATx) GetRelativeTimeoutNano() uint64 {
	return uint64(i.RelativeTimeout.Nanoseconds())
}

func FormatICAAccountOwner(chainId string, accountType string) string {
	return fmt.Sprintf("%s.%s", chainId, accountType)
}
