package v33_test

import (
	"encoding/hex"
	"testing"
	"time"

	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v33/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v33/app/upgrades/v33"
	"github.com/Stride-Labs/stride/v33/utils"
	epochstypes "github.com/Stride-Labs/stride/v33/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v33/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v33/x/stakeibc/types"
)

const (
	untouchedChainId         = "stargaze-1"
	untouchedValidatorWeight = uint64(1234)
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper

	// captured pre-upgrade state for post-upgrade comparison
	preUpgradeBondedPool        sdk.Coins
	preUpgradeStakingValidators int
	preUpgradeDelegations       int
	preUpgradeActiveProposalIDs []uint64

	// POA validator count before the upgrade handler runs (1 genesis test validator)
	initialPOAValidatorCount int
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// ----- arrange -----
	s.setupICSValidatorSet(8)
	s.setupGovenatorState(3)
	s.setupConsumerRewardAccounts()
	s.setupActiveGovProposal()
	s.setupHostZoneValidators()

	s.capturePreUpgradeState()
	s.populateValidatorMonikers()

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v33.UpgradeName)

	// ----- assert -----
	s.checkPOAValidatorsMatchICSSnapshot()
	s.checkPOAAdminSet()
	s.checkGovenatorStateUntouched()
	s.checkICSModuleAccountsDrained()
	s.checkActiveGovProposalUnaffected()
	s.checkValidatorSetContinuity()
	s.checkValidatorWeightsUpdated()
	s.checkUntouchedHostZoneUnchanged()
}

// --- assertion helpers ---

func (s *UpgradeTestSuite) checkPOAValidatorsMatchICSSnapshot() {
	ccVals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	s.Require().Len(ccVals, 8, "ICS seed should have 8 validators after seeding")

	poaVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	// Pre-existing test genesis validator + 8 newly seeded ICS validators.
	s.Require().Len(poaVals, s.initialPOAValidatorCount+8,
		"POA should have initialCount + 8 validators after upgrade")

	// Build a lookup of POA validators by their consensus address so we can
	// check each ICS validator is present with the correct pubkey and power.
	poaByConsAddr := make(map[string]poaVal, len(poaVals))
	for _, pv := range poaVals {
		var pk cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(pv.PubKey, &pk))
		poaByConsAddr[sdk.GetConsAddress(pk).String()] = poaVal{pk: pk, power: pv.Power}
	}

	for _, cv := range ccVals {
		consPub, err := cv.ConsPubKey()
		s.Require().NoError(err)
		consAddr := sdk.GetConsAddress(consPub).String()

		pv, ok := poaByConsAddr[consAddr]
		s.Require().True(ok, "ICS validator %s missing from POA after upgrade", consAddr)
		s.Require().Equal(cv.Power, pv.power,
			"power mismatch for validator %s: ICS=%d POA=%d", consAddr, cv.Power, pv.power)
		s.Require().True(pv.pk.Equals(consPub),
			"pubkey mismatch for validator %s", consAddr)
	}
}

func (s *UpgradeTestSuite) checkPOAAdminSet() {
	params, err := s.App.POAKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v33.AdminMultisigAddress, params.Admin)
}

func (s *UpgradeTestSuite) checkGovenatorStateUntouched() {
	bondedAddr := s.App.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName)
	actualBonded := s.App.BankKeeper.GetAllBalances(s.Ctx, bondedAddr)
	s.Require().Equal(s.preUpgradeBondedPool, actualBonded,
		"bonded pool balance should be unchanged after upgrade")

	vals, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(vals, s.preUpgradeStakingValidators,
		"staking validator count should be unchanged after upgrade")

	delegations, err := s.App.StakingKeeper.GetAllDelegations(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(delegations, s.preUpgradeDelegations,
		"delegation count should be unchanged after upgrade")
}

func (s *UpgradeTestSuite) checkICSModuleAccountsDrained() {
	consRedistrAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)
	consToProvAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerToSendToProviderName)
	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistrAddr).IsZero(),
		"cons_redistribute balance should be zero after sweep")
	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consToProvAddr).IsZero(),
		"cons_to_send_to_provider balance should be zero after sweep")
}

func (s *UpgradeTestSuite) checkActiveGovProposalUnaffected() {
	for _, id := range s.preUpgradeActiveProposalIDs {
		prop, err := s.App.GovKeeper.Proposals.Get(s.Ctx, id)
		s.Require().NoError(err, "proposal %d should still be readable after upgrade", id)
		s.Require().NotEqual(govtypes.StatusFailed, prop.Status,
			"proposal %d should not have been failed by the upgrade", id)
	}
}

