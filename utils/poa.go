package utils

import sdkmath "cosmossdk.io/math"

// WARNING: DO NOT MODIFY

// Validators are paid 15% of revenue
var PoaValPaymentRate = sdkmath.LegacyMustNewDecFromStr("0.15")

type PoaValidator struct {
	Moniker    string
	Operator   string // sdk.AccAddress bech32 — the payout + POA OperatorAddress
	HubAddress string
}

// REHEARSAL ONLY — DO NOT MERGE
// This branch (poa-migration-ig-tests) replaces the 8 mainnet entries with
// 3 test validators matching app/upgrades/v33/validators.json. See that
// file's sidecar REHEARSAL_ONLY marker for context.
var PoaValidatorSet = []PoaValidator{
	{Moniker: "Validator1", Operator: "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7", HubAddress: ""},
	{Moniker: "Validator2", Operator: "stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca", HubAddress: ""},
	{Moniker: "Validator3", Operator: "stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv", HubAddress: ""},
}
