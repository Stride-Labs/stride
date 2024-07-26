package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v23/utils"
	recordtypes "github.com/Stride-Labs/stride/v23/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v23/x/stakeibc/types"
	oldtypes "github.com/Stride-Labs/stride/v23/x/staketia/legacytypes"
	"github.com/Stride-Labs/stride/v23/x/staketia/types"
)

// Helper to deserialize the host zone with the old types
func (k Keeper) GetLegacyHostZone(ctx sdk.Context) (hostZone oldtypes.HostZone, err error) {
	store := ctx.KVStore(k.storeKey)
	hostZoneBz := store.Get(types.HostZoneKey)

	if len(hostZoneBz) == 0 {
		return hostZone, types.ErrHostZoneNotFound.Wrapf("No HostZone found, there must be exactly one HostZone!")
	}

	k.cdc.MustUnmarshal(hostZoneBz, &hostZone)
	return hostZone, nil
}

// Helper to deserialize store a host zone with the old type definition
// (only used for tests)
func (k Keeper) SetLegacyHostZone(ctx sdk.Context, hostZone oldtypes.HostZone) {
	store := ctx.KVStore(k.storeKey)
	hostZoneBz := k.cdc.MustMarshal(&hostZone)
	store.Set(types.HostZoneKey, hostZoneBz)
}

// Update the newly created stakeibc host zone with the accounting values from staketia
func (k Keeper) UpdateStakeibcHostZone(ctx sdk.Context, legacyHostZone oldtypes.HostZone) (stakeibctypes.HostZone, error) {
	// Grab the newly created stakeibc host zone
	stakeibcHostZone, found := k.stakeibcKeeper.GetHostZone(ctx, types.CelestiaChainId)
	if !found {
		return stakeibctypes.HostZone{}, errors.New("celestia host zone not found in stakeibc after registration")
	}

	// Disable redemptions and set the redemption rate to the one from stakeibc
	stakeibcHostZone.RedemptionsEnabled = false
	stakeibcHostZone.RedemptionRate = legacyHostZone.RedemptionRate
	stakeibcHostZone.MinInnerRedemptionRate = legacyHostZone.MinInnerRedemptionRate
	stakeibcHostZone.MaxInnerRedemptionRate = legacyHostZone.MaxInnerRedemptionRate
	stakeibcHostZone.Halted = legacyHostZone.Halted

	// Set the total delegations to the sum of the staketia total
	stakeibcHostZone.TotalDelegations = legacyHostZone.DelegatedBalance
	k.stakeibcKeeper.SetHostZone(ctx, stakeibcHostZone)

	return stakeibcHostZone, nil
}

