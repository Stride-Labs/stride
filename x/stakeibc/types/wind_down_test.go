// x/stakeibc/types/wind_down_test.go
package types_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// Every non-deprecated stakeibc zone (spec §2) and nothing else.
var inScopeChainIds = []string{
	"celestia", "cosmoshub-4", "dydx-mainnet-1", "haqq_11235-1", "injective-1", "juno-1",
	"laozi-mainnet", "osmosis-1", "phoenix-1", "sommelier-3", "ssc-1",
}

var deprecatedChainIds = []string{"comdex-1", "evmos_9001-2", "stargaze-1", "umee-1"}

func TestHostToOsmosisTransferChannel(t *testing.T) {
	for _, chainId := range inScopeChainIds {
		channelId, found := types.HostToOsmosisTransferChannel[chainId]
		require.True(t, found, "%s must have a host-side channel to osmosis", chainId)
		if chainId == types.OsmosisChainId {
			require.Empty(t, channelId, "osmosis-1 maps to an empty channel (bank send form)")
			continue
		}
		require.True(t, channeltypes.IsValidChannelID(channelId), "%s channel %q is not a channel id", chainId, channelId)
	}
	for _, chainId := range deprecatedChainIds {
		_, found := types.HostToOsmosisTransferChannel[chainId]
		require.False(t, found, "deprecated zone %s must not be sweepable", chainId)
	}
	require.Len(t, types.HostToOsmosisTransferChannel, len(inScopeChainIds), "no extra zones in the map")
}

func TestSweepUnwindChannels(t *testing.T) {
	expected := map[string]string{
		"channel-0":   "cosmos",
		"channel-162": "celestia",
		"channel-5":   "osmo",
		"channel-24":  "juno",
		"channel-150": "somm",
		"channel-213": "saga",
		"channel-160": "dydx",
	}
	require.Equal(t, expected, types.SweepUnwindChannels)
	for channelId := range types.SweepUnwindChannels {
		require.True(t, channeltypes.IsValidChannelID(channelId))
	}
	require.Equal(t, types.OsmosisBech32Prefix, types.SweepUnwindChannels[types.StrideToOsmosisTransferChannelId],
		"the Stride->Osmosis channel unwinds to the osmo prefix")
}

// The release gate fills the two operator addresses; from here on they must be set and must
// carry the right prefix and length. Which keys they are is decided in spec §4 and proven by
// the signed spends recorded on the PR, not asserted here.
func TestOperatorAddresses(t *testing.T) {
	require.NotEmpty(t, types.OsmosisVaultAddress, "OsmosisVaultAddress must be filled by the release gate")
	vaultBytes, err := sdk.GetFromBech32(types.OsmosisVaultAddress, types.OsmosisBech32Prefix)
	require.NoError(t, err, "osmosis vault must be an osmo bech32 address")
	require.Len(t, vaultBytes, 20, "osmosis vault must be a 20-byte account address")

	require.NotEmpty(t, types.SweepOperatorAddress, "SweepOperatorAddress must be filled by the release gate")
	operatorBytes, err := sdk.GetFromBech32(types.SweepOperatorAddress, "stride")
	require.NoError(t, err, "sweep operator must be a stride bech32 address")
	require.Len(t, operatorBytes, 20, "sweep operator must be a 20-byte account address")
}

// The transfer builder branches on the chain id, so nothing but osmosis-1 may map to ""
func TestOnlyOsmosisMapsToEmptyChannel(t *testing.T) {
	for chainId, channelId := range types.HostToOsmosisTransferChannel {
		if chainId == types.OsmosisChainId {
			require.Empty(t, channelId)
			continue
		}
		require.NotEmpty(t, channelId, "%s must have a channel", chainId)
	}
}
