package v33_test

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
	"github.com/Stride-Labs/stride/v32/utils"
)

// mainnetExportPath is relative to this package — read directly from the
// testdata/ checkout so we don't have to ship it in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite runs the v33 upgrade handler against real post-v32
// mainnet state, complementing the synthetic-state suite in upgrades_test.go.
//
// The synthetic suite seeds 8 ed25519 keys + arbitrary funding inside a fresh
// test app — close enough to verify the handler logic but not the validator
// addresses, balances, or governance state we'll actually meet on chain. This
// suite parses a `strided export` of mainnet (trimmed by scripts/trim_export.sh)
// and surgically populates the test app's consumer keeper, ICS module account
// balances, and (eventually) gov state with the real values, then runs the
// same handler and asserts on the same outcomes — plus the moniker join
// (mainnet hex_cons_addr → validators.json moniker → utils.PoaValidatorSet
// operator), which the synthetic test can't validate end-to-end.
//
// The fixture is large enough that it isn't always present in dev checkouts;
// when missing, the suite skips gracefully so CI doesn't break. Release flow
// should ensure the fixture is committed before tagging the binary.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper

	// preUpgradePOAValidatorCount: bootstrap POA validator(s) seeded by the
	// test app at genesis. The handler should add exactly 8 more (one per
	// mainnet ICS validator) on top of this baseline.
	preUpgradePOAValidatorCount int

	// preUpgradeICSAccountTotal: sum of coins held in the two ICS reward
	// module accounts before the handler runs. SweepICSModuleAccounts must
	// move exactly this much into the community pool.
	preUpgradeICSAccountTotal sdk.Coins

	// preUpgradeCommunityPool: community pool balance before the handler runs.
	preUpgradeCommunityPool sdk.DecCoins
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

// strideExport is a thin view over the trimmed `strided export` JSON shape:
// we only care about app_state and decode each module section on demand.
type strideExport struct {
	AppState map[string]json.RawMessage `json:"app_state"`
}

func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	export := s.loadTrimmedExport()

	// ----- arrange: replace synthetic state with real mainnet state -----
	s.populateConsumerKeeperFromExport(export)
	s.populateICSModuleBalancesFromExport(export)

	s.capturePreUpgradeState()

	// Sanity: the export must have exactly 8 validators with addresses that
	// all match validators.json keys — otherwise the moniker-join inside
	// SnapshotValidatorsFromICS will halt the upgrade before we get to assert.
	s.requireMonikersResolvable()

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v33.UpgradeName)

	// ----- assert -----
	s.assertPOAValidatorsMatchExport(export)
	s.assertPOAAdminSet()
	s.assertICSModuleAccountsDrained()
	s.assertCommunityPoolGrewByICSBalances()
}

// --- arrange helpers ---

// loadTrimmedExport reads testdata/mainnet_export.json.gz and unmarshals just
// enough of it to dispatch per-module InitGenesis-equivalent population.
func (s *MainnetExportTestSuite) loadTrimmedExport() strideExport {
	f, err := os.Open(mainnetExportPath)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = f.Close() })

	gz, err := gzip.NewReader(f)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = gz.Close() })

	var export strideExport
	s.Require().NoError(json.NewDecoder(gz).Decode(&export))
	s.Require().NotEmpty(export.AppState, "trimmed export has no app_state — check trim_export.sh output")
	return export
}

// populateConsumerKeeperFromExport wipes any test-default CCValidators and
// installs the mainnet validator set extracted from the export's
// `app_state.ccvconsumer.provider.initial_val_set`. That field is what the
// consumer module's ExportGenesis writes via MustGetCurrentValidatorsAsABCIUpdates,
// so it reflects mainnet's *current* live validator set at export height.
func (s *MainnetExportTestSuite) populateConsumerKeeperFromExport(e strideExport) {
	raw, ok := e.AppState["ccvconsumer"]
	s.Require().True(ok, "trimmed export missing ccvconsumer section")

	var gen consumertypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &gen))
	s.Require().NotEmpty(gen.Provider.InitialValSet,
		"ccvconsumer.provider.initial_val_set is empty — was the export taken from a running consumer chain?")

	// Drop any CCValidators left over from the test app's own InitChain.
	for _, existing := range s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx) {
		s.App.ConsumerKeeper.DeleteCCValidator(s.Ctx, existing.Address)
	}

	for _, vu := range gen.Provider.InitialValSet {
		consPub, err := cryptocodec.FromCmtProtoPublicKey(vu.PubKey)
		s.Require().NoError(err)
		pkAny, err := codectypes.NewAnyWithValue(consPub)
		s.Require().NoError(err)
		s.App.ConsumerKeeper.SetCCValidator(s.Ctx, consumertypes.CrossChainValidator{
			Address: consPub.Address().Bytes(),
			Power:   vu.Power,
			Pubkey:  pkAny,
		})
	}

	got := len(s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx))
	s.Require().Equal(v33.ExpectedValidatorCount, got,
		"expected %d mainnet validators in export, got %d", v33.ExpectedValidatorCount, got)
}

