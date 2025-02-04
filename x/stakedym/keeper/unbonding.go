package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v25/utils"
	"github.com/Stride-Labs/stride/v25/x/stakedym/types"
)

// Takes custody of staked tokens in an escrow account, updates the current
// accumulating UnbondingRecord with the amount taken, and creates or updates
// the RedemptionRecord for this user
func (k Keeper) RedeemStake(ctx sdk.Context, redeemer string, stTokenAmount sdkmath.Int) (nativeToken sdk.Coin, err error) {
	// Validate Basic already has ensured redeemer is legal address, stTokenAmount is above min threshold

	// Check HostZone exists, has legal redemption address for escrow, is not halted, has RR in bounds
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return nativeToken, err
	}

	escrowAccount, err := sdk.AccAddressFromBech32(hostZone.RedemptionAddress)
	if err != nil {
		return nativeToken, errorsmod.Wrapf(err, "could not bech32 decode redemption address %s on stride", hostZone.RedemptionAddress)
	}

	err = k.CheckRedemptionRateExceedsBounds(ctx)
	if err != nil {
		return nativeToken, err
	}

	// Get the current accumulating UnbondingRecord
	accUnbondingRecord, err := k.GetAccumulatingUnbondingRecord(ctx)
	if err != nil {
		return nativeToken, err
	}

	// Check redeemer owns at least stTokenAmount of stadym
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
	redeemerAccount, err := sdk.AccAddressFromBech32(redeemer)
	if err != nil {
		return nativeToken, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", redeemer)
	}

	balance := k.bankKeeper.GetBalance(ctx, redeemerAccount, stDenom)
	if balance.Amount.LT(stTokenAmount) {
		return nativeToken, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"wallet balance of stTIA is lower than redemption amount. %v < %v: ", balance.Amount, stTokenAmount)
	}

	// Estimate a placeholder native amount with current RedemptionRate
	// this estimate will be updated when the Undelegation record is finalized
	nativeAmount := sdk.NewDecFromInt(stTokenAmount).Mul(hostZone.RedemptionRate).TruncateInt()
	if nativeAmount.GT(hostZone.DelegatedBalance) {
		return nativeToken, errorsmod.Wrapf(types.ErrUnbondAmountToLarge,
			"cannot unstake an amount g.t. total staked balance: %v > %v", nativeAmount, hostZone.DelegatedBalance)
	}

	// Update the accumulating UnbondingRecord with the undelegation amounts
	accUnbondingRecord.StTokenAmount = accUnbondingRecord.StTokenAmount.Add(stTokenAmount)
	accUnbondingRecord.NativeAmount = accUnbondingRecord.NativeAmount.Add(nativeAmount)

	// Update or create the RedemptionRecord for this redeemer
	redemptionRecord, userHasActiveRedemptionRecord := k.GetRedemptionRecord(ctx, accUnbondingRecord.Id, redeemer)
	if userHasActiveRedemptionRecord {
		// Already active RedemptionRecord found for this redeemer this epoch so will update it
		redemptionRecord.StTokenAmount = redemptionRecord.StTokenAmount.Add(stTokenAmount)
		redemptionRecord.NativeAmount = redemptionRecord.NativeAmount.Add(nativeAmount)
	} else {
		// Creating new RedemptionRecord for this redeemer this epoch
		redemptionRecord = types.RedemptionRecord{
			UnbondingRecordId: accUnbondingRecord.Id,
			Redeemer:          redeemer,
			NativeAmount:      nativeAmount,
			StTokenAmount:     stTokenAmount,
		}
	}
	nativeToken = sdk.NewCoin(hostZone.NativeTokenDenom, nativeAmount) // Should it be NativeTokenIbcDenom?

	// Escrow user's stTIA balance before setting either record in the store to verify everything worked
	redeemCoins := sdk.NewCoins(sdk.NewCoin(stDenom, stTokenAmount))
	err = k.bankKeeper.SendCoins(ctx, redeemerAccount, escrowAccount, redeemCoins)
	if err != nil {
		return nativeToken, errorsmod.Wrapf(err, "couldn't send %v stadym. err: %s", stTokenAmount, err.Error())
	}

	// Now that escrow succeeded, actually set the updated records in the store
	k.SetUnbondingRecord(ctx, accUnbondingRecord)
	k.SetRedemptionRecord(ctx, redemptionRecord)

	EmitSuccessfulRedeemStakeEvent(ctx, redeemer, hostZone, nativeAmount, stTokenAmount)
	return nativeToken, nil
}

