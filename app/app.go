package app

import (
	"io"
	"os"
	"path/filepath"

	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"

	"github.com/Stride-Labs/stride/v9/utils"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	claimvesting "github.com/Stride-Labs/stride/v9/x/claim/vesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v9/x/claim/vesting/types"

	"github.com/Stride-Labs/stride/v9/x/mint"
	mintkeeper "github.com/Stride-Labs/stride/v9/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v9/x/mint/types"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/x/params"
	distrclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/spf13/cast"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	// monitoringp "github.com/tendermint/spn/x/monitoringp"
	// monitoringpkeeper "github.com/tendermint/spn/x/monitoringp/keeper"

	epochsmodule "github.com/Stride-Labs/stride/v9/x/epochs"
	epochsmodulekeeper "github.com/Stride-Labs/stride/v9/x/epochs/keeper"
	epochsmoduletypes "github.com/Stride-Labs/stride/v9/x/epochs/types"

	"github.com/Stride-Labs/stride/v9/x/interchainquery"
	interchainquerykeeper "github.com/Stride-Labs/stride/v9/x/interchainquery/keeper"
	interchainquerytypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"

	"github.com/Stride-Labs/stride/v9/x/autopilot"
	autopilotkeeper "github.com/Stride-Labs/stride/v9/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v9/x/autopilot/types"

	"github.com/Stride-Labs/stride/v9/x/claim"
	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	icacallbacksmodule "github.com/Stride-Labs/stride/v9/x/icacallbacks"
	icacallbacksmodulekeeper "github.com/Stride-Labs/stride/v9/x/icacallbacks/keeper"
	icacallbacksmoduletypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	ratelimitmodule "github.com/Stride-Labs/stride/v9/x/ratelimit"
	ratelimitclient "github.com/Stride-Labs/stride/v9/x/ratelimit/client"
	ratelimitmodulekeeper "github.com/Stride-Labs/stride/v9/x/ratelimit/keeper"
	ratelimitmoduletypes "github.com/Stride-Labs/stride/v9/x/ratelimit/types"
	recordsmodule "github.com/Stride-Labs/stride/v9/x/records"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v9/x/records/keeper"
	recordsmoduletypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibcmodule "github.com/Stride-Labs/stride/v9/x/stakeibc"
	stakeibcclient "github.com/Stride-Labs/stride/v9/x/stakeibc/client"
	stakeibcmodulekeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibcmoduletypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
)

const (
	AccountAddressPrefix = "stride"
	Name                 = "stride"
)

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		stakeibcclient.AddValidatorsProposalHandler,
		ratelimitclient.AddRateLimitProposalHandler,
		ratelimitclient.UpdateRateLimitProposalHandler,
		ratelimitclient.RemoveRateLimitProposalHandler,
		ratelimitclient.ResetRateLimitProposalHandler,
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		claimvesting.AppModuleBasic{},
		// monitoringp.AppModuleBasic{},
		stakeibcmodule.AppModuleBasic{},
		epochsmodule.AppModuleBasic{},
		interchainquery.AppModuleBasic{},
		ica.AppModuleBasic{},
		recordsmodule.AppModuleBasic{},
		ratelimitmodule.AppModuleBasic{},
		icacallbacksmodule.AppModuleBasic{},
		claim.AppModuleBasic{},
		autopilot.AppModuleBasic{},
		tendermint.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		distrtypes.ModuleName:      nil,
		// mint module needs burn access to remove excess validator tokens (it overallocates, then burns)
		minttypes.ModuleName:                    {authtypes.Minter, authtypes.Burner},
		stakingtypes.BondedPoolName:             {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:          {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                     {authtypes.Burner},
		ibctransfertypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
		stakeibcmoduletypes.ModuleName:          {authtypes.Minter, authtypes.Burner, authtypes.Staking},
		claimtypes.ModuleName:                   nil,
		interchainquerytypes.ModuleName:         nil,
		icatypes.ModuleName:                     nil,
		stakeibcmoduletypes.RewardCollectorName: nil,
	}
)

