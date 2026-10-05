package v35

import (
	"fmt"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/Stride-Labs/stride/v35/utils"
)

// SetConsensusAuthority makes the team multisig the chain-wide authority (authority spec §3).
// sdk.ValidateAuthority prefers the consensus-params authority over each module's own, so the
// multisig can submit MsgSoftwareUpgrade directly once stake is gone and gov cannot pass anything.
// The field is nil on mainnet today. The error is returned: skipping this step would leave the
// chain with no upgrade path after the undelegation.
func SetConsensusAuthority(ctx sdk.Context, k consensusparamkeeper.Keeper) error {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return err
	}
	params.Authority = &cmtproto.AuthorityParams{Authority: UpgradeAuthority}
	if err := k.ParamsStore.Set(ctx, params); err != nil {
		return err
	}
	ctx.Logger().Info(fmt.Sprintf("v35: consensus authority set to %s", UpgradeAuthority))
	return nil
}

// CloseGovSubmission raises the gov min deposit to GovUnreachableDeposit and the expedited min
// deposit to GovUnreachableExpeditedDeposit (strictly greater, as gov params validation requires),
// both far above total supply, so no proposal can be submitted once the multisig is the authority (authority spec §3). Every
// other gov param is kept as read; the multisig can lower the deposits again through gov
// MsgUpdateParams.
func CloseGovSubmission(ctx sdk.Context, k govkeeper.Keeper) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	minDeposit := sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(GovUnreachableDeposit)))
	expeditedMinDeposit := sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(GovUnreachableExpeditedDeposit)))
	params.MinDeposit = minDeposit
	params.ExpeditedMinDeposit = expeditedMinDeposit
	if err := k.Params.Set(ctx, params); err != nil {
		return err
	}
	ctx.Logger().Info(fmt.Sprintf("v35: gov min deposit and expedited min deposit set to %s and %s", minDeposit, expeditedMinDeposit))
	return nil
}

// SetPOAAdmin points the POA admin param at the same multisig (authority spec §3). The POA msg
// server gates its admin messages through sdk.ValidateAuthority too, so once the consensus
// authority is set only that address can rotate the validator set regardless of the param;
// writing it keeps state and reality in agreement. The error is returned like the other
// authority writes.
func SetPOAAdmin(ctx sdk.Context, k *poakeeper.Keeper) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	previous := params.Admin
	params.Admin = UpgradeAuthority
	if err := k.UpdateParams(ctx, params); err != nil {
		return err
	}
	ctx.Logger().Info(fmt.Sprintf("v35: POA admin set to %s (was %s)", UpgradeAuthority, previous))
	return nil
}
