package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// The RestrictedWasmHooks removes the callback functionality so that the only hook
// is the contract execution that occurs in OnRecvPacket
// All other overrides are removed
type RestrictedWasmHooks struct {
	wasmHooks   *ibchooks.WasmHooks
	ics4Wrapper porttypes.ICS4Wrapper
}

func NewRestrictedWasmHooks(
	wasmHooks *ibchooks.WasmHooks,
	ics4Wrapper porttypes.ICS4Wrapper,
) RestrictedWasmHooks {
	return RestrictedWasmHooks{
		wasmHooks:   wasmHooks,
		ics4Wrapper: ics4Wrapper,
	}
}

func (h RestrictedWasmHooks) ProperlyConfigured() bool {
	return h.wasmHooks.ProperlyConfigured()
}

// The RestrictedWasmHooks OnRecvPacketOverride is the same as the WasmHooks OnRecvPacket
func (h RestrictedWasmHooks) OnRecvPacketOverride(
	im ibchooks.IBCMiddleware,
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return h.wasmHooks.OnRecvPacketOverride(im, ctx, packet, relayer)
}

// There is no SendPacketOverride for the RestrictedWasmHooks
// It passes directly to the next module in the stack
func (h RestrictedWasmHooks) SendPacketOverride(
	i ibchooks.ICS4Middleware,
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight ibcclienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return h.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// There is no OnAcknowledgementPacketOverride for the RestrictedWasmHooks
// It passes directly to the next module in the stack
func (h RestrictedWasmHooks) OnAcknowledgementPacketOverride(
	im ibchooks.IBCMiddleware,
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.App.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// There is no OnTimeoutPacketOverride for the RestrictedWasmHooks
// It passes directly to the next module in the stack
func (h RestrictedWasmHooks) OnTimeoutPacketOverride(
	im ibchooks.IBCMiddleware,
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.App.OnTimeoutPacket(ctx, packet, relayer)
}