// Migrates the protocol owned accounts (deposit and fee) to their stakeibc counterparts
func (k Keeper) MigrateProtocolOwnedAccounts(
	ctx sdk.Context,
	legacyHostZone oldtypes.HostZone,
	stakeibcHostZone stakeibctypes.HostZone,
) error {
	// Transfer tokens from the staketia deposit account to the stakeibc deposit account
	ctx.Logger().Info("Migrating the deposit account...")
	staketiaDepositAddress, err := sdk.AccAddressFromBech32(legacyHostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid staketia deposit address")
	}
	stakeibcDepositAddress, err := sdk.AccAddressFromBech32(stakeibcHostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid stakeibc deposit address")
	}

	depositBalance := k.bankKeeper.GetBalance(ctx, staketiaDepositAddress, legacyHostZone.NativeTokenIbcDenom)
	err = k.bankKeeper.SendCoins(ctx, staketiaDepositAddress, stakeibcDepositAddress, sdk.NewCoins(depositBalance))
	if err != nil {
		return errorsmod.Wrapf(err, "unable to transfer deposit accounts")
	}

	// Add that deposit amount to the new stakeibc deposit record (in status TRANSFER_QUEUE)
	celestiaDepositRecords := []recordtypes.DepositRecord{}
	for _, depositRecord := range k.recordsKeeper.GetAllDepositRecord(ctx) {
		if depositRecord.HostZoneId == types.CelestiaChainId {
			celestiaDepositRecords = append(celestiaDepositRecords, depositRecord)
		}
	}

	if len(celestiaDepositRecords) != 1 || celestiaDepositRecords[0].Status != recordtypes.DepositRecord_TRANSFER_QUEUE {
		return errors.New("there should only be one celestia deposit record in status TRANSFER_QUEUE")
	}

	depositRecord := celestiaDepositRecords[0]
	depositRecord.Amount = depositBalance.Amount
	k.recordsKeeper.SetDepositRecord(ctx, depositRecord)

	// Transfer tokens from the staketia fee account to the stakeibc reward collector
	ctx.Logger().Info("Migrating the fee account...")
	staketiaFeeAddress := k.accountKeeper.GetModuleAddress(types.FeeAddress)
	stakeibcFeeAddress := stakeibctypes.RewardCollectorName

	feesBalance := k.bankKeeper.GetBalance(ctx, staketiaFeeAddress, legacyHostZone.NativeTokenIbcDenom)
	if feesBalance.IsZero() {
		ctx.Logger().Info("No fees to migrate")
		return nil
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, staketiaFeeAddress, stakeibcFeeAddress, sdk.NewCoins(feesBalance))
	if err != nil {
		return errorsmod.Wrapf(err, "unable to transfer fee accounts")
	}

	return nil
}

// Initiates the migration to stakeibc by registering the host zone
// and transferring funds to the new stakeibc accounts
// This will be called from the upgrade handler
func InitiateMigration(k Keeper, ctx sdk.Context) error {
	ctx.Logger().Info("Initiating staketia to stakeibc migration...")

	// Deserialize the staketia host zone with the old types (to recover the redemption rates)
	legacyHostZone, err := k.GetLegacyHostZone(ctx)
	if err != nil {
		return err
	}

	registerMsg := stakeibctypes.MsgRegisterHostZone{
		ConnectionId:                 types.CelestiaConnectionId,
		Bech32Prefix:                 types.CelestiaBechPrefix,
		HostDenom:                    legacyHostZone.NativeTokenDenom,
		IbcDenom:                     legacyHostZone.NativeTokenIbcDenom,
		TransferChannelId:            legacyHostZone.TransferChannelId,
		UnbondingPeriod:              types.CelestiaUnbondingPeriodDays,
		MinRedemptionRate:            legacyHostZone.MinRedemptionRate,
		MaxRedemptionRate:            legacyHostZone.MaxRedemptionRate,
		LsmLiquidStakeEnabled:        false,
		CommunityPoolTreasuryAddress: "",
		MaxMessagesPerIcaTx:          32,
	}

	ctx.Logger().Info("Registering the stakeibc host zone...")
	if _, err := k.stakeibcKeeper.RegisterHostZone(ctx, &registerMsg); err != nil {
		return errorsmod.Wrapf(err, "unable to register host zone with stakeibc")
	}

	ctx.Logger().Info("Updating the stakeibc host zone...")
	stakeibcHostZone, err := k.UpdateStakeibcHostZone(ctx, legacyHostZone)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to update the new stakeibc host zone")
	}

	ctx.Logger().Info("Migrating protocol owned accounts...")
	if err := k.MigrateProtocolOwnedAccounts(ctx, legacyHostZone, stakeibcHostZone); err != nil {
		return errorsmod.Wrapf(err, "unable to migrate protocol owned accounts")
	}

	ctx.Logger().Info("Done with staketia migration")
	return nil
}

