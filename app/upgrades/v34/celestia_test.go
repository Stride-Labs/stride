package v34_test

import (
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	icacallbackstypes "github.com/Stride-Labs/stride/v34/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

const (
	celestiaConnectionId        = "connection-0"
	celestiaDelegationChannelId = v34.CelestiaDelegationChannelId
	celestiaDeadChannelId       = "channel-2"
	celestiaDelegationIca       = "celestia1delegationica"
	celestiaHostDenom           = "utia"
	celestiaDepositEpoch        = uint64(1400)
)

// celestiaState is a comparable snapshot of everything ReconcileCelestia may write
type celestiaState struct {
	validatorDelegations map[string]sdkmath.Int
	changesInProgress    map[string]int64
	totalDelegations     sdkmath.Int
	records              map[uint64]recordstypes.DepositRecord
	callbackKeys         map[string]bool
}

// celestiaSeed describes the records seeded by one of the scenario helpers below
type celestiaSeed struct {
	phantomAmount sdkmath.Int
	deleted       []uint64                    // records ReconcileCelestia must delete
	untouched     []uint64                    // records ReconcileCelestia must leave exactly as seeded
	shrunk        map[uint64]sdkmath.Int      // queue record id -> expected amount after the shrink
	replaced      *recordstypes.DepositRecord // the in-progress record replaced by a fresh queue record, if any
	leftover      sdkmath.Int                 // amount of that fresh queue record
}

func celestiaDelegationPortId() string {
	owner := stakeibctypes.FormatHostZoneICAOwner(v34.CelestiaChainId, stakeibctypes.ICAAccountType_DELEGATION)
	portId, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		panic(err)
	}
	return portId
}

func celestiaPhantomAmount() sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, entry := range v34.CelestiaDelegationDeltas {
		total = total.Add(entry.Delta)
	}
	return total
}

// Seeds the celestia host zone with every validator in the delta table at a round tracked
// delegation, and registers the active DELEGATION ICA channel
func (s *UpgradeTestSuite) setupCelestiaHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for i, entry := range v34.CelestiaDelegationDeltas {
		delegation := sdkmath.NewInt(int64(1_000_000_000 + i*1000))
		validators = append(validators, &stakeibctypes.Validator{
			Name:       entry.Name,
			Address:    entry.Address,
			Delegation: delegation,
		})
		tracked[entry.Address] = delegation
		trackedTotal = trackedTotal.Add(delegation)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:              v34.CelestiaChainId,
		ConnectionId:         celestiaConnectionId,
		HostDenom:            celestiaHostDenom,
		DelegationIcaAddress: celestiaDelegationIca,
		TotalDelegations:     trackedTotal,
		Validators:           validators,
	})

	owner := stakeibctypes.FormatHostZoneICAOwner(v34.CelestiaChainId, stakeibctypes.ICAAccountType_DELEGATION)
	s.MockICAChannel(celestiaConnectionId, celestiaDelegationChannelId, owner, celestiaDelegationIca)
	return tracked, trackedTotal
}

func (s *UpgradeTestSuite) setCelestiaChangesInProgress(address string, count int64) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	_, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, address)
	s.Require().True(found)
	hostZone.Validators[index].DelegationChangesInProgress = count
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
}

func (s *UpgradeTestSuite) seedDepositRecord(chainId string, status recordstypes.DepositRecord_Status, amount sdkmath.Int) uint64 {
	return s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, recordstypes.DepositRecord{
		Amount:             amount,
		Denom:              celestiaHostDenom,
		HostZoneId:         chainId,
		Status:             status,
		Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
		DepositEpochNumber: celestiaDepositEpoch,
	})
}

// Stores a delegate callback the way SubmitTxs does and returns its key
func (s *UpgradeTestSuite) seedDelegateCallback(
	portId string,
	channelId string,
	sequence uint64,
	recordId uint64,
	validators ...string,
) string {
	splits := []*stakeibctypes.SplitDelegation{}
	for _, validator := range validators {
		splits = append(splits, &stakeibctypes.SplitDelegation{Validator: validator, Amount: sdkmath.NewInt(100)})
	}
	args, err := proto.Marshal(&stakeibctypes.DelegateCallback{
		HostZoneId:       v34.CelestiaChainId,
		DepositRecordId:  recordId,
		SplitDelegations: splits,
	})
	s.Require().NoError(err)

	key := icacallbackstypes.PacketID(portId, channelId, sequence)
	s.App.IcacallbacksKeeper.SetCallbackData(s.Ctx, icacallbackstypes.CallbackData{
		CallbackKey:  key,
		PortId:       portId,
		ChannelId:    channelId,
		Sequence:     sequence,
		CallbackId:   stakeibckeeper.ICACallbackID_Delegate,
		CallbackArgs: args,
	})
	return key
}

