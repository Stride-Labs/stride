package types

import (
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ICACallbackFunction func(sdk.Context, channeltypes.Packet, *AcknowledgementResponse, []byte) error

type ICACallback struct {
	CallbackId   string
	CallbackFunc ICACallbackFunction
}

type ModuleCallbacks []ICACallback
