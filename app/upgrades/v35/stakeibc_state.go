package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// DeprecateComdex flags comdex-1 as deprecated so it carries the same flag as the other three
// dead zones (spec §5 "Comdex"). Halted is not touched; the flag is documentation and the
// wind-down txs refuse deprecated zones.
func DeprecateComdex(ctx sdk.Context, k stakeibckeeper.Keeper) {
	hostZone, found := k.GetHostZone(ctx, ComdexChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping deprecation", ComdexChainId))
		return
	}
	hostZone.Deprecated = true
	k.SetHostZone(ctx, hostZone)
	ctx.Logger().Info(fmt.Sprintf("v35: %s marked deprecated", ComdexChainId))
}

// DeleteDydxTradeRoute removes the one live trade route (spec §5 "Trade route"). Its USDC is
// swept from the withdrawal ICA by MsgTransferFromIca; the converter ICAs are written off.
// For the record (mainnet 2026-09-29, `/Stride-Labs/stride/stakeibc/trade_routes`): the route
// is uusdc/adydx; host account (dydx-mainnet-1 WITHDRAWAL, the zone's withdrawal ICA)
// dydx1tfcqhkf4tzqupknpl6wvdnusm7jl8ah8907cqlkvm06fsff4xj5sypzmad; reward account (noble-1
// CONVERTER_UNWIND) noble19mapxcry0frl8esuga29w40sfw7wc3hcgf4l0d6jed92auravl2ssjsp6g; trade
// account (osmosis-1 CONVERTER_TRADE, holds the stale authz grant to the trade controller)
// osmo1znalva74f4e0e0flkw932ru9vlsm6dxjugpecave8x7re9ecfkdq25nnja. Deleting the route removes
// nothing from those accounts; whatever dust they hold is written off (spec §3).
func DeleteDydxTradeRoute(ctx sdk.Context, k stakeibckeeper.Keeper) {
	if _, found := k.GetTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom); !found {
		ctx.Logger().Info(fmt.Sprintf("v35: trade route %s/%s not found, skipping deletion",
			DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom))
		return
	}
	k.RemoveTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom)
	ctx.Logger().Info(fmt.Sprintf("v35: trade route %s/%s deleted", DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom))
}
