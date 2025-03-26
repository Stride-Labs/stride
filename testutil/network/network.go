package network

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	pruningtypes "cosmossdk.io/store/pruning/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	types1 "github.com/cometbft/cometbft/abci/types"
	cometbftrand "github.com/cometbft/cometbft/libs/rand"
	tmtypes "github.com/cometbft/cometbft/types"
	cosmosdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v6/x/ccv/consumer/types"

	"github.com/Stride-Labs/stride/v26/app"

	testutil "github.com/Stride-Labs/stride/v26/testutil"
)

type (
	Network   = network.Network
	Config    = network.Config
	Validator = network.Validator
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
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
	// app doesn't have this module anymore, but we need them for test setup, which uses gentx
	tempApp := app.InitStrideTestApp(false)
	encoding := app.GetEncodingConfig()
	tempApp.BasicModuleManager.RegisterInterfaces(encoding.InterfaceRegistry)

	chainId := fmt.Sprintf("stride-%d", cometbftrand.NewRand().Uint64())
	genState := tempApp.DefaultGenesis()
	return network.Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			err := modifyConsumerGenesis(val.(network.Validator))
			if err != nil {
				panic(err)
			}

			// do NOT create validators using gentxs so that val sets are applied using only ccv genesis
			err = modifyGenutilGenesis(val.(network.Validator))
			if err != nil {
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
				nil,
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

// add new val sets to consumer genesis before launching app
func modifyConsumerGenesis(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(val.PubKey)
	if err != nil {
		return errorsmod.Wrap(err, "invalid public key")
	}

	initialValset := []types1.ValidatorUpdate{{PubKey: tmProtoPublicKey, Power: 100}}
	vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
	if err != nil {
		return errorsmod.Wrap(err, "could not convert val updates to validator set")
	}

	consumerGenesisState := testutil.CreateMinimalConsumerTestGenesis()
	consumerGenesisState.Provider.InitialValSet = initialValset
	consumerGenesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()

	if err := consumerGenesisState.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid consumer genesis")
	}

	consumerGenStateBz, err := val.ClientCtx.Codec.MarshalJSON(consumerGenesisState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal consumer genesis state into JSON")
	}

	appState[ccvconsumertypes.ModuleName] = consumerGenStateBz
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal application genesis state into JSON")
	}

	genDoc.AppState = appStateJSON
	err = genutil.ExportGenesisFile(genDoc, genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to export genesis state")
	}

	return nil
}

// remove gentxs from genutil genesis
func modifyGenutilGenesis(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis from the file")
	}

	genutilGenesisState := genutiltypes.DefaultGenesisState()
	genutilGenStateBz, err := val.ClientCtx.Codec.MarshalJSON(genutilGenesisState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal consumer genesis state into JSON")
	}
	appState[genutiltypes.ModuleName] = genutilGenStateBz
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal application genesis state into JSON")
	}

	genDoc.AppState = appStateJSON
	err = genutil.ExportGenesisFile(genDoc, genFile)
	if err != nil {
		return errorsmod.Wrap(err, "failed to export genesis state")
	}

	return nil
}
