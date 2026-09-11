package v34

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v34"

	// ValidatorPower matches the uniform power of the existing POA set — the
	// value snapshotted from ICS at the v33 migration. POA only weighs relative
	// power, so incoming validators join at the same weight and total power is
	// unchanged by the swap.
	ValidatorPower = int64(274523)
)

// IncomingValidator identifies a validator added to the POA set by this
// upgrade. The payout/operator address deliberately does NOT live here — the
// handler joins it from utils.PoaValidatorSet by moniker so the address has a
// single source of truth.
type IncomingValidator struct {
	Moniker          string
	ConsPubKeyBase64 string // base64-encoded ed25519 consensus pubkey
}

// vars rather than consts so tests can substitute generated test keys.
var (
	// Pubkeys confirmed in writing by each validator (2026-09-11) from the
	// `key` field of `strided tendermint show-validator` on their running
	// nodes. Payout addresses live in utils.PoaValidatorSet.
	IncomingValidators = []IncomingValidator{
		{Moniker: "cosmosrescue", ConsPubKeyBase64: "X9ma3W9EfbHImJKJaUCoKQwDQU9eB1aGZJz2bVHfA3U="},
		{Moniker: "CitizenWeb3", ConsPubKeyBase64: "5tALxrcAfArCTEMJhCB4ISxsRogKyQzo/R7nZ5YbTMo="},
	}

	// OutgoingMonikers are resolved against live POA state at upgrade time —
	// no hand-transcribed consensus addresses to typo.
	OutgoingMonikers = []string{"Citadel.one", "Cosmostation"}
)
