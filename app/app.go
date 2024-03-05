package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit"
	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
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
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
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
	"github.com/cosmos/cosmos-sdk/x/params"
	distrclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
	ccvconsumer "github.com/cosmos/interchain-security/v3/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v3/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	ccvdistr "github.com/cosmos/interchain-security/v3/x/ccv/democracy/distribution"
	ccvgov "github.com/cosmos/interchain-security/v3/x/ccv/democracy/governance"
	ccvstaking "github.com/cosmos/interchain-security/v3/x/ccv/democracy/staking"
	evmosvesting "github.com/evmos/vesting/x/vesting"
	evmosvestingclient "github.com/evmos/vesting/x/vesting/client"
	evmosvestingkeeper "github.com/evmos/vesting/x/vesting/keeper"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v18/utils"
	"github.com/Stride-Labs/stride/v18/x/autopilot"
	autopilotkeeper "github.com/Stride-Labs/stride/v18/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v18/x/autopilot/types"
	"github.com/Stride-Labs/stride/v18/x/claim"
	claimkeeper "github.com/Stride-Labs/stride/v18/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v18/x/claim/types"
	claimvesting "github.com/Stride-Labs/stride/v18/x/claim/vesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v18/x/claim/vesting/types"
	epochsmodule "github.com/Stride-Labs/stride/v18/x/epochs"
	epochsmodulekeeper "github.com/Stride-Labs/stride/v18/x/epochs/keeper"
	epochsmoduletypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	icacallbacksmodule "github.com/Stride-Labs/stride/v18/x/icacallbacks"
	icacallbacksmodulekeeper "github.com/Stride-Labs/stride/v18/x/icacallbacks/keeper"
	icacallbacksmoduletypes "github.com/Stride-Labs/stride/v18/x/icacallbacks/types"
	icaoracle "github.com/Stride-Labs/stride/v18/x/icaoracle"
	icaoraclekeeper "github.com/Stride-Labs/stride/v18/x/icaoracle/keeper"
	icaoracletypes "github.com/Stride-Labs/stride/v18/x/icaoracle/types"
	"github.com/Stride-Labs/stride/v18/x/interchainquery"
	interchainquerykeeper "github.com/Stride-Labs/stride/v18/x/interchainquery/keeper"
	interchainquerytypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v18/x/mint"
	mintkeeper "github.com/Stride-Labs/stride/v18/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v18/x/mint/types"
	recordsmodule "github.com/Stride-Labs/stride/v18/x/records"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v18/x/records/keeper"
	recordsmoduletypes "github.com/Stride-Labs/stride/v18/x/records/types"
	stakeibcmodule "github.com/Stride-Labs/stride/v18/x/stakeibc"
	stakeibcclient "github.com/Stride-Labs/stride/v18/x/stakeibc/client"
	stakeibcmodulekeeper "github.com/Stride-Labs/stride/v18/x/stakeibc/keeper"
	stakeibcmoduletypes "github.com/Stride-Labs/stride/v18/x/stakeibc/types"
	staketia "github.com/Stride-Labs/stride/v18/x/staketia"
	staketiakeeper "github.com/Stride-Labs/stride/v18/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v18/x/staketia/types"
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
		stakeibcclient.ToggleLSMProposalHandler,
		evmosvestingclient.RegisterClawbackProposalHandler,
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
		ccvstaking.AppModuleBasic{},
		mint.AppModuleBasic{},
		ccvdistr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		consensus.AppModuleBasic{},
		vesting.AppModuleBasic{},
		claimvesting.AppModuleBasic{},
		stakeibcmodule.AppModuleBasic{},
		epochsmodule.AppModuleBasic{},
		interchainquery.AppModuleBasic{},
		ica.AppModuleBasic{},
		recordsmodule.AppModuleBasic{},
		ratelimit.AppModuleBasic{},
		icacallbacksmodule.AppModuleBasic{},
		claim.AppModuleBasic{},
		ccvconsumer.AppModuleBasic{},
		autopilot.AppModuleBasic{},
		icaoracle.AppModuleBasic{},
		tendermint.AppModuleBasic{},
		packetforward.AppModuleBasic{},
		evmosvesting.AppModuleBasic{},
		staketia.AppModuleBasic{},
		wasm.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		distrtypes.ModuleName:      nil,
		// mint module needs burn access to remove excess validator tokens (it overallocates, then burns)
		ccvconsumertypes.ConsumerRedistributeName:     nil,
		ccvconsumertypes.ConsumerToSendToProviderName: nil,
		minttypes.ModuleName:                          {authtypes.Minter, authtypes.Burner},
		stakingtypes.BondedPoolName:                   {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:                {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                           {authtypes.Burner},
		ibctransfertypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		stakeibcmoduletypes.ModuleName:                {authtypes.Minter, authtypes.Burner, authtypes.Staking},
		claimtypes.ModuleName:                         nil,
		interchainquerytypes.ModuleName:               nil,
		icatypes.ModuleName:                           nil,
		stakeibcmoduletypes.RewardCollectorName:       nil,
		staketiatypes.ModuleName:                      {authtypes.Minter, authtypes.Burner},
		staketiatypes.FeeAddress:                      nil,
		wasmtypes.ModuleName:                          {authtypes.Burner},
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
	VestingKeeper         evmosvestingkeeper.Keeper
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	ConsumerKeeper        ccvconsumerkeeper.Keeper
	AutopilotKeeper       autopilotkeeper.Keeper
	PacketForwardKeeper   *packetforwardkeeper.Keeper
	WasmKeeper            wasmkeeper.Keeper
	ContractKeeper        *wasmkeeper.PermissionedKeeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedCCVConsumerKeeper   capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper

	StakeibcKeeper stakeibcmodulekeeper.Keeper

	EpochsKeeper          epochsmodulekeeper.Keeper
	InterchainqueryKeeper interchainquerykeeper.Keeper
	ScopedRecordsKeeper   capabilitykeeper.ScopedKeeper
	RecordsKeeper         recordsmodulekeeper.Keeper
	IcacallbacksKeeper    icacallbacksmodulekeeper.Keeper
	ScopedratelimitKeeper capabilitykeeper.ScopedKeeper
	RatelimitKeeper       ratelimitkeeper.Keeper
	ClaimKeeper           claimkeeper.Keeper
	ICAOracleKeeper       icaoraclekeeper.Keeper
	StaketiaKeeper        staketiakeeper.Keeper

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
	wasmOpts []wasmkeeper.Option,
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
		ratelimittypes.StoreKey,
		icacallbacksmoduletypes.StoreKey,
		claimtypes.StoreKey,
		icaoracletypes.StoreKey,
		ccvconsumertypes.StoreKey,
		crisistypes.StoreKey,
		consensusparamtypes.StoreKey,
		packetforwardtypes.StoreKey,
		evmosvestingtypes.StoreKey,
		staketiatypes.StoreKey,
		wasmtypes.StoreKey,
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

	// TODO: Remove ParamsKeeper after v10 upgrade
	app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		keys[consensusparamtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedCCVConsumerKeeper := app.CapabilityKeeper.ScopeToModule(ccvconsumertypes.ModuleName)

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
		app.StakingKeeper, ccvconsumertypes.ConsumerRedistributeName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, encodingConfig.Amino, keys[slashingtypes.StoreKey], &app.ConsumerKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.ClaimKeeper.Hooks()),
	)

	// ... other modules keepers

	app.ConsumerKeeper = ccvconsumerkeeper.NewNonZeroKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.GetSubspace(ccvconsumertypes.ModuleName),
	)
	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		&app.ConsumerKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)

	// Create Ratelimit Keeper
	scopedratelimitKeeper := app.CapabilityKeeper.ScopeToModule(ratelimittypes.ModuleName)
	app.ScopedratelimitKeeper = scopedratelimitKeeper
	app.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
		appCodec,
		keys[ratelimittypes.StoreKey],
		app.GetSubspace(ratelimittypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
	)
	ratelimitModule := ratelimit.NewAppModule(appCodec, app.RatelimitKeeper)

	// Initialize the packet forward middleware Keeper
	// It's important to note that the PFM Keeper must be initialized before the Transfer Keeper
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		nil, // will be zero-value here, reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.DistrKeeper,
		app.BankKeeper,
		app.RatelimitKeeper, // ICS4Wrapper
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.PacketForwardKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)

	// Set the TransferKeeper reference in the PacketForwardKeeper
	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.ConsumerKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// Create CCV consumer and modules
	app.ConsumerKeeper = ccvconsumerkeeper.NewKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.GetSubspace(ccvconsumertypes.ModuleName),
		scopedCCVConsumerKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		app.SlashingKeeper,
		app.BankKeeper,
		app.AccountKeeper,
		&app.TransferKeeper,
		app.IBCKeeper,
		authtypes.FeeCollectorName,
	)
	app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)

	// register slashing module StakingHooks to the consumer keeper
	app.ConsumerKeeper = *app.ConsumerKeeper.SetHooks(app.SlashingKeeper.Hooks())
	consumerModule := ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))

	// Note: must be above app.StakeibcKeeper and app.ICAOracleKeeper
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

	// Note: Must be above stakeibc keeper
	app.ICAOracleKeeper = *icaoraclekeeper.NewKeeper(
		appCodec,
		keys[icaoracletypes.StoreKey],
		app.GetSubspace(icaoracletypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper - Note: this technically should be ICAController but it doesn't implement ICS4
		app.IBCKeeper.ClientKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.ICAControllerKeeper,
		app.IcacallbacksKeeper,
	)
	icaoracleModule := icaoracle.NewAppModule(appCodec, app.ICAOracleKeeper)

	stakeibcKeeper := stakeibcmodulekeeper.NewKeeper(
		appCodec,
		keys[stakeibcmoduletypes.StoreKey],
		keys[stakeibcmoduletypes.MemStoreKey],
		app.GetSubspace(stakeibcmoduletypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper,
		app.BankKeeper,
		app.ICAControllerKeeper,
		*app.IBCKeeper,
		app.InterchainqueryKeeper,
		app.RecordsKeeper,
		app.StakingKeeper,
		app.IcacallbacksKeeper,
		app.RatelimitKeeper,
		app.ICAOracleKeeper,
		app.ConsumerKeeper,
	)
	app.StakeibcKeeper = *stakeibcKeeper.SetHooks(
		stakeibcmoduletypes.NewMultiStakeIBCHooks(app.ClaimKeeper.Hooks()),
	)

	stakeibcModule := stakeibcmodule.NewAppModule(appCodec, app.StakeibcKeeper, app.AccountKeeper, app.BankKeeper)

	app.AutopilotKeeper = *autopilotkeeper.NewKeeper(
		appCodec,
		keys[autopilottypes.StoreKey],
		app.GetSubspace(autopilottypes.ModuleName),
		app.BankKeeper,
		app.StakeibcKeeper,
		app.ClaimKeeper,
		app.TransferKeeper,
	)
	autopilotModule := autopilot.NewAppModule(appCodec, app.AutopilotKeeper)

	// Staketia Keeper must be initialized after TransferKeeper
	app.StaketiaKeeper = *staketiakeeper.NewKeeper(
		appCodec,
		keys[staketiatypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.ICAOracleKeeper,
		app.RatelimitKeeper,
		app.TransferKeeper,
	)
	stakeTiaModule := staketia.NewAppModule(appCodec, app.StaketiaKeeper)

	app.VestingKeeper = evmosvestingkeeper.NewKeeper(
		keys[evmosvestingtypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName), appCodec,
		app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.StakingKeeper,
	)

	// Add wasm keeper
	wasmCapabilities := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,cosmwasm_1_4"
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)

	// Register Gov (must be registered after stakeibc)
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(stakeibcmoduletypes.RouterKey, stakeibcmodule.NewStakeibcProposalHandler(app.StakeibcKeeper)).
		AddRoute(evmosvestingtypes.RouterKey, evmosvesting.NewVestingProposalHandler(&app.VestingKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.MsgServiceRouter(), govtypes.DefaultConfig(), authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper

	// Register ICQ callbacks
	err = app.InterchainqueryKeeper.SetCallbackHandler(stakeibcmoduletypes.ModuleName, app.StakeibcKeeper.ICQCallbackHandler())
	if err != nil {
		return nil
	}

	app.EpochsKeeper = *epochsKeeper.SetHooks(
		epochsmoduletypes.NewMultiEpochHooks(
			app.StakeibcKeeper.Hooks(),
			app.MintKeeper.Hooks(),
			app.ClaimKeeper.Hooks(),
			app.StaketiaKeeper.Hooks(),
		),
	)
	epochsModule := epochsmodule.NewAppModule(appCodec, app.EpochsKeeper)

	icacallbacksModule := icacallbacksmodule.NewAppModule(appCodec, app.IcacallbacksKeeper, app.AccountKeeper, app.BankKeeper)
	icacallbacksIBCModule := icacallbacksmodule.NewIBCModule(app.IcacallbacksKeeper)

	// Register IBC calllbacks
	if err := app.IcacallbacksKeeper.SetICACallbacks(
		app.StakeibcKeeper.Callbacks(),
		app.RecordsKeeper.Callbacks(),
		app.ICAOracleKeeper.Callbacks(),
	); err != nil {
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

	// Stack two (ICACallbacks Stack) contains
	// - IBC
	// - ICAController
	// - ICAOracle
	// - stakeibc
	// - ICACallbacks
	// - base app
	var icacallbacksStack porttypes.IBCModule = icacallbacksIBCModule
	icacallbacksStack = stakeibcmodule.NewIBCMiddleware(icacallbacksStack, app.StakeibcKeeper)
	icacallbacksStack = icaoracle.NewIBCMiddleware(icacallbacksStack, app.ICAOracleKeeper)
	icacallbacksStack = icacontroller.NewIBCMiddleware(icacallbacksStack, app.ICAControllerKeeper)

	// SendPacket originates from the base app and work up the stack to core IBC
	// RecvPacket originates from core IBC and goes down the stack

	// Stack three contains
	// - core ibc
	// - autopilot
	// - records
	// - staketia
	// - ratelimit
	// - pfm
	// - transfer
	// - base app
	// Note: Traffic up the stack does not pass through records or autopilot,
	// as defined via the ICS4Wrappers of each keeper
	var transferStack porttypes.IBCModule = transferIBCModule
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0, // retries on timeout
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,  // refund timeout
	)
	transferStack = ratelimit.NewIBCMiddleware(app.RatelimitKeeper, transferStack)
	transferStack = staketia.NewIBCMiddleware(app.StaketiaKeeper, transferStack)
	transferStack = recordsmodule.NewIBCModule(app.RecordsKeeper, transferStack)
	transferStack = autopilot.NewIBCModule(app.AutopilotKeeper, transferStack)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		// ICAHost Stack
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		// ICACallbacks Stack
		AddRoute(icacontrollertypes.SubModuleName, icacallbacksStack).
		// Transfer stack
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		// Consumer stack
		AddRoute(ccvconsumertypes.ModuleName, consumerModule)

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
		evmosvesting.NewAppModule(app.VestingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		claimvesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		ccvgov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, IsProposalWhitelisted, app.GetSubspace(govtypes.ModuleName), IsModuleWhiteList),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.ConsumerKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		ccvdistr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, app.GetSubspace(distrtypes.ModuleName)),
		ccvstaking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		claim.NewAppModule(appCodec, app.ClaimKeeper),
		// technically, app.GetSubspace(packetforwardtypes.ModuleName) will never be run https://github.com/cosmos/ibc-apps/issues/146#issuecomment-1839242144
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.BaseApp.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		transferModule,
		// monitoringModule,
		stakeibcModule,
		epochsModule,
		interchainQueryModule,
		icaModule,
		recordsModule,
		ratelimitModule,
		icacallbacksModule,
		consumerModule,
		autopilotModule,
		icaoracleModule,
		stakeTiaModule,
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
		evmosvestingtypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimittypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		ccvconsumertypes.ModuleName,
		autopilottypes.ModuleName,
		icaoracletypes.ModuleName,
		consensusparamtypes.ModuleName,
		packetforwardtypes.ModuleName,
		staketiatypes.ModuleName,
		wasmtypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
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
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		evmosvestingtypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimittypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		ccvconsumertypes.ModuleName,
		autopilottypes.ModuleName,
		icaoracletypes.ModuleName,
		consensusparamtypes.ModuleName,
		packetforwardtypes.ModuleName,
		staketiatypes.ModuleName,
		wasmtypes.ModuleName,
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
		genutiltypes.ModuleName,
		ibchost.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		evmosvestingtypes.ModuleName,
		icatypes.ModuleName,
		stakeibcmoduletypes.ModuleName,
		epochsmoduletypes.ModuleName,
		interchainquerytypes.ModuleName,
		recordsmoduletypes.ModuleName,
		ratelimittypes.ModuleName,
		icacallbacksmoduletypes.ModuleName,
		claimtypes.ModuleName,
		ccvconsumertypes.ModuleName,
		autopilottypes.ModuleName,
		icaoracletypes.ModuleName,
		consensusparamtypes.ModuleName,
		packetforwardtypes.ModuleName,
		staketiatypes.ModuleName,
		wasmtypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)
	app.setupUpgradeHandlers(appOpts)

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

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:         app.IBCKeeper,
			ConsumerKeeper:    app.ConsumerKeeper,
			WasmConfig:        &wasmConfig,
			TXCounterStoreKey: keys[wasmtypes.StoreKey],
			WasmKeeper:        &app.WasmKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	// Register snapshot extensions to enable state-sync for wasm.
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		// Initialize pinned codes in wasmvm as they are not persisted there
		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedCCVConsumerKeeper = scopedCCVConsumerKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper

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
		if acc == stakeibcmoduletypes.ModuleName ||
			acc == stakeibcmoduletypes.RewardCollectorName ||
			acc == ccvconsumertypes.ConsumerToSendToProviderName ||
			acc == staketiatypes.ModuleName ||
			acc == staketiatypes.FeeAddress {
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
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
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
// TODO: Remove ParamsKeeper after v10 upgrade
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable())   //nolint:staticcheck
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())       //nolint:staticcheck
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())     //nolint:staticcheck
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(stakeibcmoduletypes.ModuleName)
	paramsKeeper.Subspace(epochsmoduletypes.ModuleName)
	paramsKeeper.Subspace(interchainquerytypes.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(recordsmoduletypes.ModuleName)
	paramsKeeper.Subspace(ratelimittypes.ModuleName)
	paramsKeeper.Subspace(icacallbacksmoduletypes.ModuleName)
	paramsKeeper.Subspace(ccvconsumertypes.ModuleName)
	paramsKeeper.Subspace(autopilottypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(icaoracletypes.ModuleName)
	paramsKeeper.Subspace(claimtypes.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *StrideApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// ConsumerApp interface implementations for e2e tests
// GetConsumerKeeper implements the ConsumerApp interface.
func (app *StrideApp) GetConsumerKeeper() ccvconsumerkeeper.Keeper {
	return app.ConsumerKeeper
}