func (s *UpgradeTestSuite) checkValidatorWeightsUpdated() {
	for chainId, weights := range v33.TargetWeights {
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(found, "host zone %s should exist", chainId)

		actualWeights := map[string]uint64{}
		for _, val := range hostZone.Validators {
			actualWeights[val.Address] = val.Weight
		}

		s.Require().Equal(len(weights), len(actualWeights),
			"%s: validator count mismatch", chainId)

		for _, w := range weights {
			actual, exists := actualWeights[w.Address]
			s.Require().True(exists, "%s: validator %s should exist", chainId, w.Address)
			s.Require().Equal(w.Weight, actual, "%s: weight mismatch for %s", chainId, w.Address)
		}
	}
}

func (s *UpgradeTestSuite) checkUntouchedHostZoneUnchanged() {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, untouchedChainId)
	s.Require().True(found, "untouched host zone %s should exist", untouchedChainId)
	s.Require().Len(hostZone.Validators, 1, "untouched host zone validator count")
	s.Require().Equal(untouchedValidatorWeight, hostZone.Validators[0].Weight,
		"untouched host zone validator weight should be unchanged")
}

func (s *UpgradeTestSuite) checkValidatorSetContinuity() {
	// In a real upgrade, POA's transient store is cleared at each block
	// boundary, so EndBlocker at block N+1 would return [] (nothing queued).
	// In the unit-test environment we stay in the same sdk.Context across the
	// upgrade, so the transient store is never wiped and EndBlocker instead
	// returns the 9 updates queued during InitGenesis (1 genesis + 8 ICS).
	//
	// We therefore assert on _content_ rather than emptiness: the updates must
	// correspond 1-to-1 with the ICS validators we seeded (subset check), plus
	// the pre-existing genesis validator. That confirms InitializePOA wrote the
	// correct set into the transient store — any extra or missing entry would
	// fail the length check below.
	updates, err := s.App.POAKeeper.EndBlocker(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(updates, s.initialPOAValidatorCount+8,
		"POA EndBlocker should have exactly initialCount+8 queued updates (1 genesis + 8 ICS); got %d", len(updates))

	// Build a set of CometBFT pubkey bytes from updates for O(1) lookup.
	updatePubKeySet := make(map[string]int64, len(updates))
	for _, u := range updates {
		pkBytes, err := u.PubKey.Marshal()
		s.Require().NoError(err)
		updatePubKeySet[string(pkBytes)] = u.Power
	}

	// Every ICS validator must appear in the update set with non-zero power.
	ccVals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	for _, cv := range ccVals {
		consPub, err := cv.ConsPubKey()
		s.Require().NoError(err)

		cmtPub, err := cryptocodec.ToCmtProtoPublicKey(consPub)
		s.Require().NoError(err)
		pkBytes, err := cmtPub.Marshal()
		s.Require().NoError(err)

		power, ok := updatePubKeySet[string(pkBytes)]
		s.Require().True(ok,
			"ICS validator %x missing from POA queued updates", cv.Address)
		s.Require().Positive(power,
			"ICS validator %x should have positive power in queued updates", cv.Address)
	}
}

// --- setup helpers ---

// setupICSValidatorSet seeds count CCValidators into the ConsumerKeeper,
// first clearing any validators already present from the genesis state.
// Mirrors the seedConsumerValidators helper in helpers_test.go.
func (s *UpgradeTestSuite) setupICSValidatorSet(count int) {
	for _, existing := range s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx) {
		s.App.ConsumerKeeper.DeleteCCValidator(s.Ctx, existing.Address)
	}

	for i := 0; i < count; i++ {
		privKey := ed25519.GenPrivKeyFromSecret([]byte{byte(i + 1)})
		pubKey := privKey.PubKey()
		addr := pubKey.Address().Bytes()

		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		s.Require().NoError(err)

		ccVal := ccvconsumertypes.CrossChainValidator{
			Address: addr,
			Power:   100,
			Pubkey:  pkAny,
		}
		s.App.ConsumerKeeper.SetCCValidator(s.Ctx, ccVal)
	}
}