// Queue records cover the phantom amount by themselves: two whole deletes and one shrink, with an
// earlier in-progress record (and its callback) left alone because queue records go first
func (s *UpgradeTestSuite) seedCelestiaQueueShrinkScenario() celestiaSeed {
	phantom := celestiaPhantomAmount()
	first := phantom.QuoRaw(4)
	second := phantom.QuoRaw(2)
	remainder := phantom.Sub(first).Sub(second)

	otherChain := s.seedDepositRecord(v34.InjectiveChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, sdkmath.NewInt(777))
	transfer := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_TRANSFER_QUEUE, sdkmath.NewInt(999))
	queue1 := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, first)
	inProgress := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, phantom)
	queue2 := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, second)
	queue3 := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, phantom)

	s.seedDelegateCallback(celestiaDelegationPortId(), celestiaDelegationChannelId, 1, inProgress, v34.CelestiaDelegationDeltas[0].Address)
	s.setCelestiaChangesInProgress(v34.CelestiaDelegationDeltas[0].Address, 1)

	return celestiaSeed{
		phantomAmount: phantom,
		deleted:       []uint64{queue1, queue2},
		untouched:     []uint64{otherChain, transfer, inProgress},
		shrunk:        map[uint64]sdkmath.Int{queue3: phantom.Sub(remainder)},
	}
}

// Queue records fall short so the remainder lands on in-progress records: one is deleted whole
// and the next is replaced by a fresh queue record for its leftover. Callbacks: the deleted
// records have entries on the active channel (decremented) and a dead channel (just removed);
// an untouched record, a non-delegate callback and another host zone's callback all survive.
func (s *UpgradeTestSuite) seedCelestiaInProgressScenario() (celestiaSeed, celestiaCallbackExpectations) {
	phantom := celestiaPhantomAmount()
	portId := celestiaDelegationPortId()
	valA := v34.CelestiaDelegationDeltas[0].Address
	valB := v34.CelestiaDelegationDeltas[1].Address
	valC := v34.CelestiaDelegationDeltas[2].Address

	queue := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, phantom.QuoRaw(2))
	deletedInProgress := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, phantom.QuoRaw(4))
	replacedInProgress := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, phantom)
	laterInProgress := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, sdkmath.NewInt(5))
	replaced, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, replacedInProgress)
	s.Require().True(found)

	// valB sits at zero to prove the decrement guard; valA and valC are decremented once each
	s.setCelestiaChangesInProgress(valA, 2)
	s.setCelestiaChangesInProgress(valB, 0)
	s.setCelestiaChangesInProgress(valC, 1)

	removed := []string{
		s.seedDelegateCallback(portId, celestiaDelegationChannelId, 1, deletedInProgress, valA, valB),
		s.seedDelegateCallback(portId, celestiaDeadChannelId, 1, deletedInProgress, valA),
		s.seedDelegateCallback(portId, celestiaDelegationChannelId, 2, replacedInProgress, valC),
	}
	kept := []string{
		s.seedDelegateCallback(portId, celestiaDelegationChannelId, 3, laterInProgress, valA),
		s.seedDelegateCallback("icacontroller-injective-1.DELEGATION", celestiaDelegationChannelId, 1, deletedInProgress, valA),
	}
	reinvestKey := icacallbackstypes.PacketID(portId, celestiaDelegationChannelId, 4)
	s.App.IcacallbacksKeeper.SetCallbackData(s.Ctx, icacallbackstypes.CallbackData{
		CallbackKey: reinvestKey,
		PortId:      portId,
		ChannelId:   celestiaDelegationChannelId,
		Sequence:    4,
		CallbackId:  stakeibckeeper.ICACallbackID_Reinvest,
	})
	kept = append(kept, reinvestKey)

	seed := celestiaSeed{
		phantomAmount: phantom,
		deleted:       []uint64{queue, deletedInProgress, replacedInProgress},
		untouched:     []uint64{laterInProgress},
		shrunk:        map[uint64]sdkmath.Int{},
		replaced:      &replaced,
		leftover:      phantom.QuoRaw(2).Add(phantom.QuoRaw(4)), // record amount P minus the remaining P/4
	}
	expectations := celestiaCallbackExpectations{
		removedKeys:       removed,
		keptKeys:          kept,
		changesInProgress: map[string]int64{valA: 1, valB: 0, valC: 0},
	}
	return seed, expectations
}

