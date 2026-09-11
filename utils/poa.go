package utils

import sdkmath "cosmossdk.io/math"

// WARNING: DO NOT MODIFY outside of a coordinated validator-set upgrade.
// This registry drives the stToken reward payout in
// x/stakeibc/keeper/reward_allocation.go AND is the source of truth for POA
// operator (payout) addresses joined by upgrade handlers. Entries must stay in
// sync with the on-chain POA validator set.

// Validators are paid 15% of revenue
var PoaValPaymentRate = sdkmath.LegacyMustNewDecFromStr("0.15")

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
	// v34 additions
	{Moniker: "cosmosrescue", Operator: "stride19397kzaerflj7m5ll5rdeap5necvt3r258zt7a"},
	{Moniker: "CitizenWeb3", Operator: "stride1dlmvgvnfp27h4c789a4y09e4hradenyl3grcg4"},
}
