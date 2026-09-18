package v34_test

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	"github.com/Stride-Labs/stride/v34/utils"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// mainnetExportPath is relative to this package — read directly from the
// testdata/ checkout, not shipped in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite replays the v34 handler against real post-v33
// mainnet state: the POA validator set, the injective-1, celestia and
// cosmoshub-4 host zones, the celestia deposit records, the cosmoshub-4 LSM
// deposits and the staketia host zone. Unlike the synthetic suite, it runs
// with the REAL constants — no test-key substitution — so it is the release
// gate: it verifies the confirmed pubkeys and payout addresses against actual
// mainnet state, including the POA-set ≡ payout-registry invariant, and that
// every accounting fix applied in full against real state (the Injective and
// Celestia delta tables, the staketia balance delta and the Cosmos Hub LSM
// deposit constant). Each fix deliberately skips rather than errors when its
// constants no longer describe chain state, so asserting the applied effects
// here is what turns a stale constant into a red build instead of a silently
// deferred reconciliation. The fixture is only committed during release prep;
// the suite skips when it is absent so CI stays green in the meantime.
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
	// ----- arrange: seed POA, the host zones, records and staketia from real mainnet state -----
	export := s.loadTrimmedExport()
	exportValidators := s.populatePOAFromExport(export)
	exportHostZones := s.populateHostZonesFromExport(export)
	exportRecords := s.populateRecordsFromExport(export)
	exportStaketiaHostZone := s.populateStaketiaHostZoneFromExport(export)
	celestiaNumeratorBefore := celestiaRateNumerator(&s.AppTestHelper)

	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeTotalPower = totalPower
	s.preUpgradeUpdateCount = len(s.App.POAKeeper.ReapValidatorUpdates(s.Ctx))

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert: every accounting fix applied in full with the real constants -----
	s.assertInjectiveReconciled(exportHostZones[v34.InjectiveChainId])
	celestiaDelta := s.assertCelestiaReconciled(exportHostZones[v34.CelestiaChainId], exportRecords, celestiaNumeratorBefore)
	s.assertStaketiaAdjusted(exportStaketiaHostZone, exportHostZones[v34.CelestiaChainId], celestiaDelta)
	s.assertCosmosHubLsmDepositClosed(exportHostZones[v34.CosmosHubChainId], exportRecords)

	// ----- assert: POA swap -----
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

// populateHostZonesFromExport seeds the injective-1, celestia and cosmoshub-4
// host zones from the export's app_state.stakeibc section and returns them by
// chain id for later comparison.
func (s *MainnetExportTestSuite) populateHostZonesFromExport(export strideExport) map[string]stakeibctypes.HostZone {
	raw, ok := export.AppState["stakeibc"]
	s.Require().True(ok, "trimmed export missing stakeibc section — regenerate per testdata/README.md")

	var genesis stakeibctypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &genesis))
	s.Require().Len(genesis.HostZoneList, 3, "export should contain exactly the injective-1, celestia and cosmoshub-4 host zones")

	hostZones := map[string]stakeibctypes.HostZone{}
	for _, hostZone := range genesis.HostZoneList {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		hostZones[hostZone.ChainId] = hostZone
	}
	for _, chainId := range []string{v34.InjectiveChainId, v34.CelestiaChainId, v34.CosmosHubChainId} {
		s.Require().Contains(hostZones, chainId, "export must contain the %s host zone", chainId)
	}
	return hostZones
}

// populateRecordsFromExport seeds the celestia deposit records and the
// cosmoshub-4 LSM token deposits from the export's app_state.records section,
// with their real statuses and ids, and returns the section for later
// comparison. No icacallbacks entries are seeded: ReconcileCelestia only
// deletes callbacks it finds, so in-progress records reconcile without them.
func (s *MainnetExportTestSuite) populateRecordsFromExport(export strideExport) recordstypes.GenesisState {
	raw, ok := export.AppState["records"]
	s.Require().True(ok, "trimmed export missing records section — regenerate per testdata/README.md")

	var genesis recordstypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &genesis))
	s.Require().NotEmpty(genesis.DepositRecordList, "export should contain the celestia deposit records")
	s.Require().NotEmpty(genesis.LsmTokenDepositList, "export should contain the cosmoshub-4 LSM deposits")

	// A real export carries deposit_record_count above every id; mirror that so a record
	// ReconcileCelestia appends gets a fresh id instead of id 0
	maxId := uint64(0)
	for _, record := range genesis.DepositRecordList {
		s.Require().Equal(v34.CelestiaChainId, record.HostZoneId, "deposit record %d is not a celestia record", record.Id)
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, record)
		if record.Id > maxId {
			maxId = record.Id
		}
	}
	s.App.RecordsKeeper.SetDepositRecordCount(s.Ctx, maxId+1)

	for _, deposit := range genesis.LsmTokenDepositList {
		s.Require().Equal(v34.CosmosHubChainId, deposit.ChainId, "LSM deposit %s is not a cosmoshub-4 deposit", deposit.Denom)
		s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	}
	return genesis
}