type celestiaCallbackExpectations struct {
	removedKeys       []string
	keptKeys          []string
	changesInProgress map[string]int64
}

func (s *UpgradeTestSuite) snapshotCelestiaState() celestiaState {
	state := celestiaState{
		validatorDelegations: map[string]sdkmath.Int{},
		changesInProgress:    map[string]int64{},
		records:              map[uint64]recordstypes.DepositRecord{},
		callbackKeys:         map[string]bool{},
	}
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	if found {
		state.totalDelegations = hostZone.TotalDelegations
		for _, validator := range hostZone.Validators {
			state.validatorDelegations[validator.Address] = validator.Delegation
			state.changesInProgress[validator.Address] = validator.DelegationChangesInProgress
		}
	}
	for _, record := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		state.records[record.Id] = record
	}
	for _, callback := range s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx) {
		state.callbackKeys[callback.CallbackKey] = true
	}
	return state
}

func (s *UpgradeTestSuite) assertCelestiaStateUnchanged(before celestiaState) {
	after := s.snapshotCelestiaState()
	s.Require().Equal(len(before.validatorDelegations), len(after.validatorDelegations), "validator count")
	for address, delegation := range before.validatorDelegations {
		s.Require().Equal(delegation.String(), after.validatorDelegations[address].String(), "%s delegation must be untouched", address)
		s.Require().Equal(before.changesInProgress[address], after.changesInProgress[address], "%s changes in progress must be untouched", address)
	}
	if !before.totalDelegations.IsNil() {
		s.Require().Equal(before.totalDelegations.String(), after.totalDelegations.String(), "TotalDelegations must be untouched")
	}
	s.Require().Equal(len(before.records), len(after.records), "record count")
	for id, record := range before.records {
		s.Require().Equal(record.Amount.String(), after.records[id].Amount.String(), "record %d amount must be untouched", id)
		s.Require().Equal(record.Status, after.records[id].Status, "record %d status must be untouched", id)
	}
	s.Require().Equal(before.callbackKeys, after.callbackKeys, "callbacks must be untouched")
}

// The redemption rate numerator components: deposit records awaiting delegation plus what is
// already delegated. A pure bucket move leaves the sum unchanged to the utia. Shared with the
// mainnet export suite, hence a free function over the common test helper.
func celestiaRateNumerator(s *apptesting.AppTestHelper) sdkmath.LegacyDec {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	undelegated := s.App.StakeibcKeeper.GetUndelegatedBalance(v34.CelestiaChainId, s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx))
	return undelegated.Add(sdkmath.LegacyNewDecFromInt(hostZone.TotalDelegations))
}

