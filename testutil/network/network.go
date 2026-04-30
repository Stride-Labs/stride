package network

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cometbftrand "github.com/cometbft/cometbft/libs/rand"
	cosmosdb "github.com/cosmos/cosmos-db"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/Stride-Labs/stride/v32/app"
)

type (
	Network   = network.Network
	Config    = network.Config
	Validator = network.Validator
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	t.Helper()
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, _ := network.New(t, t.TempDir(), cfg)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	tempApp := app.InitStrideTestApp(false)
	encoding := app.MakeEncodingConfig()

	chainId := fmt.Sprintf("stride-%d", cometbftrand.NewRand().Uint64())
	genState := tempApp.DefaultGenesis()
	return network.Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			// POA is the InitChain validator-update source post-v33 (ICS →
			// POA migration). Without this seed CometBFT panics on an empty
			// validator set update. See seedNetworkValidator for details.
			if err := seedNetworkValidator(val.(network.Validator)); err != nil {
				panic(err)
			}

			// Drop gentxs so the chain's validator set comes only from POA's
			// genesis, not from a framework-generated staking gentx.
			if err := clearGentxs(val.(network.Validator)); err != nil {
				panic(err)
			}

			// A custom home directory is needed for wasm tests since wasmvm locks the directory
			tempHomeDir, err := os.MkdirTemp("", "stride-unit-test")
			if err != nil {
				panic(err)
			}

			return app.NewStrideApp(
				val.GetCtx().Logger,
				cosmosdb.NewMemDB(),
				true,
				simtestutil.NewAppOptionsWithFlagHome(tempHomeDir),
				[]wasmkeeper.Option{},
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(chainId),
			)
		},
		GenesisState:    genState,
		TimeoutCommit:   2 * time.Second,
		ChainID:         chainId,
		NumValidators:   1,
		BondDenom:       sdk.DefaultBondDenom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}

// seedNetworkValidator seeds POA's genesis with the network's validator so
// InitChain returns at least one ValidatorUpdate. Post-v33 (ICS → POA
// migration) ccvconsumer is no longer in the module manager, so POA is the
// only module that can produce InitChain validator updates — without this
// seed CometBFT panics on an empty validator set.
//
// We deliberately do NOT seed a shadow staking validator here. Block production
// works fine without one because distribution is wired through
// app/distrwrapper, which iterates bonded staking validators by stake (no
// per-voter ValidatorByConsAddr lookup). With zero bonded validators,
// distrwrapper routes the staking share to the community pool — fine for
// these CLI tests since they don't assert on per-validator rewards.
func seedNetworkValidator(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	pkAny, err := codectypes.NewAnyWithValue(val.PubKey)
	if err != nil {
		return errorsmod.Wrap(err, "failed to wrap validator pubkey as Any")
	}

	poaGenesis := &poatypes.GenesisState{
		Params: poatypes.Params{Admin: authtypes.NewModuleAddress("gov").String()},
		Validators: []poatypes.Validator{
			{
				PubKey: pkAny,
				// Power=100 matches what modifyConsumerGenesis used pre-v33 and
				// what the network framework's BondedTokens implies.
				Power: 100,
				Metadata: &poatypes.ValidatorMetadata{
					OperatorAddress: val.Address.String(),
					Moniker:         "test-validator",
				},
			},
		},
	}
	poaBz, err := val.ClientCtx.Codec.MarshalJSON(poaGenesis)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal POA genesis state into JSON")
	}
	appState[poatypes.ModuleName] = poaBz

	return writeGenesis(genDoc, appState, genFile)
}

// clearGentxs replaces the genutil genesis with the default (empty gentx list).
// Called so the chain's validator set is taken from POA's genesis only, not
// from a framework-generated staking gentx that would compete with POA.
func clearGentxs(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	genutilBz, err := val.ClientCtx.Codec.MarshalJSON(genutiltypes.DefaultGenesisState())
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal genutil genesis state into JSON")
	}
	appState[genutiltypes.ModuleName] = genutilBz

	return writeGenesis(genDoc, appState, genFile)
}

// writeGenesis serializes appState back into genDoc and writes it to genFile.
func writeGenesis(genDoc *genutiltypes.AppGenesis, appState map[string]json.RawMessage, genFile string) error {
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal application genesis state into JSON")
	}
	genDoc.AppState = appStateJSON
	if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
		return errorsmod.Wrap(err, "failed to export genesis state")
	}
	return nil
}
