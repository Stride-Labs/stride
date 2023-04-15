package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/golang/protobuf/proto" //nolint:staticcheck

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

// Submits an ICA to "Redeem" an LSM Token - meaning converting the token into native stake
// This function is called in the EndBlocker which means if the ICA submission fails,
//   any modified state is not reverted
// The deposit Status is intentionally updated before the ICA is submitted even though it will NOT be reverted
//   if the ICA fails to send. This is because a failure is likely caused by a closed ICA channel, and the status
//   update will prevent the ICA from being continuously re-submitted. When the ICA channel is restored, the
//   deposit status will get reset, and the ICA will be attempted again.
func (k Keeper) DetokenizeLSMDeposit(ctx sdk.Context, hostZone types.HostZone, deposit types.LSMTokenDeposit) error {
	// Get the delegation account (which owns the LSM token)
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Build the detokenization ICA message
	token := sdk.NewCoin(deposit.Denom, deposit.Amount)
	detokenizeMsg := []sdk.Msg{&types.MsgRedeemTokensforShares{
		DelegatorAddress: delegationAccount.Address,
		Amount:           token,
	}}

	// Store the LSMTokenDeposit for the callback
	callbackArgs := types.DetokenizeSharesCallback{
		Deposit: &deposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	if err != nil {
		return err
	}

	// Mark the deposit as IN_PROGRESS
	deposit.Status = types.DETOKENIZATION_IN_PROGRESS
	k.SetLSMTokenDeposit(ctx, deposit)

	// Submit the ICA with a 24 hour timeout
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour * 24).Nanoseconds()) // 1 day
	if _, err := k.SubmitTxs(ctx, hostZone.ConnectionId, detokenizeMsg, *delegationAccount, timeout, ICACallbackID_Detokenize, callbackArgsBz); err != nil {
		return errorsmod.Wrapf(err, "unable to submit detokenization ICA for %s", deposit.Denom)
	}

	return nil
}
