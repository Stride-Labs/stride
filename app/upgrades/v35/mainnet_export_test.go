package v35_test

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/testdata"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icagenesistypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v35/x/autopilot/types"
	icaoracletypes "github.com/Stride-Labs/stride/v35/x/icaoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v35/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v35/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// mainnetExportPath is relative to this package: read from the testdata checkout, never
// shipped in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// evmosChainId is deprecated on mainnet; the suite gives it an open, empty delegation channel.
const evmosChainId = "evmos_9001-2"

// expectedResetZones is the zones with an open delegation channel and no packet in flight at the
// fixture's height (testdata/README.md, "Committed fixture provenance"): the only zones whose
// flags the handler may reset. juno-1 has packets in flight; celestia, haqq and laozi have no
// open channel; deprecated zones (evmos here) are skipped even with an open channel.
var expectedResetZones = []string{
	"cosmoshub-4", "dydx-mainnet-1", "injective-1", "osmosis-1", "phoenix-1", "sommelier-3", "ssc-1",
}

// MainnetExportTestSuite replays the v35 handler against real mainnet state with the REAL
// constants (no test-key substitution) and asserts every effect of spec §5 on it. It is the
// release gate: a constant that no longer matches state (the haqq delta table, the contract
// list, the channel picture) turns into a red build here instead of a silently skipped step.
// The fixture is committed during release prep only; the suite skips while it is absent.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper
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

// strideExport is a thin view over the trimmed export JSON shape.
type strideExport struct {
	AppState map[string]json.RawMessage `json:"app_state"`
}

// delegationChannel is the fixture's synthetic per-zone delegation ICA picture (README).
type delegationChannel struct {
	ConnectionId      string `json:"connection_id"`
	ChannelId         string `json:"channel_id"`
	PacketCommitments int    `json:"packet_commitments"`
}

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

func (s *MainnetExportTestSuite) section(export strideExport, name string) json.RawMessage {
	raw, ok := export.AppState[name]
	s.Require().True(ok, "trimmed export missing %s section — regenerate per testdata/README.md", name)
	return raw
}

// populateStakeibcFromExport seeds every host zone and trade route and returns the zones by
// chain id for later comparison.
func (s *MainnetExportTestSuite) populateStakeibcFromExport(export strideExport) map[string]stakeibctypes.HostZone {
	var genesis stakeibctypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "stakeibc"), &genesis))
	s.Require().GreaterOrEqual(len(genesis.HostZoneList), 12, "export should carry the eleven in-scope zones and comdex-1 at least")

	hostZones := map[string]stakeibctypes.HostZone{}
	for _, hostZone := range genesis.HostZoneList {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		hostZones[hostZone.ChainId] = hostZone
	}
	for _, tradeRoute := range genesis.TradeRoutes {
		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)
	}
	s.Require().Contains(hostZones, v35.HaqqChainId)
	s.Require().Contains(hostZones, v35.ComdexChainId)
	s.Require().Len(genesis.TradeRoutes, 1, "mainnet has exactly the dYdX trade route")
	return hostZones
}

func (s *MainnetExportTestSuite) populateQueriesFromExport(export strideExport) []icqtypes.Query {
	var genesis icqtypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "interchainquery"), &genesis))
	for _, query := range genesis.Queries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}
	return genesis.Queries
}

func (s *MainnetExportTestSuite) populateAutopilotFromExport(export strideExport) {
	var genesis autopilottypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "autopilot"), &genesis))
	s.Require().True(genesis.Params.StakeibcActive, "autopilot stakeibc is active on mainnet before the upgrade")
	s.App.AutopilotKeeper.SetParams(s.Ctx, genesis.Params)
}

func (s *MainnetExportTestSuite) populateICAHostFromExport(export strideExport) []string {
	var genesis icagenesistypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "interchainaccounts"), &genesis))
	params := genesis.HostGenesisState.Params
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}))
	// Authority spec §2: mainnet allows delegate and redelegate (but not create validator or cancel unbonding)
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}))
	s.App.ICAHostKeeper.SetParams(s.Ctx, params)
	return params.AllowMessages
}

