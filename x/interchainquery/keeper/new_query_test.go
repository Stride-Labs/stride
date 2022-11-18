package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	icqtypes "github.com/Stride-Labs/stride/v3/x/interchainquery/types"
)

type NewQueryTestCase struct {
	module       string
	connectionId string
	chainId      string
	queryType    string
	request      []byte
	period       sdk.Int
	callbackId   string
	ttl          uint64
	height       int64
}

func (suite *KeeperTestSuite) SetupNewQuery() NewQueryTestCase {
	// module is the name of the module invoking the query, used to find the callback upon response
	module := "stakeibc"
	// connectionId of the target chain you're querying
	connectionId := "connection-0"
	// chain Id of the target chain you're querying
	chainId := "GAIA"
	// QueryType is a string that is used to identify the store you'd like to query, as well as whether you'd like a proof returned alongside the result
	// use "staking" store to access validator which lives in the staking module
	// use "key" suffix to retrieve a proof alongside the query result
	queryType := icqtypes.STAKING_STORE_QUERY_WITH_PROOF
	// request is a byte array that points to the entry in the target store you'd like to query on the host zone
	//	 e.g. for querying a validator, `data := stakingtypes.GetValidatorKey(valAddr)`
	request := []byte{0x01, 0x02, 0x03}
	// period is the frequency at which you'd like the query to be executed, if you're executing a period query (note -1 specifies a one-time query)
	period := sdk.NewInt(-1)
	// callbackId is a string that is used to identify the callback you'd like to execute upon receiving the result of the query
	callbackId := "validator"
	// ttl is the expiry time of the query, in absolute units of time, unix nanos
	ttl := uint64(0) // ttl
	// height is the height at which you'd like the query to be executed on the host zone. height=0 means execute at the latest height
	height := int64(0) // height always 0 (which means current height)

	return NewQueryTestCase{
		module:       module,
		connectionId: connectionId,
		chainId:      chainId,
		queryType:    queryType,
		request:      request,
		period:       period,
		callbackId:   callbackId,
		ttl:          ttl,
		height:       height,
	}
}

func (s *KeeperTestSuite) TestNewQuerySuccessful() {
	tc := s.SetupNewQuery()

	actualQuery := s.App.InterchainqueryKeeper.NewQuery(
		s.Ctx(),
		tc.module,
		tc.connectionId,
		tc.chainId,
		tc.queryType,
		tc.request,
		tc.period,
		tc.callbackId,
		tc.ttl,
		tc.height,
	)

	// this hash is testing `GenerateQueryHash`.
	//  note: the module gets hashed in the `GenerateQueryHash` function, so hashes will be unique to each module
	//  note: the queryID does NOT distinguish between two queries issued with the same (connection_id, chain_id, query_type, request, module, height, request)
	//      practically the differentiator will be `request`, but it DOES mean that if we submit more than 1 automated query to the same value on the same host account simultaneously, we would get callback conflicts!
	expectedId := "9792c1d779a3846a8de7ae82f31a74d308b279a521fa9e0d5c4f08917117bf3e"
	s.Require().Equal(expectedId, actualQuery.Id)

	// lastHeight should be 0
	expectedLastHeight := sdk.NewInt(0)
	s.Require().Equal(expectedLastHeight, actualQuery.LastHeight)

	// all other arguments should be the same as the input
	s.Require().Equal(tc.connectionId, actualQuery.ConnectionId)
	s.Require().Equal(tc.chainId, actualQuery.ChainId)
	s.Require().Equal(tc.queryType, actualQuery.QueryType)
	s.Require().Equal(tc.request, actualQuery.Request)
	s.Require().Equal(tc.period, actualQuery.Period)
	s.Require().Equal(tc.callbackId, actualQuery.CallbackId)
	s.Require().Equal(tc.ttl, actualQuery.Ttl)
	s.Require().Equal(tc.height, actualQuery.Height)
}
