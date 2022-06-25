package stakeibc

import (
	"time"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/telemetry"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	epochInBlocks := int64(k.GetParam(ctx, types.KeyBlockBasedEpochInterval))
	epochNumber := ctx.BlockHeight() / epochInBlocks
	if ctx.BlockHeight() % int64(epochInBlocks) == 0 {
		k.Logger(ctx).Info("Triggering deposits")
		// Create a new deposit record for each host zone for the upcoming epoch
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)
	
		depositRecords := k.GetAllDepositRecord(ctx)
		depositInterval := int64(k.GetParam(ctx, types.KeyDepositInterval))
		if epochNumber%depositInterval == 0 {
			// process previous deposit records
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}
		
		// NOTE: the stake ICA timeout *must* be l.t. the staking epoch length, otherwise
		// we could send a stake ICA call (which could succeed), without deleting the record.
		// This could happen if the ack doesn't return by the next epoch. We would then send
		// *another* stake ICA call, for a portion of the balance which has *already* been staked,
		// which is very bad! This could result in the protocol becoming insolvent, by staking balances
		// that were earmarked for another purpose, e.g. redemptions.
		// The same holds true for IBC transfers.
		// Given these assumptions, the order of staking / transfers is not important, because stride deposit
		// records always accurately reflect the state of the controller / host chain by the next epoch.
		// Put another way, all outstanding ICA calls / IBC transfers must be settled on the controller
		// chain before the next epoch begins.
		delegationInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegationInterval == 0 {
			k.StakeExistingDepositsOnHostZones(ctx, epochNumber, depositRecords)
		}
	
		// TODO(TEST-88): Close this ticket
		// UNCOMMENT
		// reinvestInterval := int64(k.GetParam(ctx, types.KeyReinvestInterval))
		// if epochNumber%reinvestInterval == 0 {
		// 	k.ProcessRewardsInterval(ctx)
		// }
	}

}