// Freezes the ACCUMULATING record by changing the status to UNBONDING_QUEUE
// and updating the native token amounts on the unbonding and redemption records
func (k Keeper) PrepareUndelegation(ctx sdk.Context, epochNumber uint64) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.DymensionChainId, "Preparing undelegation for epoch %d", epochNumber))

	// Get the redemption record from the host zone (to calculate the native tokens)
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return err
	}
	redemptionRate := hostZone.RedemptionRate

	// Get the one accumulating record that has the redemptions for the past epoch
	unbondingRecord, err := k.GetAccumulatingUnbondingRecord(ctx)
	if err != nil {
		return err
	}

	// Create the new accumulating record for this epoch
	newUnbondingRecord := types.UnbondingRecord{
		Id:            epochNumber,
		Status:        types.ACCUMULATING_REDEMPTIONS,
		StTokenAmount: sdkmath.ZeroInt(),
		NativeAmount:  sdkmath.ZeroInt(),
	}
	if err := k.SafelySetUnbondingRecord(ctx, newUnbondingRecord); err != nil {
		return err
	}

	// Update the number of native tokens for all the redemption records
	// Keep track of the total for the unbonding record
	totalNativeTokens := sdkmath.ZeroInt()
	for _, redemptionRecord := range k.GetRedemptionRecordsFromUnbondingId(ctx, unbondingRecord.Id) {
		nativeAmount := sdk.NewDecFromInt(redemptionRecord.StTokenAmount).Mul(redemptionRate).TruncateInt()
		redemptionRecord.NativeAmount = nativeAmount
		k.SetRedemptionRecord(ctx, redemptionRecord)
		totalNativeTokens = totalNativeTokens.Add(nativeAmount)
	}

	// If there were no unbondings this epoch, archive the current record
	if totalNativeTokens.IsZero() {
		k.ArchiveUnbondingRecord(ctx, unbondingRecord)
		return nil
	}

	// Update the total on the record and change the status to QUEUE
	unbondingRecord.Status = types.UNBONDING_QUEUE
	unbondingRecord.NativeAmount = totalNativeTokens
	k.SetUnbondingRecord(ctx, unbondingRecord)

	return nil
}

// Confirms that an undelegation has been completed on the host zone
// Updates the record status to UNBONDING_IN_PROGRESS, decrements the delegated balance and burns stTokens
func (k Keeper) ConfirmUndelegation(ctx sdk.Context, recordId uint64, txHash string, sender string) (err error) {
	// grab unbonding record, verify it's in the right state, and no tx hash has been submitted yet
	record, found := k.GetUnbondingRecord(ctx, recordId)
	if !found {
		return errorsmod.Wrapf(types.ErrUnbondingRecordNotFound, "couldn't find unbonding record with id: %d", recordId)
	}
	if record.Status != types.UNBONDING_QUEUE {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d is not ready to be undelegated", recordId)
	}
	if record.UndelegationTxHash != "" {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d already has undelegation tx hash set", recordId)
	}
	if record.UnbondedTokenSweepTxHash != "" {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d already has token sweep tx hash set", recordId)
	}

	// if there are no tokens to unbond (or negative on the record): throw an error!
	noTokensUnbondedOrNegative := record.NativeAmount.LTE(sdk.ZeroInt()) || record.StTokenAmount.LTE(sdk.ZeroInt())
	if noTokensUnbondedOrNegative {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d has no tokens to unbond (or negative)", recordId)
	}

	// Note: we're intentionally not checking that the host zone is halted, because we still want to process this tx in that case
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	// sanity check: store down the stToken supply and DelegatedBalance for checking against after burn
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
	stTokenSupplyBefore := k.bankKeeper.GetSupply(ctx, stDenom).Amount
	delegatedBalanceBefore := hostZone.DelegatedBalance

	// update the record's txhash, status, and unbonding completion time
	unbondingLength := time.Duration(hostZone.UnbondingPeriodSeconds) * time.Second         // 21 days
	unbondingCompletionTime := utils.IntToUint(ctx.BlockTime().Add(unbondingLength).Unix()) // now + 21 days

	record.UndelegationTxHash = txHash
	record.Status = types.UNBONDING_IN_PROGRESS
	record.UnbondingCompletionTimeSeconds = unbondingCompletionTime
	k.SetUnbondingRecord(ctx, record)

	// update host zone struct's delegated balance
	amountAddedToDelegation := record.NativeAmount
	newDelegatedBalance := hostZone.DelegatedBalance.Sub(amountAddedToDelegation)

	// sanity check: if the new balance is negative, throw an error
	if newDelegatedBalance.IsNegative() {
		return errorsmod.Wrapf(types.ErrNegativeNotAllowed, "host zone's delegated balance would be negative after undelegation")
	}
	hostZone.DelegatedBalance = newDelegatedBalance
	k.SetHostZone(ctx, hostZone)

	// burn the corresponding stTokens from the redemptionAddress
	stTokensToBurn := sdk.NewCoins(sdk.NewCoin(stDenom, record.StTokenAmount))
	if err := k.BurnRedeemedStTokens(ctx, stTokensToBurn, hostZone.RedemptionAddress); err != nil {
		return errorsmod.Wrapf(err, "unable to burn stTokens in ConfirmUndelegation")
	}

	// sanity check: check that (DelegatedBalance increment / stToken supply decrement) is within outer bounds
	if err := k.VerifyImpliedRedemptionRateFromUnbonding(ctx, stTokenSupplyBefore, delegatedBalanceBefore); err != nil {
		return errorsmod.Wrap(err, "ratio of delegation change to burned tokens exceeds redemption rate bounds")
	}

	EmitSuccessfulConfirmUndelegationEvent(ctx, recordId, record.NativeAmount, txHash, sender)
	return nil
}