// Asserts the validator half (every delta applied, TotalDelegations up by the sum) and the record
// half (exactly the phantom amount removed, as the seed describes) of a completed reconciliation
func (s *UpgradeTestSuite) assertCelestiaReconciled(seed celestiaSeed, tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int, before celestiaState) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	for _, entry := range v34.CelestiaDelegationDeltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta).String(), validator.Delegation.String(),
			"%s delegation should move by its delta", entry.Name)
	}
	s.Require().Equal(trackedTotal.Add(seed.phantomAmount).String(), hostZone.TotalDelegations.String(),
		"TotalDelegations up by the phantom amount")
	s.Require().Len(hostZone.Validators, len(v34.CelestiaDelegationDeltas), "no validators should be added")

	for _, id := range seed.deleted {
		_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, id)
		s.Require().False(found, "record %d should be deleted", id)
	}
	for _, id := range seed.untouched {
		record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, id)
		s.Require().True(found, "record %d should survive", id)
		s.Require().Equal(before.records[id].Amount.String(), record.Amount.String(), "record %d amount", id)
		s.Require().Equal(before.records[id].Status, record.Status, "record %d status", id)
	}
	for id, expectedAmount := range seed.shrunk {
		record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, id)
		s.Require().True(found, "shrunk record %d should survive", id)
		s.Require().Equal(expectedAmount.String(), record.Amount.String(), "record %d shrunk amount", id)
		s.Require().Equal(recordstypes.DepositRecord_DELEGATION_QUEUE, record.Status, "only queue records are shrunk")
	}

	// A replaced in-progress record leaves a brand new queue record for its leftover
	if seed.replaced != nil {
		replacement, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, s.App.RecordsKeeper.GetDepositRecordCount(s.Ctx)-1)
		s.Require().True(found, "replacement record should be appended")
		s.Require().Greater(replacement.Id, seed.replaced.Id, "replacement is a new record")
		s.Require().Equal(seed.leftover.String(), replacement.Amount.String(), "replacement amount is the leftover")
		s.Require().Equal(recordstypes.DepositRecord_DELEGATION_QUEUE, replacement.Status)
		s.Require().Equal(seed.replaced.HostZoneId, replacement.HostZoneId)
		s.Require().Equal(seed.replaced.Denom, replacement.Denom)
		s.Require().Equal(seed.replaced.Source, replacement.Source)
		s.Require().Equal(seed.replaced.DepositEpochNumber, replacement.DepositEpochNumber)
		s.Require().Zero(replacement.DelegationTxsInProgress)
	}

	// Exactly the phantom amount left the records in total
	beforeRecordsTotal := sdkmath.ZeroInt()
	for _, record := range before.records {
		if record.HostZoneId == v34.CelestiaChainId {
			beforeRecordsTotal = beforeRecordsTotal.Add(record.Amount)
		}
	}
	afterRecordsTotal := sdkmath.ZeroInt()
	for _, record := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		if record.HostZoneId == v34.CelestiaChainId {
			afterRecordsTotal = afterRecordsTotal.Add(record.Amount)
		}
	}
	s.Require().Equal(beforeRecordsTotal.Sub(seed.phantomAmount).String(), afterRecordsTotal.String(),
		"exactly the phantom amount is removed from the records")
}

func (s *UpgradeTestSuite) assertCelestiaCallbacks(expectations celestiaCallbackExpectations) {
	for _, key := range expectations.removedKeys {
		_, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, key)
		s.Require().False(found, "callback %s should be removed", key)
	}
	for _, key := range expectations.keptKeys {
		_, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, key)
		s.Require().True(found, "callback %s should be kept", key)
	}

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	for address, expected := range expectations.changesInProgress {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, address)
		s.Require().True(found)
		s.Require().Equal(expected, validator.DelegationChangesInProgress, "%s delegation changes in progress", address)
	}
}

func (s *UpgradeTestSuite) TestReconcileCelestia_QueueRecordsShrunk() {
	tracked, trackedTotal := s.setupCelestiaHostZone()
	seed := s.seedCelestiaQueueShrinkScenario()
	before := s.snapshotCelestiaState()
	numeratorBefore := celestiaRateNumerator(&s.AppTestHelper)

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().True(applied)

	s.assertCelestiaReconciled(seed, tracked, trackedTotal, before)
	s.Require().Equal(numeratorBefore.String(), celestiaRateNumerator(&s.AppTestHelper).String(), "redemption rate components unchanged")

	// The in-progress record was never reached, so its callback and validator counter are untouched
	s.Require().Equal(before.callbackKeys, s.snapshotCelestiaState().callbackKeys, "no callbacks removed")
	s.assertCelestiaCallbacks(celestiaCallbackExpectations{
		changesInProgress: map[string]int64{v34.CelestiaDelegationDeltas[0].Address: 1},
	})
}

func (s *UpgradeTestSuite) TestReconcileCelestia_InProgressRecordsDeleted() {
	tracked, trackedTotal := s.setupCelestiaHostZone()
	seed, expectations := s.seedCelestiaInProgressScenario()
	before := s.snapshotCelestiaState()
	numeratorBefore := celestiaRateNumerator(&s.AppTestHelper)

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().True(applied)

	s.assertCelestiaReconciled(seed, tracked, trackedTotal, before)
	s.assertCelestiaCallbacks(expectations)
	s.Require().Equal(numeratorBefore.String(), celestiaRateNumerator(&s.AppTestHelper).String(), "redemption rate components unchanged")

	// No in-progress record survives with a reduced amount
	for _, record := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		if record.Status != recordstypes.DepositRecord_DELEGATION_IN_PROGRESS {
			continue
		}
		s.Require().Equal(before.records[record.Id].Amount.String(), record.Amount.String(),
			"in-progress record %d must never be shrunk", record.Id)
	}
}

