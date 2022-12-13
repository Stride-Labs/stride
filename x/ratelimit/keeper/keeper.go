package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

type (
	Keeper struct {
		storeKey storetypes.StoreKey
		cdc      codec.BinaryCodec

		ics4Wrapper porttypes.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey,
	ics4Wrapper porttypes.ICS4Wrapper,
) *Keeper {
	return &Keeper{
		cdc:         cdc,
		storeKey:    key,
		ics4Wrapper: ics4Wrapper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
}

func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
