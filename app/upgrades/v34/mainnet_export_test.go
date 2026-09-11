package v34_test

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	"github.com/Stride-Labs/stride/v34/utils"
)

// mainnetExportPath is relative to this package — read directly from the
// testdata/ checkout, not shipped in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite replays the v34 handler against real post-v33
// mainnet POA state. Unlike the synthetic suite, it runs with the REAL
// constants — no test-key substitution — so it is the release gate: it
// verifies the confirmed pubkeys and payout addresses against actual mainnet
// state, including the POA-set ≡ payout-registry invariant. The fixture is
// only committed during release prep; the suite skips when it is absent so
// CI stays green in the meantime.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper

	preUpgradeTotalPower  int64
	preUpgradeUpdateCount int
}

func (s *MainnetExportTestSuite) SetupTest() {
	s.Setup()
}

func TestMainnetExportTestSuite(t *testing.T) {
	if _, err := os.Stat(mainnetExportPath); errors.Is(err, os.ErrNotExist) {
		t.Skipf("skipping: mainnet export fixture not present at %s — see testdata/README.md to generate it", mainnetExportPath)
	}
	suite.Run(t, new(MainnetExportTestSuite))
}

// strideExport is a thin view over the trimmed `strided export` JSON shape.
type strideExport struct {
	AppState map[string]json.RawMessage `json:"app_state"`
}

func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	// ----- arrange: seed POA with the real mainnet validator set -----
	export := s.loadTrimmedExport()
	exportValidators := s.populatePOAFromExport(export)

	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeTotalPower = totalPower
	s.preUpgradeUpdateCount = len(s.App.POAKeeper.ReapValidatorUpdates(s.Ctx))

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	byMoniker := s.validatorsByMoniker()

	operatorByMoniker := map[string]string{}
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
	for _, entry := range v34.IncomingValidators {
		validator, ok := byMoniker[entry.Moniker]
		s.Require().True(ok, "incoming validator %s missing from POA", entry.Moniker)
		s.Require().Equal(v34.ValidatorPower, validator.Power)
		s.Require().Equal(operatorByMoniker[entry.Moniker], validator.Metadata.OperatorAddress)

		// The real constant must decode to the real key.
		expectedKeyBytes, err := base64.StdEncoding.DecodeString(entry.ConsPubKeyBase64)
		s.Require().NoError(err)
		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		s.Require().True(pubKey.Equals(&ed25519.PubKey{Key: expectedKeyBytes}))
	}

	for _, moniker := range v34.OutgoingMonikers {
		validator, ok := byMoniker[moniker]
		s.Require().True(ok, "outgoing validator %s record should be retained", moniker)
		s.Require().Zero(validator.Power)
	}

	// Continuing mainnet validators untouched.
	for moniker, exportPower := range exportValidators {
		if moniker == "Citadel.one" || moniker == "Cosmostation" {
			continue
		}
		s.Require().Equal(exportPower, byMoniker[moniker].Power, "continuing validator %s power changed", moniker)
	}

	postTotalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.preUpgradeTotalPower, postTotalPower)

	// Exactly 4 new updates, unique pubkeys — the chain-halt guard.
	// ReapValidatorUpdates does NOT drain the queue despite its doc comment
	// (the transient store clears per block, not per call) — so updates
	// accumulate across the whole test and slicing off the pre-upgrade count
	// is intentional, not a workaround.
	allUpdates := s.App.POAKeeper.ReapValidatorUpdates(s.Ctx)
	newUpdates := allUpdates[s.preUpgradeUpdateCount:]
	s.Require().Len(newUpdates, 4)
	seen := map[string]bool{}
	for _, update := range newUpdates {
		key := update.PubKey.String()
		s.Require().False(seen[key], "duplicate consensus pubkey in emitted updates")
		seen[key] = true
	}

	// The central invariant this upgrade must preserve: POA's active
	// (power > 0) validator set and utils.PoaValidatorSet's payout registry
	// must name exactly the same validators. Diverge and either stTokens get
	// burned (a signer with no registry entry) or the registry pays out to a
	// validator that no longer signs.
	registryByMoniker := map[string]string{}
	for _, v := range utils.PoaValidatorSet {
		registryByMoniker[v.Moniker] = v.Operator
	}

	// (a) every active POA validator must have a matching registry entry,
	// with the operator address agreeing. Skip the apptesting genesis
	// validator (moniker "test-validator") — it's a fixture of the local
	// test app, not part of the real mainnet POA set, and has no registry
	// entry.
	for moniker, validator := range byMoniker {
		if moniker == "test-validator" || validator.Power == 0 {
			continue
		}
		operator, ok := registryByMoniker[moniker]
		s.Require().True(ok, "active POA validator %s has no utils.PoaValidatorSet entry", moniker)
		s.Require().Equal(operator, validator.Metadata.OperatorAddress,
			"POA validator %s operator address disagrees with utils.PoaValidatorSet", moniker)
	}

	// (b) every registry entry must correspond to an active POA validator.
	s.Require().Len(utils.PoaValidatorSet, 8, "registry should have exactly 8 entries post-swap")
	for _, entry := range utils.PoaValidatorSet {
		validator, ok := byMoniker[entry.Moniker]
		s.Require().True(ok, "registry entry %s has no POA validator", entry.Moniker)
		s.Require().NotZero(validator.Power, "registry entry %s corresponds to an inactive (power 0) POA validator", entry.Moniker)
	}
}

// loadTrimmedExport reads testdata/mainnet_export.json.gz.
func (s *MainnetExportTestSuite) loadTrimmedExport() strideExport {
	f, err := os.Open(mainnetExportPath)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = f.Close() })

	gz, err := gzip.NewReader(f)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = gz.Close() })

	var export strideExport
	s.Require().NoError(json.NewDecoder(gz).Decode(&export))
	s.Require().NotEmpty(export.AppState, "trimmed export has no app_state — check testdata/README.md")
	return export
}

// populatePOAFromExport seeds POA with every validator from the export's
// app_state.poa section and returns moniker → power for later comparison.
// The test app's genesis POA validator (moniker "test-validator") is left in
// place; its consensus key cannot collide with mainnet keys.
func (s *MainnetExportTestSuite) populatePOAFromExport(export strideExport) map[string]int64 {
	raw, ok := export.AppState["poa"]
	s.Require().True(ok, "trimmed export missing poa section")

	var genesis poatypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &genesis))
	s.Require().Len(genesis.Validators, 8, "mainnet export should contain exactly 8 POA validators")

	powers := map[string]int64{}
	for _, validator := range genesis.Validators {
		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		err := s.App.POAKeeper.CreateValidator(s.Ctx, sdk.GetConsAddress(pubKey), validator, true)
		s.Require().NoError(err)
		powers[validator.Metadata.Moniker] = validator.Power
	}

	s.Require().Contains(powers, "Citadel.one", "export must contain the outgoing validators")
	s.Require().Contains(powers, "Cosmostation", "export must contain the outgoing validators")
	return powers
}

// validatorsByMoniker unpacks every POA validator into a moniker-keyed map.
func (s *MainnetExportTestSuite) validatorsByMoniker() map[string]poatypes.Validator {
	validators, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)

	byMoniker := map[string]poatypes.Validator{}
	for _, validator := range validators {
		s.Require().NotNil(validator.Metadata)
		byMoniker[validator.Metadata.Moniker] = validator
	}
	return byMoniker
}
