// x/stakeibc/types/wind_down.go
package types

import "time"

// Wind-down operator addresses (spec §4). Vars rather than consts so tests can substitute
// them; every use fails closed if one is ever emptied again. The vault is deliberately the
// protocol-admin multisig re-encoded with the osmo prefix (same signer set on both chains).
// Proving each address by a test transfer and a signed spend is an ops step (spec §9) recorded
// on the PR, not something a test asserts.
var (
	SweepOperatorAddress = "stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9" // the only signer of MsgSweepTokensOffStride
	OsmosisVaultAddress  = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"   // receiver of every MsgTransferFromIca; pool admin and moderator
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

// The staketia (S0-S2) and stakedym (S4-S6) multisigs, copied here because those packages import
// this one. A keeper test asserts they equal the constants in x/staketia/types/celestia.go and
// x/stakedym/types/dymension.go, so a drift fails the build of the suite
const (
	StaketiaDepositAddress    = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04"
	StaketiaRedemptionAddress = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g"
	StaketiaClaimAddress      = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd"
	StakedymDepositAddress    = "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj"
	StakedymRedemptionAddress = "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n"
	StakedymClaimAddress      = "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras"
	StaketiaSafeAddress       = "stride18p7xg4hj2u3zpk0v9gq68pjyuuua5wa387sjjc"
	StakedymSafeAddress       = "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7"
	StaketiaOperatorAddress   = "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw"
)

// SweepProtocolAddresses lists the addresses the sweep must never touch even though each is a
// plain 20-byte BaseAccount that passes every account-type rule: the staketia and stakedym
// deposit, redemption and claim multisigs, their safes and the staketia operator, which module code spends from, plus the sweep
// operator when one is set. Read per tx because the operator is a var filled by the release gate
func SweepProtocolAddresses() []string {
	addresses := []string{
		StaketiaDepositAddress,
		StaketiaRedemptionAddress,
		StaketiaClaimAddress,
		StakedymDepositAddress,
		StakedymRedemptionAddress,
		StakedymClaimAddress,
		StaketiaSafeAddress,
		StakedymSafeAddress,
		StaketiaOperatorAddress,
	}
	if SweepOperatorAddress != "" {
		addresses = append(addresses, SweepOperatorAddress)
	}
	return addresses
}