// setupGovenatorState creates count staking validators with a minimal
// self-delegation so the bonded-pool and delegation counts are non-zero.
// The upgrade must not touch any of this state.
func (s *UpgradeTestSuite) setupGovenatorState(count int) {
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	delegationAmt := sdkmath.NewInt(1_000_000)

	for i := 0; i < count; i++ {
		privKey := ed25519.GenPrivKeyFromSecret([]byte{byte(i + 100)}) // offset from ICS seeds
		pubKey := privKey.PubKey()

		valAddr := sdk.ValAddress(pubKey.Address())
		delAddr := sdk.AccAddress(pubKey.Address())

		// Create the validator record in Bonded status with tokens and shares.
		validator, err := stakingtypes.NewValidator(
			valAddr.String(),
			pubKey,
			stakingtypes.Description{Moniker: fmtMoniker(i + 20)},
		)
		s.Require().NoError(err)
		validator.Status = stakingtypes.Bonded
		validator.Tokens = delegationAmt
		validator.DelegatorShares = sdkmath.LegacyNewDecFromInt(delegationAmt)

		s.Require().NoError(s.App.StakingKeeper.SetValidator(s.Ctx, validator))
		s.Require().NoError(s.App.StakingKeeper.SetValidatorByConsAddr(s.Ctx, validator))
		s.Require().NoError(s.App.StakingKeeper.SetValidatorByPowerIndex(s.Ctx, validator))

		// Set a self-delegation.
		delegation := stakingtypes.NewDelegation(
			delAddr.String(),
			valAddr.String(),
			sdkmath.LegacyNewDecFromInt(delegationAmt),
		)
		s.Require().NoError(s.App.StakingKeeper.SetDelegation(s.Ctx, delegation))

		// Fund and send to the bonded pool so the bank module stays consistent.
		coin := sdk.NewCoin(bondDenom, delegationAmt)
		s.FundModuleAccount(stakingtypes.BondedPoolName, coin)
	}
}

// setupHostZoneValidators seeds each in-scope host zone with its current
// on-chain validator set (from OldValidators), plus one out-of-scope host zone
// that the upgrade must leave untouched.
func (s *UpgradeTestSuite) setupHostZoneValidators() {
	for chainId, vals := range v33.OldValidators {
		var validators []*stakeibctypes.Validator
		for _, val := range vals {
			validators = append(validators, &stakeibctypes.Validator{
				Name:                      val.Name,
				Address:                   val.Address,
				Weight:                    val.Weight,
				Delegation:                sdkmath.ZeroInt(),
				SlashQueryProgressTracker: sdkmath.ZeroInt(),
				SlashQueryCheckpoint:      sdkmath.ZeroInt(),
				SharesToTokensRate:        sdkmath.LegacyOneDec(),
			})
		}

		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
			ChainId:    chainId,
			Validators: validators,
		})
	}

	// Host zone with no entry in TargetWeights — must not be modified
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: untouchedChainId,
		Validators: []*stakeibctypes.Validator{{
			Name:                      "untouchedval",
			Address:                   "stargazevaloper1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			Weight:                    untouchedValidatorWeight,
			Delegation:                sdkmath.ZeroInt(),
			SlashQueryProgressTracker: sdkmath.ZeroInt(),
			SlashQueryCheckpoint:      sdkmath.ZeroInt(),
			SharesToTokensRate:        sdkmath.LegacyOneDec(),
		}},
	})
}

// setupConsumerRewardAccounts pre-funds the two ICS reward module accounts
// so the sweep step has something to drain.
func (s *UpgradeTestSuite) setupConsumerRewardAccounts() {
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	coin := sdk.NewInt64Coin(bondDenom, 500_000)
	s.FundModuleAccount(ccvconsumertypes.ConsumerRedistributeName, coin)
	s.FundModuleAccount(ccvconsumertypes.ConsumerToSendToProviderName, coin)
}

// setupActiveGovProposal inserts one proposal in VotingPeriod status so we
// can verify the upgrade handler does not accidentally modify its state.
func (s *UpgradeTestSuite) setupActiveGovProposal() {
	proposalID := uint64(1)
	votingEndTime := s.Ctx.BlockTime().Add(time.Hour * 24 * 7)

	submitTime := s.Ctx.BlockTime()
	proposal := govtypes.Proposal{
		Id:               proposalID,
		Status:           govtypes.StatusVotingPeriod,
		VotingEndTime:    &votingEndTime,
		SubmitTime:       &submitTime,
		FinalTallyResult: &govtypes.TallyResult{},
	}
	err := s.App.GovKeeper.SetProposal(s.Ctx, proposal)
	s.Require().NoError(err)

	s.preUpgradeActiveProposalIDs = append(s.preUpgradeActiveProposalIDs, proposalID)
}

