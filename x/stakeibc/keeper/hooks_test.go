package keeper_test

import (
	"os"
	"strings"
	"time"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v35/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v35/x/mint/types"
	recordtypes "github.com/Stride-Labs/stride/v35/x/records/types"
	"github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// Every test in this file runs the real BeforeEpochStart on non-halted host zones and asserts
// two things per epoch type: the flows the wind-down keeps still do their work (an ICA or a
// transfer packet was actually sent and the record moved), and the flows it deletes leave no
// trace (the rate, the record set and the ICA channel sequence are exactly what the kept
// flows alone produce). Nothing here sets HostZone.Halted.

// epochInfoForHook builds the EpochInfo the epochs module hands to BeforeEpochStart. The
// current epoch starts at the block time so the tracker's next start time is in the future,
// which the ICA timeouts of the kept flows need.
func (s *KeeperTestSuite) epochInfoForHook(identifier string, epochNumber int64) epochstypes.EpochInfo {
	return epochstypes.EpochInfo{
		Identifier:            identifier,
		Duration:              time.Hour,
		CurrentEpoch:          epochNumber,
		CurrentEpochStartTime: s.Ctx.BlockTime(),
		EpochCountingStarted:  true,
	}
}

// The day epoch keeps three flows: submitting queued redemption records as undelegate ICAs,
// submitting a pending (upgrade-queued) undelegation, and deleting fully claimed epoch
// records. It no longer creates an epoch unbonding record for the new epoch.
//
// Fixture: SetupInitiateAllHostZoneUnbondings seeds GAIA (unbonding period 14 → every 3rd
// day epoch) and OSMO (21 → every 4th) with a queued record at epoch 5 and a GAIA.DELEGATION
// ICA channel, and a day tracker at epoch 12, on which both zones unbond. On top of that a
// third zone, JUNO, carries a pending undelegation and an unbonding period of 28 (every 5th
// day epoch, and 12 % 5 != 0), so SubmitPendingUndelegations submits it on this epoch instead
// of deferring, and a fully claimed epoch-2 record exists for the cleanup to delete.
//
// ICA routing in this suite: SubmitTxs derives the ICA owner from the chain id of the
// connection's tendermint client (GetChainIdFromConnectionId in ibc.go), not from the host
// zone, and all three zones sit on connection-0 whose client chain id is GAIA. So every
// undelegate ICA, OSMO's and JUNO's included, is sent on the single GAIA.DELEGATION channel
// (TestInitiateAllHostZoneUnbondings_Successful relies on the same thing: it opens only that
// channel and still sees OSMO's undelegation event). One channel, three ICAs expected.
func (s *KeeperTestSuite) TestBeforeEpochStart_DayEpoch_KeptFlowsRunNoNewRecord() {
	s.SetupInitiateAllHostZoneUnbondings()
	gaiaOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	gaiaPortId, err := icatypes.NewControllerPortID(gaiaOwner)
	s.Require().NoError(err, "GAIA delegation port id")
	gaiaChannelId, found := s.App.ICAControllerKeeper.GetOpenActiveChannel(s.Ctx, ibctesting.FirstConnectionID, gaiaPortId)
	s.Require().True(found, "GAIA delegation channel open")

	// A third zone with a pending undelegation that submits on this epoch (12 % 5 != 0).
	// Same shape as SetupSubmitPendingUndelegations: only val1 has capacity for the 200.
	// No channel of its own: its ICA goes out on the GAIA delegation channel (see above)
	junoChainId := "JUNO"
	pendingAmount := sdkmath.NewInt(200)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              junoChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            "ujuno",
		DelegationIcaAddress: "juno_DELEGATION",
		Validators: []*types.Validator{
			{Address: "juno_val1", Weight: 50, Delegation: sdkmath.NewInt(600)},
			{Address: "juno_val2", Weight: 50, Delegation: sdkmath.NewInt(400)},
		},
		TotalDelegations:    sdkmath.NewInt(1000),
		RedemptionRate:      sdkmath.LegacyOneDec(),
		MaxMessagesPerIcaTx: 32,
		UnbondingPeriod:     28,
	})
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, junoChainId, pendingAmount)

	// A fully claimed old epoch record, which the cleanup deletes
	claimedEpoch := uint64(2)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: claimedEpoch,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{HostZoneId: HostChainId, ClaimableNativeTokens: sdkmath.ZeroInt(), Status: recordtypes.HostZoneUnbonding_CLAIMABLE},
		},
	})

	dayEpoch := int64(12)
	_, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(dayEpoch))
	s.Require().False(found, "no epoch unbonding record for the new epoch before the hook")
	gaiaStartSequence := s.MustGetNextSequenceNumber(gaiaPortId, gaiaChannelId)

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.DAY_EPOCH, dayEpoch))

	// Kept: InitiateAllHostZoneUnbondings submitted GAIA's and OSMO's queued records as one
	// undelegate ICA each (batch size 32), and SubmitPendingUndelegations submitted JUNO's
	// pending amount as a third. All three land on the GAIA delegation channel (routing note
	// above), so the sequence advanced by exactly 3: no fourth ICA of any kind went out
	s.Require().Equal(gaiaStartSequence+3, s.MustGetNextSequenceNumber(gaiaPortId, gaiaChannelId),
		"exactly three undelegate ICAs (GAIA record, OSMO record, JUNO pending) on the GAIA delegation channel")
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
	gaiaUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, 5, HostChainId)
	s.Require().True(found, "GAIA epoch-5 unbonding record found")
	s.Require().Equal(recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS, gaiaUnbonding.Status, "GAIA record submitted")
	gaiaHostZone := s.MustGetHostZone(HostChainId)
	for _, validator := range gaiaHostZone.Validators {
		s.Require().Equal(int64(1), validator.DelegationChangesInProgress, "GAIA validator %s flagged by the batch", validator.Address)
	}

	// Kept: SubmitPendingUndelegations marked JUNO's batch in flight (its ICA is the third of
	// the three counted above); the amount stays stored until the ack
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, junoChainId), "JUNO batch in flight")
	stillPending, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, junoChainId)
	s.Require().True(found, "JUNO pending undelegation kept until the ack")
	s.Require().Equal(pendingAmount, stillPending, "JUNO pending amount unchanged")
	junoHostZone := s.MustGetHostZone(junoChainId)
	s.Require().Equal(int64(1), junoHostZone.Validators[0].DelegationChangesInProgress, "juno_val1 flagged")
	s.Require().Zero(junoHostZone.Validators[1].DelegationChangesInProgress, "juno_val2 had no capacity, not flagged")

	// Kept: CleanupEpochUnbondingRecords deleted the fully claimed record and nothing else
	_, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, claimedEpoch)
	s.Require().False(found, "fully claimed epoch record deleted")
	_, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 5)
	s.Require().True(found, "in-progress epoch record kept")

	// Deleted: CreateEpochUnbondingRecord no longer runs, so the new epoch has no record
	_, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(dayEpoch))
	s.Require().False(found, "no epoch unbonding record created for the new epoch")

	// Kept: the day tracker advanced
	tracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, epochstypes.DAY_EPOCH)
	s.Require().True(found, "day epoch tracker set")
	s.Require().Equal(uint64(dayEpoch), tracker.EpochNumber, "day tracker updated by the hook")
}

