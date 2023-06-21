package ics_test

// import (
// 	"encoding/json"
// 	"testing"

// 	"cosmossdk.io/simapp"
// 	"github.com/cometbft/cometbft/libs/log"
// 	ibctesting "github.com/cosmos/ibc-go/v7/testing"
// 	appProvider "github.com/cosmos/interchain-security/v3/app/provider"
// 	"github.com/cosmos/interchain-security/v3/tests/e2e"
// 	e2etestutil "github.com/cosmos/interchain-security/v3/testutil/e2e"
// 	icssimapp "github.com/cosmos/interchain-security/v3/testutil/simapp"
// 	"github.com/stretchr/testify/suite"
// 	dbm "github.com/tendermint/tm-db"

// 	appConsumer "github.com/Stride-Labs/stride/v10/app"
// )

// // Executes the standard group of ccv tests against a consumer and provider app.go implementation.
// func TestCCVTestSuite(t *testing.T) {

// 	t.Skip("This test is skipped for now. Once the interchain security version is updated to/newer then 3b562dd87397 the test can be run")
// 	ccvSuite := e2e.NewCCVTestSuite(
// 		func(t *testing.T) (
// 			*ibctesting.Coordinator,
// 			*ibctesting.TestChain,
// 			*ibctesting.TestChain,
// 			e2etestutil.ProviderApp,
// 			e2etestutil.ConsumerApp,
// 		) {
// 			// Here we pass the concrete types that must implement the necessary interfaces
// 			// to be ran with e2e tests.
// 			coord, prov, cons := NewProviderConsumerCoordinator(t)
// 			return coord, prov, cons, prov.App.(*appProvider.App), cons.App.(*appConsumer.StrideApp)
// 		},
// 	)
// 	suite.Run(t, ccvSuite)
// }

// // NewCoordinator initializes Coordinator with interchain security dummy provider and noble consumer chain
// func NewProviderConsumerCoordinator(t *testing.T) (*ibctesting.Coordinator, *ibctesting.TestChain, *ibctesting.TestChain) {
// 	coordinator := icssimapp.NewBasicCoordinator(t)
// 	chainID := ibctesting.GetChainID(1)
// 	coordinator.Chains[chainID] = ibctesting.NewTestChain(t, coordinator, icssimapp.SetupTestingappProvider, chainID)
// 	providerChain := coordinator.GetChain(chainID)
// 	chainID = ibctesting.GetChainID(2)
// 	coordinator.Chains[chainID] = ibctesting.NewTestChainWithValSet(t, coordinator,
// 		SetupTestingAppConsumer, chainID, providerChain.Vals, providerChain.Signers)
// 	consumerChain := coordinator.GetChain(chainID)
// 	return coordinator, providerChain, consumerChain
// }

// func SetupTestingAppConsumer() (ibctesting.TestingApp, map[string]json.RawMessage) {
// 	db := dbm.NewMemDB()
// 	testApp := appConsumer.NewStrideApp(
// 		log.NewNopLogger(),
// 		db,
// 		nil,
// 		true,
// 		map[int64]bool{},
// 		appConsumer.DefaultNodeHome,
// 		5,
// 		appConsumer.MakeEncodingConfig(),
// 		simapp.EmptyAppOptions{},
// 	)

// 	return testApp, appConsumer.NewDefaultGenesisState()
// }
