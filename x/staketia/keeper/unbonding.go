package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v17/utils"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"
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

	// Check redeemer owns at least stTokenAmount of stutia
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
	nativeAmount := sdk.NewDecFromInt(stTokenAmount).Mul(hostZone.RedemptionRate).RoundInt()
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
		return nativeToken, errorsmod.Wrapf(err, "couldn't send %v stutia. err: %s", stTokenAmount, err.Error())
	}

	// Now that escrow succeeded, actually set the updated records in the store
	k.SetUnbondingRecord(ctx, accUnbondingRecord)
	k.SetRedemptionRecord(ctx, redemptionRecord)

	// TODO: emit events for redeem stake
	return nativeToken, nil
}

// Freezes the ACCUMULATING record by changing the status to UNBONDING_QUEUE
// and updating the native token amounts on the unbonding and redemption records
func (k Keeper) PrepareUndelegation(ctx sdk.Context, epochNumber uint64) error {
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
		nativeAmount := sdk.NewDecFromInt(redemptionRecord.StTokenAmount).Mul(redemptionRate).RoundInt()
		redemptionRecord.NativeAmount = nativeAmount
		k.SetRedemptionRecord(ctx, redemptionRecord)
		totalNativeTokens = totalNativeTokens.Add(nativeAmount)
	}

	// If there were no unbondings this epoch, archive the current record
	if totalNativeTokens.IsZero() {
		k.ArchiveUnbondingRecord(ctx, unbondingRecord.Id)
		return nil
	}

	// Update the total on the record and change the status to QUEUE
	unbondingRecord.Status = types.UNBONDING_QUEUE
	unbondingRecord.NativeAmount = totalNativeTokens
	k.SetUnbondingRecord(ctx, unbondingRecord)

	return nil
}

// Checks for any unbonding records that have finished unbonding,
// identified by having status UNBONDING_IN_PROGRESS and an
// unbonding that's older than the current time.
// Records are annotated with a new status UNBONDED
func (k Keeper) CheckUnbondingFinished(ctx sdk.Context) {
	for _, unbondingRecord := range k.GetAllUnbondingRecordsByStatus(ctx, types.UNBONDING_IN_PROGRESS) {
		if ctx.BlockTime().Unix() > int64(unbondingRecord.UnbondingCompletionTimeSeconds) {
			unbondingRecord.Status = types.UNBONDED
			k.SetUnbondingRecord(ctx, unbondingRecord)
		}
	}
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
		k.ArchiveUnbondingRecord(ctx, unbondingRecord.Id)
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
	// For each redemption record, bank send from the claim address to the user address
	for _, redemptionRecord := range k.GetRedemptionRecordsFromUnbondingId(ctx, unbondingRecordId) {
		userAddress, err := sdk.AccAddressFromBech32(redemptionRecord.Redeemer)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid redeemer address %s", userAddress)
		}

		nativeTokens := sdk.NewCoin(hostNativeIbcDenom, redemptionRecord.NativeAmount)
		if err := k.bankKeeper.SendCoins(ctx, claimAddress, userAddress, sdk.NewCoins(nativeTokens)); err != nil {
			return errorsmod.Wrapf(err, "unable to send %v from claim address to %s", nativeTokens, redemptionRecord.Redeemer)
		}
	}
	return nil
}
