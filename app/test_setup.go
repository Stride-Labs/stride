package app

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/abci/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmtypes "github.com/cometbft/cometbft/types"
	cosmosdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	appconsumer "github.com/cosmos/interchain-security/v6/app/consumer"
	consumertypes "github.com/cosmos/interchain-security/v6/x/ccv/consumer/types"
	ccvtypes "github.com/cosmos/interchain-security/v6/x/ccv/types"

	cmdcfg "github.com/Stride-Labs/stride/v28/cmd/strided/config"
	testutil "github.com/Stride-Labs/stride/v28/testutil"
)

const Bech32Prefix = "stride"

func init() {
	SetupConfig()
}

func SetupConfig() {
	config := sdk.GetConfig()
	valoper := sdk.PrefixValidator + sdk.PrefixOperator
	valoperpub := sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	config.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32Prefix+valoper, Bech32Prefix+valoperpub)
	cmdcfg.SetAddressPrefixes(config)
}

// Initializes a new StrideApp without IBC functionality
func InitStrideTestApp(initChain bool) *StrideApp {
	db := cosmosdb.NewMemDB()

	// A custom home directory is needed for wasm tests since wasmvm locks the directory
	tempHomeDir, err := os.MkdirTemp("", "stride-unit-test")
	if err != nil {
		panic(err)
	}
	appopts := simtestutil.NewAppOptionsWithFlagHome(tempHomeDir)

	app := NewStrideApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		appopts,
		[]wasmkeeper.Option{},
	)

	if initChain {
		genesisState := GenesisStateWithConsumerValSet(app)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		reqInitChain := &abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		}
		if _, err := app.InitChain(reqInitChain); err != nil {
			panic(err)
		}
	}

	return app
}

// Iniitalize default genesis state as a consumer chain with 1 validator
func GenesisStateWithConsumerValSet(app *StrideApp) GenesisState {
	// Create the default genesis state
	genesisState := app.DefaultGenesis()

	// Create a new validator and base account
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()
	validator := tmtypes.NewValidator(pubKey, 1)
	account := authtypes.NewBaseAccountWithAddress(secp256k1.GenPrivKey().PubKey().Address().Bytes())

	// Prepare the list of 1 validator and 1 account for the initial genesis state
	validatorSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	genesisAccounts := []authtypes.GenesisAccount{account}
	accountBalance := banktypes.Balance{
		Address: account.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	// Initialize the genesis state with the validator and account
	genesisState, err := simtestutil.GenesisStateWithValSet(
		app.appCodec,
		genesisState,
		validatorSet,
		genesisAccounts,
		accountBalance,
	)
	if err != nil {
		panic(err)
	}

	// Set the consumer state
	validatorProto, _ := validator.ToProto()
	initialValidatorPowers := []abci.ValidatorUpdate{{
		Power:  validator.VotingPower,
		PubKey: validatorProto.PubKey,
	}}
	vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValidatorPowers)
	if err != nil {
		panic(fmt.Sprintf("failed to get vals: %s", err.Error()))
	}

	consumerGenesisState := testutil.CreateMinimalConsumerTestGenesis()
	consumerGenesisState.Provider.InitialValSet = initialValidatorPowers
	consumerGenesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()
	consumerGenesisState.Params.Enabled = true
	genesisState[consumertypes.ModuleName] = app.AppCodec().MustMarshalJSON(consumerGenesisState)

	return genesisState
}

// Initializes a new Stride App casted as a TestingApp for IBC support
func InitStrideIBCTestingApp(initValPowers []types.ValidatorUpdate) func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		encoding := appconsumer.MakeTestEncodingConfig()
		app := InitStrideTestApp(false)
		genesisState := app.DefaultGenesis()

		// Feed consumer genesis with provider validators
		var consumerGenesis ccvtypes.ConsumerGenesisState
		encoding.Codec.MustUnmarshalJSON(genesisState[consumertypes.ModuleName], &consumerGenesis)
		consumerGenesis.Provider.InitialValSet = initValPowers
		consumerGenesis.Params.Enabled = true
		genesisState[consumertypes.ModuleName] = encoding.Codec.MustMarshalJSON(&consumerGenesis)

		return app, genesisState
	}
}
