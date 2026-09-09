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
	// the v34.IncomingValidators placeholders by fillPlaceholders
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

// fillPlaceholders substitutes test values for the release-time placeholders:
// generated ed25519 keys for the incoming consensus pubkeys, and random
// accounts for the placeholder payout addresses in utils.PoaValidatorSet.
// (Package-level vars are mutated; tests in this package run serially and
// every test that needs filled values calls this in arrange.)
func (s *UpgradeTestSuite) fillPlaceholders() {
	s.incomingPubKeys = map[string]cryptotypes.PubKey{}
	for i := range v34.IncomingValidators {
		moniker := v34.IncomingValidators[i].Moniker
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("incoming-" + moniker)).PubKey()
		s.incomingPubKeys[moniker] = pubKey
		v34.IncomingValidators[i].ConsPubKeyBase64 = base64.StdEncoding.EncodeToString(pubKey.Bytes())
	}
	for i := range utils.PoaValidatorSet {
		if utils.IsPlaceholderOperator(utils.PoaValidatorSet[i].Operator) {
			utils.PoaValidatorSet[i].Operator = apptesting.CreateRandomAccounts(1)[0].String()
		}
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
	s.fillPlaceholders()
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

func (s *UpgradeTestSuite) TestSwapFailsWithPlaceholderPubkey() {
	s.fillPlaceholders()
	v34.IncomingValidators[0].ConsPubKeyBase64 = v34.PlaceholderConsPubKey
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "placeholder consensus pubkey")
}

func (s *UpgradeTestSuite) TestSwapFailsWithPlaceholderPayoutAddress() {
	s.fillPlaceholders()
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "cosmosrescue" {
			utils.PoaValidatorSet[i].Operator = utils.PlaceholderOperatorCosmosRescue
		}
	}
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "placeholder payout address")
}

func (s *UpgradeTestSuite) TestSwapFailsWhenOutgoingValidatorMissing() {
	s.fillPlaceholders()
	// Seed everyone except Citadel.one.
	s.seedPOASet(append(append([]string{}, continuingMonikers...), "Cosmostation"))

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, `"Citadel.one" not found`)
}

func (s *UpgradeTestSuite) TestSwapFailsWhenIncomingMissingFromRegistry() {
	s.fillPlaceholders()
	// Break the moniker join for cosmosrescue.
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "cosmosrescue" {
			utils.PoaValidatorSet[i].Moniker = "not-cosmosrescue"
		}
	}
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "no entry in utils.PoaValidatorSet")

	// Restore for subsequent tests.
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "not-cosmosrescue" {
			utils.PoaValidatorSet[i].Moniker = "cosmosrescue"
		}
	}
}