// Burn stTokens from the redemption account
// - this requires sending them to an module account first, then burning them from there.
// - we use the stakedym module account
func (k Keeper) BurnRedeemedStTokens(ctx sdk.Context, stTokensToBurn sdk.Coins, redemptionAddress string) error {
	acctAddressRedemption, err := sdk.AccAddressFromBech32(redemptionAddress)
	if err != nil {
		return fmt.Errorf("could not bech32 decode address %s", redemptionAddress)
	}

	// send tokens from the EOA to the stakedym module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, acctAddressRedemption, types.ModuleName, stTokensToBurn)
	if err != nil {
		return errorsmod.Wrapf(err, "could not send coins from account %s to module %s. err: %s", redemptionAddress, types.ModuleName, err)
	}

	// burn the stTokens from the stakedym module account
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, stTokensToBurn)
	if err != nil {
		return errorsmod.Wrapf(err, "couldn't burn %v tokens in module account", stTokensToBurn)
	}

	return nil
}

// Sanity check helper for checking diffs on delegated balance and stToken supply are within outer RR bounds
func (k Keeper) VerifyImpliedRedemptionRateFromUnbonding(ctx sdk.Context, stTokenSupplyBefore sdkmath.Int, delegatedBalanceBefore sdkmath.Int) error {
	hostZoneAfter, err := k.GetHostZone(ctx)
	if err != nil {
		return types.ErrHostZoneNotFound
	}
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZoneAfter.NativeTokenDenom)

	// grab the delegated balance and token supply after the burn
	delegatedBalanceAfter := hostZoneAfter.DelegatedBalance
	stTokenSupplyAfter := k.bankKeeper.GetSupply(ctx, stDenom).Amount

	// calculate the delta for both the delegated balance and stToken burn
	delegatedBalanceDecremented := delegatedBalanceBefore.Sub(delegatedBalanceAfter)
	stTokenSupplyBurned := stTokenSupplyBefore.Sub(stTokenSupplyAfter)

	// It shouldn't be possible for this to be zero, but this will prevent a division by zero error
	if stTokenSupplyBurned.IsZero() {
		return types.ErrDivisionByZero
	}

	// calculate the ratio of delegated balance change to stToken burn - it should be close to the redemption rate
	ratio := sdk.NewDecFromInt(delegatedBalanceDecremented).Quo(sdk.NewDecFromInt(stTokenSupplyBurned))

	// check ratio against bounds
	if ratio.LT(hostZoneAfter.MinRedemptionRate) || ratio.GT(hostZoneAfter.MaxRedemptionRate) {
		return types.ErrRedemptionRateOutsideSafetyBounds
	}
	return nil
}

// Checks for any unbonding records that have finished unbonding,
// identified by having status UNBONDING_IN_PROGRESS and an
// unbonding that's older than the current time.
// Records are annotated with a new status UNBONDED
func (k Keeper) MarkFinishedUnbondings(ctx sdk.Context) {
	for _, unbondingRecord := range k.GetAllUnbondingRecordsByStatus(ctx, types.UNBONDING_IN_PROGRESS) {
		if ctx.BlockTime().Unix() > utils.UintToInt(unbondingRecord.UnbondingCompletionTimeSeconds) {
			unbondingRecord.Status = types.UNBONDED
			k.SetUnbondingRecord(ctx, unbondingRecord)
		}
	}
}

