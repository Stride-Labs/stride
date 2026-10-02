package v35

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v35"

	// UpgradeAuthority is the 2-of-3 team multisig that becomes the chain-wide consensus-params
	// authority, so it can submit MsgSoftwareUpgrade (and any other authority-gated message)
	// directly once stake is gone and gov can no longer pass anything (authority spec §3).
	UpgradeAuthority = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"

	// GovUnreachableDeposit (ustrd) closes gov submission: ~26,000x total supply. The multisig
	// can lower it again through gov MsgUpdateParams (authority spec §3).
	GovUnreachableDeposit = 1_000_000_000_000_000_000

	// StakingMaxEntries replaces mainnet's 7 so the mass undelegation never hits
	// ErrMaxUnbondingDelegationEntries on a pair that already holds 7 entries (authority spec §3).
	StakingMaxEntries = 100

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

// BlockedStakingMsgTypeUrls are the staking messages that could re-lock STRD after the mass
// undelegation. The ante decorator rejects them at the tx level and the handler drops them from
// the ICA host allow-list (authority spec §4).
var BlockedStakingMsgTypeUrls = []string{
	sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
	sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
}
