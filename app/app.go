package app

import (
	"io"
	"os"
	"path/filepath"

	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"

	"github.com/Stride-Labs/stride/v4/utils"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
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

	claimvesting "github.com/Stride-Labs/stride/v4/x/claim/vesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v4/x/claim/vesting/types"

	"github.com/Stride-Labs/stride/v4/x/mint"
	mintkeeper "github.com/Stride-Labs/stride/v4/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v4/x/mint/types"

	"github.com/cosmos/cosmos-sdk/x/params"
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
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v3/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	// monitoringp "github.com/tendermint/spn/x/monitoringp"
	// monitoringpkeeper "github.com/tendermint/spn/x/monitoringp/keeper"

	epochsmodule "github.com/Stride-Labs/stride/v4/x/epochs"
	epochsmodulekeeper "github.com/Stride-Labs/stride/v4/x/epochs/keeper"
	epochsmoduletypes "github.com/Stride-Labs/stride/v4/x/epochs/types"

	"github.com/Stride-Labs/stride/v4/x/interchainquery"
	interchainquerykeeper "github.com/Stride-Labs/stride/v4/x/interchainquery/keeper"
	interchainquerytypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"

	"github.com/Stride-Labs/stride/v4/x/claim"
	claimkeeper "github.com/Stride-Labs/stride/v4/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacksmodule "github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbacksmodulekeeper "github.com/Stride-Labs/stride/v4/x/icacallbacks/keeper"
	icacallbacksmoduletypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordsmodule "github.com/Stride-Labs/stride/v4/x/records"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	recordsmoduletypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibcmodule "github.com/Stride-Labs/stride/v4/x/stakeibc"
	stakeibcclient "github.com/Stride-Labs/stride/v4/x/stakeibc/client"
	stakeibcmodulekeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibcmoduletypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	// this line is used by starport scaffolding # stargate/app/moduleImport
)

const (
	AccountAddressPrefix = "stride"
	Name                 = "stride"
	Version              = "4.0.3"
)