// populateICSModuleBalancesFromExport finds the two ICS reward module account
// addresses in the bank section's balances list, then mints + transfers
// matching coins into the test app's same module accounts. We do not
// repopulate non-ICS accounts — the handler doesn't read them and rebuilding
// the full bank state would conflict with the test app's existing supply.
func (s *MainnetExportTestSuite) populateICSModuleBalancesFromExport(e strideExport) {
	raw, ok := e.AppState["bank"]
	s.Require().True(ok, "trimmed export missing bank section")

	var bankGen banktypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &bankGen))

	moduleAccountTargets := map[string]string{
		s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerRedistributeName).String():     consumertypes.ConsumerRedistributeName,
		s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerToSendToProviderName).String(): consumertypes.ConsumerToSendToProviderName,
	}

	matched := 0
	for _, b := range bankGen.Balances {
		moduleName, isTarget := moduleAccountTargets[b.Address]
		if !isTarget {
			continue
		}
		if b.Coins.IsZero() {
			continue
		}
		for _, coin := range b.Coins {
			s.FundModuleAccount(moduleName, coin)
		}
		matched++
	}
	// It's fine for the export to have these accounts at zero balance (the
	// sweep is then a no-op on this run). We only fail if the bank section
	// is structurally broken.
	s.Require().LessOrEqual(matched, len(moduleAccountTargets),
		"bank.balances had duplicate entries for an ICS module account")
}

// requireMonikersResolvable fails the test early if any CCValidator's
// hex_cons_addr is not present in v33.ValidatorMonikers. The moniker map is
// embedded from validators.json at build time, so this assertion validates
// that validators.json is in sync with the export's actual validator set —
// the most prod-faithful thing this suite checks.
func (s *MainnetExportTestSuite) requireMonikersResolvable() {
	for _, cv := range s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx) {
		hexAddr := fmt.Sprintf("%x", cv.Address)
		_, ok := v33.ValidatorMonikers[hexAddr]
		s.Require().True(ok,
			"hex_cons_addr %s from mainnet export has no entry in v33 validators.json — "+
				"either the export is from a different validator set or validators.json is stale",
			hexAddr)
	}
	// Also check the operator-address side of the join — if validators.json
	// names a moniker that no longer appears in utils.PoaValidatorSet, the
	// handler would halt with "no entry in utils.PoaValidatorSet".
	knownMonikers := make(map[string]bool, len(utils.PoaValidatorSet))
	for _, v := range utils.PoaValidatorSet {
		knownMonikers[v.Moniker] = true
	}
	for hexAddr, moniker := range v33.ValidatorMonikers {
		s.Require().True(knownMonikers[moniker],
			"validators.json maps %s → %q, but %q has no entry in utils.PoaValidatorSet",
			hexAddr, moniker, moniker)
	}
}

// capturePreUpgradeState snapshots state we need to compare against after
// the handler runs.
func (s *MainnetExportTestSuite) capturePreUpgradeState() {
	initialVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradePOAValidatorCount = len(initialVals)

	consRedistr := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerRedistributeName)
	consToProv := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerToSendToProviderName)
	s.preUpgradeICSAccountTotal = s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistr).
		Add(s.App.BankKeeper.GetAllBalances(s.Ctx, consToProv)...)

	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeCommunityPool = feePool.CommunityPool
}

// --- assertion helpers ---