// The stride epoch keeps two flows: carrying TRANSFER_QUEUE deposits to the delegation ICA,
// and sweeping completed unbondings to the redemption ICA. Everything that compounded is
// gone, and the redemption rate is the
// proof: with deposit records and rewards present the old formula would have moved it and
// the frozen hook leaves it exactly where it was.
//
// Old rate formula (UpdateRedemptionRateForHostZone):
//
//	(TotalDelegations + DELEGATION_QUEUE deposits + TRANSFER_QUEUE deposits + LSM) / stSupply
//	= (5 + 3 + 4 + 0) / 10 = 1.2, from an initial 1.0
//
// The DELEGATION_QUEUE record is the "accrued rewards" position: the reinvest pipeline's
// withdrawal-balance callback books withdrawn rewards as a DELEGATION_QUEUE deposit record,
// and that is how rewards entered the rate. The old hook would also have delegated it
// (StakeExistingDepositsOnHostZones), which is why the record must be untouched after.
func (s *KeeperTestSuite) TestBeforeEpochStart_StrideEpoch_RateFrozenKeptFlowsRun() {
	// ICA channel (this also opens the transfer channel channel-0 on the same connection)
	delegationOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelId, delegationPortId := s.CreateICAChannel(delegationOwner)
	delegationAddress := s.IcaAddresses[delegationOwner]

	// A real IBC denom for the deposit transfer, and a funded deposit address
	ibcDenomTrace := s.GetIBCDenomTrace(Atom)
	s.App.TransferKeeper.SetDenom(s.Ctx, ibcDenomTrace)
	ibcDenom := ibcDenomTrace.IBCDenom()
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)
	initialDepositBalance := sdkmath.NewInt(15_000)
	s.FundAccount(depositAddress, sdk.NewCoin(ibcDenom, initialDepositBalance))

	initialRate := sdkmath.LegacyNewDec(1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		IbcDenom:             ibcDenom,
		ConnectionId:         ibctesting.FirstConnectionID,
		TransferChannelId:    ibctesting.FirstChannelID,
		DepositAddress:       depositAddress.String(),
		DelegationIcaAddress: delegationAddress,
		WithdrawalIcaAddress: "cosmos_WITHDRAWAL",
		RedemptionIcaAddress: "cosmos_REDEMPTION",
		UnbondingPeriod:      14,
		RedemptionRate:       initialRate,
		TotalDelegations:     sdkmath.NewInt(5),
		Validators: []*types.Validator{
			{Address: ValAddress, Delegation: sdkmath.NewInt(5), Weight: 1, SharesToTokensRate: sdkmath.LegacyOneDec()},
		},
		MaxMessagesPerIcaTx: 10,
	})

	// stToken supply of 10
	s.Require().NoError(s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(StAtom, sdkmath.NewInt(10)))))

	// Deposit records from the previous epoch: one to transfer (kept flow) and the rewards
	// position the deleted delegate flow would have staked
	strideEpoch := int64(7)
	previousEpoch := uint64(strideEpoch - 1)
	rewardsRecord := recordtypes.DepositRecord{
		Id:                 1,
		HostZoneId:         HostChainId,
		Denom:              Atom,
		Amount:             sdkmath.NewInt(3),
		Status:             recordtypes.DepositRecord_DELEGATION_QUEUE,
		DepositEpochNumber: previousEpoch,
	}
	transferRecord := recordtypes.DepositRecord{
		Id:                 2,
		HostZoneId:         HostChainId,
		Denom:              Atom,
		Amount:             sdkmath.NewInt(4),
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
		DepositEpochNumber: previousEpoch,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, rewardsRecord)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, transferRecord)

	// A completed unbonding waiting for the sweep
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{{
			HostZoneId:        HostChainId,
			NativeTokenAmount: sdkmath.NewInt(1_000_000),
			Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
			UnbondingTime:     uint64(s.Ctx.BlockTime().Add(-1 * time.Minute).UnixNano()),
		}},
	})

	// Sanity: the old formula's inputs are all present, so the old hook would have moved the rate
	s.Require().Len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx), 2, "two deposit records seeded")
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "no ICQ before the hook")
	delegationStartSequence := s.MustGetNextSequenceNumber(delegationPortId, delegationChannelId)
	transferStartSequence := s.MustGetNextSequenceNumber(ibctesting.TransferPort, ibctesting.FirstChannelID)

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.STRIDE_EPOCH, strideEpoch))

	// Frozen: the rate is exactly the seeded value, not the 1.2 the old formula computes
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(initialRate, hostZone.RedemptionRate, "redemption rate must not move on a stride epoch")

	// Only the redemption sweep uses the delegation ICA. Reward claims, withdrawal-address
	// changes, and delegating the DELEGATION_QUEUE record must not add packets.
	s.Require().Equal(delegationStartSequence+1, s.MustGetNextSequenceNumber(delegationPortId, delegationChannelId),
		"delegation channel carries only the redemption sweep")
	s.CheckEventValueEmitted(types.EventTypeRedemptionSweep, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeRedemptionSweep, types.AttributeKeySweptAmount, "1000000")
	sweptRecord, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, 1, HostChainId)
	s.Require().True(found, "swept unbonding record found")
	s.Require().Equal(recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS, sweptRecord.Status,
		"sweep ICA in flight (the record becomes CLAIMABLE on its ack)")

	// Kept, on the transfer channel: exactly one packet, the deposit transfer, and the record
	// and the deposit balance moved with it
	s.Require().Equal(transferStartSequence+1, s.MustGetNextSequenceNumber(ibctesting.TransferPort, ibctesting.FirstChannelID),
		"one ICS-20 packet for the deposit transfer")
	transferredRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, transferRecord.Id)
	s.Require().True(found, "transfer deposit record found")
	s.Require().Equal(recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, transferredRecord.Status, "deposit transfer in progress")
	s.Require().Equal(initialDepositBalance.Sub(transferRecord.Amount),
		s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, ibcDenom).Amount, "deposit address debited by the transfer")

	// Deleted: the rewards position was not delegated, no deposit record was created for the
	// epoch, and the reinvest pipeline's withdrawal-balance ICQ was not submitted
	untouchedRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, rewardsRecord.Id)
	s.Require().True(found, "rewards deposit record found")
	s.Require().Equal(rewardsRecord, untouchedRecord, "DELEGATION_QUEUE record untouched: no delegate flow ran")
	s.Require().Len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx), 2, "no deposit record created for the epoch")
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "no withdrawal-balance ICQ submitted")

	// Kept: the stride tracker advanced
	tracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, epochstypes.STRIDE_EPOCH)
	s.Require().True(found, "stride epoch tracker set")
	s.Require().Equal(uint64(strideEpoch), tracker.EpochNumber, "stride tracker updated by the hook")
}

