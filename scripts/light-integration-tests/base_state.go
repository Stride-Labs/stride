package lighttest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var STRIDE_ACCT = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"
var CHAIN_ID = "stride"
var BASE_HEADER = tmproto.Header{Height: 1, ChainID: CHAIN_ID, Time: time.Now().UTC()}

func BaseAppStateOld(t *testing.T) (*strideapp.StrideApp, sdk.Context) {
	isCheckTx := false
	app := strideapp.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, BASE_HEADER)

	return app, ctx
}

func BaseAppState(t *testing.T) *strideapp.StrideApp {
	db := dbm.NewMemDB()
	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	app := strideapp.NewStrideApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		filepath.Join(userHomeDir, ".stride"),
		5,
		strideapp.MakeEncodingConfig(),
		simapp.EmptyAppOptions{},
	)

	// ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: CHAIN_ID, Time: time.Now().UTC()})

	return app
}

func GetModuleAddress(t *testing.T, moduleName string) string {
	relAddr := authtypes.NewModuleAddress(moduleName)
	outAddr, err := bech32.ConvertAndEncode("stride", relAddr.Bytes())
	require.NoError(t, err)
	return outAddr
}

func getPrivKey(t *testing.T, hexStr string) secp256k1.PrivKey {
	key, err := hex.DecodeString(hexStr)
	require.NoError(t, err)
	senderPrivKey := secp256k1.PrivKey{Key: key}
	return senderPrivKey
}

func SetSDKDefaults() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("stride", "stridepub")
	cfg.SetBech32PrefixForConsensusNode("stride", "stridecons")
	cfg.SetBech32PrefixForValidator("stridevaloper", "stridevaloperpub")
}

func SetupBaseApp(t *testing.T) (*strideapp.StrideApp, sdk.Context, map[string]tmtypes.PrivValidator) {
	// Be warned, I wasted 2 hours because I put `app := ...` before the cfg changes.
	// DO NOT MOVE SETTING DEFAULTS
	SetSDKDefaults()
	app := BaseAppState(t)
	// set genesis accounts
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}

	allKeys := []string{"87b02600b8e300691689b51254c3fd23fa2d381d6e18a59583d0d483d549ce0f"}
	for _, hexKey := range allKeys {
		senderPrivKey := getPrivKey(t, hexKey)

		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), uint64(0), 0)
		amount, ok := sdk.NewIntFromString("100000000000000000")
		require.True(t, ok)
		balance := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("ustrd", amount)),
		}

		genAccs = append(genAccs, acc)
		genBals = append(genBals, balance)
	}
	// set genesis state
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState := strideapp.NewDefaultGenesisState()
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	valAddrs := []string{"stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm"}
	numVals := len(valAddrs)
	validators := make([]stakingtypes.Validator, 0, numVals)
	delegations := make([]stakingtypes.Delegation, 0, numVals)
	signersByAddr := make(map[string]tmtypes.PrivValidator, numVals)
	bondAmt := sdk.TokensFromConsensusPower(1, sdk.NewInt(1000000))

	for _, valAddrBech32 := range valAddrs {
		privKey := getPrivKey(t, allKeys[0])
		pubKey := privKey.PubKey()
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		valAddr, err := sdk.ValAddressFromBech32(valAddrBech32)
		require.NoError(t, err)
		valDescr := stakingtypes.Description{
			Moniker:  "Stride Validator",
			Identity: "Stride Validator",
			Website:  "https://stride.zone/",
			Details:  "Stride Validator",
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   valAddr.String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       valDescr,
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}

		validators = append(validators, validator)
		cryptoPrivKey := cryptotypes.PrivKey{}
		signersByAddr[pubKey.Address().String()] = ibcmock.PV{PrivKey: cryptoPrivKey}
		// privKey
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), valAddr.Bytes(), sdk.OneDec()))
	}

	// set validators and delegations
	var stakingGenesis stakingtypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesis)
	// get bondedPool address
	bondedAddr := GetModuleAddress(t, stakingtypes.BondedPoolName)

	stakingGenesis.Params.BondDenom = "ustrd"
	// add bonded amount to bonded pool module account
	genBals = append(genBals, banktypes.Balance{
		Address: bondedAddr,
		Coins:   sdk.Coins{sdk.NewCoin("ustrd", bondAmt.Mul(sdk.NewInt(int64(1))))},
	})

	// set validators and delegations
	stakingGenesis = *stakingtypes.NewGenesisState(stakingGenesis.Params, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&stakingGenesis)

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, genBals, sdk.NewCoins(), []banktypes.Metadata{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         CHAIN_ID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(
		abci.RequestBeginBlock{
			Header: tmproto.Header{
				ChainID: CHAIN_ID,
				Height:  app.LastBlockHeight() + 1,
				AppHash: app.LastCommitID().Hash,
			},
		},
	)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: CHAIN_ID, Time: time.Now().UTC()})

	return app, ctx, signers
}