// populateStaketiaHostZoneFromExport seeds the staketia host zone from the
// export's app_state.staketia section and returns it for later comparison.
func (s *MainnetExportTestSuite) populateStaketiaHostZoneFromExport(export strideExport) staketiatypes.HostZone {
	raw, ok := export.AppState["staketia"]
	s.Require().True(ok, "trimmed export missing staketia section — regenerate per testdata/README.md")

	var genesis staketiatypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &genesis))
	s.Require().Equal(staketiatypes.CelestiaChainId, genesis.HostZone.ChainId, "export should contain the staketia host zone")
	s.Require().False(genesis.HostZone.RemainingDelegatedBalance.IsNil(), "staketia host zone has no remaining delegated balance")

	s.App.StaketiaKeeper.SetHostZone(s.Ctx, genesis.HostZone)
	return genesis.HostZone
}

// assertInjectiveReconciled checks the delta table applied in full against the
// real host zone: the pending undelegation equals the whole table's sum (which
// can only happen if every validator was found and no result went negative),
// and each validator moved by exactly its delta.
func (s *MainnetExportTestSuite) assertInjectiveReconciled(exportHostZone stakeibctypes.HostZone) {
	expectedDelta := sdkmath.ZeroInt()
	for _, entry := range v34.InjectiveDelegationDeltas {
		_, _, found := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, entry.Address)
		s.Require().True(found, "delta table entry %s (%s) is not on the mainnet host zone — re-measure the table", entry.Name, entry.Address)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().True(expectedDelta.IsPositive(), "the table should net to the staked redemption excess")

	pending, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, v34.InjectiveChainId)
	s.Require().True(found, "reconciliation was skipped against mainnet state — a delta would drive a delegation negative, re-measure the table")
	s.Require().Equal(expectedDelta, pending, "pending undelegation should equal the full table sum")

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().True(found)
	s.Require().Equal(exportHostZone.TotalDelegations.Add(expectedDelta), hostZone.TotalDelegations)
	for _, entry := range v34.InjectiveDelegationDeltas {
		before, _, _ := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, entry.Address)
		after, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().Equal(before.Delegation.Add(entry.Delta), after.Delegation, "%s delegation should move by its delta", entry.Name)
		s.Require().False(after.Delegation.IsNegative(), "%s delegation went negative", entry.Name)
	}
}

// assertCelestiaReconciled checks both halves of the reconciliation applied in
// full against the real state: every table validator moved by exactly its
// delta and TotalDelegations rose by the table sum (which can only happen if
// every address was found, no result went negative and the records covered
// the sum), exactly the table sum left the celestia deposit records, transfer
// records were never touched, and the redemption rate components are
// unchanged. Returns the table sum for the staketia assertion.
func (s *MainnetExportTestSuite) assertCelestiaReconciled(
	exportHostZone stakeibctypes.HostZone,
	exportRecords recordstypes.GenesisState,
	numeratorBefore sdkmath.LegacyDec,
) sdkmath.Int {
	phantomAmount := sdkmath.ZeroInt()
	for _, entry := range v34.CelestiaDelegationDeltas {
		_, _, found := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, entry.Address)
		s.Require().True(found, "delta table entry %s (%s) is not on the mainnet host zone — re-measure the table", entry.Name, entry.Address)
		phantomAmount = phantomAmount.Add(entry.Delta)
	}
	s.Require().True(phantomAmount.IsPositive(), "the table should net to the unacknowledged stake")

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	s.Require().Equal(exportHostZone.TotalDelegations.Add(phantomAmount).String(), hostZone.TotalDelegations.String(),
		"reconciliation was skipped against mainnet state — a delta would drive a delegation negative or the records "+
			"no longer cover the table sum, re-measure the table")
	for _, entry := range v34.CelestiaDelegationDeltas {
		before, _, _ := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, entry.Address)
		after, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().Equal(before.Delegation.Add(entry.Delta).String(), after.Delegation.String(), "%s delegation should move by its delta", entry.Name)
		s.Require().False(after.Delegation.IsNegative(), "%s delegation went negative", entry.Name)
	}

	// Exactly the table sum left the records (deletes, one shrink or split), nothing else
	beforeRecordsTotal := sdkmath.ZeroInt()
	for _, record := range exportRecords.DepositRecordList {
		beforeRecordsTotal = beforeRecordsTotal.Add(record.Amount)
	}
	afterRecordsTotal := sdkmath.ZeroInt()
	for _, record := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		if record.HostZoneId == v34.CelestiaChainId {
			afterRecordsTotal = afterRecordsTotal.Add(record.Amount)
		}
	}
	s.Require().Equal(beforeRecordsTotal.Sub(phantomAmount).String(), afterRecordsTotal.String(),
		"exactly the table sum should be removed from the celestia deposit records")
	for _, record := range exportRecords.DepositRecordList {
		if record.Status != recordstypes.DepositRecord_TRANSFER_QUEUE && record.Status != recordstypes.DepositRecord_TRANSFER_IN_PROGRESS {
			continue
		}
		after, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, record.Id)
		s.Require().True(found, "transfer record %d should never be touched", record.Id)
		s.Require().Equal(record.Amount.String(), after.Amount.String(), "transfer record %d amount", record.Id)
		s.Require().Equal(record.Status, after.Status, "transfer record %d status", record.Id)
	}

	s.Require().Equal(numeratorBefore.String(), celestiaRateNumerator(&s.AppTestHelper).String(),
		"celestia redemption rate components (undelegated records + TotalDelegations) should be unchanged")
	return phantomAmount
}

