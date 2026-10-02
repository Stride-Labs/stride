// x/stakeibc/types/wind_down.go
package types

import "time"

// Wind-down operator addresses (spec §4). Vars rather than consts so tests can substitute
// them; every use fails closed if one is ever emptied again. The vault is deliberately the
// protocol-admin multisig re-encoded with the osmo prefix (same signer set on both chains).
// Proving each address by a test transfer and a signed spend is an ops step (spec §9) recorded
// on the PR, not something a test asserts.
var (
	SweepOperatorAddress = "stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy" // the only signer of MsgSweepTokensOffStride (rehearsal: k8s test operator)
	OsmosisVaultAddress  = "osmo1mymazvsd79f9yhjq4n84dchyf6zvfmd8jaelyx"   // receiver of every MsgTransferFromIca; pool admin and moderator (rehearsal: k8s test vault)
)

const (
	OsmosisChainId                   = "osmosis-test-1" // rehearsal: mainnet is osmosis-1
	OsmosisBech32Prefix              = "osmo"
	StrideToOsmosisTransferChannelId = "channel-1"      // rehearsal: mainnet is channel-5
	WindDownTransferTimeout          = 15 * time.Minute // rehearsal: mainnet is 24h
)

// Host-side transfer channel to osmosis per in-scope zone (rehearsal: k8s test network; mainnet map
// covered every in-scope zone). The osmosis chain maps to "" which selects an ICA bank send.
var HostToOsmosisTransferChannel = map[string]string{
	"cosmoshub-test-1": "channel-1",
	"osmosis-test-1":   "",
}

// Stride transfer channels a voucher may be unwound over, with the counterparty's bech32 prefix:
// exactly the chains whose wallets derive the same address bytes as Stride (spec §3, §7).
var SweepUnwindChannels = map[string]string{
	"channel-0": "cosmos",
	"channel-1": "osmo",
}

// The staketia (S0-S2) and stakedym (S4-S6) multisigs, copied here because those packages import
// this one. A keeper test asserts they equal the constants in x/staketia/types/celestia.go and
// x/stakedym/types/dymension.go, so a drift fails the build of the suite
const (
	StaketiaDepositAddress    = "stride1ju3xt2f8xuhzxqg6590sazctlz6l4md0wc5w6c"
	StaketiaRedemptionAddress = "stride19ksqv50zmntzjfflfmnegj75tdfkk89vl2q5yu"
	StaketiaClaimAddress      = "stride1pjw24gg0fm26758hxee3wta35kq9jpszcslm6z"
	StakedymDepositAddress    = "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj"
	StakedymRedemptionAddress = "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n"
	StakedymClaimAddress      = "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras"
	StaketiaSafeAddress       = "stride1tpzfseenwg4kq54sf9hdp3mkra652fvqtsuclq"
	StakedymSafeAddress       = "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7"
	StaketiaOperatorAddress   = "stride19xm04qaah8t2eupyeglz63vkaxzytpyc8m7kk4"
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
