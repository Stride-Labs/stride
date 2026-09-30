// x/stakeibc/types/wind_down.go
package types

import "time"

// Wind-down constants (spec §4, §7). The two addresses are vars so tests can set them; they
// ship empty in this PR and are filled by the release gate, and every use fails closed while
// they are empty (the transfer tx errors, the sweep gate rejects every signer).
var (
	SweepOperatorAddress = "" // stride1..., the only signer of MsgSweepTokensOffStride
	OsmosisVaultAddress  = "" // osmo1..., receiver of every MsgTransferFromIca
)

const (
	OsmosisChainId                   = "osmosis-1"
	OsmosisBech32Prefix              = "osmo"
	StrideToOsmosisTransferChannelId = "channel-5"
	WindDownTransferTimeout          = 24 * time.Hour
	MaxSweepAddressesPerTx           = 100
)

// Host-side transfer channel to osmosis-1 per in-scope zone (chain registry 2026-09-24, re-verified
// against each host before the proposal). osmosis-1 maps to "" which selects an ICA bank send.
var HostToOsmosisTransferChannel = map[string]string{
	"celestia":       "channel-2",
	"cosmoshub-4":    "channel-141",
	"dydx-mainnet-1": "channel-3",
	"haqq_11235-1":   "channel-2",
	"injective-1":    "channel-8",
	"juno-1":         "channel-0",
	"laozi-mainnet":  "channel-83",
	"phoenix-1":      "channel-1",
	"sommelier-3":    "channel-0",
	"ssc-1":          "channel-1",
	"osmosis-1":      "",
}

// Stride transfer channels a voucher may be unwound over, with the counterparty's bech32 prefix:
// exactly the chains whose wallets derive the same address bytes as Stride (spec §3, §7).
var SweepUnwindChannels = map[string]string{
	"channel-0":   "cosmos",
	"channel-162": "celestia",
	"channel-5":   "osmo",
	"channel-24":  "juno",
	"channel-150": "somm",
	"channel-213": "saga",
	"channel-160": "dydx",
}
