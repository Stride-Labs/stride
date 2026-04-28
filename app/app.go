package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v3"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	tmos "github.com/cometbft/cometbft/libs/os"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
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
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v11"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v11/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v11/types"
	ibcwasm "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v11"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v11/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v11/types"
	ica "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v11/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v11/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v11/modules/core"
	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v11/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v11/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v11/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
	ccvconsumer "github.com/cosmos/interchain-security/v7/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	ccvdistr "github.com/cosmos/interchain-security/v7/x/ccv/democracy/distribution"
	ccvstaking "github.com/cosmos/interchain-security/v7/x/ccv/democracy/staking"
	evmosvesting "github.com/evmos/vesting/x/vesting"
	evmosvestingkeeper "github.com/evmos/vesting/x/vesting/keeper"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v32/utils"
	airdrop "github.com/Stride-Labs/stride/v32/x/airdrop"
	airdropkeeper "github.com/Stride-Labs/stride/v32/x/airdrop/keeper"
	airdroptypes "github.com/Stride-Labs/stride/v32/x/airdrop/types"
	auction "github.com/Stride-Labs/stride/v32/x/auction"
	auctionkeeper "github.com/Stride-Labs/stride/v32/x/auction/keeper"
	auctiontypes "github.com/Stride-Labs/stride/v32/x/auction/types"
	"github.com/Stride-Labs/stride/v32/x/autopilot"
	autopilotkeeper "github.com/Stride-Labs/stride/v32/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v32/x/autopilot/types"
	"github.com/Stride-Labs/stride/v32/x/claim"
	claimkeeper "github.com/Stride-Labs/stride/v32/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v32/x/claim/types"
	claimvesting "github.com/Stride-Labs/stride/v32/x/claim/vesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v32/x/claim/vesting/types"
	epochsmodule "github.com/Stride-Labs/stride/v32/x/epochs"
	epochsmodulekeeper "github.com/Stride-Labs/stride/v32/x/epochs/keeper"
	epochsmoduletypes "github.com/Stride-Labs/stride/v32/x/epochs/types"
	icacallbacksmodule "github.com/Stride-Labs/stride/v32/x/icacallbacks"
	icacallbacksmodulekeeper "github.com/Stride-Labs/stride/v32/x/icacallbacks/keeper"
	icacallbacksmoduletypes "github.com/Stride-Labs/stride/v32/x/icacallbacks/types"
	icaoracle "github.com/Stride-Labs/stride/v32/x/icaoracle"
	icaoraclekeeper "github.com/Stride-Labs/stride/v32/x/icaoracle/keeper"
	icaoracletypes "github.com/Stride-Labs/stride/v32/x/icaoracle/types"
	icqoracle "github.com/Stride-Labs/stride/v32/x/icqoracle"
	icqoraclekeeper "github.com/Stride-Labs/stride/v32/x/icqoracle/keeper"
	icqoracletypes "github.com/Stride-Labs/stride/v32/x/icqoracle/types"
	"github.com/Stride-Labs/stride/v32/x/interchainquery"
	interchainquerykeeper "github.com/Stride-Labs/stride/v32/x/interchainquery/keeper"
	interchainquerytypes "github.com/Stride-Labs/stride/v32/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v32/x/mint"
	mintkeeper "github.com/Stride-Labs/stride/v32/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v32/x/mint/types"
	recordsmodule "github.com/Stride-Labs/stride/v32/x/records"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v32/x/records/keeper"
	recordsmoduletypes "github.com/Stride-Labs/stride/v32/x/records/types"
	stakedym "github.com/Stride-Labs/stride/v32/x/stakedym"
	stakedymkeeper "github.com/Stride-Labs/stride/v32/x/stakedym/keeper"
	stakedymtypes "github.com/Stride-Labs/stride/v32/x/stakedym/types"
	stakeibcmodule "github.com/Stride-Labs/stride/v32/x/stakeibc"
	stakeibcclient "github.com/Stride-Labs/stride/v32/x/stakeibc/client"
	stakeibcmodulekeeper "github.com/Stride-Labs/stride/v32/x/stakeibc/keeper"
	stakeibcmoduletypes "github.com/Stride-Labs/stride/v32/x/stakeibc/types"
	staketia "github.com/Stride-Labs/stride/v32/x/staketia"
	staketiakeeper "github.com/Stride-Labs/stride/v32/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v32/x/staketia/types"
	strdburner "github.com/Stride-Labs/stride/v32/x/strdburner"
	strdburnerkeeper "github.com/Stride-Labs/stride/v32/x/strdburner/keeper"
	strdburnertypes "github.com/Stride-Labs/stride/v32/x/strdburner/types"
)