// populateWasmFromExport seeds the real upload-access params and one stand-in contract per
// deploy-key contract in the fixture. The fixture carries the contracts' infos but not their
// (Hyperlane) code, and wasmd's InitGenesis refuses a contract without a code history, so the
// suite stores hackatom once (through the gov permission keeper, which bypasses upload
// access) and instantiates it once per fixture entry with the deploy key as admin: the
// handler only reads and rewrites ContractInfo.Admin, so the code behind the address does
// not matter. What the fixture proves is the count and the admin of the real contracts;
// what the suite proves is that every contract with that admin ends with gov as admin.
func (s *MainnetExportTestSuite) populateWasmFromExport(export strideExport) wasmSeed {
	var genesis wasmtypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "wasm"), &genesis))
	s.Require().Len(genesis.Contracts, 4, "spec §3: the Hyperlane IGP, IGP hook, aggregate hook and multisig ISM")
	s.Require().NoError(s.App.WasmKeeper.SetParams(s.Ctx, genesis.Params))

	govKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	creator := apptesting.CreateRandomAccounts(1)[0]
	codeId, _, err := govKeeper.Create(s.Ctx, creator, testdata.HackatomContractWasm(), nil)
	s.Require().NoError(err, "store hackatom")
	initMsg, err := json.Marshal(map[string]string{"verifier": creator.String(), "beneficiary": creator.String()})
	s.Require().NoError(err)

	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	seed := wasmSeed{controlAdmin: apptesting.CreateRandomAccounts(1)[0]}
	for i, contract := range genesis.Contracts {
		s.Require().Equal(v35.WasmDeployKey, contract.ContractInfo.Admin, "fixture contract %d (%s) is not deploy-key administered", i, contract.ContractAddress)
		address, _, err := govKeeper.Instantiate(s.Ctx, codeId, creator, deployKey, initMsg, "hackatom", sdk.NewCoins())
		s.Require().NoError(err, "instantiate stand-in for %s", contract.ContractAddress)
		seed.deployKeyContracts = append(seed.deployKeyContracts, address)
	}

	// A control contract under another admin proves the handler moves only the deploy key's contracts
	seed.controlContract, _, err = govKeeper.Instantiate(s.Ctx, codeId, creator, seed.controlAdmin, initMsg, "hackatom-control", sdk.NewCoins())
	s.Require().NoError(err, "instantiate control contract")

	params := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, params.CodeUploadAccess.Permission)
	s.Require().Len(params.CodeUploadAccess.Addresses, 2, "spec §3: upload restricted to two addresses before the upgrade")
	return seed
}

// wasmSeed is the stand-in contracts the suite instantiates: one per deploy-key contract in the
// fixture, plus a control contract administered by an unrelated address.
type wasmSeed struct {
	deployKeyContracts []sdk.AccAddress
	controlContract    sdk.AccAddress
	controlAdmin       sdk.AccAddress
}

func (s *MainnetExportTestSuite) populateOraclesFromExport(export strideExport) []icaoracletypes.Oracle {
	var genesis icaoracletypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "icaoracle"), &genesis))
	for _, oracle := range genesis.Oracles {
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}
	return genesis.Oracles
}

func (s *MainnetExportTestSuite) populateRateLimitsFromExport(export strideExport) ratelimittypes.GenesisState {
	var genesis ratelimittypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "ratelimit"), &genesis))
	s.Require().NotEmpty(genesis.RateLimits, "spec §3: stToken rate limits exist before the upgrade")
	for _, rateLimit := range genesis.RateLimits {
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
	}
	for _, denom := range genesis.BlacklistedDenoms {
		s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, denom)
	}
	for _, pair := range genesis.WhitelistedAddressPairs {
		s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, pair)
	}
	return genesis
}

func (s *MainnetExportTestSuite) populateRecordsFromExport(export strideExport) recordstypes.GenesisState {
	var genesis recordstypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "records"), &genesis))
	for _, record := range genesis.EpochUnbondingRecordList {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, record)
	}
	return genesis
}

