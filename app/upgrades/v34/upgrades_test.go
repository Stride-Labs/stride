package v34_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v33/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v33/app/upgrades/v34"
	"github.com/Stride-Labs/stride/v33/utils"
)

// continuingMonikers are the POA validators the upgrade must not touch.
var continuingMonikers = []string{"Polkachu", "L5", "Imperator", "Keplr", "Stakecito", "CryptoCrew"}

var outgoingMonikers = []string{"Citadel.one", "Cosmostation"}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper

	// consensus pubkeys generated for the incoming validators, injected into
	// v34.IncomingValidators by useTestConsensusKeys
	incomingPubKeys map[string]cryptotypes.PubKey

	// consensus pubkeys and operator addresses of the seeded current set
	seededPubKeys   map[string]cryptotypes.PubKey
	seededOperators map[string]string

	preUpgradeTotalPower  int64
	preUpgradeUpdateCount int
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// useTestConsensusKeys swaps generated ed25519 keys into
// v34.IncomingValidators so the synthetic suite doesn't depend on the real
// (mainnet) consensus keys baked into the constants.
//
// Package-level vars are mutated, so this snapshots both globals and
// registers a cleanup to restore them once the test completes. Without this,
// a test that runs before MainnetExportTestSuite — an ordering the Go test
// runner does NOT guarantee, especially under `-shuffle=on` — would leave
// the export suite (the release gate) reading test-filled values instead of
// the real constants and passing vacuously.
func (s *UpgradeTestSuite) useTestConsensusKeys() {
	incomingSnapshot := append([]v34.IncomingValidator{}, v34.IncomingValidators...)
	registrySnapshot := append([]utils.PoaValidator{}, utils.PoaValidatorSet...)
	s.T().Cleanup(func() {
		copy(v34.IncomingValidators, incomingSnapshot)
		copy(utils.PoaValidatorSet, registrySnapshot)
	})

	s.incomingPubKeys = map[string]cryptotypes.PubKey{}
	for i := range v34.IncomingValidators {
		moniker := v34.IncomingValidators[i].Moniker
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("incoming-" + moniker)).PubKey()
		s.incomingPubKeys[moniker] = pubKey
		v34.IncomingValidators[i].ConsPubKeyBase64 = base64.StdEncoding.EncodeToString(pubKey.Bytes())
	}
}

// seedPOASet seeds one POA validator per moniker at ValidatorPower, on top of
// the single power-1 genesis test validator.
func (s *UpgradeTestSuite) seedPOASet(monikers []string) {
	s.seededPubKeys = map[string]cryptotypes.PubKey{}
	s.seededOperators = map[string]string{}
	operators := apptesting.CreateRandomAccounts(len(monikers))

	for i, moniker := range monikers {
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("existing-" + moniker)).PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		s.Require().NoError(err)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  v34.ValidatorPower,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         moniker,
				OperatorAddress: operators[i].String(),
			},
		}
		err = s.App.POAKeeper.CreateValidator(s.Ctx, sdk.GetConsAddress(pubKey), validator, true)
		s.Require().NoError(err)

		s.seededPubKeys[moniker] = pubKey
		s.seededOperators[moniker] = operators[i].String()
	}
}

func (s *UpgradeTestSuite) seedCurrentPOASet() {
	s.seedPOASet(append(append([]string{}, continuingMonikers...), outgoingMonikers...))
}

func (s *UpgradeTestSuite) capturePreUpgradeState() {
	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeTotalPower = totalPower
	s.preUpgradeUpdateCount = len(s.App.POAKeeper.ReapValidatorUpdates(s.Ctx))
}

// validatorsByMoniker unpacks every POA validator into a moniker-keyed map.
func (s *UpgradeTestSuite) validatorsByMoniker() map[string]poatypes.Validator {
	validators, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)

	byMoniker := map[string]poatypes.Validator{}
	for _, validator := range validators {
		s.Require().NotNil(validator.Metadata)
		byMoniker[validator.Metadata.Moniker] = validator
	}
	return byMoniker
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// ----- arrange -----
	s.useTestConsensusKeys()
	s.seedCurrentPOASet()
	s.capturePreUpgradeState()

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	byMoniker := s.validatorsByMoniker()

	// Incoming: present at ValidatorPower with the generated pubkey and the
	// (test-filled) payout address from utils.PoaValidatorSet.
	operatorByMoniker := map[string]string{}
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
	for _, entry := range v34.IncomingValidators {
		validator, ok := byMoniker[entry.Moniker]
		s.Require().True(ok, "incoming validator %s missing from POA", entry.Moniker)
		s.Require().Equal(v34.ValidatorPower, validator.Power)
		s.Require().Equal(operatorByMoniker[entry.Moniker], validator.Metadata.OperatorAddress)

		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		s.Require().True(pubKey.Equals(s.incomingPubKeys[entry.Moniker]))
	}

	// Outgoing: soft-deleted — record retained at power 0 with metadata intact.
	for _, moniker := range outgoingMonikers {
		validator, ok := byMoniker[moniker]
		s.Require().True(ok, "outgoing validator %s record should be retained", moniker)
		s.Require().Zero(validator.Power)
		s.Require().Equal(s.seededOperators[moniker], validator.Metadata.OperatorAddress)
	}

	// Continuing: untouched.
	for _, moniker := range continuingMonikers {
		s.Require().Equal(v34.ValidatorPower, byMoniker[moniker].Power, "continuing validator %s power changed", moniker)
	}

	// Total power unchanged (+2*274523 from creates, -2*274523 from removals).
	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.preUpgradeTotalPower, totalPower)

	s.checkEmittedUpdates()
}