const (
	AccountAddressPrefix = "stride"
	Name                 = "stride"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	// mint module needs burn access to remove excess validator tokens (it overallocates, then burns)
	// strdburner module needs burn access to burn STRD tokens that are sent to it
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:                    nil,
		distrtypes.ModuleName:                         nil,
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
		stakedymtypes.ModuleName:                      {authtypes.Minter, authtypes.Burner},
		stakedymtypes.FeeAddress:                      nil,
		wasmtypes.ModuleName:                          {authtypes.Burner},
		icqoracletypes.ModuleName:                     nil,
		auctiontypes.ModuleName:                       nil,
		strdburnertypes.ModuleName:                    {authtypes.Burner},
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

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys  map[string]*storetypes.KVStoreKey
	tkeys map[string]*storetypes.TransientStoreKey

	// Dependency keepers
	// Note: IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	VestingKeeper         evmosvestingkeeper.Keeper
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper //nolint:staticcheck // SA1019: x/params removal scheduled for follow-up
	IBCKeeper             *ibckeeper.Keeper
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
	IBCHooksKeeper        ibchookskeeper.Keeper
	WasmClientKeeper      ibcwasmkeeper.Keeper

	// Middleware for IBCHooks
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware

	// Stride keepers
	StakeibcKeeper        stakeibcmodulekeeper.Keeper
	EpochsKeeper          epochsmodulekeeper.Keeper
	InterchainqueryKeeper interchainquerykeeper.Keeper
	RecordsKeeper         recordsmodulekeeper.Keeper
	IcacallbacksKeeper    icacallbacksmodulekeeper.Keeper
	RatelimitKeeper       ratelimitkeeper.Keeper
	ClaimKeeper           claimkeeper.Keeper
	ICAOracleKeeper       icaoraclekeeper.Keeper
	StaketiaKeeper        staketiakeeper.Keeper
	StakedymKeeper        stakedymkeeper.Keeper
	AirdropKeeper         airdropkeeper.Keeper
	ICQOracleKeeper       icqoraclekeeper.Keeper
	AuctionKeeper         auctionkeeper.Keeper
	StrdBurnerKeeper      strdburnerkeeper.Keeper

	// Module managers
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// module configurator
	configurator module.Configurator
}