// populateDelegationChannelsFromExport registers each zone's open delegation channel the way
// mainnet has it and plants one packet commitment per recorded in-flight packet, so the
// stale-flag reset sees the real picture. Returns the chain ids whose flags must be reset.
func (s *MainnetExportTestSuite) populateDelegationChannelsFromExport(export strideExport, hostZones map[string]stakeibctypes.HostZone) (resettable []string) {
	var channels map[string]delegationChannel
	s.Require().NoError(json.Unmarshal(s.section(export, "delegation_channels"), &channels))

	for chainId, channel := range channels {
		if channel.ChannelId == "" {
			continue
		}
		owner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_DELEGATION)
		s.MockICAChannel(channel.ConnectionId, channel.ChannelId, owner, hostZones[chainId].DelegationIcaAddress)
		portId, _ := icatypes.NewControllerPortID(owner)
		for sequence := 1; sequence <= channel.PacketCommitments; sequence++ {
			s.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(s.Ctx, portId, channel.ChannelId, uint64(sequence), []byte{1})
		}
		if channel.PacketCommitments == 0 {
			resettable = append(resettable, chainId)
		}
	}
	s.Require().NotEmpty(resettable, "at least one zone should have an open channel with nothing in flight")
	return resettable
}

// setSyntheticInProgressFlags puts DelegationChangesInProgress = 1 on the first validator of
// every zone. The fixture's resettable zones already carry zero flags, so without this the
// reset would be unobservable; the flag is synthetic, not mainnet state.
func (s *MainnetExportTestSuite) setSyntheticInProgressFlags(hostZones map[string]stakeibctypes.HostZone) {
	for chainId, hostZone := range hostZones {
		s.Require().NotEmpty(hostZone.Validators, "%s has validators", chainId)
		hostZone.Validators[0].DelegationChangesInProgress = 1
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}
}

// mockDeprecatedZoneChannel opens a delegation channel with zero commitments for a deprecated
// zone, so the reset's Deprecated skip is what keeps its flag (the fixture has no such channel).
func (s *MainnetExportTestSuite) mockDeprecatedZoneChannel(hostZone stakeibctypes.HostZone) {
	owner := stakeibctypes.FormatHostZoneICAOwner(hostZone.ChainId, stakeibctypes.ICAAccountType_DELEGATION)
	s.Require().NotEmpty(hostZone.ConnectionId)
	s.MockICAChannel(hostZone.ConnectionId, "channel-9999", owner, "evmos-delegation-ica")
}