// assertStaketiaAdjusted checks the staketia remaining delegated balance moved
// by exactly the constant delta (a negative result would have been skipped),
// and that the step left stakeibc's celestia TotalDelegations alone: it still
// equals the export value plus only the Celestia reconciliation's delta.
func (s *MainnetExportTestSuite) assertStaketiaAdjusted(
	exportStaketiaHostZone staketiatypes.HostZone,
	exportCelestiaHostZone stakeibctypes.HostZone,
	celestiaDelta sdkmath.Int,
) {
	staketiaHostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(exportStaketiaHostZone.RemainingDelegatedBalance.Add(v34.StaketiaRemainingDelegatedBalanceDelta).String(),
		staketiaHostZone.RemainingDelegatedBalance.String(),
		"adjustment was skipped against mainnet state — the result would be negative, re-measure the delta")
	s.Require().False(staketiaHostZone.RemainingDelegatedBalance.IsNegative())

	celestiaHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CelestiaChainId)
	s.Require().True(found)
	s.Require().Equal(exportCelestiaHostZone.TotalDelegations.Add(celestiaDelta).String(), celestiaHostZone.TotalDelegations.String(),
		"the staketia adjustment must not mirror to stakeibc TotalDelegations")
}

// assertCosmosHubLsmDepositClosed checks the export carried the stranded LSM
// deposit exactly as the constant describes it, and that the close-out
// applied: the deposit is gone, the other cosmoshub-4 deposits are untouched,
// and stakewithus plus TotalDelegations rose by exactly the constant amount.
func (s *MainnetExportTestSuite) assertCosmosHubLsmDepositClosed(exportHostZone stakeibctypes.HostZone, exportRecords recordstypes.GenesisState) {
	stranded := v34.CosmosHubStrandedLsmDeposit
	var exportDeposit *recordstypes.LSMTokenDeposit
	for i := range exportRecords.LsmTokenDepositList {
		if exportRecords.LsmTokenDepositList[i].Denom == stranded.Denom {
			exportDeposit = &exportRecords.LsmTokenDepositList[i]
		}
	}
	s.Require().NotNil(exportDeposit, "LSM deposit %s is not in the mainnet export — re-verify the constant", stranded.Denom)
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, exportDeposit.Status, "mainnet deposit status drifted from the constant")
	s.Require().Equal(stranded.Amount.String(), exportDeposit.Amount.String(), "mainnet deposit amount drifted from the constant")
	s.Require().Equal(stranded.ValidatorAddress, exportDeposit.ValidatorAddress, "mainnet deposit validator drifted from the constant")
	_, _, found := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, stranded.ValidatorAddress)
	s.Require().True(found, "validator %s is not on the mainnet host zone — re-verify the constant", stranded.ValidatorAddress)

	_, found = s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, stranded.Denom)
	s.Require().False(found, "close-out was skipped against mainnet state — the deposit should be removed")
	for _, deposit := range exportRecords.LsmTokenDepositList {
		if deposit.Denom == stranded.Denom {
			continue
		}
		after, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, deposit.Denom)
		s.Require().True(found, "LSM deposit %s should be untouched", deposit.Denom)
		s.Require().Equal(deposit, after, "LSM deposit %s should be untouched", deposit.Denom)
	}

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	s.Require().True(found)
	before, _, _ := stakeibckeeper.GetValidatorFromAddress(exportHostZone.Validators, stranded.ValidatorAddress)
	after, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, stranded.ValidatorAddress)
	s.Require().Equal(before.Delegation.Add(stranded.Amount).String(), after.Delegation.String(), "stakewithus delegation up by exactly the closed amount")
	s.Require().Equal(exportHostZone.TotalDelegations.Add(stranded.Amount).String(), hostZone.TotalDelegations.String(),
		"cosmoshub-4 TotalDelegations up by exactly the closed amount")
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