var (
	_ servertypes.Application = (*StrideApp)(nil)
	_ ibctesting.TestingApp   = (*StrideApp)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// StrideApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type StrideApp struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper   evidencekeeper.Keeper
	TransferKeeper   ibctransferkeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	// MonitoringKeeper    monitoringpkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	AutopilotKeeper     autopilotkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	// ScopedMonitoringKeeper capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper

	ScopedStakeibcKeeper capabilitykeeper.ScopedKeeper
	StakeibcKeeper       stakeibcmodulekeeper.Keeper

	EpochsKeeper          epochsmodulekeeper.Keeper
	InterchainqueryKeeper interchainquerykeeper.Keeper
	ScopedRecordsKeeper   capabilitykeeper.ScopedKeeper
	RecordsKeeper         recordsmodulekeeper.Keeper
	IcacallbacksKeeper    icacallbacksmodulekeeper.Keeper
	ScopedratelimitKeeper capabilitykeeper.ScopedKeeper
	RatelimitKeeper       ratelimitmodulekeeper.Keeper
	ClaimKeeper           claimkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	mm           *module.Manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// RUN GOSEC
// New returns a reference to an initialized blockchain app
func NewStrideApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *StrideApp {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey, // monitoringptypes.StoreKey,
		stakeibcmoduletypes.StoreKey,
		autopilottypes.StoreKey,
		epochsmoduletypes.StoreKey,
		interchainquerytypes.StoreKey,
		icacontrollertypes.StoreKey, icahosttypes.StoreKey,
		recordsmoduletypes.StoreKey,
		ratelimitmoduletypes.StoreKey,
		icacallbacksmoduletypes.StoreKey,
		claimtypes.StoreKey,
		crisistypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &StrideApp{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[upgradetypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], authtypes.ProtoBaseAccount, maccPerms, AccountAddressPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.AccountKeeper, app.BlacklistedModuleAccountAddrs(), authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.StakingKeeper = *stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	epochsKeeper := epochsmodulekeeper.NewKeeper(appCodec, keys[epochsmoduletypes.StoreKey])
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.EpochsKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, encodingConfig.Amino, keys[slashingtypes.StoreKey], app.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, keys[crisistypes.StoreKey],
		invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ClaimKeeper = *claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper, app.StakingKeeper, app.DistrKeeper, epochsKeeper)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks(), app.ClaimKeeper.Hooks()),
	)

	// ... other modules keepers

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), app.StakingKeeper, app.UpgradeKeeper, scopedIBCKeeper,
	)

	// Create Ratelimit Keeper
	scopedratelimitKeeper := app.CapabilityKeeper.ScopeToModule(ratelimitmoduletypes.ModuleName)
	app.ScopedratelimitKeeper = scopedratelimitKeeper
	app.RatelimitKeeper = *ratelimitmodulekeeper.NewKeeper(
		appCodec,
		keys[ratelimitmoduletypes.StoreKey],
		app.GetSubspace(ratelimitmoduletypes.ModuleName),
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		// TODO: Implement ICS4Wrapper in Records and pass records keeper here
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
	)
	ratelimitModule := ratelimitmodule.NewAppModule(appCodec, app.RatelimitKeeper)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.RatelimitKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// TODO(TEST-20): look for all lines that include 'monitoring' in this file! there are a few places this
	// is commented out
	// scopedMonitoringKeeper := app.CapabilityKeeper.ScopeToModule(monitoringptypes.ModuleName)
	// app.MonitoringKeeper = *monitoringpkeeper.NewKeeper(
	// 	appCodec,
	// 	keys[monitoringptypes.StoreKey],
	// 	keys[monitoringptypes.MemStoreKey],
	// 	app.GetSubspace(monitoringptypes.ModuleName),
	// 	app.StakingKeeper,
	// 	app.IBCKeeper.ClientKeeper,
	// 	app.IBCKeeper.ConnectionKeeper,
	// 	app.IBCKeeper.ChannelKeeper,
	// 	&app.IBCKeeper.PortKeeper,
	// 	scopedMonitoringKeeper,
	// )
	// monitoringModule := monitoringp.NewAppModule(appCodec, app.MonitoringKeeper)

	// Note: must be above app.StakeibcKeeper
	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey], app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, app.MsgServiceRouter(),
	)

	app.IcacallbacksKeeper = *icacallbacksmodulekeeper.NewKeeper(
		appCodec,
		keys[icacallbacksmoduletypes.StoreKey],
		keys[icacallbacksmoduletypes.MemStoreKey],
		app.GetSubspace(icacallbacksmoduletypes.ModuleName),
		*app.IBCKeeper,
	)

	app.InterchainqueryKeeper = interchainquerykeeper.NewKeeper(appCodec, keys[interchainquerytypes.StoreKey], app.IBCKeeper)
	interchainQueryModule := interchainquery.NewAppModule(appCodec, app.InterchainqueryKeeper)

	app.RecordsKeeper = *recordsmodulekeeper.NewKeeper(
		appCodec,
		keys[recordsmoduletypes.StoreKey],
		keys[recordsmoduletypes.MemStoreKey],
		app.GetSubspace(recordsmoduletypes.ModuleName),
		app.AccountKeeper,
		app.TransferKeeper,
		*app.IBCKeeper,
		app.IcacallbacksKeeper,
	)
	recordsModule := recordsmodule.NewAppModule(appCodec, app.RecordsKeeper, app.AccountKeeper, app.BankKeeper)

	stakeibcKeeper := stakeibcmodulekeeper.NewKeeper(
		appCodec,
		keys[stakeibcmoduletypes.StoreKey],
		keys[stakeibcmoduletypes.MemStoreKey],
		app.GetSubspace(stakeibcmoduletypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.ICAControllerKeeper,
		*app.IBCKeeper,
		app.InterchainqueryKeeper,
		app.RecordsKeeper,
		app.StakingKeeper,
		app.IcacallbacksKeeper,
		app.RatelimitKeeper,
	)
	app.StakeibcKeeper = *stakeibcKeeper.SetHooks(
		stakeibcmoduletypes.NewMultiStakeIBCHooks(app.ClaimKeeper.Hooks()),
	)

	stakeibcModule := stakeibcmodule.NewAppModule(appCodec, app.StakeibcKeeper, app.AccountKeeper, app.BankKeeper)
	stakeibcIBCModule := stakeibcmodule.NewIBCModule(app.StakeibcKeeper)

	app.AutopilotKeeper = *autopilotkeeper.NewKeeper(
		appCodec,
		keys[autopilottypes.StoreKey],
		app.GetSubspace(autopilottypes.ModuleName),
		app.StakeibcKeeper,
		app.ClaimKeeper)
	autopilotModule := autopilot.NewAppModule(appCodec, app.AutopilotKeeper)

	// Register Gov (must be registerd after stakeibc)
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(stakeibcmoduletypes.RouterKey, stakeibcmodule.NewStakeibcProposalHandler(app.StakeibcKeeper)).
		AddRoute(ratelimitmoduletypes.RouterKey, ratelimitmodule.NewRateLimitProposalHandler(app.RatelimitKeeper, app.IBCKeeper.ChannelKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.MsgServiceRouter(), govtypes.DefaultConfig(), authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper

	// Register ICQ callbacks
	err := app.InterchainqueryKeeper.SetCallbackHandler(stakeibcmoduletypes.ModuleName, app.StakeibcKeeper.ICQCallbackHandler())
	if err != nil {
		return nil
	}

	app.EpochsKeeper = *epochsKeeper.SetHooks(
		epochsmoduletypes.NewMultiEpochHooks(
			app.StakeibcKeeper.Hooks(),
			app.MintKeeper.Hooks(),
			app.ClaimKeeper.Hooks(),
			app.RatelimitKeeper.Hooks(),
		),
	)
	epochsModule := epochsmodule.NewAppModule(appCodec, app.EpochsKeeper)

	icacallbacksModule := icacallbacksmodule.NewAppModule(appCodec, app.IcacallbacksKeeper, app.AccountKeeper, app.BankKeeper)

	// Register ICA calllbacks
	// NOTE: The icacallbacks struct implemented below provides a mapping from ICA channel owner to ICACallback handler,
	// where the callback handler stores and routes to the various callback functions for a particular module.
	// However, as of ibc-go v6, the icacontroller module owns the ICA channel. A consequence of this is that there can
	// be no more than one module that implements ICA callbacks. Should we add an new module with ICA support in the future,
	// we'll need to refactor this
	err = app.IcacallbacksKeeper.SetICACallbackHandler(icacontrollertypes.SubModuleName, app.StakeibcKeeper.ICACallbackHandler())
	if err != nil {
		return nil
	}
	err = app.IcacallbacksKeeper.SetICACallbackHandler(ibctransfertypes.ModuleName, app.RecordsKeeper.ICACallbackHandler())
	if err != nil {
		return nil
	}

	// create IBC middleware stacks by combining middleware with base application
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)
	// Create the middleware stacks
	// Stack one (ICAHost Stack) contains:
	// - IBC
	// - ICAHost
	// - base app
	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	// Stack two (Stakeibc Stack) contains
	// - IBC
	// - ICA
	// - stakeibc
	// - base app
	var stakeibcStack porttypes.IBCModule = stakeibcIBCModule
	stakeibcStack = icacontroller.NewIBCMiddleware(stakeibcStack, app.ICAControllerKeeper)

	// Stack three contains
	// - IBC
	// - autopilot
	// - records
	// - ratelimit
	// - transfer
	// - base app
	var transferStack porttypes.IBCModule = transferIBCModule
	transferStack = ratelimitmodule.NewIBCMiddleware(app.RatelimitKeeper, transferStack)
	transferStack = recordsmodule.NewIBCModule(app.RecordsKeeper, transferStack)
	transferStack = autopilot.NewIBCModule(app.AutopilotKeeper, transferStack)

	// Create static IBC router, add transfer route, then set and seal it
	// Two routes are included for the ICAController because of the following procedure when registering an ICA
	//     1. RegisterInterchainAccount binds the new portId to the icacontroller module and initiates a channel opening
	//     2. MsgChanOpenInit is invoked from the IBC message server.  The message server identifies that the
	//        icacontroller module owns the portID and routes to the stakeibc stack (the "icacontroller" route below)
	//     3. The stakeibc stack works top-down, first in the ICAController's OnChanOpenInit, and then in stakeibc's OnChanOpenInit
	//     4. In stakeibc's OnChanOpenInit, the stakeibc module steals the portId from the icacontroller module
	//     5. Now in OnChanOpenAck and any other subsequent IBC callback, the message server will identify
	//        the portID owner as stakeibc and route to the same stakeibcStack, this time using the "stakeibc" route instead
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		// ICAHost Stack
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		// Stakeibc Stack
		AddRoute(icacontrollertypes.SubModuleName, stakeibcStack).
		// Transfer stack
		AddRoute(ibctransfertypes.ModuleName, transferStack)

	app.IBCKeeper.SetRouter(ibcRouter)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		claimvesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, &app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		claim.NewAppModule(appCodec, app.ClaimKeeper),
		transferModule,
		// monitoringModule,
		stakeibcModule,
		epochsModule,
		interchainQueryModule,
		icaModule,
		recordsModule,
		ratelimitModule,
		icacallbacksModule,
		autopilotModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		// monitoringptypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		autopilottypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		// monitoringptypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		autopilottypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		// monitoringptypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		autopilottypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)
	app.setupUpgradeHandlers()

	// create the simulation manager and define the order of the modules for deterministic simulations
	// app.sm = module.NewSimulationManager(
	// 	auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
	// 	bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
	// 	capability.NewAppModule(appCodec, *app.CapabilityKeeper),
	// 	feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
	// 	gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
	// 	mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
	// 	staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
	// 	distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
	// 	slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
	// 	params.NewAppModule(app.ParamsKeeper),
	// 	evidence.NewAppModule(app.EvidenceKeeper),
	// 	ibc.NewAppModule(app.IBCKeeper),
	// 	transferModule,
	// 	// monitoringModule,
	// 	stakeibcModule,
	// 	epochsModule,
	// 	interchainQueryModule,
	// 	recordsModule,
	// icacallbacksModule,
	// )
	// app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// app.ScopedMonitoringKeeper = scopedMonitoringKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper

	return app
}

// Name returns the name of the App
func (app *StrideApp) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *StrideApp) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// GetStakingKeeper implements the TestingApp interface.
func (app *StrideApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *StrideApp) GetTransferKeeper() *ibctransferkeeper.Keeper {
	return &app.TransferKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *StrideApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *StrideApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *StrideApp) GetTxConfig() client.TxConfig {
	cfg := MakeEncodingConfig()
	return cfg.TxConfig
}

