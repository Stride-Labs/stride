package ratelimit_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	ratelimit "github.com/Stride-Labs/stride/v4/x/ratelimit"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

const (
	transferPort = "transfer"
	uosmo        = "uosmo"
	ujuno        = "ujuno"
	ustrd        = "ustrd"
	stuatom      = "stuatom"
)

func hashDenomTrace(denomTrace string) string {
	trace32byte := sha256.Sum256([]byte(denomTrace))
	var traceTmByte tmbytes.HexBytes = trace32byte[:]
	return fmt.Sprintf("ibc/%s", traceTmByte)
}

func TestParseDenomFromSendPacket(t *testing.T) {
	testCases := []struct {
		name             string
		packetDenomTrace string
		expectedDenom    string
	}{
		// Native assets stay as is
		{
			name:             "ustrd",
			packetDenomTrace: ustrd,
			expectedDenom:    ustrd,
		},
		{
			name:             "stuatom",
			packetDenomTrace: stuatom,
			expectedDenom:    stuatom,
		},
		// Non-native assets are hashed
		{
			name:             "uosmo_one_hop",
			packetDenomTrace: "transfer/channel-0/usomo",
			expectedDenom:    hashDenomTrace("transfer/channel-0/usomo"),
		},
		{
			name:             "uosmo_two_hops",
			packetDenomTrace: "transfer/channel-2/transfer/channel-1/usomo",
			expectedDenom:    hashDenomTrace("transfer/channel-2/transfer/channel-1/usomo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet := transfertypes.FungibleTokenPacketData{
				Denom: tc.packetDenomTrace,
			}

			parsedDenom := ratelimit.ParseDenomFromSendPacket(packet)
			require.Equal(t, tc.expectedDenom, parsedDenom, tc.name)
		})
	}
}

func TestParseDenomFromRecvPacket(t *testing.T) {
	osmoChannelOnStride := "channel-0"
	strideChannelOnOsmo := "channel-100"
	junoChannelOnOsmo := "channel-200"
	junoChannelOnStride := "channel-300"

	testCases := []struct {
		name               string
		packetDenomTrace   string
		sourceChannel      string
		destinationChannel string
		expectedDenom      string
	}{
		// Sink asset one hop away:
		//   uosmo sent from Osmosis to Stride (uosmo)
		//   -> tack on prefix (transfer/channel-0/uosmo) and hash
		{
			name:               "sink_one_hop",
			packetDenomTrace:   uosmo,
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace(fmt.Sprintf("%s/%s/%s", transferPort, osmoChannelOnStride, uosmo)),
		},
		// Sink asset two hops away:
		//   ujuno sent from Juno to Osmosis to Stride (transfer/channel-200/ujuno)
		//   -> tack on prefix (transfer/channel-0/transfer/channel-200/ujuno) and hash
		{
			name:               "sink_two_hops",
			packetDenomTrace:   fmt.Sprintf("%s/%s/%s", transferPort, junoChannelOnOsmo, ujuno),
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace(fmt.Sprintf("%s/%s/%s/%s/%s", transferPort, osmoChannelOnStride, transferPort, junoChannelOnOsmo, ujuno)),
		},
		// Native source assets
		//    ustrd sent from Stride to Osmosis and then back to Stride (transfer/channel-0/ustrd)
		//    -> remove prefix and leave as is (ustrd)
		{
			name:               "native_source",
			packetDenomTrace:   fmt.Sprintf("%s/%s/%s", transferPort, strideChannelOnOsmo, ustrd),
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      ustrd,
		},
		// Non-native source assets
		//    ujuno was sent from Juno to Stride, then to Osmosis, then back to Stride (transfer/channel-0/transfer/channel-300/ujuno)
		//    -> remove prefix (transfer/channel-300/ujuno) and hash
		{
			name:               "non_native_source",
			packetDenomTrace:   fmt.Sprintf("%s/%s/%s/%s/%s", transferPort, strideChannelOnOsmo, transferPort, junoChannelOnStride, ujuno),
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace(fmt.Sprintf("%s/%s/%s", transferPort, junoChannelOnStride, ujuno)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet := channeltypes.Packet{
				SourcePort:         transferPort,
				DestinationPort:    transferPort,
				SourceChannel:      tc.sourceChannel,
				DestinationChannel: tc.destinationChannel,
			}
			packetData := transfertypes.FungibleTokenPacketData{
				Denom: tc.packetDenomTrace,
			}

			parsedDenom := ratelimit.ParseDenomFromRecvPacket(packet, packetData)
			require.Equal(t, tc.expectedDenom, parsedDenom, tc.name)
		})
	}
}

