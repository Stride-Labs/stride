package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v17/utils"
	stakeibctypes "github.com/Stride-Labs/stride/v17/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"
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

// IBC transfers all TIA in the deposit account and sends it to the delegation account
func (k Keeper) PrepareDelegation(ctx sdk.Context, epochNumber uint64, epochDuration time.Duration) error {
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
	nativeCoins := k.bankKeeper.GetBalance(ctx, depositAddress, hostZone.NativeTokenIbcDenom)

	// Create a new delgation record with status TRANSFER IN PROGRESS
	delegationRecord := types.DelegationRecord{
		Id:           epochNumber,
		NativeAmount: nativeCoins.Amount,
		Status:       types.TRANSFER_IN_PROGRESS,
	}
	err = k.SafelySetDelegationRecord(ctx, delegationRecord)
	if err != nil {
		return err
	}

	// Timeout the transfer at the end of the epoch
	timeoutTimestamp := uint64(ctx.BlockTime().Add(epochDuration).UnixNano())

	// Transfer the native tokens to the host chain
	transferMsgDepositToDelegation := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            nativeCoins,
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
