package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	recordtypes "github.com/Stride-Labs/stride/v22/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v22/x/stakeibc/types"
	oldtypes "github.com/Stride-Labs/stride/v22/x/staketia/legacytypes"
	"github.com/Stride-Labs/stride/v22/x/staketia/types"
)

// TODO [UPGRADE HANDLER]: Migrate stakeibc host zone (set redemptions enabled to true on each host zone)

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

// Update the newly created stakeibc host zone with the accounting values from staketia
func (k Keeper) UpdateStakeibcHostZone(ctx sdk.Context, legacyHostZone oldtypes.HostZone) (stakeibctypes.HostZone, error) {
	// Grab the newly created stakeibc host zone
	stakeibcHostZone, found := k.stakeibcKeeper.GetHostZone(ctx, types.CelestiaChainId)
	if !found {
		return stakeibctypes.HostZone{}, errors.New("celestia host zone not found in stakeibc after registration")
	}

	// Disable redemptions and set the redemption rate to the one from stakeibc
	stakeibcHostZone.RedemptionsEnabled = false
	stakeibcHostZone.RedemptionRate = legacyHostZone.LastRedemptionRate

	// Set the total delegations to the sum of the staketia total, plus any delegation records
	// This is so we don't have to trigger any stakeibc account changes when delegations are
	// confirmed from staketia
	// In practice, if timed right, there should be no delegation records
	pendingDelegations := sdkmath.ZeroInt()
	for _, delegationRecord := range k.GetAllActiveDelegationRecords(ctx) {
		pendingDelegations = pendingDelegations.Add(delegationRecord.NativeAmount)
	}
	stakeibcHostZone.TotalDelegations = legacyHostZone.DelegatedBalance.Add(pendingDelegations)
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