// capturePreUpgradeState snapshots bonded pool balance, validator count, and
// delegation count so checkGovenatorStateUntouched can assert nothing changed.
func (s *UpgradeTestSuite) capturePreUpgradeState() {
	// Record how many POA validators exist before the upgrade handler adds the
	// 8 ICS-migrated ones (the genesis bootstrap seeds 1 test validator).
	initialVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.initialPOAValidatorCount = len(initialVals)

	bondedAddr := s.App.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName)
	s.preUpgradeBondedPool = s.App.BankKeeper.GetAllBalances(s.Ctx, bondedAddr)

	vals, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeStakingValidators = len(vals)

	delegations, err := s.App.StakingKeeper.GetAllDelegations(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeDelegations = len(delegations)
}

// populateValidatorMonikers maps each seeded ICS validator to one of the real
// monikers in utils.PoaValidatorSet so SnapshotValidatorsFromICS can complete
// the hex_cons_addr → moniker → operator join. Cleaned up after the test run.
func (s *UpgradeTestSuite) populateValidatorMonikers() {
	vals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	s.Require().Len(vals, len(utils.PoaValidatorSet),
		"test seeds must match utils.PoaValidatorSet length so every validator gets a real moniker")

	for i, v := range vals {
		v33.ValidatorMonikers[hex.EncodeToString(v.Address)] = utils.PoaValidatorSet[i].Moniker
	}
	s.T().Cleanup(func() {
		for _, v := range vals {
			delete(v33.ValidatorMonikers, hex.EncodeToString(v.Address))
		}
	})
}

// poaVal is a small struct to hold a POA validator's pubkey + power, used in
// checkPOAValidatorsMatchICSSnapshot's lookup map.
type poaVal struct {
	pk    cryptotypes.PubKey
	power int64
}

// Tracked (pre-migration) delegations for the three phantom validators; each is > its on-chain target
var (
	cosmostation = v33.OsmosisPhantomDelegations[0].Address
	chorusOne    = v33.OsmosisPhantomDelegations[1].Address
	pryzm        = v33.OsmosisPhantomDelegations[2].Address

	trackedCosmostation = sdkmath.NewInt(258_638_300000)
	trackedChorusOne    = sdkmath.NewInt(231_057_000000)
	trackedPryzm        = sdkmath.NewInt(185_390_600000)
)

func (s *UpgradeTestSuite) setupOsmosisPhantomState() (trackedTotal, expectedReduction sdkmath.Int) {
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochstypes.STRIDE_EPOCH,
		EpochNumber:     42,
	})

	trackedTotal = trackedCosmostation.Add(trackedChorusOne).Add(trackedPryzm)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v33.OsmosisChainId,
		HostDenom:        "uosmo",
		TotalDelegations: trackedTotal,
		Validators: []*stakeibctypes.Validator{
			{Address: cosmostation, Delegation: trackedCosmostation},
			{Address: chorusOne, Delegation: trackedChorusOne},
			{Address: pryzm, Delegation: trackedPryzm},
		},
	})

	// reduction = tracked - on-chain target, summed
	expectedReduction = trackedCosmostation.
		Add(trackedChorusOne).
		Add(trackedPryzm.Sub(v33.OsmosisPhantomDelegations[2].OnChainDelegation))
	return trackedTotal, expectedReduction
}