// RUN GOSEC
// New returns a reference to an initialized blockchain app
func NewStrideApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *StrideApp {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibcexported.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		// monitoringptypes.StoreKey,
		stakeibcmoduletypes.StoreKey,
		autopilottypes.StoreKey,
		epochsmoduletypes.StoreKey,
		interchainquerytypes.StoreKey,
		icacontrollertypes.StoreKey,
		icahosttypes.StoreKey,
		recordsmoduletypes.StoreKey,
		ratelimittypes.StoreKey,
		icacallbacksmoduletypes.StoreKey,
		claimtypes.StoreKey,
		icaoracletypes.StoreKey,
		ccvconsumertypes.StoreKey,
		consensusparamtypes.StoreKey,
		packetforwardtypes.StoreKey,
		evmosvestingtypes.StoreKey,
		staketiatypes.StoreKey,
		stakedymtypes.StoreKey,
		wasmtypes.StoreKey,
		ibchookstypes.StoreKey,
		ibcwasmtypes.StoreKey,
		airdroptypes.StoreKey,
		icqoracletypes.StoreKey,
		auctiontypes.StoreKey,
		strdburnertypes.StoreKey,
	)

	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &StrideApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
	}

	// TODO: Remove ParamsKeeper after v10 upgrade
	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.BlacklistedModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		bApp.Logger(),
	)
	app.StakingKeeper = *stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	epochsKeeper := epochsmodulekeeper.NewKeeper(
		appCodec,
		keys[epochsmoduletypes.StoreKey],
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.GetSubspace(minttypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.EpochsKeeper,
		authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		ccvconsumertypes.ConsumerRedistributeName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ClaimKeeper = *claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		epochsKeeper,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)

	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.ClaimKeeper.Hooks()),
	)

	// Add airdrop keeper
	app.AirdropKeeper = airdropkeeper.NewKeeper(
		appCodec,
		keys[airdroptypes.StoreKey],
		app.BankKeeper,
	)
	airdropModule := airdrop.NewAppModule(appCodec, app.AirdropKeeper)

	// Add ICS Consumer Keeper
	app.ConsumerKeeper = ccvconsumerkeeper.NewNonZeroKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
	)

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ibcexported.StoreKey]),
		app.UpgradeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the tendermint light client module with the client keeper
	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := clientKeeper.GetStoreProvider()
	tmLightClientModule := ibctm.NewLightClientModule(appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &tmLightClientModule)

	// Create IBC Hooks Keeper (must be defined after the wasmKeeper is created)
	app.IBCHooksKeeper = ibchookskeeper.NewKeeper(
		keys[ibchookstypes.StoreKey],
	)
	// v11 NewWasmHooks returns *WasmHooks (was a value in v10).
	app.Ics20WasmHooks = ibchooks.NewWasmHooks(&app.IBCHooksKeeper, nil, AccountAddressPrefix) // the contract keeper is set later
	app.Ics20WasmHooks.ContractKeeper = &app.WasmKeeper                                        // wasm keeper initialized below

	// Create Ratelimit Keeper
	// v11 dropped the legacy paramtypes.Subspace argument.
	app.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ratelimittypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ClientKeeper,
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
	)
	ratelimitModule := ratelimit.NewAppModule(appCodec, &app.RatelimitKeeper)

	// Create the ICS4 wrapper which routes up the stack from ibchooks -> ratelimit
	// (see full stack definition below)
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.RatelimitKeeper, // ICS4Wrapper
		app.Ics20WasmHooks,
	)

	// Initialize the packet forward middleware Keeper
	// It's important to note that the PFM Keeper must be initialized before the Transfer Keeper
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[packetforwardtypes.StoreKey]),
		nil, // will be zero-value here, reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.HooksICS4Wrapper, // ICS4Wrapper
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create Transfer Keepers
	// ICS4Wrapper is PacketForwardKeeper so outbound sends route up through pfm → ibchooks → ratelimit → core IBC.
	// The autopilot/records/staketia/stakedym middlewares short-circuit SendPacket and must not sit above the
	// TransferKeeper on the outbound path, so we do NOT overwrite this with the full transferStack below.
	app.TransferKeeper = *ibctransferkeeper.NewKeeper(
		appCodec,
		app.AccountKeeper.AddressCodec(),
		runtime.NewKVStoreService(keys[ibctransfertypes.StoreKey]),
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// Override the default ICS4Wrapper so outbound sends route up through pfm → ibchooks → ratelimit → core IBC.
	app.TransferKeeper.WithICS4Wrapper(app.PacketForwardKeeper)

	// Set the TransferKeeper reference in the PacketForwardKeeper
	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	// Add wasm keeper and wasm client keeper (must be after IBCKeeper and TransferKeeper)
	wasmContractMemoryLimit := uint32(32)
	wasmDir := filepath.Join(homePath, "wasm")
	wasmVmDir := filepath.Join(wasmDir, "wasm")
	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	wasmer, err := wasmvm.NewVM(
		wasmVmDir,
		wasmkeeper.BuiltInCapabilities(),
		wasmContractMemoryLimit,
		wasmConfig.ContractDebugMode,
		wasmConfig.MemoryCacheSize,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize WasmVM: %v", err))
	}
	wasmOpts = append(wasmOpts, wasmkeeper.WithWasmEngine(wasmer))

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeperV2,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmtypes.VMConfig{},
		wasmkeeper.BuiltInCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)
	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)

	app.WasmClientKeeper = ibcwasmkeeper.NewKeeperWithVM(
		appCodec,
		runtime.NewKVStoreService(keys[ibcwasmtypes.StoreKey]),
		app.IBCKeeper.ClientKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmer,
		app.GRPCQueryRouter(),
	)

	// Register the 08-wasm light client module with the client keeper
	wasmLightClientModule := ibcwasm.NewLightClientModule(app.WasmClientKeeper, app.IBCKeeper.ClientKeeper.GetStoreProvider())
	app.IBCKeeper.ClientKeeper.AddRoute(ibcwasmtypes.ModuleName, &wasmLightClientModule)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		&app.ConsumerKeeper,
		app.SlashingKeeper,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// Create CCV consumer and modules
	app.ConsumerKeeper = ccvconsumerkeeper.NewKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		app.SlashingKeeper,
		app.BankKeeper,
		app.AccountKeeper,
		&app.TransferKeeper,
		app.IBCKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)

	// register slashing module StakingHooks to the consumer keeper
	app.ConsumerKeeper = *app.ConsumerKeeper.SetHooks(app.SlashingKeeper.Hooks())
	consumerModule := ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))

	// Note: must be above app.StakeibcKeeper and app.ICAOracleKeeper
	app.ICAControllerKeeper = *icacontrollerkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[icacontrollertypes.StoreKey]),
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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
		app.RecordsKeeper,
		app.StakeibcKeeper,
		app.TransferKeeper,
	)
	stakeTiaModule := staketia.NewAppModule(appCodec, app.StaketiaKeeper)

	// Stakedym Keeper must be initialized after TransferKeeper
	app.StakedymKeeper = *stakedymkeeper.NewKeeper(
		appCodec,
		keys[stakedymtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.ICAOracleKeeper,
		app.RatelimitKeeper,
		app.TransferKeeper,
	)
	stakeDymModule := stakedym.NewAppModule(appCodec, app.StakedymKeeper)

	app.VestingKeeper = evmosvestingkeeper.NewKeeper(
		keys[evmosvestingtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName), appCodec,
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.StakingKeeper,
	)

	app.ICQOracleKeeper = *icqoraclekeeper.NewKeeper(
		appCodec,
		keys[icqoracletypes.StoreKey],
		&app.InterchainqueryKeeper,
		app.TransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	icqOracleModule := icqoracle.NewAppModule(appCodec, app.ICQOracleKeeper)

	app.AuctionKeeper = *auctionkeeper.NewKeeper(
		appCodec,
		keys[auctiontypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.ICQOracleKeeper,
	)
	auctionModule := auction.NewAppModule(appCodec, app.AuctionKeeper)

	app.StrdBurnerKeeper = *strdburnerkeeper.NewKeeper(
		appCodec,
		keys[strdburnertypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
	)
	strdburnerModule := strdburner.NewAppModule(appCodec, app.StrdBurnerKeeper)

	// Register Gov (must be registered after stakeibc)
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(stakeibcmoduletypes.RouterKey, stakeibcmodule.NewStakeibcProposalHandler(app.StakeibcKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(app.StakingKeeper),
	)
	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper

	// Register ICQ callbacks
	err = app.InterchainqueryKeeper.SetCallbackHandler(stakeibcmoduletypes.ModuleName, app.StakeibcKeeper.ICQCallbackHandler())
	if err != nil {
		return nil
	}
	err = app.InterchainqueryKeeper.SetCallbackHandler(icqoracletypes.ModuleName, app.ICQOracleKeeper.ICQCallbackHandler())
	if err != nil {
		return nil
	}

	app.EpochsKeeper = *epochsKeeper.SetHooks(
		epochsmoduletypes.NewMultiEpochHooks(
			app.StakeibcKeeper.Hooks(),
			app.MintKeeper.Hooks(),
			app.ClaimKeeper.Hooks(),
			app.StaketiaKeeper.Hooks(),
			app.StakedymKeeper.Hooks(),
		),
	)
	epochsModule := epochsmodule.NewAppModule(appCodec, app.EpochsKeeper)

	icacallbacksModule := icacallbacksmodule.NewAppModule(
		appCodec,
		app.IcacallbacksKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)
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
	app.ICAHostKeeper = *icahostkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[icahosttypes.StoreKey]),
		app.IBCKeeper.ChannelKeeper,
		app.AccountKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)

	// Create the middleware stacks
	// Stack one (ICAHost Stack) contains:
	// - IBC
	// - ICAHost
	// - base app
	icaHostIBCModule := icahost.NewIBCModule(&app.ICAHostKeeper)

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
	icacallbacksStack = icacontroller.NewIBCMiddlewareWithAuth(icacallbacksStack, &app.ICAControllerKeeper)

	// SendPacket originates from the base app and work up the stack to core IBC
	// RecvPacket originates from core IBC and goes down the stack

	// Stack three contains
	// - core ibc
	// - autopilot
	// - records
	// - staketia
	// - stakedym
	// - ratelimit
	// - ibchooks
	// - pfm
	// - transfer
	// - base app
	// Note: Traffic up the stack does not pass through records or autopilot,
	// as defined via the ICS4Wrappers of each keeper
	var transferStack porttypes.IBCModule = transfer.NewIBCModule(&app.TransferKeeper)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0, // retries on timeout
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout
	)
	transferStack = ibchooks.NewIBCMiddleware(transferStack, &app.HooksICS4Wrapper)
	transferStack = ratelimit.NewIBCMiddleware(&app.RatelimitKeeper, transferStack)
	transferStack = staketia.NewIBCMiddleware(app.StaketiaKeeper, transferStack)
	transferStack = stakedym.NewIBCMiddleware(app.StakedymKeeper, transferStack)
	transferStack = recordsmodule.NewIBCModule(app.RecordsKeeper, transferStack)
	transferStack = autopilot.NewIBCModule(app.AutopilotKeeper, transferStack)

	// Do NOT call TransferKeeper.WithICS4Wrapper(transferStack) here — see the comment at
	// TransferKeeper construction for why PacketForwardKeeper must remain the ICS4Wrapper.

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

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig),
		evmosvesting.NewAppModule(app.VestingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		claimvesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.ConsumerKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		ccvdistr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, app.GetSubspace(distrtypes.ModuleName)),
		ccvstaking.NewAppModule(appCodec, &app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper), //nolint:staticcheck // SA1019: x/params removal scheduled for follow-up
		claim.NewAppModule(appCodec, app.ClaimKeeper),
		// technically, app.GetSubspace(packetforwardtypes.ModuleName) will never be run https://github.com/cosmos/ibc-apps/issues/146#issuecomment-1839242144
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.BaseApp.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibchooks.NewAppModule(app.AccountKeeper),
		transfer.NewAppModule(&app.TransferKeeper),
		ibcwasm.NewAppModule(app.WasmClientKeeper),
		ibctm.NewAppModule(tmLightClientModule),
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
		stakeDymModule,
		airdropModule,
		stakeTiaModule,
		icqOracleModule,
		auctionModule,
		strdburnerModule,
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
					distrclient.ProposalHandler,
					stakeibcclient.AddValidatorsProposalHandler,
					stakeibcclient.ToggleLSMProposalHandler,
				},
			),
		})

	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authtypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
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
		stakedymtypes.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		ibcwasmtypes.ModuleName,
		airdroptypes.ModuleName,
		icqoracletypes.ModuleName,
		auctiontypes.ModuleName,
		strdburnertypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		minttypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibcexported.ModuleName,
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
		stakedymtypes.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		ibcwasmtypes.ModuleName,
		airdroptypes.ModuleName,
		icqoracletypes.ModuleName,
		auctiontypes.ModuleName,
		strdburnertypes.ModuleName, // strdburner should be last
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.ModuleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		claimvestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		ibcexported.ModuleName,
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
		stakedymtypes.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		ibcwasmtypes.ModuleName,
		airdroptypes.ModuleName,
		icqoracletypes.ModuleName,
		auctiontypes.ModuleName,
		strdburnertypes.ModuleName,
	)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		return nil
	}
	app.setupUpgradeHandlers(appOpts)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(txConfig, wasmConfig, keys[wasmtypes.StoreKey])
	app.setPostHandler()
	app.SetPrecommiter(app.Precommitter)
	app.SetPrepareCheckStater(app.PrepareCheckStater)

	// Register AutoCLI query service to automatically generate query endpoints for all modules
	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	// Setup gRPC reflection, enabling runtime service/method discovery for clients
	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

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
			panic(fmt.Errorf("error loading last version: %w", err))
		}

		// Initialize pinned codes in wasmvm as they are not persisted there
		ctx := app.BaseApp.NewContext(true)
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
		if err := app.WasmClientKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("WasmClientKeeper failed initialize pinned codes %s", err))
		}
	}

	return app
}

