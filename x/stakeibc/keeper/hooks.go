package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/utils"
	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const StrideEpochsPerDayEpoch = uint64(4)

// BeforeEpochStart runs the flows that finish open redemptions and move what is already in
// flight (reward claims, queued deposit transfers) after the v35 upgrade. The protocol is
// winding down (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md, §6):
// the calls that compounded or moved stake (rate refresh, reinvest, delegate, rebalance,
// reward-token transfer, withdrawal-address set, deposit and epoch-unbonding record creation,
// the reward-collector auction) were deleted rather than gated, so the redemption rate of every
// host zone is frozen at its last pre-upgrade value and cannot be toggled back. Their keeper
// functions remain defined until the follow-up cleanup.
func (k Keeper) BeforeEpochStart(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)

	// Update the stakeibc epoch tracker: the day tracker feeds the ICA timeout of every
	// undelegate batch and the stride tracker feeds the redemption sweep ICA
	epochNumber, err := k.UpdateEpochTracker(ctx, epochInfo)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to update epoch tracker, err: %s", err.Error()))
		return
	}

	// Day Epoch - submit and clean up unbondings
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		// Submit the queued redemption records of any host zone that unbonds this epoch
		k.InitiateAllHostZoneUnbondings(ctx, epochNumber)
		// Submit any one-shot undelegations queued by an upgrade handler (e.g. the v34 Injective
		// reconciliation). Store-driven, so this is a no-op when nothing is pending. A host zone
		// that unbonds this epoch is deferred so the two flows don't compete for validator capacity
		k.SubmitPendingUndelegations(ctx, epochNumber)
		// Delete epoch unbonding records once every host zone's unbonding on them is claimed
		k.CleanupEpochUnbondingRecords(ctx, epochNumber)
	}

	// Stride Epoch - move what is already in flight toward the redemption accounts
	if epochInfo.Identifier == epochstypes.STRIDE_EPOCH {
		depositInterval := k.GetParam(ctx, types.KeyDepositInterval)

		// Withdraw accrued staking rewards to the withdrawal ICA, where the wind-down sweeps them
		k.ClaimAccruedStakingRewards(ctx)

		// Transfer the amount of any remaining TRANSFER_QUEUE deposit record to the delegation ICA
		// so it leaves with the ICA balance instead of being stranded on Stride
		if epochNumber%depositInterval == 0 {
			depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}

		// Check previous epochs to see if unbondings finished, and send the relevant tokens
		// to the redemption account
		k.SweepUnbondedTokensAllHostZones(ctx)
	}
}

func (k Keeper) AfterEpochEnd(context context.Context, epochInfo epochstypes.EpochInfo) {}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)

	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)

	h.k.AfterEpochEnd(ctx, epochInfo)
}

// Set the withdrawal account address for each host zone
func (k Keeper) SetWithdrawalAddress(context context.Context) {
	ctx := sdk.UnwrapSDKContext(context)

	k.Logger(ctx).Info("Setting Withdrawal Addresses...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		err := k.SetWithdrawalAddressOnHost(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to set withdrawal address on %s, err: %s", hostZone.ChainId, err))
		}
	}
}

// Claim staking rewards for each host zone
func (k Keeper) ClaimAccruedStakingRewards(context context.Context) {
	ctx := sdk.UnwrapSDKContext(context)

	k.Logger(ctx).Info("Claiming Accrued Staking Rewards...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		err := k.ClaimAccruedStakingRewardsOnHost(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to claim accrued staking rewards on %s, err: %s", hostZone.ChainId, err))
		}
	}
}

// TODO [cleanup]: Remove after v17 upgrade
func (k Keeper) DisableHubTokenization(context context.Context) {
	ctx := sdk.UnwrapSDKContext(context)

	k.Logger(ctx).Info("Disabling the ability to tokenize Gaia delegations")

	chainId := "cosmoshub-4"
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		k.Logger(ctx).Error("Gaia host zone not found, unable to disable tokenization")
		return
	}

	// Build the msg for the disable tokenization ICA tx
	var msgs []proto.Message
	msgs = append(msgs, &types.MsgDisableTokenizeShares{
		DelegatorAddress: hostZone.DelegationIcaAddress,
	})

	// Send the ICA tx to disable tokenization
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(24 * time.Hour).UnixNano())
	delegationOwner := types.FormatHostZoneICAOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, delegationOwner, msgs, timeoutTimestamp)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to submit ICA tx to disable tokenization for gaia: %s", err.Error()))
		return
	}
}