func (s *UpgradeTestSuite) osmoDelegationQueueRecords() []recordstypes.DepositRecord {
	var out []recordstypes.DepositRecord
	for _, r := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		if r.HostZoneId == v33.OsmosisChainId && r.Status == recordstypes.DepositRecord_DELEGATION_QUEUE {
			out = append(out, r)
		}
	}
	return out
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations() {
	trackedTotal, expectedReduction := s.setupOsmosisPhantomState()

	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "reconciliation should succeed")

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)
	s.Require().True(found)

	// Each phantom validator's Delegation is set to its on-chain target
	delegations := map[string]sdkmath.Int{}
	for _, v := range hostZone.Validators {
		delegations[v.Address] = v.Delegation
	}
	s.Require().Equal(sdkmath.ZeroInt(), delegations[cosmostation], "cosmostation reconciled to on-chain 0")
	s.Require().Equal(sdkmath.ZeroInt(), delegations[chorusOne], "chorus_one reconciled to on-chain 0")
	s.Require().Equal(v33.OsmosisPhantomDelegations[2].OnChainDelegation, delegations[pryzm], "pryzmstakedrop reconciled to on-chain")

	// TotalDelegations reduced by exactly the phantom
	s.Require().Equal(trackedTotal.Sub(expectedReduction), hostZone.TotalDelegations, "TotalDelegations reduced by phantom")

	// A single DELEGATION_QUEUE deposit record credits the reconciled amount
	records := s.osmoDelegationQueueRecords()
	s.Require().Len(records, 1, "exactly one deposit record should be created")
	s.Require().Equal(expectedReduction, records[0].Amount, "deposit record credits the reconciled amount")
	s.Require().Equal("uosmo", records[0].Denom, "deposit record denom")

	// Rate-preservation invariant: the credit exactly offsets the TotalDelegations reduction
	s.Require().Equal(trackedTotal.Sub(hostZone.TotalDelegations), records[0].Amount,
		"credit must exactly offset the TotalDelegations reduction (rate preserved)")

	// Internal consistency: TotalDelegations == sum of validator delegations
	sum := sdkmath.ZeroInt()
	for _, v := range hostZone.Validators {
		sum = sum.Add(v.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_MissingHostZone() {
	// Without the osmosis-1 host zone (e.g. dockernet or a testnet), reconciliation must
	// no-op instead of erroring - an error would fail the upgrade and halt the chain
	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "missing host zone should be skipped, not an error")
	s.Require().Empty(s.osmoDelegationQueueRecords(), "no deposit record should be created")
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_Idempotent() {
	s.setupOsmosisPhantomState()

	s.Require().NoError(v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper))
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)
	firstTotal := hostZone.TotalDelegations

	// Running again must be a no-op: records already match on-chain, so no reduction and no new credit
	s.Require().NoError(v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper))
	hostZoneAfter, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)

	s.Require().Equal(firstTotal, hostZoneAfter.TotalDelegations, "second run must not change TotalDelegations")
	s.Require().Len(s.osmoDelegationQueueRecords(), 1, "second run must not create another deposit record")
}

// A drifted constant (tracked delegation below the expected on-chain amount) must skip that
// validator and still reconcile the rest, rather than halting the upgrade.
func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_SkipsDriftedValidator() {
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochstypes.STRIDE_EPOCH,
		EpochNumber:     42,
	})

	// pryzm tracked below its on-chain constant → drift → must be skipped
	driftedPryzm := v33.OsmosisPhantomDelegations[2].OnChainDelegation.Sub(sdkmath.NewInt(1))
	trackedTotal := trackedCosmostation.Add(trackedChorusOne).Add(driftedPryzm)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v33.OsmosisChainId,
		HostDenom:        "uosmo",
		TotalDelegations: trackedTotal,
		Validators: []*stakeibctypes.Validator{
			{Address: cosmostation, Delegation: trackedCosmostation},
			{Address: chorusOne, Delegation: trackedChorusOne},
			{Address: pryzm, Delegation: driftedPryzm},
		},
	})

	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "drift must be skipped, not an error")

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)
	delegations := map[string]sdkmath.Int{}
	for _, v := range hostZone.Validators {
		delegations[v.Address] = v.Delegation
	}
	s.Require().Equal(sdkmath.ZeroInt(), delegations[cosmostation], "cosmostation still reconciled")
	s.Require().Equal(sdkmath.ZeroInt(), delegations[chorusOne], "chorus_one still reconciled")
	s.Require().Equal(driftedPryzm, delegations[pryzm], "drifted pryzmstakedrop left untouched")

	// Credit reflects only the two reductions that were actually applied
	appliedReduction := trackedCosmostation.Add(trackedChorusOne)
	s.Require().Equal(trackedTotal.Sub(appliedReduction), hostZone.TotalDelegations, "TotalDelegations reduced only by applied reductions")
	records := s.osmoDelegationQueueRecords()
	s.Require().Len(records, 1)
	s.Require().Equal(appliedReduction, records[0].Amount, "credit equals only the applied reductions")
}

// A phantom validator absent from the host zone must be skipped, with the rest still reconciled.
func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_SkipsMissingValidator() {
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochstypes.STRIDE_EPOCH,
		EpochNumber:     42,
	})

	// pryzm is not on the host zone at all → skipped
	trackedTotal := trackedCosmostation.Add(trackedChorusOne)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v33.OsmosisChainId,
		HostDenom:        "uosmo",
		TotalDelegations: trackedTotal,
		Validators: []*stakeibctypes.Validator{
			{Address: cosmostation, Delegation: trackedCosmostation},
			{Address: chorusOne, Delegation: trackedChorusOne},
		},
	})

	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "missing validator must be skipped, not an error")

	records := s.osmoDelegationQueueRecords()
	s.Require().Len(records, 1)
	s.Require().Equal(trackedTotal, records[0].Amount, "credit equals the reductions for the present validators")
}
