package v34

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v34"

	// PlaceholderConsPubKey marks a consensus pubkey the incoming validator has
	// not yet confirmed (via `strided tendermint show-validator`). The upgrade
	// handler refuses to run while any incoming validator still carries it.
	// The real value is the "key" field of that command's JSON output, e.g.
	// {"@type":"/cosmos.crypto.ed25519.PubKey","key":"<base64>"} — copy the
	// base64 string, not the whole JSON object.
	PlaceholderConsPubKey = "PLACEHOLDER"

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

// vars rather than consts so tests can substitute filled-in values.
var (
	IncomingValidators = []IncomingValidator{
		{Moniker: "cosmosrescue", ConsPubKeyBase64: PlaceholderConsPubKey},
		{Moniker: "CitizenWeb3", ConsPubKeyBase64: PlaceholderConsPubKey},
	}

	// OutgoingMonikers are resolved against live POA state at upgrade time —
	// no hand-transcribed consensus addresses to typo.
	OutgoingMonikers = []string{"Citadel.one", "Cosmostation"}
)