// assertPOAValidatorsMatchExport verifies that every mainnet ICS validator
// landed in the POA store with the right pubkey, power, real moniker, and
// real stride1... operator address. This is the strongest end-to-end check
// in the suite — it validates that validators.json + utils.PoaValidatorSet +
// the handler all agree on every single mainnet validator.
func (s *MainnetExportTestSuite) assertPOAValidatorsMatchExport(e strideExport) {
	ccVals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
	s.Require().Len(ccVals, v33.ExpectedValidatorCount)

	poaVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(poaVals, s.preUpgradePOAValidatorCount+v33.ExpectedValidatorCount,
		"POA should have bootstrap + %d mainnet validators after upgrade", v33.ExpectedValidatorCount)

	// Build operator → POA validator lookup so we can match by canonical
	// stride1... address (unique across the set, unlike Moniker).
	poaByOperator := make(map[string]struct {
		pk    cryptotypes.PubKey
		power int64
	}, len(poaVals))
	for _, pv := range poaVals {
		var pk cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(pv.PubKey, &pk))
		if pv.Metadata == nil {
			continue // bootstrap validator from test_setup may have no metadata
		}
		poaByOperator[pv.Metadata.OperatorAddress] = struct {
			pk    cryptotypes.PubKey
			power int64
		}{pk: pk, power: pv.Power}
	}

	// Every entry in utils.PoaValidatorSet should have ended up in POA store
	// with the matching pubkey + power from the ICS export.
	for _, expected := range utils.PoaValidatorSet {
		got, ok := poaByOperator[expected.Operator]
		s.Require().True(ok, "operator %s (%s) missing from POA store after upgrade",
			expected.Operator, expected.Moniker)

		// Look up the corresponding CCValidator by hex_cons_addr → moniker join.
		var matched *consumertypes.CrossChainValidator
		for i := range ccVals {
			hexAddr := fmt.Sprintf("%x", ccVals[i].Address)
			if v33.ValidatorMonikers[hexAddr] == expected.Moniker {
				matched = &ccVals[i]
				break
			}
		}
		s.Require().NotNil(matched,
			"no CCValidator maps to moniker %q via validators.json", expected.Moniker)

		consPub, err := matched.ConsPubKey()
		s.Require().NoError(err)
		s.Require().True(got.pk.Equals(consPub),
			"pubkey mismatch for %s (%s): POA pubkey != ICS pubkey",
			expected.Operator, expected.Moniker)
		s.Require().Equal(matched.Power, got.power,
			"power mismatch for %s (%s): POA=%d ICS=%d",
			expected.Operator, expected.Moniker, got.power, matched.Power)
	}
}

func (s *MainnetExportTestSuite) assertPOAAdminSet() {
	params, err := s.App.POAKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v33.AdminMultisigAddress, params.Admin)
}

func (s *MainnetExportTestSuite) assertICSModuleAccountsDrained() {
	consRedistr := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerRedistributeName)
	consToProv := s.App.AccountKeeper.GetModuleAddress(consumertypes.ConsumerToSendToProviderName)
	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistr).IsZero(),
		"cons_redistribute should be drained after sweep")
	s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consToProv).IsZero(),
		"cons_to_send_to_provider should be drained after sweep")
}

// assertCommunityPoolGrewByICSBalances verifies the sweep transferred exactly
// the pre-upgrade ICS module account totals into the community pool, no more
// and no less. We compare per-denom to handle multi-denom balances cleanly.
func (s *MainnetExportTestSuite) assertCommunityPoolGrewByICSBalances() {
	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	post := feePool.CommunityPool

	// Δ = post − pre  must equal pre-upgrade ICS account totals (cast to DecCoins).
	delta := post.Sub(s.preUpgradePOACommunityPool())
	expected := sdk.NewDecCoinsFromCoins(s.preUpgradeICSAccountTotal...)
	s.Require().True(delta.Equal(expected),
		"community pool delta mismatch:\n  expected: %s\n  actual:   %s",
		expected, delta)
}

// preUpgradePOACommunityPool is a tiny accessor that returns the pre-upgrade
// community pool we captured. Helper exists purely to keep the assertion
// site readable.
func (s *MainnetExportTestSuite) preUpgradePOACommunityPool() sdk.DecCoins {
	return s.preUpgradeCommunityPool
}
