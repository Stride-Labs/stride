package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

type (
	Keeper struct {
		storeKey storetypes.StoreKey
		cdc      codec.BinaryCodec

		ics4Wrapper    porttypes.ICS4Wrapper
		transferKeeper transferkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey,
	ics4Wrapper porttypes.ICS4Wrapper,
	transferKeeper transferkeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       key,
		ics4Wrapper:    ics4Wrapper,
		transferKeeper: transferKeeper,
	}
}

func (k Keeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
}

func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