// Initialize a local testnet using mainnet state
// The Staking, Slashing, and Distribution changes are required - everything beyond that is custom
func InitStrideAppForTestnet(app *StrideApp, newValAddr bytes.HexBytes, newValPubKey crypto.PubKey, newOperatorAddress, upgradeToTrigger string) *StrideApp {
	// Create a new account that will be used in the validator
	// This does not match the actual operator keys, but it's not required that they match
	ctx := app.BaseApp.NewContext(true)
	pubkey := &ed25519.PubKey{Key: newValPubKey.Bytes()}
	pubkeyAny, err := types.NewAnyWithValue(pubkey)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// Fund the operator account
	operatorAddress := sdk.MustAccAddressFromBech32(newOperatorAddress)
	initialBalance := sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(100_000_000)))
	if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialBalance); err != nil {
		tmos.Exit(err.Error())
	}
	if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, operatorAddress, initialBalance); err != nil {
		tmos.Exit(err.Error())
	}

	// Create Validator struct for our new validator.
	// The validator will have all the bonded tokens
	vs := app.StakingKeeper.GetValidatorSet()
	totalBondedTokens, err := vs.TotalBondedTokens(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	_, bz, err := bech32.DecodeAndConvert(newOperatorAddress)
	if err != nil {
		tmos.Exit(err.Error())
	}
	bech32Addr, err := bech32.ConvertAndEncode(sdk.GetConfig().GetBech32ValidatorAddrPrefix(), bz)
	if err != nil {
		tmos.Exit(err.Error())
	}
	newVal := stakingtypes.Validator{
		OperatorAddress: bech32Addr,
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          totalBondedTokens,
		DelegatorShares: sdkmath.LegacyMustNewDecFromStr("10000000"),
		Description: stakingtypes.Description{
			Moniker: "Testnet Validator",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          sdkmath.LegacyMustNewDecFromStr("0.05"),
				MaxRate:       sdkmath.LegacyMustNewDecFromStr("0.1"),
				MaxChangeRate: sdkmath.LegacyMustNewDecFromStr("0.05"),
			},
		},
		MinSelfDelegation: sdkmath.OneInt(),
	}

	// Remove all validators from power store
	stakingKey := app.GetKey(stakingtypes.ModuleName)
	stakingStore := ctx.KVStore(stakingKey)
	iterator, err := app.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all valdiators from last validators store
	iterator, err = app.StakingKeeper.LastValidatorsIterator(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from validators store
	iterator = storetypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from unbonding queue
	iterator = storetypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorQueueKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Add our validator to power and last validators store
	err = app.StakingKeeper.SetValidator(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetValidatorByConsAddr(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetValidatorByPowerIndex(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	valAddr, err := sdk.ValAddressFromBech32(newVal.GetOperator())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetLastValidatorPower(ctx, valAddr, 0)
	if err != nil {
		tmos.Exit(err.Error())
	}
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		panic(err)
	}

	// Initialize records for this validator across all distribution stores
	valAddr, err = sdk.ValAddressFromBech32(newVal.GetOperator())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorHistoricalRewards(ctx, valAddr, 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorCurrentRewards(ctx, valAddr, distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, distrtypes.InitialValidatorAccumulatedCommission())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})
	if err != nil {
		tmos.Exit(err.Error())
	}

	// Set validator signing info for our new validator.
	newConsAddr := sdk.ConsAddress(newValAddr.Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}
	err = app.SlashingKeeper.SetValidatorSigningInfo(ctx, newConsAddr, newValidatorSigningInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// Shorten the gov voting period
	newExpeditedVotingPeriod := time.Minute
	newVotingPeriod := time.Minute * 2

	govParams, err := app.GovKeeper.Params.Get(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	govParams.ExpeditedVotingPeriod = &newExpeditedVotingPeriod
	govParams.VotingPeriod = &newVotingPeriod
	govParams.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 100000000))
	govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 150000000))

	err = app.GovKeeper.Params.Set(ctx, govParams)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// Reset each epoch
	for _, epoch := range app.EpochsKeeper.AllEpochInfos(ctx) {
		epoch.CurrentEpochStartTime = time.Now().UTC()
		epoch.CurrentEpochStartHeight = app.LastBlockHeight()
		app.EpochsKeeper.SetEpochInfo(ctx, epoch)
	}

	// Optionally play the latest upgrade on top of the mainnet state
	if upgradeToTrigger != "" {
		upgradePlan := upgradetypes.Plan{
			Name:   upgradeToTrigger,
			Height: app.LastBlockHeight() + 10,
		}
		err = app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// Sets all anti-handlers to run before the tx
func (app *StrideApp) setAnteHandler(
	txConfig client.TxConfig,
	wasmConfig wasmtypes.NodeConfig,
	txCounterStoreKey *storetypes.KVStoreKey,
) {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:         app.IBCKeeper,
			ConsumerKeeper:    app.ConsumerKeeper,
			WasmConfig:        &wasmConfig,
			TXCounterStoreKey: runtime.NewKVStoreService(txCounterStoreKey),
			WasmKeeper:        &app.WasmKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

// Sets post handlers (new as of SDK 46) which run after the tx
func (app *StrideApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *StrideApp) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *StrideApp) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// GetIBCKeeper implements the TestingApp interface.
func (app *StrideApp) GetTransferKeeper() *ibctransferkeeper.Keeper {
	return &app.TransferKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *StrideApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *StrideApp) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *StrideApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	appOptions := autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}

	// Disable the autocli command for ICQ SubmitQueryResponse which has a conflict
	// with the chain-id flag and also doesn't need a CLI function
	moduleOptions, exists := appOptions.ModuleOptions[interchainquerytypes.ModuleName]
	if !exists {
		moduleOptions = &autocliv1.ModuleOptions{}
		appOptions.ModuleOptions[interchainquerytypes.ModuleName] = moduleOptions
	}
	if moduleOptions.Tx == nil {
		moduleOptions.Tx = &autocliv1.ServiceCommandDescriptor{}
	}
	moduleOptions.Tx.RpcCommandOptions = append(moduleOptions.Tx.RpcCommandOptions, &autocliv1.RpcCommandOptions{
		RpcMethod: "SubmitQueryResponse",
		Skip:      true,
	})

	return appOptions
}