// checkEmittedUpdates is the direct regression guard against the chain-halt
// hazard: the upgrade must queue exactly 4 ABCI updates (2 creates + 2
// removals), each for a distinct consensus pubkey. A duplicate pubkey in one
// block's update set panics every CometBFT node.
func (s *UpgradeTestSuite) checkEmittedUpdates() {
	// ReapValidatorUpdates does NOT drain the queue despite its doc comment
	// (the transient store clears per block, not per call), so updates
	// accumulate across the whole test — slicing off the pre-upgrade count
	// is intentional, not a workaround.
	allUpdates := s.App.POAKeeper.ReapValidatorUpdates(s.Ctx)
	s.Require().GreaterOrEqual(len(allUpdates), s.preUpgradeUpdateCount)
	newUpdates := allUpdates[s.preUpgradeUpdateCount:]
	s.Require().Len(newUpdates, 4, "upgrade must emit exactly 4 validator updates")

	powersByPubKey := map[string]int64{}
	for _, update := range newUpdates {
		key := update.PubKey.String()
		_, duplicate := powersByPubKey[key]
		s.Require().False(duplicate, "duplicate consensus pubkey in emitted updates — this would halt every node")
		powersByPubKey[key] = update.Power
	}

	for moniker, pubKey := range s.incomingPubKeys {
		cmtKey, err := cryptocodec.ToCmtProtoPublicKey(pubKey)
		s.Require().NoError(err)
		s.Require().Equal(v34.ValidatorPower, powersByPubKey[cmtKey.String()],
			"incoming validator %s should be created at ValidatorPower", moniker)
	}
	for _, moniker := range outgoingMonikers {
		cmtKey, err := cryptocodec.ToCmtProtoPublicKey(s.seededPubKeys[moniker])
		s.Require().NoError(err)
		power, ok := powersByPubKey[cmtKey.String()]
		s.Require().True(ok, "outgoing validator %s should have a removal update", moniker)
		s.Require().Zero(power)
	}
}

// TestOutgoingValidatorFeesRemainWithdrawable proves the accrued-fees
// guarantee documented on SwapPoaValidators: fees an outgoing validator
// accrued before the swap must still be withdrawable afterward, even though
// its power drops to 0. This depends on checkpoint ordering inside the
// handler — CreateValidator runs with checkpoint=true for the incoming
// validators while the outgoing validators still hold power, so the
// outgoing validators' pending share is checkpointed (and thus locked in)
// before they are later zeroed out.
func (s *UpgradeTestSuite) TestOutgoingValidatorFeesRemainWithdrawable() {
	// ----- arrange -----
	s.useTestConsensusKeys()
	s.seedCurrentPOASet()

	// Fund the POA module account directly — POA's own account balance
	// (minus what's already allocated) is what getUnallocatedFees treats as
	// pending fees to checkpoint out to validators.
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	s.FundModuleAccount(poatypes.ModuleName, sdk.NewInt64Coin(bondDenom, 1_000_000))

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	outgoingOperator, err := sdk.AccAddressFromBech32(s.seededOperators["Citadel.one"])
	s.Require().NoError(err)

	payout, err := s.App.POAKeeper.WithdrawValidatorFees(s.Ctx, outgoingOperator)
	s.Require().NoError(err)
	s.Require().False(payout.IsZero(), "outgoing validator's pre-upgrade accrued fees should still be withdrawable")
}

func (s *UpgradeTestSuite) TestSwapFailsWhenOutgoingValidatorMissing() {
	s.useTestConsensusKeys()
	// Seed everyone except Citadel.one.
	s.seedPOASet(append(append([]string{}, continuingMonikers...), "Cosmostation"))

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, `"Citadel.one" not found`)
}

func (s *UpgradeTestSuite) TestSwapFailsWhenIncomingMissingFromRegistry() {
	s.useTestConsensusKeys()
	// Break the moniker join for cosmosrescue. Restore is registered
	// immediately so a failed assertion below can't leave
	// "not-cosmosrescue" in the registry for the rest of the binary.
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "cosmosrescue" {
			utils.PoaValidatorSet[i].Moniker = "not-cosmosrescue"
		}
	}
	s.T().Cleanup(func() {
		for i := range utils.PoaValidatorSet {
			if utils.PoaValidatorSet[i].Moniker == "not-cosmosrescue" {
				utils.PoaValidatorSet[i].Moniker = "cosmosrescue"
			}
		}
	})
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "no entry in utils.PoaValidatorSet")
}
