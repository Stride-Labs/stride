package ratelimit_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	ratelimit "github.com/Stride-Labs/stride/v4/x/ratelimit"
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
			packetDenomTrace: "ustrd",
			expectedDenom:    "ustrd",
		},
		{
			name:             "statom",
			packetDenomTrace: "ustrd",
			expectedDenom:    "ustrd",
		},
		// Non-native assets are hashed
		{
			name:             "usomo_one_hop",
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
		//   uosmo sent from Osmosis to Stride
		//   -> tack on prefix and hash
		{
			name:               "sink_one_hop",
			packetDenomTrace:   "uosmo",
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace("transfer/" + osmoChannelOnStride + "/uosmo"),
		},
		// Sink asset two hops away:
		//   ujuno sent from Juno to Osmosis to Stride
		//   -> tack on prefix and hash
		{
			name:               "sink_two_hops",
			packetDenomTrace:   "transfer/" + junoChannelOnOsmo + "/ujuno",
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace("transfer/" + osmoChannelOnStride + "/transfer/" + junoChannelOnOsmo + "/ujuno"),
		},
		// Native source assets
		//    ustrd sent from Stride to Osmosis and then back to Stride
		//    -> remove prefix and leave as is
		{
			name:               "native_source",
			packetDenomTrace:   "transfer/" + strideChannelOnOsmo + "/ustrd",
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      "ustrd",
		},
		// Non-native source assets
		//    ujuno was sent from Juno to Stride, then to Osmosis, then back to Stride
		//    -> remove prefix and hash
		{
			name:               "non_native_source",
			packetDenomTrace:   "transfer/" + strideChannelOnOsmo + "/transfer/" + junoChannelOnStride + "/ujuno",
			sourceChannel:      strideChannelOnOsmo,
			destinationChannel: osmoChannelOnStride,
			expectedDenom:      hashDenomTrace("transfer/" + junoChannelOnStride + "/ujuno"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet := channeltypes.Packet{
				SourcePort:         "transfer",
				DestinationPort:    "transfer",
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
