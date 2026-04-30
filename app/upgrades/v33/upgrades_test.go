package v33_test

import (
	"encoding/hex"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
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

// populateValidatorMonikers fills v33.ValidatorMonikers for each seeded ICS
// validator so SnapshotValidatorsFromICS has monikers to record. Cleaned up
// after the test run.
func (s *UpgradeTestSuite) populateValidatorMonikers() {
	vals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	for i, v := range vals {
		v33.ValidatorMonikers[hex.EncodeToString(v.Address)] = fmtMoniker(i)
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