// Confirms that unbonded tokens have been sent back to stride and marks the unbonding record CLAIMABLE
func (k Keeper) ConfirmUnbondedTokenSweep(ctx sdk.Context, recordId uint64, txHash string, sender string) (err error) {
	// grab unbonding record and verify the record is ready to be swept, and has not been swept yet
	record, found := k.GetUnbondingRecord(ctx, recordId)
	if !found {
		return errorsmod.Wrapf(types.ErrUnbondingRecordNotFound, "couldn't find unbonding record with id: %d", recordId)
	}
	if record.Status != types.UNBONDED {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d is not ready to be swept", recordId)
	}
	if record.UnbondedTokenSweepTxHash != "" {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d already has a tx hash set", recordId)
	}

	// verify amount to sweep is positive
	unbondingRecordIsNonPositive := !record.NativeAmount.IsPositive() || !record.StTokenAmount.IsPositive()
	if unbondingRecordIsNonPositive {
		return errorsmod.Wrapf(types.ErrInvalidUnbondingRecord, "unbonding record with id: %d has non positive amount to sweep", recordId)
	}

	// grab claim address from host zone
	// note: we're intentionally not checking that the host zone is halted, because we still want to process this tx in that case
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	claimAddress, err := sdk.AccAddressFromBech32(hostZone.ClaimAddress)
	if err != nil {
		return err
	}

	// verify the claim address has the same or more tokens than the record (necessary condition if sweep was successful)
	claimAddressBalance := k.bankKeeper.GetBalance(ctx, claimAddress, hostZone.NativeTokenIbcDenom)
	if claimAddressBalance.Amount.LT(record.NativeAmount) {
		return errorsmod.Wrapf(types.ErrInsufficientFunds, "claim address %s has insufficient funds to confirm sweep unbonded tokens", hostZone.ClaimAddress)
	}

	// update record status to CLAIMABLE
	record.Status = types.CLAIMABLE
	record.UnbondedTokenSweepTxHash = txHash
	k.SetUnbondingRecord(ctx, record)

	EmitSuccessfulConfirmUnbondedTokenSweepEvent(ctx, recordId, record.NativeAmount, txHash, sender)
	return nil
}

// Iterates all unbonding records and distributes unbonded tokens to redeemers
// This function will operate atomically by using a cache context wrapper when
// it's invoked. This means that if any redemption send fails across any unbonding
// records, all partial state will be reverted
func (k Keeper) DistributeClaims(ctx sdk.Context) error {
	// Get the claim address which will be the sender
	// The token denom will be the native host zone token in it's IBC form as it lives on stride
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return err
	}
	nativeTokenIbcDenom := hostZone.NativeTokenIbcDenom

	claimAddress, err := sdk.AccAddressFromBech32(hostZone.ClaimAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid host zone claim address %s", hostZone.ClaimAddress)
	}

	// Loop through each claimable unbonding record and send out all the relevant claims
	for _, unbondingRecord := range k.GetAllUnbondingRecordsByStatus(ctx, types.CLAIMABLE) {
		if err := k.DistributeClaimsForUnbondingRecord(ctx, nativeTokenIbcDenom, claimAddress, unbondingRecord.Id); err != nil {
			return errorsmod.Wrapf(err, "Unable to distribute claims for unbonding record %d: %s",
				unbondingRecord.Id, err.Error())
		}

		// Once all claims have been distributed for a record, archive the record
		unbondingRecord.Status = types.CLAIMED
		k.ArchiveUnbondingRecord(ctx, unbondingRecord)
	}

	return nil
}

// Distribute claims for a given unbonding record
func (k Keeper) DistributeClaimsForUnbondingRecord(
	ctx sdk.Context,
	hostNativeIbcDenom string,
	claimAddress sdk.AccAddress,
	unbondingRecordId uint64,
) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.DymensionChainId,
		"Distributing claims for unbonding record %d", unbondingRecordId))

	// For each redemption record, bank send from the claim address to the user address and then delete the record
	for _, redemptionRecord := range k.GetRedemptionRecordsFromUnbondingId(ctx, unbondingRecordId) {
		userAddress, err := sdk.AccAddressFromBech32(redemptionRecord.Redeemer)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid redeemer address %s", userAddress)
		}

		nativeTokens := sdk.NewCoin(hostNativeIbcDenom, redemptionRecord.NativeAmount)
		if err := k.bankKeeper.SendCoins(ctx, claimAddress, userAddress, sdk.NewCoins(nativeTokens)); err != nil {
			return errorsmod.Wrapf(err, "unable to send %v from claim address to %s",
				nativeTokens, redemptionRecord.Redeemer)
		}

		k.RemoveRedemptionRecord(ctx, unbondingRecordId, redemptionRecord.Redeemer)
	}
	return nil
}

// Runs prepare undelegations with a cache context wrapper so revert any partial state changes
func (k Keeper) SafelyPrepareUndelegation(ctx sdk.Context, epochNumber uint64) error {
	return utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.PrepareUndelegation(ctx, epochNumber)
	})
}

// Runs distribute claims with a cache context wrapper so revert any partial state changes
func (k Keeper) SafelyDistributeClaims(ctx sdk.Context) error {
	return utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.DistributeClaims(ctx)
	})
}