// On mainnet the open records total roughly the phantom amount, so once the first run has retired
// it the survivors cannot cover a second application and the check in step 3 skips everything
func (s *UpgradeTestSuite) TestReconcileCelestia_Idempotent() {
	tracked, trackedTotal := s.setupCelestiaHostZone()
	phantom := celestiaPhantomAmount()
	first := phantom.QuoRaw(4)
	survivor := sdkmath.NewInt(10)
	queue1 := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, first)
	queue2 := s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, phantom.Sub(first).Add(survivor))
	seed := celestiaSeed{
		phantomAmount: phantom,
		deleted:       []uint64{queue1},
		shrunk:        map[uint64]sdkmath.Int{queue2: survivor},
	}
	before := s.snapshotCelestiaState()

	s.Require().True(v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper))
	afterFirst := s.snapshotCelestiaState()

	s.Require().False(v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper),
		"second run must be a no-op")
	s.assertCelestiaStateUnchanged(afterFirst)
	s.assertCelestiaReconciled(seed, tracked, trackedTotal, before)
}

func (s *UpgradeTestSuite) TestReconcileCelestia_MissingHostZone() {
	s.seedCelestiaQueueShrinkScenarioWithoutCallbacks()
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied)

	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().False(found, "the reconciliation should not create the host zone")
	s.assertCelestiaStateUnchanged(before)
}

// Records only, for the skip tests that have no host zone to hang a callback counter on
func (s *UpgradeTestSuite) seedCelestiaQueueShrinkScenarioWithoutCallbacks() {
	phantom := celestiaPhantomAmount()
	s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, phantom)
	s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, phantom)
}

func (s *UpgradeTestSuite) TestReconcileCelestia_MissingValidatorSkipsAll() {
	_, trackedTotal := s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	removed := hostZone.Validators[len(hostZone.Validators)-1]
	hostZone.Validators = hostZone.Validators[:len(hostZone.Validators)-1]
	hostZone.TotalDelegations = trackedTotal.Sub(removed.Delegation)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied)
	s.assertCelestiaStateUnchanged(before)
}

// The table is all positive, so a negative outcome can only be simulated by swapping one entry's
// delta; the global is restored on cleanup so the mainnet export gate never sees the test value
func (s *UpgradeTestSuite) TestReconcileCelestia_NegativeResultSkipsAll() {
	snapshot := append([]v34.DelegationDelta{}, v34.CelestiaDelegationDeltas...)
	s.T().Cleanup(func() { copy(v34.CelestiaDelegationDeltas, snapshot) })

	tracked, _ := s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()
	target := &v34.CelestiaDelegationDeltas[len(v34.CelestiaDelegationDeltas)-1]
	target.Delta = tracked[target.Address].AddRaw(1).Neg()
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied)
	s.assertCelestiaStateUnchanged(before)
}

func (s *UpgradeTestSuite) TestReconcileCelestia_RecordsCannotCoverSkipsAll() {
	s.setupCelestiaHostZone()
	phantom := celestiaPhantomAmount()
	s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_QUEUE, phantom.QuoRaw(2))
	s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, phantom.Sub(phantom.QuoRaw(2)).SubRaw(1))
	s.seedDepositRecord(v34.CelestiaChainId, recordstypes.DepositRecord_TRANSFER_QUEUE, phantom) // never counts
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied)
	s.assertCelestiaStateUnchanged(before)
}

// seedMalformedDelegateCallback stores unparseable bytes as a delegate callback's args, the way
// a schema mismatch or data corruption might, so tests can assert ReconcileCelestia's pre-scan
// catches it before any write instead of surfacing later as a failed ack unmarshal
func (s *UpgradeTestSuite) seedMalformedDelegateCallback(portId, channelId string, sequence uint64) string {
	key := icacallbackstypes.PacketID(portId, channelId, sequence)
	s.App.IcacallbacksKeeper.SetCallbackData(s.Ctx, icacallbackstypes.CallbackData{
		CallbackKey:  key,
		PortId:       portId,
		ChannelId:    channelId,
		Sequence:     sequence,
		CallbackId:   stakeibckeeper.ICACallbackID_Delegate,
		CallbackArgs: []byte{0xff, 0xff, 0xff, 0xff},
	})
	return key
}

func (s *UpgradeTestSuite) TestReconcileCelestia_RestoredChannelSkipsAll() {
	s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()

	// A restore after the measurement moves the active channel to a fresh id
	owner := stakeibctypes.FormatHostZoneICAOwner(v34.CelestiaChainId, stakeibctypes.ICAAccountType_DELEGATION)
	s.MockICAChannel(celestiaConnectionId, "channel-999", owner, celestiaDelegationIca)
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied, "an active channel other than the pinned one must skip the whole reconciliation")
	s.assertCelestiaStateUnchanged(before)
}