// PreBlocker application updates every pre block
func (app *StrideApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *StrideApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *StrideApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// Precommitter application updates before the commital of a block after all transactions have been delivered.
func (app *StrideApp) Precommitter(ctx sdk.Context) {
	if err := app.ModuleManager.Precommit(ctx); err != nil {
		panic(err)
	}
}

func (app *StrideApp) PrepareCheckStater(ctx sdk.Context) {
	if err := app.ModuleManager.PrepareCheckState(ctx); err != nil {
		panic(err)
	}
}

// InitChainer application update at chain initialization
func (app *StrideApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	if err != nil {
		return nil, err
	}
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
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
			acc == staketiatypes.FeeAddress ||
			acc == stakedymtypes.ModuleName ||
			acc == stakedymtypes.FeeAddress {
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
	return app.legacyAmino
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

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *StrideApp) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.appCodec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetStoreKeys returns all the stored store keys.
func (app *StrideApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}

	return keys
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *StrideApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
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
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *StrideApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *StrideApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, cmtApp.Query)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *StrideApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(
		clientCtx,
		app.GRPCQueryRouter(),
		cfg,
		func() int64 { return app.CommitMultiStore().EarliestVersion() },
	)
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
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper { //nolint:staticcheck // SA1019: x/params removal scheduled for follow-up
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey) //nolint:staticcheck // SA1019: x/params removal scheduled for follow-up

	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable())   //nolint:staticcheck
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())       //nolint:staticcheck
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(stakeibcmoduletypes.ModuleName)
	paramsKeeper.Subspace(epochsmoduletypes.ModuleName)
	paramsKeeper.Subspace(interchainquerytypes.ModuleName)
	paramsKeeper.Subspace(recordsmoduletypes.ModuleName)
	paramsKeeper.Subspace(ratelimittypes.ModuleName)
	paramsKeeper.Subspace(icacallbacksmoduletypes.ModuleName)
	paramsKeeper.Subspace(ccvconsumertypes.ModuleName)
	paramsKeeper.Subspace(autopilottypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName)
	paramsKeeper.Subspace(icaoracletypes.ModuleName)
	paramsKeeper.Subspace(claimtypes.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)

	// IBC legacy subspaces (param key tables removed in ibc-go v11; subspaces kept for migration handlers)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *StrideApp) SimulationManager() *module.SimulationManager {
	return nil
}

// ConsumerApp interface implementations for e2e tests
// GetConsumerKeeper implements the ConsumerApp interface.
func (app *StrideApp) GetConsumerKeeper() ccvconsumerkeeper.Keeper {
	return app.ConsumerKeeper
}
