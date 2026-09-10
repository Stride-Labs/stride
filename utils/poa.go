package utils

import sdkmath "cosmossdk.io/math"

// WARNING: DO NOT MODIFY outside of a coordinated validator-set upgrade.
// This registry drives the stToken reward payout in
// x/stakeibc/keeper/reward_allocation.go AND is the source of truth for POA
// operator (payout) addresses joined by upgrade handlers. Entries must stay in
// sync with the on-chain POA validator set.

// Validators are paid 15% of revenue
var PoaValPaymentRate = sdkmath.LegacyMustNewDecFromStr("0.15")

// Placeholder payout addresses for incoming validators whose payout address is
// not yet confirmed. Deterministic sha256-derived addresses with no known
// private key — valid bech32 (reward-allocation code parses every operator
// with MustAccAddressFromBech32) but unmistakably not a real wallet.
// The v34 upgrade handler refuses to run while any of these remain in the set.
//
// Derivation: bech32("stride", sha256("v34-placeholder-payout-<name>")[:20])
const (
	PlaceholderOperatorCosmosRescue = "stride1kddnkeu5ccca350thdhs2087268x4w3mfxy8s5"
	PlaceholderOperatorCitizenWeb3  = "stride1yqestx8f9z4sx9yct5ew4jkk45ntqevrwkkcpy"
)

// IsPlaceholderOperator reports whether the operator address is one of the
// not-yet-confirmed placeholder payout addresses.
func IsPlaceholderOperator(operator string) bool {
	return operator == PlaceholderOperatorCosmosRescue || operator == PlaceholderOperatorCitizenWeb3
}

type PoaValidator struct {
	Moniker  string
	Operator string // sdk.AccAddress bech32 — the payout + POA OperatorAddress
}

var PoaValidatorSet = []PoaValidator{
	{Moniker: "Polkachu", Operator: "stride1gp957czryfgyvxwn3tfnyy2f0t9g2p4pxxdj7c"},
	{Moniker: "L5", Operator: "stride1wj9ckvakuzgvlgw3hwpmsfjxvsc7uke73ps4u8"},
	{Moniker: "Imperator", Operator: "stride13u4dsapth4m3hef3z8qgjtdnv06predefnndkw"},
	{Moniker: "Keplr", Operator: "stride1j79tw5chf34u88s30gxchzx2cu080elm4hqg5j"},
	{Moniker: "Stakecito", Operator: "stride1qe8uuf5x69c526h4nzxwv4ltftr73v7qr7y9vq"},
	{Moniker: "CryptoCrew", Operator: "stride1smuvvnjj6w7x6ytq9kdgvlj6er99y6648s3der"},
	// v34 additions — placeholder payout addresses until the validators
	// confirm real ones (see PlaceholderOperator* above).
	{Moniker: "cosmosrescue", Operator: PlaceholderOperatorCosmosRescue},
	{Moniker: "CitizenWeb3", Operator: PlaceholderOperatorCitizenWeb3},
}