func (s *UpgradeTestSuite) TestReconcileCelestia_PendingPacketSkipsAll() {
	s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()

	// An unacknowledged delegate on the pinned channel may still execute on the host
	s.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(s.Ctx, celestiaDelegationPortId(), celestiaDelegationChannelId, 7, []byte("commitment"))
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied, "a pending packet on the delegation channel must skip the whole reconciliation")
	s.assertCelestiaStateUnchanged(before)
}

// The pinned channel may be CLOSED (its timeouts relayed) at the upgrade block: that is the
// quietest possible state and must still apply
func (s *UpgradeTestSuite) TestReconcileCelestia_ClosedQuietChannelApplies() {
	s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()
	s.UpdateChannelState(celestiaDelegationPortId(), celestiaDelegationChannelId, channeltypes.CLOSED)

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().True(applied, "a closed pinned channel with no pending packets must still apply")
}

// A delegate callback on the port that cannot be unmarshalled must skip the whole reconciliation
// before any write, rather than deleting the deposit record it (unreadably) references and
// leaving the callback behind to fail unmarshalling on its eventual ack instead of being a no-op
func (s *UpgradeTestSuite) TestReconcileCelestia_MalformedCallbackSkipsAll() {
	s.setupCelestiaHostZone()
	s.seedCelestiaQueueShrinkScenario()
	malformedKey := s.seedMalformedDelegateCallback(celestiaDelegationPortId(), celestiaDelegationChannelId, 99)
	before := s.snapshotCelestiaState()

	applied := v34.ReconcileCelestia(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper, s.App.IcacallbacksKeeper)
	s.Require().False(applied, "a malformed delegate callback must skip the whole reconciliation")
	s.assertCelestiaStateUnchanged(before)

	_, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, malformedKey)
	s.Require().True(found, "the malformed callback itself must also be left untouched")
}

func (s *UpgradeTestSuite) setupStaketiaHostZone(remainingDelegatedBalance sdkmath.Int) {
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, staketiatypes.HostZone{
		ChainId:                   staketiatypes.CelestiaChainId,
		NativeTokenDenom:          celestiaHostDenom,
		RemainingDelegatedBalance: remainingDelegatedBalance,
	})
}

func (s *UpgradeTestSuite) TestAdjustStaketiaRemainingDelegatedBalance() {
	_, trackedTotal := s.setupCelestiaHostZone()
	initial := sdkmath.NewInt(200_000_000_000)
	s.setupStaketiaHostZone(initial)
	s.Require().True(v34.StaketiaRemainingDelegatedBalanceDelta.IsNegative(), "test relies on a negative constant")

	applied := v34.AdjustStaketiaRemainingDelegatedBalance(s.Ctx, s.App.StaketiaKeeper, v34.StaketiaRemainingDelegatedBalanceDelta)
	s.Require().True(applied)

	staketiaHostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(initial.Add(v34.StaketiaRemainingDelegatedBalanceDelta).String(), staketiaHostZone.RemainingDelegatedBalance.String())

	// Unlike MsgAdjustDelegatedBalance, stakeibc is not mirrored
	stakeibcHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	s.Require().Equal(trackedTotal.String(), stakeibcHostZone.TotalDelegations.String(), "stakeibc TotalDelegations untouched")
}

func (s *UpgradeTestSuite) TestAdjustStaketiaRemainingDelegatedBalance_MissingHostZone() {
	applied := v34.AdjustStaketiaRemainingDelegatedBalance(s.Ctx, s.App.StaketiaKeeper, v34.StaketiaRemainingDelegatedBalanceDelta)
	s.Require().False(applied)

	_, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().Error(err, "the adjustment should not create the host zone")
}

func (s *UpgradeTestSuite) TestAdjustStaketiaRemainingDelegatedBalance_NegativeResultSkips() {
	initial := sdkmath.NewInt(1)
	s.setupStaketiaHostZone(initial)

	applied := v34.AdjustStaketiaRemainingDelegatedBalance(s.Ctx, s.App.StaketiaKeeper, sdkmath.NewInt(-2))
	s.Require().False(applied)

	staketiaHostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(initial.String(), staketiaHostZone.RemainingDelegatedBalance.String(), "balance untouched")
}