func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	// ----- arrange: seed every section the handler touches from real mainnet state -----
	export := s.loadTrimmedExport()
	hostZones := s.populateStakeibcFromExport(export)
	queriesBefore := s.populateQueriesFromExport(export)
	s.populateAutopilotFromExport(export)
	allowBefore := s.populateICAHostFromExport(export)
	wasmSeed := s.populateWasmFromExport(export)
	oraclesBefore := s.populateOraclesFromExport(export)
	s.populateRateLimitsFromExport(export)
	recordsBefore := s.populateRecordsFromExport(export)
	resettableZones := s.populateDelegationChannelsFromExport(export, hostZones)
	s.setSyntheticInProgressFlags(hostZones)
	s.mockDeprecatedZoneChannel(hostZones[evmosChainId])

	// The fixture's channel picture: an open delegation channel with nothing in flight
	s.Require().ElementsMatch(expectedResetZones, resettableZones, "fixture channel data implies exactly the README's resettable zones")
	s.Require().True(hostZones[evmosChainId].Deprecated, "evmos is deprecated: its open channel must not trigger a reset")

	haqqBefore := hostZones[v35.HaqqChainId]
	flagsBefore := map[string][]int64{}
	for chainId, hostZone := range hostZones {
		flagsBefore[chainId] = delegationChangeFlags(hostZone)
	}

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// ----- assert: entry points -----
	s.Require().False(s.App.AutopilotKeeper.GetParams(s.Ctx).StakeibcActive, "autopilot stakeibc off")
	allowAfter := s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}))
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}))
	s.Require().Contains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}))
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}))
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}))
	s.Require().Contains(allowAfter, sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}))
	s.Require().Len(allowAfter, len(allowBefore)-4, "exactly the two stakeibc and the two staking messages mainnet allows leave the allow-list")

	// ----- assert: wasm -----
	uploadAccess := s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, uploadAccess.Permission)
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, uploadAccess.Addresses, "upload access is gov only")
	for _, address := range wasmSeed.deployKeyContracts {
		info := s.App.WasmKeeper.GetContractInfo(s.Ctx, address)
		s.Require().NotNil(info)
		s.Require().Equal(v35.GovModuleAddress().String(), info.Admin, "contract %s admin moved to gov", address)
	}
	controlInfo := s.App.WasmKeeper.GetContractInfo(s.Ctx, wasmSeed.controlContract)
	s.Require().NotNil(controlInfo)
	s.Require().Equal(wasmSeed.controlAdmin.String(), controlInfo.Admin, "a contract with another admin is untouched")

	// ----- assert: stakeibc state flips -----
	comdex, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().True(comdex.Deprecated, "comdex-1 deprecated")
	s.Require().False(comdex.Halted, "comdex-1 Halted untouched")
	s.Require().Empty(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), "dYdX trade route deleted")

	// ----- assert: oracles and rate limits -----
	for _, before := range oraclesBefore {
		after, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, before.ChainId)
		s.Require().True(found)
		s.Require().False(after.Active, "oracle %s inactive", before.ChainId)
	}
	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx))
	// The handler empties the blacklist. The test app runs stakedym at default genesis, so nothing
	// re-adds `stadym` here; on mainnet the list ends as exactly `stadym`, because stakedym
	// (deprecated, halted, rate above its max bound) re-adds it every block.
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx))

	// ----- assert: stale flags reset exactly where nothing is in flight -----
	for chainId, before := range flagsBefore {
		after, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(found)
		if contains(expectedResetZones, chainId) {
			for _, flag := range delegationChangeFlags(after) {
				s.Require().Zero(flag, "%s: every DelegationChangesInProgress reset", chainId)
			}
			continue
		}
		s.Require().Equal(before, delegationChangeFlags(after), "%s: flags untouched (channel missing, packets in flight or deprecated)", chainId)
	}

	// ----- assert: ICQ purges -----
	remaining := map[string]icqtypes.Query{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		remaining[query.Id] = query
	}
	for _, query := range queriesBefore {
		_, stillThere := remaining[query.Id]
		isHaqqSlashPath := query.ChainId == v35.HaqqChainId && (query.CallbackId == stakeibckeeper.ICQCallbackID_Delegation ||
			query.CallbackId == stakeibckeeper.ICQCallbackID_Validator || query.CallbackId == stakeibckeeper.ICQCallbackID_Calibrate)
		isWithdrawalBalance := query.CallbackId == stakeibckeeper.ICQCallbackID_WithdrawalHostBalance
		isCalibration := query.CallbackId == stakeibckeeper.ICQCallbackID_Calibrate
		shouldPurge := query.CallbackModule == stakeibctypes.ModuleName && (isHaqqSlashPath || isWithdrawalBalance || isCalibration)
		s.Require().Equal(!shouldPurge, stillThere,
			"query %s (%s %s) purge decision", query.Id, query.ChainId, query.CallbackId)
	}
	haqqAfter, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range haqqAfter.Validators {
		s.Require().False(validator.SlashQueryInProgress, "haqq %s slash flag cleared", validator.Name)
	}

	// ----- assert: haqq deltas applied in full with the real table -----
	// The table is pinned to the tracked delegations it was measured against and skipped whole on
	// any mismatch, so check the pin against the fixture first: a mismatch means the table and the
	// fixture come from different heights
	for _, entry := range v35.HaqqDelegationDeltas {
		before, _, found := stakeibckeeper.GetValidatorFromAddress(haqqBefore.Validators, entry.Address)
		s.Require().True(found, "table validator %s is on the mainnet haqq zone", entry.Name)
		s.Require().True(trackedDelegation(before).Equal(v35.HaqqExpectedTrackedDelegations[entry.Address]),
			"haqq %s tracked %s vs pin %s: the table and the fixture are from different heights",
			entry.Name, trackedDelegation(before), v35.HaqqExpectedTrackedDelegations[entry.Address])
	}
	expectedNet := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		before, _, found := stakeibckeeper.GetValidatorFromAddress(haqqBefore.Validators, entry.Address)
		s.Require().True(found, "table validator %s is on the mainnet haqq zone", entry.Name)
		after, _, _ := stakeibckeeper.GetValidatorFromAddress(haqqAfter.Validators, entry.Address)
		s.Require().Equal(trackedDelegation(before).Add(entry.Delta), trackedDelegation(after), "haqq %s delta applied", entry.Name)
		expectedNet = expectedNet.Add(entry.Delta)
	}
	s.Require().True(expectedNet.IsNegative(), "the haqq table nets to a decrease (spec §5)")
	s.Require().Equal(haqqBefore.TotalDelegations.Add(expectedNet), haqqAfter.TotalDelegations, "haqq TotalDelegations moved by the net delta")
	sum := sdkmath.ZeroInt()
	for _, validator := range haqqAfter.Validators {
		sum = sum.Add(trackedDelegation(*validator))
	}
	s.Require().Equal(sum, haqqAfter.TotalDelegations, "haqq TotalDelegations == sum of validators")

	// ----- assert: upgrade path without gov (authority spec §3) -----
	consensusParams, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotNil(consensusParams.Authority, "consensus authority set")
	s.Require().Equal(v35.UpgradeAuthority, consensusParams.Authority.Authority, "consensus authority is the multisig")
	govParams, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(unreachableDeposit().Equal(govParams.MinDeposit), "gov min deposit %s", govParams.MinDeposit)
	s.Require().True(unreachableDeposit().Equal(govParams.ExpeditedMinDeposit), "gov expedited min deposit %s", govParams.ExpeditedMinDeposit)

	// ----- assert: nothing the handler must not touch -----
	// Compared by String(): Equal on structs holding sdkmath.Int is a DeepEqual over big.Int
	// internals and can differ for equal values (v34's suite compares the same way)
	recordsAfter := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().Len(recordsAfter, len(recordsBefore.EpochUnbondingRecordList), "unbonding record count untouched")
	for i, before := range recordsBefore.EpochUnbondingRecordList {
		s.Require().Equal(before.String(), recordsAfter[i].String(), "unbonding record %d untouched", before.EpochNumber)
	}
	for chainId, before := range hostZones {
		after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().Equal(before.RedemptionRate, after.RedemptionRate, "%s redemption rate untouched", chainId)
		s.Require().Equal(before.Halted, after.Halted, "%s Halted untouched", chainId)

		// Only comdex-1 flips to deprecated; only haqq's tracked delegations and total move
		if chainId != v35.ComdexChainId {
			s.Require().Equal(before.Deprecated, after.Deprecated, "%s Deprecated untouched", chainId)
		}
		if chainId == v35.HaqqChainId {
			s.assertHaqqOutsideTableUntouched(before, after)
			continue
		}
		s.Require().True(before.TotalDelegations.Equal(after.TotalDelegations), "%s TotalDelegations untouched", chainId)
		s.Require().Len(after.Validators, len(before.Validators), "%s validator count untouched", chainId)
		for _, validator := range before.Validators {
			afterValidator, _, found := stakeibckeeper.GetValidatorFromAddress(after.Validators, validator.Address)
			s.Require().True(found, "%s validator %s still tracked", chainId, validator.Name)
			s.Require().True(trackedDelegation(*validator).Equal(trackedDelegation(afterValidator)),
				"%s validator %s delegation untouched", chainId, validator.Name)
		}
	}
}

