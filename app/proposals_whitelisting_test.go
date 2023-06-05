package app_test

// import (
// 	"encoding/json"
// 	"testing"

// 	"cosmossdk.io/simapp"
// 	"github.com/cometbft/cometbft/libs/log"
// 	ibctesting "github.com/cosmos/ibc-go/v7/testing"
// 	icssimapp "github.com/cosmos/interchain-security/testutil/simapp"
// 	"github.com/stretchr/testify/require"
// 	dbm "github.com/tendermint/tm-db"

// 	"github.com/Stride-Labs/stride/v9/app"
// )

// func TestConsumerWhitelistingKeys(t *testing.T) {
// 	chain := ibctesting.NewTestChain(t, icssimapp.NewBasicCoordinator(t), SetupTestingAppConsumer, "test")
// 	paramKeeper := chain.App.(*app.StrideApp).ParamsKeeper
// 	for paramKey := range app.WhitelistedParams {
// 		ss, ok := paramKeeper.GetSubspace(paramKey.Subspace)
// 		require.True(t, ok, "Unknown subspace %s", paramKey.Subspace)
// 		hasKey := ss.Has(chain.GetContext(), []byte(paramKey.Key))
// 		require.True(t, hasKey, "Invalid key %s for subspace %s", paramKey.Key, paramKey.Subspace)
// 	}
// }

// func SetupTestingAppConsumer() (ibctesting.TestingApp, map[string]json.RawMessage) {
// 	db := dbm.NewMemDB()
// 	testApp := app.NewStrideApp(
// 		log.NewNopLogger(),
// 		db,
// 		nil,
// 		true,
// 		map[int64]bool{},
// 		app.DefaultNodeHome,
// 		5,
// 		app.MakeEncodingConfig(),
// 		simapp.EmptyAppOptions{},
// 	)

// 	return testApp, app.NewDefaultGenesisState()
// }
