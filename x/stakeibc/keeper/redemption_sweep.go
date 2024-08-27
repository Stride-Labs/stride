package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v23/utils"
	recordstypes "github.com/Stride-Labs/stride/v23/x/records/types"
	"github.com/Stride-Labs/stride/v23/x/stakeibc/types"
)

// Gets the total unbonded amount for the host zone that has finished unbonding
func (k Keeper) GetTotalRedemptionSweepAmountAndRecordIds(
	ctx sdk.Context,
	chainId string,
	hostBlockTime uint64,
) (totalSweepAmount sdkmath.Int, unbondingRecordIds []uint64) {
	// Sum the total unbonded amount for each unbonding record
	totalSweepAmount = sdkmath.ZeroInt()
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		// Get all the unbondings associated with the epoch + host zone pair
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, chainId)
		if !found {
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Epoch %d - Status: %s, Amount: %v, Unbonding Time: %d, Block Time: %d",
			epochUnbondingRecord.EpochNumber, hostZoneUnbonding.Status.String(),
			hostZoneUnbonding.NativeTokenAmount, hostZoneUnbonding.UnbondingTime, hostBlockTime))

		// If the unbonding period has elapsed, then we can send the ICA call to sweep this
		//   hostZone's unbondings to the redemption account (in a batch).
		// Verify:
		//      1. the unbonding time is set (g.t. 0)
		//      2. the unbonding time is less than the current block time
		//      3. the host zone is in the EXIT_TRANSFER_QUEUE state, meaning it's ready to be transferred
		inTransferQueue := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		validUnbondingTime := hostZoneUnbonding.UnbondingTime > 0 && hostZoneUnbonding.UnbondingTime < hostBlockTime
		if inTransferQueue && validUnbondingTime {
			k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "  %v%s included in sweep", hostZoneUnbonding.NativeTokenAmount, hostZoneUnbonding.Denom))

			totalSweepAmount = totalSweepAmount.Add(hostZoneUnbonding.NativeTokenAmount)
			unbondingRecordIds = append(unbondingRecordIds, epochUnbondingRecord.EpochNumber)
		}
	}

	return totalSweepAmount, unbondingRecordIds
}

// Batch transfers any unbonded tokens from the delegation account to the redemption account
func (k Keeper) SweepUnbondedTokensForHostZone(ctx sdk.Context, hostZone types.HostZone) error {
	chainId := hostZone.ChainId
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Sweeping unbonded tokens"))

	// Confirm the delegation (destination) and redemption (source) accounts are registered
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}
	if hostZone.RedemptionIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no redemption account found for %s", chainId)
	}

	// Get latest blockTime from light client
	hostBlockTime, err := k.GetLightClientTime(ctx, hostZone.ConnectionId)
	if err != nil {
		return errorsmod.Wrapf(err, "could not get light client block time for host zone")
	}

	// Determine the total unbonded amount that has finished unbonding
	totalSweepAmount, epochUnbondingRecordIds := k.GetTotalRedemptionSweepAmountAndRecordIds(ctx, chainId, hostBlockTime)

	// If we have any amount to sweep, then we can send the ICA call to sweep them
	if totalSweepAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "No tokens ready for sweep"))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Batch transferring %v to host zone", totalSweepAmount))

	// Build transfer message to transfer from the delegation account to redemption account
	sweepCoin := sdk.NewCoin(hostZone.HostDenom, totalSweepAmount)
	msgs := []proto.Message{
		&banktypes.MsgSend{
			FromAddress: hostZone.DelegationIcaAddress,
			ToAddress:   hostZone.RedemptionIcaAddress,
			Amount:      sdk.NewCoins(sweepCoin),
		},
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Preparing MsgSend from Delegation Account to Redemption Account"))

	// Store the epoch numbers in the callback to identify the epoch unbonding records
	redemptionCallback := types.RedemptionCallback{
		HostZoneId:              chainId,
		EpochUnbondingRecordIds: epochUnbondingRecordIds,
	}
	marshalledCallbackArgs, err := proto.Marshal(&redemptionCallback)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal redemption callback")
	}

	// Send the bank send ICA
	_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, types.ICAAccountType_DELEGATION, ICACallbackID_Redemption, marshalledCallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to submit redemption ICA for %s", chainId)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "ICA MsgSend Successfully Sent"))

	// Update the host zone unbonding records to status IN_PROGRESS
	err = k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, chainId, epochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS)
	if err != nil {
		return err
	}

	EmitRedemptionSweepEvent(ctx, hostZone, totalSweepAmount)

	return nil
}

// Sends all unbonded tokens that have finished unbonding to the redemption account
// Each host zone acts atomically - if an error is thrown, the state changes are discarded
func (k Keeper) SweepUnbondedTokensAllHostZones(ctx sdk.Context) {
	k.Logger(ctx).Info("Sweeping All Unbonded Tokens...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		err := utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.SweepUnbondedTokensForHostZone(ctx, hostZone)
		})
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error initiating redemption sweep for host zone %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
	}
}
