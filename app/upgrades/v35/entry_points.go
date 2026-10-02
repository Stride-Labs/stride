package v35

import (
	"fmt"

	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v35/x/autopilot/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// DisableAutopilotStakeibc turns off the autopilot liquid-stake and redeem paths, which
// bypass the msg service router that PR 1 emptied (spec §5 "Entry points that bypass the router").
func DisableAutopilotStakeibc(ctx sdk.Context, k autopilotkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.StakeibcActive = false
	k.SetParams(ctx, params)
	ctx.Logger().Info("v35: autopilot StakeibcActive set to false")
}

// RemoveStakeibcFromICAHostAllowList drops MsgLiquidStake and MsgRedeemStake from the ICA host
// allow-list so interchain accounts on Stride cannot reach the removed handlers either.
// MsgClaimUndelegatedTokens stays: ICA-originated claims keep paying the open redemptions.
// The existing list is filtered in place, never rewritten from a constant.
func RemoveStakeibcFromICAHostAllowList(ctx sdk.Context, k *icahostkeeper.Keeper) {
	removed := map[string]bool{
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}): true,
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}): true,
	}

	params := k.GetParams(ctx)
	kept := make([]string, 0, len(params.AllowMessages))
	for _, msgTypeUrl := range params.AllowMessages {
		if removed[msgTypeUrl] {
			ctx.Logger().Info(fmt.Sprintf("v35: removing %s from the ICA host allow-list", msgTypeUrl))
			continue
		}
		kept = append(kept, msgTypeUrl)
	}
	params.AllowMessages = kept
	k.SetParams(ctx, params)
}

// RemoveStakingFromICAHostAllowList drops the BlockedStakingMsgTypeUrls (delegate, redelegate,
// create validator, cancel unbonding) from the ICA host allow-list so an interchain account on
// Stride cannot re-lock STRD after the mass undelegation; the ante decorator closes the tx path
// (authority spec §3-4). MsgUndelegate stays. The existing list is filtered in place.
func RemoveStakingFromICAHostAllowList(ctx sdk.Context, k *icahostkeeper.Keeper) {
	removed := map[string]bool{}
	for _, msgTypeUrl := range BlockedStakingMsgTypeUrls {
		removed[msgTypeUrl] = true
	}

	params := k.GetParams(ctx)
	kept := make([]string, 0, len(params.AllowMessages))
	for _, msgTypeUrl := range params.AllowMessages {
		if removed[msgTypeUrl] {
			ctx.Logger().Info(fmt.Sprintf("v35: removing %s from the ICA host allow-list", msgTypeUrl))
			continue
		}
		kept = append(kept, msgTypeUrl)
	}
	params.AllowMessages = kept
	k.SetParams(ctx, params)
}
