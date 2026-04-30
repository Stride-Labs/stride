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
	{Moniker: "Polkachu", Operator: "stride1gp957czryfgyvxwn3tfnyy2f0t9g2p4pxxdj7c", HubAddress: "cosmosvalcons1m7fg8k39k2tyym5hgwrpf5wx9hqsr8vywuyrtm"},
	{Moniker: "L5", Operator: "stride1wj9ckvakuzgvlgw3hwpmsfjxvsc7uke73ps4u8", HubAddress: "cosmosvalcons1c5e86exd7jsyhcfqdejltdsagjfrvv8xv22368"},
	{Moniker: "Imperator", Operator: "stride13u4dsapth4m3hef3z8qgjtdnv06predefnndkw", HubAddress: "cosmosvalcons1pdpwglc4fcjdzqvyhvfwxg684trpc6uqck5sxk"},
	{Moniker: "Cosmostation", Operator: "stride1jj9z2xwxesuy65n90dujsak554eqkrr2ygyan2", HubAddress: "cosmosvalcons1px0zkz2cxvc6lh34uhafveea9jnaagckmrlsye"},
	{Moniker: "Keplr", Operator: "stride1j79tw5chf34u88s30gxchzx2cu080elm4hqg5j", HubAddress: "cosmosvalcons1vz42ewp04wwepjed7z4qenj925gpakgvap4q2u"},
	{Moniker: "Stakecito", Operator: "stride1qe8uuf5x69c526h4nzxwv4ltftr73v7qr7y9vq", HubAddress: "cosmosvalcons1upc05nc9pwhhagnkr3f2dft327qxsxfeyvajsu"},
	{Moniker: "Citadel.one", Operator: "stride1rgwn0h67xmuluymk4vvhtl4tqtgfg39j9zuk2z", HubAddress: "cosmosvalcons1f6cjsfn47ujttypx7gtncglsmjvndugc2zelqx"},
	{Moniker: "CryptoCrew", Operator: "stride1smuvvnjj6w7x6ytq9kdgvlj6er99y6648s3der", HubAddress: "cosmosvalcons1jufcrrd9gze26sxd82ppse03eg5g5xa2gplt6p"},
}