func (s *KeeperTestSuite) TestBeforeEpochStart_StrideEpoch_NoRewardICA() {
	delegationOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelId, delegationPortId := s.CreateICAChannel(delegationOwner)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		ConnectionId:         ibctesting.FirstConnectionID,
		DelegationIcaAddress: s.IcaAddresses[delegationOwner],
		WithdrawalIcaAddress: "cosmos_WITHDRAWAL",
		RedemptionRate:       sdkmath.LegacyOneDec(),
		TotalDelegations:     sdkmath.NewInt(5),
		Validators: []*types.Validator{{
			Address: ValAddress, Delegation: sdkmath.NewInt(5), SharesToTokensRate: sdkmath.LegacyOneDec(),
		}},
	})
	before := s.MustGetHostZone(HostChainId)
	s.Require().False(before.Halted, "active validator remains funded so reward claims would send a packet")
	delegationStartSequence := s.MustGetNextSequenceNumber(delegationPortId, delegationChannelId)
	transferStartSequence := s.MustGetNextSequenceNumber(ibctesting.TransferPort, ibctesting.FirstChannelID)

	// With no deposit or unbonding records, repeated epochs have no reason to send traffic.
	for epochNumber := int64(7); epochNumber <= 9; epochNumber++ {
		s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.STRIDE_EPOCH, epochNumber))
		s.Require().Equal(delegationStartSequence, s.MustGetNextSequenceNumber(delegationPortId, delegationChannelId),
			"epoch %d must not submit a staking reward ICA", epochNumber)
		s.Require().Equal(transferStartSequence, s.MustGetNextSequenceNumber(ibctesting.TransferPort, ibctesting.FirstChannelID))
		s.Require().Equal(before, s.MustGetHostZone(HostChainId), "validator delegation and rate remain unchanged")
		s.Require().Empty(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx))
		s.Require().Empty(s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx))
		s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx))
		tracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, epochstypes.STRIDE_EPOCH)
		s.Require().True(found)
		s.Require().Equal(uint64(epochNumber), tracker.EpochNumber, "tracker advances even without ICA traffic")
	}
}

