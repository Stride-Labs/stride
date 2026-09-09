package v33

// FrozenValidator pins the moniker → operator join the v33 upgrade handler
// used when it executed on mainnet. utils.PoaValidatorSet is a live registry
// that later upgrades edit (v34 swaps two entries); v33 is historical and must
// not track it. Copied verbatim from utils/poa.go as of the v33 release
// (HubAddress omitted — v33's join never read it).
type FrozenValidator struct {
	Moniker  string
	Operator string
}

var FrozenValidatorSet = []FrozenValidator{
	{Moniker: "Polkachu", Operator: "stride1gp957czryfgyvxwn3tfnyy2f0t9g2p4pxxdj7c"},
	{Moniker: "L5", Operator: "stride1wj9ckvakuzgvlgw3hwpmsfjxvsc7uke73ps4u8"},
	{Moniker: "Imperator", Operator: "stride13u4dsapth4m3hef3z8qgjtdnv06predefnndkw"},
	{Moniker: "Cosmostation", Operator: "stride1jj9z2xwxesuy65n90dujsak554eqkrr2ygyan2"},
	{Moniker: "Keplr", Operator: "stride1j79tw5chf34u88s30gxchzx2cu080elm4hqg5j"},
	{Moniker: "Stakecito", Operator: "stride1qe8uuf5x69c526h4nzxwv4ltftr73v7qr7y9vq"},
	{Moniker: "Citadel.one", Operator: "stride1rgwn0h67xmuluymk4vvhtl4tqtgfg39j9zuk2z"},
	{Moniker: "CryptoCrew", Operator: "stride1smuvvnjj6w7x6ytq9kdgvlj6er99y6648s3der"},
}