func TestSendRateLimitedPacket(t *testing.T) {
	s := apptesting.SetupSuitelessTestHelper()

	denom := "denom"
	sourceChannel := "channel-0"
	destinationChannel := "channel-1"

	// Create rate limit
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
		Path: &types.Path{
			Denom:     denom,
			ChannelId: sourceChannel, // for send, use source channel
		},
		Quota: &types.Quota{
			MaxPercentSend: sdk.NewInt(10),
			MaxPercentRecv: sdk.NewInt(10),
			DurationHours:  uint64(10),
		},
		Flow: &types.Flow{
			Inflow:       sdk.NewInt(0),
			Outflow:      sdk.NewInt(9), // outflow almost at threshold
			ChannelValue: sdk.NewInt(100),
		},
	})

	// This packet should cause an Outflow quota exceed error
	data, err := json.Marshal(transfertypes.FungibleTokenPacketData{Denom: denom, Amount: "5"})
	require.NoError(t, err)

	// We check for a quota error because it doesn't appear until the end of the function
	// We're avoiding checking for a success here because we can get a false positive if the rate limit doesn't exist
	err = ratelimit.SendRateLimitedPacket(s.Ctx, s.App.RatelimitKeeper, channeltypes.Packet{
		SourcePort:         transferPort,
		SourceChannel:      sourceChannel,
		DestinationPort:    transferPort,
		DestinationChannel: destinationChannel,
		Data:               data,
	})
	require.ErrorIs(t, err, types.ErrQuotaExceeded, "error type")
	require.ErrorContains(t, err, "Outflow exceeds quota", "error text")
}

func TestReceiveRateLimitedPacket(t *testing.T) {
	s := apptesting.SetupSuitelessTestHelper()

	denom := "denom"
	sourceChannel := "channel-1"
	destinationChannel := "channel-0"

	// Create rate limit
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
		Path: &types.Path{
			Denom:     hashDenomTrace(fmt.Sprintf("%s/%s/%s", transferPort, destinationChannel, denom)),
			ChannelId: destinationChannel, // for receive, use destination channel
		},
		Quota: &types.Quota{
			MaxPercentSend: sdk.NewInt(10),
			MaxPercentRecv: sdk.NewInt(10),
			DurationHours:  uint64(10),
		},
		Flow: &types.Flow{
			Inflow:       sdk.NewInt(9), // outflow almost at threshold
			Outflow:      sdk.NewInt(0),
			ChannelValue: sdk.NewInt(100),
		},
	})

	// This packet should cause an Outflow quota exceed error
	data, err := json.Marshal(transfertypes.FungibleTokenPacketData{Denom: denom, Amount: "5"})
	require.NoError(t, err)

	// We check for a quota error because it doesn't appear until the end of the function
	// We're avoiding checking for a success here because we can get a false positive if the rate limit doesn't exist
	err = ratelimit.ReceiveRateLimitedPacket(s.Ctx, s.App.RatelimitKeeper, channeltypes.Packet{
		SourcePort:         transferPort,
		SourceChannel:      sourceChannel,
		DestinationPort:    transferPort,
		DestinationChannel: destinationChannel,
		Data:               data,
	})
	require.ErrorIs(t, err, types.ErrQuotaExceeded, "error type")
	require.ErrorContains(t, err, "Inflow exceeds quota", "error text")
}
