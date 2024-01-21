package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// TODO [sttia]
func (k Keeper) RedeemStake(ctx sdk.Context, staker string, stTokenAmount sdkmath.Int) error {
	return nil
}

// TODO [sttia]
func (k Keeper) PrepareUndelegation(ctx sdk.Context, epochNumber uint64) error {
	return nil
}

// TODO [sttia]
func (k Keeper) CheckUnbondingFinished(ctx sdk.Context) error {
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