// The mint epoch used to liquid stake 15% of the reward collector's host fees for the POA
// validators and send the rest to the auction module. Both stop: the balance stays put and no
// stTokens are minted.
func (s *KeeperTestSuite) TestBeforeEpochStart_MintEpoch_RewardCollectorUntouched() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(1000)
	s.FundModuleAccount(types.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.MINT_EPOCH, 3))

	s.checkModuleAccountBalance(types.RewardCollectorName, IbcAtom, rewardAmount)
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StAtom).Int64(),
		"no stTokens minted to the POA validators")
}

// Source-level guard against the deleted calls being re-added. The behavioural tests above
// prove what the hook does; this one only pins the set of call sites by name, so a merge that
// resurrects one fails with the offending identifier in the message instead of a subtle
// balance drift. It is a tripwire, not evidence that the kept calls execute. It reads
// hooks.go relative to the package directory, which is where `go test` runs.
func (s *KeeperTestSuite) TestBeforeEpochStart_SourceHasNoCompoundingCalls() {
	source, err := os.ReadFile("hooks.go")
	s.Require().NoError(err, "hooks.go readable from the package directory")

	start := strings.Index(string(source), "func (k Keeper) BeforeEpochStart(")
	end := strings.Index(string(source), "func (k Keeper) AfterEpochEnd(")
	s.Require().True(start >= 0 && end > start, "BeforeEpochStart and AfterEpochEnd both defined, in that order")
	body := string(source[start:end])

	deletedCalls := []string{
		"k.ClaimAccruedStakingRewards(",
		"k.UpdateRedemptionRates(",
		"k.ReinvestRewards(",
		"k.StakeExistingDepositsOnHostZones(",
		"k.RebalanceAllHostZones(",
		"k.TransferAllRewardTokens(",
		"k.SetWithdrawalAddress(",
		"k.CreateDepositRecordsForEpoch(",
		"k.CreateEpochUnbondingRecord(",
		"k.AuctionOffRewardCollectorBalance(",
	}
	for _, call := range deletedCalls {
		s.Require().NotContains(body, call, "%s must not be called from BeforeEpochStart (spec §6)", call)
	}

	keptCalls := []string{
		"k.UpdateEpochTracker(",
		"k.InitiateAllHostZoneUnbondings(",
		"k.SubmitPendingUndelegations(",
		"k.CleanupEpochUnbondingRecords(",
		"k.TransferExistingDepositsToHostZones(",
		"k.SweepUnbondedTokensAllHostZones(",
	}
	for _, call := range keptCalls {
		s.Require().Contains(body, call, "%s must still be called from BeforeEpochStart (spec §6)", call)
	}

	// The mint-epoch branch had exactly one call and it was deleted, so the branch is gone too
	s.Require().NotContains(body, "MINT_EPOCH", "no mint-epoch branch remains")
}

// Compile-time pin that the keeper functions behind the deleted calls still exist: they are
// dead from the hook but kept until the follow-up cleanup (spec §6, §12), and other tests
// call them directly.
var (
	_ = keeper.Keeper.ClaimAccruedStakingRewards
	_ = keeper.Keeper.UpdateRedemptionRates
	_ = keeper.Keeper.ReinvestRewards
	_ = keeper.Keeper.StakeExistingDepositsOnHostZones
	_ = keeper.Keeper.RebalanceAllHostZones
	_ = keeper.Keeper.TransferAllRewardTokens
	_ = keeper.Keeper.SetWithdrawalAddress
	_ = keeper.Keeper.CreateDepositRecordsForEpoch
	_ = keeper.Keeper.CreateEpochUnbondingRecord
	_ = keeper.Keeper.AuctionOffRewardCollectorBalance
)