// Direct copy of the, now deprecated, redemption rate update function that was in staketia
// This is used to verify nothing went wrong during the migration
func (k Keeper) UpdateRedemptionRate(ctx sdk.Context) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "Updating redemption rate"))

	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	// Get the number of stTokens from the supply
	stTokenSupply := k.bankKeeper.GetSupply(ctx, utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)).Amount
	if stTokenSupply.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"No st%s in circulation - redemption rate is unchanged", hostZone.NativeTokenDenom))
		return nil
	}

	// Get the balance of the deposit address
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid deposit address")
	}
	depositAccountBalance := k.bankKeeper.GetBalance(ctx, depositAddress, hostZone.NativeTokenIbcDenom)

	// Then add that to the sum of the delegation records to get the undelegated balance
	// Delegation records are only created once the tokens leave the deposit address
	// and the record is deleted once the tokens are delegated
	undelegatedBalance := sdkmath.ZeroInt()
	for _, delegationRecord := range k.GetAllActiveDelegationRecords(ctx) {
		undelegatedBalance = undelegatedBalance.Add(delegationRecord.NativeAmount)
	}

	// Finally, calculated the redemption rate as the native tokens locked divided by the stTokens
	nativeTokensLocked := depositAccountBalance.Amount.Add(undelegatedBalance).Add(hostZone.DelegatedBalance)
	if !nativeTokensLocked.IsPositive() {
		return errors.New("Non-zero stToken supply, yet the zero delegated and undelegated balance")
	}
	redemptionRate := sdk.NewDecFromInt(nativeTokensLocked).Quo(sdk.NewDecFromInt(stTokenSupply))

	// Set the old and update redemption rate on the host
	hostZone.LastRedemptionRate = hostZone.RedemptionRate
	hostZone.RedemptionRate = redemptionRate
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "Redemption rate updated from %v to %v",
		hostZone.LastRedemptionRate, hostZone.RedemptionRate))
	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId,
		"Deposit Account Balance: %v, Undelegated Balance: %v, Delegated Balance: %v, StToken Supply: %v",
		depositAccountBalance.Amount, undelegatedBalance, hostZone.DelegatedBalance, stTokenSupply))

	return nil
}

// Direct copy of the redemption rate update function in stakeibc
// This is used to verify nothing went wrong during the migration
func (k Keeper) UpdateRedemptionRateForHostZone(ctx sdk.Context, hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) {
	// Gather redemption rate components
	stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)).Amount
	if stSupply.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"No st%s in circulation - redemption rate is unchanged", hostZone.HostDenom))
		return
	}

	depositAccountBalance := k.GetDepositAccountBalance(hostZone.ChainId, depositRecords)
	undelegatedBalance := k.GetUndelegatedBalance(hostZone.ChainId, depositRecords)
	tokenizedDelegation := k.GetTotalTokenizedDelegations(ctx, hostZone)
	nativeDelegation := sdk.NewDecFromInt(hostZone.TotalDelegations)

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Redemption Rate Components - Deposit Account Balance: %v, Undelegated Balance: %v, "+
			"LSM Delegated Balance: %v, Native Delegations: %v, stToken Supply: %v",
		depositAccountBalance, undelegatedBalance, tokenizedDelegation,
		nativeDelegation, stSupply))

	// Calculate the redemption rate
	nativeTokensLocked := depositAccountBalance.Add(undelegatedBalance).Add(tokenizedDelegation).Add(nativeDelegation)
	redemptionRate := nativeTokensLocked.Quo(sdk.NewDecFromInt(stSupply))

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"New Redemption Rate: %v (vs Prev Rate: %v)", redemptionRate, hostZone.RedemptionRate))

	// Update the host zone
	hostZone.LastRedemptionRate = hostZone.RedemptionRate
	hostZone.RedemptionRate = redemptionRate
	k.SetHostZone(ctx, hostZone)

	// If the redemption rate is outside of safety bounds, exit so the redemption rate is not pushed to the oracle
	redemptionRateSafe, _ := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !redemptionRateSafe {
		return
	}

	// Otherwise, submit the redemption rate to the oracle
	if err := k.PostRedemptionRateToOracles(ctx, hostZone.HostDenom, redemptionRate); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to send redemption rate to oracle: %s", err.Error()))
		return
	}
}
