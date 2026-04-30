package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/abci/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmtypes "github.com/cometbft/cometbft/types"
	cosmosdb "github.com/cosmos/cosmos-db"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
	consumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	cmdcfg "github.com/Stride-Labs/stride/v32/cmd/strided/config"
	testutil "github.com/Stride-Labs/stride/v32/testutil"
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

	// Seed the POA genesis with the test validator so that POA's InitGenesis
	// returns a non-empty ValidatorUpdate (the module manager requires exactly
	// one module to do so after the ccvconsumer module was removed).
	pubKeyAny, err := codectypes.NewAnyWithValue(privVal.PrivKey.PubKey())
	if err != nil {
		panic(fmt.Sprintf("failed to pack poa validator pubkey: %s", err.Error()))
	}
	poaGenesisState := &poatypes.GenesisState{
		Params: poatypes.Params{Admin: authtypes.NewModuleAddress("gov").String()},
		Validators: []poatypes.Validator{
			{
				PubKey: pubKeyAny,
				Power:  1,
				Metadata: &poatypes.ValidatorMetadata{
					OperatorAddress: account.GetAddress().String(),
					Moniker:         "test-validator",
				},
			},
		},
	}
	genesisState[poatypes.ModuleName] = app.AppCodec().MustMarshalJSON(poaGenesisState)

	return genesisState
}

// Initializes a new Stride App casted as a TestingApp for IBC support
func InitStrideIBCTestingApp(initValPowers []types.ValidatorUpdate) func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		app := InitStrideTestApp(false)
		genesisState := app.DefaultGenesis()

		// ccvconsumer is no longer in the module manager (post-v33 ICS→POA migration),
		// so its genesis is not processed by InitChain. POA is now the source of
		// validator-set updates at genesis; seed it with the IBC test's validator set
		// so CometBFT's InitChain validator set matches what the test framework expects.
		poaGenesis := buildPOAGenesisFromAbciValSet(initValPowers)
		genesisState[poatypes.ModuleName] = app.AppCodec().MustMarshalJSON(poaGenesis)

		return app, genesisState
	}
}

// buildPOAGenesisFromAbciValSet converts a slice of CometBFT ValidatorUpdates into a
// POA GenesisState. This is used by the IBC testing path, which passes in validators
// originating from the provider chain simulator rather than the normal test setup.
func buildPOAGenesisFromAbciValSet(valSet []types.ValidatorUpdate) *poatypes.GenesisState {
	poaVals := make([]poatypes.Validator, 0, len(valSet))
	for i, vu := range valSet {
		sdkPK, err := cryptocodec.FromCmtProtoPublicKey(vu.PubKey)
		if err != nil {
			panic(fmt.Sprintf("ibc test: convert cmt pubkey: %s", err))
		}
		pkAny, err := codectypes.NewAnyWithValue(sdkPK)
		if err != nil {
			panic(fmt.Sprintf("ibc test: wrap pubkey: %s", err))
		}

		// Derive a deterministic operator address distinct from the consensus address.
		// POA enforces OperatorAddress uniqueness across validators.
		consAddr := sdkPK.Address().Bytes()
		hash := sha256.Sum256(append([]byte(fmt.Sprintf("ibc-test-op:%d:", i)), consAddr...))
		opAddr := sdk.AccAddress(hash[:20]).String()

		poaVals = append(poaVals, poatypes.Validator{
			PubKey: pkAny,
			Power:  vu.Power,
			Metadata: &poatypes.ValidatorMetadata{
				OperatorAddress: opAddr,
				Moniker:         fmt.Sprintf("ibc-test-validator-%d", i),
			},
		})
	}
	return &poatypes.GenesisState{
		Params:     poatypes.Params{Admin: authtypes.NewModuleAddress("gov").String()},
		Validators: poaVals,
	}
}
