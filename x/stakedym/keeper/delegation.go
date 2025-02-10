package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v25/utils"
	"github.com/Stride-Labs/stride/v25/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
)

// Liquid stakes native tokens and returns stTokens to the user
// The staker's native tokens (which exist as an IBC denom on stride) are escrowed
// in the deposit account
// StTokens are minted at the current redemption rate
func (k Keeper) LiquidStake(ctx sdk.Context, liquidStaker string, nativeAmount sdkmath.Int) (stToken sdk.Coin, err error) {
	// Get the host zone and verify it's unhalted
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return stToken, err
	}

	// Get user and deposit account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(liquidStaker)
	if err != nil {
		return stToken, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return stToken, errorsmod.Wrapf(err, "host zone deposit address is invalid")
	}

	// Check redemption rates are within safety bounds
	if err := k.CheckRedemptionRateExceedsBounds(ctx); err != nil {
		return stToken, err
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeToken := sdk.NewCoin(hostZone.NativeTokenIbcDenom, nativeAmount)
	if !utils.IsIBCToken(hostZone.NativeTokenIbcDenom) {
		return stToken, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken,
			"denom is not an IBC token (%s)", hostZone.NativeTokenIbcDenom)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdk.NewDecFromInt(nativeAmount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return stToken, errorsmod.Wrapf(stakeibctypes.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", nativeAmount.String(), hostZone.NativeTokenDenom)
	}

	// Transfer the native tokens from the user to module account
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "failed to send tokens from liquid staker %s to deposit address", liquidStaker)
	}

	// Mint the stTokens and transfer them to the user
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
	stToken = sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "Failed to mint stTokens")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "Failed to send %s from deposit address to liquid staker", stToken.String())
	}

	// Emit liquid stake event with the same schema as stakeibc
	EmitSuccessfulLiquidStakeEvent(ctx, liquidStaker, hostZone, nativeAmount, stAmount)

	return stToken, nil
}

// IBC transfers all DYM in the deposit account and sends it to the delegation account
func (k Keeper) PrepareDelegation(ctx sdk.Context, epochNumber uint64, epochDuration time.Duration) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.DymensionChainId, "Preparing delegation for epoch %d", epochNumber))

	// Only send the transfer if the host zone isn't halted
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return err
	}

	// safety check: if any delegation records are in progress, do not allow another transfer
	delegationRecords := k.GetAllActiveDelegationRecords(ctx)
	for _, record := range delegationRecords {
		if record.Status == types.TRANSFER_IN_PROGRESS {
			return errorsmod.Wrapf(types.ErrInvariantBroken,
				"cannot prepare delegation while a transfer is in progress, record ID %d", record.Id)
		}
	}

	// Transfer the full deposit balance which will include new liquid stakes, as well as reinvestment
	depositAddress := sdk.MustAccAddressFromBech32(hostZone.DepositAddress)
	nativeTokens := k.bankKeeper.GetBalance(ctx, depositAddress, hostZone.NativeTokenIbcDenom)

	// If there's nothing to delegate, exit early - no need to create a new record
	if nativeTokens.Amount.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(types.DymensionChainId, "No new liquid stakes for epoch %d", epochNumber))
		return nil
	}

	// Create a new delgation record with status TRANSFER IN PROGRESS
	delegationRecord := types.DelegationRecord{
		Id:           epochNumber,
		NativeAmount: nativeTokens.Amount,
		Status:       types.TRANSFER_IN_PROGRESS,
	}
	err = k.SafelySetDelegationRecord(ctx, delegationRecord)
	if err != nil {
		return err
	}

	// Timeout the transfer at the end of the epoch
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(epochDuration).UnixNano())

	// Transfer the native tokens to the host chain
	transferMsgDepositToDelegation := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            nativeTokens,
		Sender:           hostZone.DepositAddress,
		Receiver:         hostZone.DelegationAddress,
		TimeoutTimestamp: timeoutTimestamp,
	}
	msgResponse, err := k.transferKeeper.Transfer(ctx, &transferMsgDepositToDelegation)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to submit transfer from deposit to delegation acct in PrepareDelegation")
	}

	// Store the record ID so that we can access it during the packet callback to update the record status
	k.SetTransferInProgressRecordId(ctx, hostZone.TransferChannelId, msgResponse.Sequence, delegationRecord.Id)

	return nil
}

// Confirms a delegation has completed on the host zone, increments the internal delegated balance,
// and archives the record
func (k Keeper) ConfirmDelegation(ctx sdk.Context, recordId uint64, txHash string, sender string) (err error) {
	// grab unbonding record, verify record is ready to be delegated, and a hash hasn't already been posted
	delegationRecord, found := k.GetDelegationRecord(ctx, recordId)
	if !found {
		return types.ErrDelegationRecordNotFound.Wrapf("delegation record not found for %v", recordId)
	}
	if delegationRecord.Status != types.DELEGATION_QUEUE {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v is not in the correct state", recordId)
	}
	if delegationRecord.TxHash != "" {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v already has a txHash", recordId)
	}

	// note: we're intentionally not checking that the host zone is halted, because we still want to process this tx in that case
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	// verify delegation record is nonzero
	if !delegationRecord.NativeAmount.IsPositive() {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v has non positive delegation", recordId)
	}

	// update delegation record to archive it
	delegationRecord.TxHash = txHash
	delegationRecord.Status = types.DELEGATION_COMPLETE
	k.ArchiveDelegationRecord(ctx, delegationRecord)

	// increment delegation on Host Zone
	hostZone.DelegatedBalance = hostZone.DelegatedBalance.Add(delegationRecord.NativeAmount)
	k.SetHostZone(ctx, hostZone)

	EmitSuccessfulConfirmDelegationEvent(ctx, recordId, delegationRecord.NativeAmount, txHash, sender)
	return nil
}

// Liquid stakes tokens in the fee account and distributes them to the fee collector
func (k Keeper) LiquidStakeAndDistributeFees(ctx sdk.Context) error {
	// Get the fee address from the host zone
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return err
	}

	// Get the balance of native tokens in the fee address, if there are no tokens, no action is necessary
	feeAddress := k.accountKeeper.GetModuleAddress(types.FeeAddress)
	feesBalance := k.bankKeeper.GetBalance(ctx, feeAddress, hostZone.NativeTokenIbcDenom)
	if feesBalance.IsZero() {
		k.Logger(ctx).Info("No fees generated this epoch")
		return nil
	}

	// Liquid stake those native tokens
	stTokens, err := k.LiquidStake(ctx, feeAddress.String(), feesBalance.Amount)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to liquid stake fees")
	}

	// Send the stTokens to the fee collector
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.FeeAddress, authtypes.FeeCollectorName, sdk.NewCoins(stTokens))
	if err != nil {
		return errorsmod.Wrapf(err, "unable to send liquid staked tokens to fee collector")
	}
	k.Logger(ctx).Info(fmt.Sprintf("Liquid staked and sent %v to fee collector", stTokens))

	return nil
}

// Runs prepare delegations with a cache context wrapper so revert any partial state changes
func (k Keeper) SafelyPrepareDelegation(ctx sdk.Context, epochNumber uint64, epochDuration time.Duration) error {
	return utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.PrepareDelegation(ctx, epochNumber, epochDuration)
	})
}

// Liquid stakes fees with a cache context wrapper so revert any partial state changes
func (k Keeper) SafelyLiquidStakeAndDistributeFees(ctx sdk.Context) error {
	return utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.LiquidStakeAndDistributeFees(ctx)
	})
}