func CreateGaiaIBCClient(t *testing.T, app *strideapp.StrideApp, ctx sdk.Context) (*testing.T, *strideapp.StrideApp, sdk.Context) {
	k := app.IBCKeeper
	goCtx := sdk.WrapSDKContext(ctx)

	clientState := codectypes.Any{}
	consensusState := codectypes.Any{}
	resp, err := k.CreateClient(goCtx, &clienttypes.MsgCreateClient{
		ClientState:    &clientState,
		ConsensusState: &consensusState,
		Signer:         "NIL",
	})
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	app.StakeibcKeeper.Logger(ctx).Info(fmt.Sprintf("%v", resp))
	_, _ = k, goCtx

	return t, app, ctx
}

// func OpenGaiaConnection(app *strideapp.StrideApp, ctx sdk.Context, t testing.T) (*strideapp.StrideApp, sdk.Context) {
// 	k := app.IBCKeeper
// 	goCtx := sdk.WrapSDKContext(ctx)

// 	k.ConnectionOpenInit(goCtx, &connectiontypes.MsgConnectionOpenInit{
// 		ClientId: "nil",
// 		Counterparty: "nil",
// 		Version      *Version,
// 		DelayPeriod  uint64       `protobuf:"varint,4,opt,name=delay_period,json=delayPeriod,proto3" json:"delay_period,omitempty" yaml:"delay_period"`
// 		Signer       string
// 	})
// 	_, _ = k, goCtx

// 	return app, ctx
// }

// func OpenGaiaChannel(app *strideapp.StrideApp, ctx sdk.Context, t testing.T) (*strideapp.StrideApp, sdk.Context) {
// 	k := app.StakeibcKeeper
// 	goCtx := sdk.WrapSDKContext(ctx)

// 	counterparty := channeltypes.CounterParty

// 	relChannel := channeltypes.Channel{
// 		State:    channeltypes.UNINITIALIZED,
// 		Ordering: channeltypes.ORDERED,
// 		// counterparty channel end
// 		Counterparty: nil,
// 		// list of connection identifiers, in order, along which packets sent on
// 		// this cheannel will travel
// 		ConnectionHops: []string{"connection-0"},
// 		// opaque channel version, which is agreed upon during the handshake
// 		Version: "HI",
// 	}

// 	app.IBCKeeper.ChannelOpenInit(goCtx, &channeltypes.MsgChannelOpenInit{
// 		PortId:  "port-1",
// 		Channel: "channel-1",
// 		Signer:  STRIDE_ACCT,
// 	})
// 	return app, ctx
// }

func RegisterHostZone(app *strideapp.StrideApp, ctx sdk.Context, t *testing.T) (*strideapp.StrideApp, sdk.Context, *testing.T) {
	k := app.StakeibcKeeper
	goCtx := sdk.WrapSDKContext(ctx)
	msg := types.MsgRegisterHostZone{
		ConnectionId:       "connection-0",
		Bech32Prefix:       "cosmos",
		HostDenom:          "uatom",
		IbcDenom:           "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Creator:            "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq",
		TransferChannelId:  "channel-1",
		UnbondingFrequency: 3,
	}
	_, err := k.RegisterHostZone(goCtx, &msg)
	if err != nil {
		t.Fatalf("error registering host zone: %v", err)
	}
	return app, ctx, t
}

func InitialBasicSetup(t *testing.T) (*testing.T, *ibctesting.Coordinator, sdk.Context) {
	// Create a coordinator
	coord := ibctesting.NewCoordinator(t, 1)
	// gaia := coordinator.GetChain(ibctesting.GetChainID(1))

	// Create Stride's app
	app, ctx, signers := SetupBaseApp(t)
	if app == nil {
		t.Error("app is nil")
	}
	//coordinator.Chains[CHAIN_ID] =
	chain := &ibctesting.TestChain{
		T:              t,
		Coordinator:    coord,
		ChainID:        CHAIN_ID,
		App:            app,
		CurrentHeader:  BASE_HEADER,
		QueryServer:    app.GetIBCKeeper(),
		TxConfig:       app.GetTxConfig(),
		Codec:          app.AppCodec(),
		Vals:           valSet,
		NextVals:       valSet,
		Signers:        signers,
		SenderPrivKey:  senderAccs[0].SenderPrivKey,
		SenderAccount:  senderAccs[0].SenderAccount,
		SenderAccounts: senderAccs,
	}
	coord.Chains[CHAIN_ID] = chain
	return t, coord, ctx
}
