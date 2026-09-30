package v35

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v35"

	HaqqChainId   = "haqq_11235-1"
	ComdexChainId = "comdex-1"

	// The one live trade route (dYdX USDC rewards → DYDX), keyed by the denoms as they
	// appear on the reward and host zones, not the IBC hashes (spec §13)
	DydxTradeRouteRewardDenom = "uusdc"
	DydxTradeRouteHostDenom   = "adydx"

	// The Stride deploy key that is admin of four Hyperlane contracts (spec §3); the
	// handler moves those admins to the gov module
	WasmDeployKey = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"
)

// InScopeChainIds are the non-deprecated stakeibc host zones the wind-down drains (spec §2).
// Used only by tests and the export README; the handler reads Deprecated from state.
var InScopeChainIds = []string{
	"celestia",
	"cosmoshub-4",
	"dydx-mainnet-1",
	"haqq_11235-1",
	"injective-1",
	"juno-1",
	"laozi-mainnet",
	"osmosis-1",
	"phoenix-1",
	"sommelier-3",
	"ssc-1",
}