// BeginBlocker application updates every begin block
func (app *StrideApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *StrideApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *StrideApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *StrideApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *StrideApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for _, acc := range utils.StringMapKeys(maccPerms) {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *StrideApp) BlacklistedModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for _, acc := range utils.StringMapKeys(maccPerms) {
		// don't blacklist stakeibc module account, so that it can ibc transfer tokens
		if acc == stakeibcmoduletypes.ModuleName || acc == stakeibcmoduletypes.RewardCollectorName {
			continue
		}
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *StrideApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *StrideApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *StrideApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *StrideApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *StrideApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *StrideApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *StrideApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *StrideApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	// paramsKeeper.Subspace(monitoringptypes.ModuleName)
	paramsKeeper.Subspace(stakeibcmoduletypes.ModuleName)
	paramsKeeper.Subspace(epochsmoduletypes.ModuleName)
	paramsKeeper.Subspace(interchainquerytypes.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(recordsmoduletypes.ModuleName)
	paramsKeeper.Subspace(ratelimitmoduletypes.ModuleName)
	paramsKeeper.Subspace(icacallbacksmoduletypes.ModuleName)
	paramsKeeper.Subspace(autopilottypes.ModuleName)

	paramsKeeper.Subspace(claimtypes.ModuleName)
	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *StrideApp) SimulationManager() *module.SimulationManager {
	return app.sm
}