// this line is used by starport scaffolding # stargate/wasm/app/enabledProposals

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.ProposalHandler,
		upgradeclient.CancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		stakeibcclient.AddValidatorProposalHandler,
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
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
		authzmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()...),
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
		icacallbacksmodule.AppModuleBasic{},
		claim.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		distrtypes.ModuleName:      nil,
		// mint module needs burn access to remove excess validator tokens (it overallocates, then burns)
		minttypes.ModuleName:            {authtypes.Minter, authtypes.Burner},
		stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:             {authtypes.Burner},
		ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
		stakeibcmoduletypes.ModuleName:  {authtypes.Minter, authtypes.Burner, authtypes.Staking},
		claimtypes.ModuleName:           nil,
		interchainquerytypes.ModuleName: nil,
		icatypes.ModuleName:             nil,
		// this line is used by starport scaffolding # stargate/app/maccPerms
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
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	AuthzKeeper      authzkeeper.Keeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper   evidencekeeper.Keeper
	TransferKeeper   ibctransferkeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	// MonitoringKeeper    monitoringpkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	// ScopedMonitoringKeeper capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper

	ScopedStakeibcKeeper capabilitykeeper.ScopedKeeper
	StakeibcKeeper       stakeibcmodulekeeper.Keeper

	EpochsKeeper             epochsmodulekeeper.Keeper
	InterchainqueryKeeper    interchainquerykeeper.Keeper
	ScopedRecordsKeeper      capabilitykeeper.ScopedKeeper
	RecordsKeeper            recordsmodulekeeper.Keeper
	ScopedIcacallbacksKeeper capabilitykeeper.ScopedKeeper
	IcacallbacksKeeper       icacallbacksmodulekeeper.Keeper
	ClaimKeeper              claimkeeper.Keeper
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

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
	bApp.SetVersion(Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey, // monitoringptypes.StoreKey,
		stakeibcmoduletypes.StoreKey,
		epochsmoduletypes.StoreKey,
		interchainquerytypes.StoreKey,
		icacontrollertypes.StoreKey, icahosttypes.StoreKey,
		recordsmoduletypes.StoreKey,
		icacallbacksmoduletypes.StoreKey,
		authzkeeper.StoreKey,
		claimtypes.StoreKey,
		// this line is used by starport scaffolding # stargate/app/storeKey
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
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	// this line is used by starport scaffolding # stargate/app/scopedKeeper

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		app.BaseApp.MsgServiceRouter(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.AccountKeeper, app.GetSubspace(banktypes.ModuleName), app.BlacklistedModuleAccountAddrs(),
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)
	epochsKeeper := epochsmodulekeeper.NewKeeper(appCodec, keys[epochsmoduletypes.StoreKey])
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.EpochsKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)

	app.ClaimKeeper = *claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper, app.StakingKeeper, app.DistrKeeper, epochsKeeper)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks(), app.ClaimKeeper.Hooks()),
	)

	// ... other modules keepers

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), app.StakingKeeper, app.UpgradeKeeper, scopedIBCKeeper,
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, app.BankKeeper, scopedTransferKeeper,
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

	scopedIcacallbacksKeeper := app.CapabilityKeeper.ScopeToModule(icacallbacksmoduletypes.ModuleName)
	app.ScopedIcacallbacksKeeper = scopedIcacallbacksKeeper
	app.IcacallbacksKeeper = *icacallbacksmodulekeeper.NewKeeper(
		appCodec,
		keys[icacallbacksmoduletypes.StoreKey],
		keys[icacallbacksmoduletypes.MemStoreKey],
		app.GetSubspace(icacallbacksmoduletypes.ModuleName),
		scopedIcacallbacksKeeper,
		*app.IBCKeeper,
		app.ICAControllerKeeper,
	)

	app.InterchainqueryKeeper = interchainquerykeeper.NewKeeper(appCodec, keys[interchainquerytypes.StoreKey], app.IBCKeeper)
	interchainQueryModule := interchainquery.NewAppModule(appCodec, app.InterchainqueryKeeper)

	scopedRecordsKeeper := app.CapabilityKeeper.ScopeToModule(recordsmoduletypes.ModuleName)
	app.ScopedRecordsKeeper = scopedRecordsKeeper
	app.RecordsKeeper = *recordsmodulekeeper.NewKeeper(
		appCodec,
		keys[recordsmoduletypes.StoreKey],
		keys[recordsmoduletypes.MemStoreKey],
		app.GetSubspace(recordsmoduletypes.ModuleName),
		scopedRecordsKeeper,
		app.AccountKeeper,
		app.TransferKeeper,
		*app.IBCKeeper,
		app.IcacallbacksKeeper,
	)
	recordsModule := recordsmodule.NewAppModule(appCodec, app.RecordsKeeper, app.AccountKeeper, app.BankKeeper)

	scopedStakeibcKeeper := app.CapabilityKeeper.ScopeToModule(stakeibcmoduletypes.ModuleName)
	app.ScopedStakeibcKeeper = scopedStakeibcKeeper
	stakeibcKeeper := stakeibcmodulekeeper.NewKeeper(
		appCodec,
		keys[stakeibcmoduletypes.StoreKey],
		keys[stakeibcmoduletypes.MemStoreKey],
		app.GetSubspace(stakeibcmoduletypes.ModuleName),
		// app.IBCKeeper.ChannelKeeper,
		// &app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ICAControllerKeeper,
		*app.IBCKeeper,
		scopedStakeibcKeeper,
		app.InterchainqueryKeeper,
		app.RecordsKeeper,
		app.StakingKeeper,
		app.IcacallbacksKeeper,
	)
	app.StakeibcKeeper = *stakeibcKeeper.SetHooks(
		stakeibcmoduletypes.NewMultiStakeIBCHooks(app.ClaimKeeper.Hooks()),
	)

	stakeibcModule := stakeibcmodule.NewAppModule(appCodec, app.StakeibcKeeper, app.AccountKeeper, app.BankKeeper)
	stakeibcIBCModule := stakeibcmodule.NewIBCModule(app.StakeibcKeeper)

	// Register Gov (must be registerd after stakeibc)
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(stakeibcmoduletypes.RouterKey, stakeibcmodule.NewStakeibcProposalHandler(app.StakeibcKeeper))

	app.GovKeeper = govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, govRouter,
	)

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
		),
	)
	epochsModule := epochsmodule.NewAppModule(appCodec, app.EpochsKeeper)

	icacallbacksModule := icacallbacksmodule.NewAppModule(appCodec, app.IcacallbacksKeeper, app.AccountKeeper, app.BankKeeper)
	// Register ICA calllbacks
	// stakeibc
	err = app.IcacallbacksKeeper.SetICACallbackHandler(stakeibcmoduletypes.ModuleName, app.StakeibcKeeper.ICACallbackHandler())
	if err != nil {
		return nil
	}
	// records
	err = app.IcacallbacksKeeper.SetICACallbackHandler(recordsmoduletypes.ModuleName, app.RecordsKeeper.ICACallbackHandler())
	if err != nil {
		return nil
	}

	// this line is used by starport scaffolding # stargate/app/keeperDefinition

	// create IBC middleware stacks by combining middleware with base application
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, scopedICAHostKeeper, app.MsgServiceRouter(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)

	// Create the middleware stacks

	// Stack one contains
	// - IBC
	// - ICA
	// - icacallbacks
	// - stakeibc
	// - base app
	var icamiddlewareStack porttypes.IBCModule
	icamiddlewareStack = icacallbacksmodule.NewIBCModule(app.IcacallbacksKeeper, stakeibcIBCModule)
	icamiddlewareStack = icacontroller.NewIBCModule(app.ICAControllerKeeper, icamiddlewareStack)
	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	// Stack two contains
	// - IBC
	// - records
	// - transfer
	// - base app
	recordsStack := recordsmodule.NewIBCModule(app.RecordsKeeper, transferIBCModule)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, recordsStack).
		AddRoute(icacontrollertypes.SubModuleName, icamiddlewareStack).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		// Note, authentication module packets are routed to the top level of the middleware stack
		AddRoute(stakeibcmoduletypes.ModuleName, icamiddlewareStack).
		AddRoute(icacallbacksmoduletypes.ModuleName, icamiddlewareStack)
	// this line is used by starport scaffolding # ibc/app/router
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
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		claimvesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
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
		icacallbacksModule,
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		// this line is used by starport scaffolding # stargate/app/appModule
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
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		authz.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
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
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		authz.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
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
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		authz.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
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
	// this line is used by starport scaffolding # stargate/app/appModule
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
	app.ScopedStakeibcKeeper = scopedStakeibcKeeper
	// this line is used by starport scaffolding # stargate/app/beforeInitReturn

	return app
}

// Name returns the name of the App
func (app *StrideApp) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *StrideApp) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// GetStakingKeeper implements the TestingApp interface.
func (app *StrideApp) GetStakingKeeper() stakingkeeper.Keeper {
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
		if acc == "stakeibc" {
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
func (app *StrideApp) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *StrideApp) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
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
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *StrideApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *StrideApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
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
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
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
	paramsKeeper.Subspace(icacallbacksmoduletypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	paramsKeeper.Subspace(claimtypes.ModuleName)
	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *StrideApp) SimulationManager() *module.SimulationManager {
	return app.sm
}
