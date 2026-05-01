// REHEARSAL ONLY — DO NOT MERGE
// PoaValidatorSet has been replaced with test operator addresses generated
// from the rehearsal validator mnemonics in
// integration-tests/network/configs/keys.json. The mainnet set lives on the
// `main` branch.

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

var PoaValidatorSet = []PoaValidator{
	{Moniker: "Validator1", Operator: "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7", HubAddress: ""},
	{Moniker: "Validator2", Operator: "stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca", HubAddress: ""},
	{Moniker: "Validator3", Operator: "stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv", HubAddress: ""},
	{Moniker: "Validator4", Operator: "stride1py0fvhdtq4au3d9l88rec6vyda3e0wtt9szext", HubAddress: ""},
	{Moniker: "Validator5", Operator: "stride1c5jnf370kaxnv009yhc3jt27f549l5u36chzem", HubAddress: ""},
	{Moniker: "Validator6", Operator: "stride1eudadx6z3dp6pa4sgqx740tyvuasfh4nrt7rc2", HubAddress: ""},
	{Moniker: "Validator7", Operator: "stride1fm497naw27ck2d4z6z4ujcwq929ad5gexwvz8f", HubAddress: ""},
	{Moniker: "Validator8", Operator: "stride1gfvjzmrucy9xemzqktvg28m8n40wpv6l4fam6r", HubAddress: ""},
}
