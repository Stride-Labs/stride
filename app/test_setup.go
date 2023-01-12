package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	cmdcfg "github.com/Stride-Labs/stride/v4/cmd/strided/config"
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
	db := dbm.NewMemDB()
	app := NewStrideApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		5,
		MakeEncodingConfig(),
		simapp.EmptyAppOptions{},
	)
	if initChain {
		genesisState := NewDefaultGenesisState()
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// Initializes a new Stride App casted as a TestingApp for IBC support
func InitStrideIBCTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	app := InitStrideTestApp(false)
	return app, NewDefaultGenesisState()
}