// assertHaqqOutsideTableUntouched checks every haqq validator the delta table does not name.
func (s *MainnetExportTestSuite) assertHaqqOutsideTableUntouched(before, after stakeibctypes.HostZone) {
	inTable := map[string]bool{}
	for _, entry := range v35.HaqqDelegationDeltas {
		inTable[entry.Address] = true
	}
	for _, validator := range before.Validators {
		if inTable[validator.Address] {
			continue
		}
		afterValidator, _, found := stakeibckeeper.GetValidatorFromAddress(after.Validators, validator.Address)
		s.Require().True(found, "haqq validator %s still tracked", validator.Name)
		s.Require().True(trackedDelegation(*validator).Equal(trackedDelegation(afterValidator)),
			"haqq validator %s is outside the delta table and must be unchanged", validator.Name)
	}
}

// trackedDelegation reads a validator's tracked delegation with a nil (never set) as zero, the
// way the handler does.
func trackedDelegation(validator stakeibctypes.Validator) sdkmath.Int {
	if validator.Delegation.IsNil() {
		return sdkmath.ZeroInt()
	}
	return validator.Delegation
}

// DelegationChangesInProgress is an int64 on the proto (validator.pb.go).
func delegationChangeFlags(hostZone stakeibctypes.HostZone) []int64 {
	flags := make([]int64, 0, len(hostZone.Validators))
	for _, validator := range hostZone.Validators {
		flags = append(flags, validator.DelegationChangesInProgress)
	}
	return flags
}

func contains(list []string, item string) bool {
	for _, entry := range list {
		if entry == item {
			return true
		}
	}
	return false
}
