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
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/cosmos/ibc-go/v3/testing/mock"
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
	cfg.SetAddressVerifier(func([]byte) (err error) { return nil })
}

func SetupBaseApp(t *testing.T) (*strideapp.StrideApp, sdk.Context, *tmtypes.ValidatorSet, map[string]tmtypes.PrivValidator, []ibctesting.SenderAccount) {
	// Be warned, I wasted 2 hours because I put `app := ...` before the cfg changes.
	// DO NOT MOVE SETTING DEFAULTS
	SetSDKDefaults()
	app := BaseAppState(t)
	SetSDKDefaults()
	// set genesis accounts
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}
	senderAccs := []ibctesting.SenderAccount{}

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

		senderAcc := ibctesting.SenderAccount{
			SenderAccount: acc,
			SenderPrivKey: &senderPrivKey,
		}
		senderAccs = append(senderAccs, senderAcc)

	}
	// set genesis state
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState := strideapp.NewDefaultGenesisState()
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	numVals := 1
	tmValidators := make([]*tmtypes.Validator, 0, numVals)
	validators := make([]stakingtypes.Validator, 0, numVals)
	delegations := make([]stakingtypes.Delegation, 0, numVals)
	signers := make(map[string]tmtypes.PrivValidator, numVals)
	bondAmt := sdk.TokensFromConsensusPower(1, sdk.NewInt(1000000))

	for i := 0; i < numVals; i++ {
		privKey := mock.NewPV()
		pubKey, err := privKey.GetPubKey()
		require.NoError(t, err)
		val := tmtypes.NewValidator(pubKey, 1)
		pk, err := cryptocodec.FromTmPubKeyInterface(pubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		valDescr := stakingtypes.Description{
			Moniker:  "Stride Validator",
			Identity: "Stride Validator",
			Website:  "https://stride.zone/",
			Details:  "Stride Validator",
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
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
		tmValidators = append(tmValidators, tmtypes.NewValidator(pubKey, 1))
		validators = append(validators, validator)
		signers[pubKey.Address().String()] = privKey
		// privKey
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}

	// set validators and delegations
	var stakingGenesis stakingtypes.GenesisState
	stakingGenesis.Params.BondDenom = "ustrd"
	app.AppCodec().MustUnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesis)
	stakingGenesis.Params.BondDenom = "ustrd"
	// get bondedPool address
	bondedAddr := GetModuleAddress(t, stakingtypes.BondedPoolName)

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

	// update distribution supply
	defaultDistr := distrtypes.DefaultGenesisState()
	fp := distrtypes.InitialFeePool()
	for _, bal := range genBals {
		coinList := []sdk.Coin{}
		for _, coin := range bal.Coins {
			coinList = append(coinList, coin)
		}
		decCoins := sdk.NewDecCoinsFromCoins(coinList...)
		fp.CommunityPool.Add(decCoins...)
	}
	distrGenesis := distrtypes.NewGenesisState(defaultDistr.Params,
		fp, defaultDistr.DelegatorWithdrawInfos, sdk.ConsAddress(defaultDistr.PreviousProposer),
		defaultDistr.OutstandingRewards, defaultDistr.ValidatorAccumulatedCommissions,
		defaultDistr.ValidatorHistoricalRewards, defaultDistr.ValidatorCurrentRewards,
		defaultDistr.DelegatorStartingInfos, defaultDistr.ValidatorSlashEvents)
	genesisState[distrtypes.ModuleName] = app.AppCodec().MustMarshalJSON(distrGenesis)

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
	valSet := tmtypes.NewValidatorSet(tmValidators)

	return app, ctx, valSet, signers, senderAccs
}

func IterateBlock(coord *ibctesting.Coordinator) {
	coord.CommitBlock(coord.GetChain(CHAIN_ID), coord.GetChain(ibctesting.GetChainID(1)))
}

func CreateIBCConnection(t *testing.T, coord *ibctesting.Coordinator, ctx sdk.Context) (*testing.T, *ibctesting.Coordinator, sdk.Context) {
	stride := coord.Chains[CHAIN_ID]
	gaia := coord.GetChain(ibctesting.GetChainID(1))
	// app := (stride.App).(*strideapp.StrideApp)
	// k := app.IBCKeeper
	// goCtx := sdk.WrapSDKContext(ctx)

	// create transfer channel
	path := ibctesting.NewPath(stride, gaia)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"
	path.EndpointA.ConnectionID = "connection-0"
	path.EndpointB.ConnectionID = "connection-0"

	coord.Setup(path)
	coord.CommitBlock()

	IterateBlock(coord)
	return t, coord, ctx
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

func RegisterHostZone(t *testing.T, coord *ibctesting.Coordinator, ctx sdk.Context) (*testing.T, *ibctesting.Coordinator, sdk.Context) {
	stride := coord.Chains[CHAIN_ID]
	app := (stride.App).(*strideapp.StrideApp)
	// k := app.StakeibcKeeper
	// goCtx := sdk.WrapSDKContext(ctx)
	msg := &types.MsgRegisterHostZone{
		ConnectionId:       "connection-0",
		Bech32Prefix:       "cosmos",
		IbcDenom:           "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Creator:            STRIDE_ACCT,
		TransferChannelId:  "transfer",
		UnbondingFrequency: 3,
	}
	handler := app.MsgServiceRouter().Handler(msg)
	resp, err := handler(ctx, msg)
	require.NoError(t, err)
	fmt.Printf("Response: %v\n", resp)
	IterateBlock(coord)
	IterateBlock(coord)
	return t, coord, ctx
}

func InitialBasicSetup(t *testing.T) (*testing.T, *ibctesting.Coordinator, sdk.Context) {
	// Create Stride's app
	app, ctx, valSet, signersMap, senderAccs := SetupBaseApp(t)
	if app == nil {
		t.Error("app is nil")
	}
	// create signers indexed by the valSet validators's order
	signers := []tmtypes.PrivValidator{}
	for _, val := range valSet.Validators {
		signers = append(signers, signersMap[val.PubKey.Address().String()])
	}

	// this makes no sense to me, but if you move the below line to the top
	// of this function, you will get an INCREDIBLY arcane error.
	// Create a coordinator
	coord := ibctesting.NewCoordinator(t, 1)
	// gaia := coordinator.GetChain(ibctesting.GetChainID(1))
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
		Signers:        signers,
		SenderPrivKey:  senderAccs[0].SenderPrivKey,
		SenderAccount:  senderAccs[0].SenderAccount,
		SenderAccounts: senderAccs,
	}
	coord.CommitBlock(chain)
	coord.Chains[CHAIN_ID] = chain

	CreateIBCConnection(t, coord, ctx)

	return t, coord, ctx
}
