package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

type BankKeeper interface {
	SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type IBCCallbackKeeper interface {
	SetCallbackData(ctx sdk.Context, callbackData icacallbacktypes.CallbackData)
	CallRegisteredICACallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbacktypes.AcknowledgementResponse) error
}
