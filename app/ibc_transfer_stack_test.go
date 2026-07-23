package app_test

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11"
	packetforward "github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware"
	packetforwardkeeper "github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/keeper"
	ratelimit "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting"
	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"
	"github.com/cosmos/ibc-go/v11/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v11/modules/core/04-channel/keeper"
	"github.com/stretchr/testify/require"

	strideapp "github.com/Stride-Labs/stride/v33/app"
	"github.com/Stride-Labs/stride/v33/x/autopilot"
	recordsmodule "github.com/Stride-Labs/stride/v33/x/records"
	"github.com/Stride-Labs/stride/v33/x/stakedym"
	"github.com/Stride-Labs/stride/v33/x/staketia"
)

// These tests pin the IBC transfer middleware sequencing in both directions so that
// dependency upgrades (e.g. swapping a forked middleware for the upstream version)
// can't silently reorder the stack.
//
// The middlewares store their neighbors in unexported fields, so the tests walk the
// chains via reflection and assert the concrete type at each hop.

// unwrapField returns the value of the named (possibly unexported) field of module
func unwrapField(t *testing.T, module any, fieldName string) any {
	t.Helper()
	value := reflect.ValueOf(module)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Unexported fields can only be read through an addressable value
	if !value.CanAddr() {
		addressable := reflect.New(value.Type()).Elem()
		addressable.Set(value)
		value = addressable
	}

	field := value.FieldByName(fieldName)
	require.Truef(t, field.IsValid(), "field %q not found on %v", fieldName, value.Type())

	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

// Inbound packets (RecvPacket) enter at the top of the transfer stack and flow down:
//
//	core IBC -> autopilot -> records -> stakedym -> staketia -> ratelimit
//	         -> ibchooks -> packetforward -> transfer
func TestTransferStackInboundOrder(t *testing.T) {
	app := strideapp.InitStrideTestApp(false)

	stack, found := app.IBCKeeper.PortKeeper.Router.Route(ibctransfertypes.ModuleName)
	require.True(t, found, "transfer route should be registered on the IBC router")

	var module any = stack
	require.IsType(t, &autopilot.IBCModule{}, module)

	module = unwrapField(t, module, "app")
	require.IsType(t, &recordsmodule.IBCModule{}, module)

	module = unwrapField(t, module, "app")
	require.IsType(t, &stakedym.IBCMiddleware{}, module)

	module = unwrapField(t, module, "app")
	require.IsType(t, &staketia.IBCMiddleware{}, module)

	module = unwrapField(t, module, "app")
	require.IsType(t, &ratelimit.IBCMiddleware{}, module)

	// Rate limiting wraps ibc-hooks through the app's packet unmarshaler adapter,
	// which embeds the ibc-hooks middleware and delegates all IBCModule callbacks to it
	module = unwrapField(t, module, "app")
	require.Contains(t, fmt.Sprintf("%T", module), "ibcHooksWithPacketUnmarshaler")

	module = unwrapField(t, module, "IBCMiddleware")
	require.IsType(t, &ibchooks.IBCMiddleware{}, module)

	module = unwrapField(t, module, "App")
	require.IsType(t, &packetforward.IBCMiddleware{}, module)

	module = unwrapField(t, module, "app")
	require.IsType(t, &transfer.IBCModule{}, module)
}

// Outbound packets (SendPacket) start at the transfer keeper and flow up through
// each keeper's ICS4Wrapper:
//
//	transfer -> packetforward -> ibchooks -> ratelimit -> core IBC channel keeper
//
// Note: this is intentionally a different sequence than the inbound stack — the
// autopilot/records/staketia/stakedym middlewares short-circuit SendPacket and
// must not sit on the outbound path (see the transfer stack wiring in app.go)
func TestTransferStackOutboundOrder(t *testing.T) {
	app := strideapp.InitStrideTestApp(false)

	// transfer keeper sends through the packet forward keeper (pointer identity)
	ics4Wrapper := unwrapField(t, app.TransferKeeper, "ics4Wrapper")
	require.IsType(t, &packetforwardkeeper.Keeper{}, ics4Wrapper)
	require.Same(t, app.PacketForwardKeeper, ics4Wrapper)

	// packet forward keeper sends through the ibchooks ICS4 middleware
	ics4Wrapper = unwrapField(t, app.PacketForwardKeeper, "ics4Wrapper")
	require.IsType(t, ibchooks.ICS4Middleware{}, ics4Wrapper)

	// ibchooks sends through the ratelimit keeper (pointer identity)
	ics4Wrapper = unwrapField(t, ics4Wrapper, "channel")
	require.IsType(t, &ratelimitkeeper.Keeper{}, ics4Wrapper)
	require.Same(t, &app.RatelimitKeeper, ics4Wrapper)

	// ratelimit sends through the core IBC channel keeper (pointer identity)
	ics4Wrapper = unwrapField(t, ics4Wrapper, "ics4Wrapper")
	require.IsType(t, &channelkeeper.Keeper{}, ics4Wrapper)
	require.Same(t, app.IBCKeeper.ChannelKeeper, ics4Wrapper)
}
