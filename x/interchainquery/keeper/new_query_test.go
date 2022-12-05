package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
)

type NewQueryTestCase struct {
	module       string
	callbackId   string
	chainId      string
	connectionId string
	queryType    string
	request      []byte
	ttl          uint64
}

func (suite *KeeperTestSuite) SetupNewQuery() NewQueryTestCase {
	// module is the name of the module invoking the query, used to find the callback upon response
	module := "stakeibc"
	// callbackId is a string that is used to identify the callback you'd like to execute upon receiving the result of the query
	callbackId := "validator"
	// chain Id of the target chain you're querying
	chainId := "GAIA"
	// connectionId of the target chain you're querying
	connectionId := "connection-0"
	// QueryType is a string that is used to identify the store you'd like to query, as well as whether you'd like a proof returned alongside the result
	// use "staking" store to access validator which lives in the staking module
	// use "key" suffix to retrieve a proof alongside the query result
	queryType := icqtypes.STAKING_STORE_QUERY_WITH_PROOF
	// request is a byte array that points to the entry in the target store you'd like to query on the host zone
	//	 e.g. for querying a validator, `data := stakingtypes.GetValidatorKey(valAddr)`
	request := []byte{0x01, 0x02, 0x03}
	// ttl is the expiry time of the query, in absolute units of time, unix nanos
	ttl := uint64(0) // ttl

	return NewQueryTestCase{
		module:       module,
		callbackId:   callbackId,
		chainId:      chainId,
		connectionId: connectionId,
		queryType:    queryType,
		request:      request,
		ttl:          ttl,
	}
}

func (s *KeeperTestSuite) TestNewQuerySuccessful() {
	tc := s.SetupNewQuery()

	actualQuery := s.App.InterchainqueryKeeper.NewQuery(
		s.Ctx,
		tc.module,
		tc.callbackId,
		tc.chainId,
		tc.connectionId,
		tc.queryType,
		tc.request,
		tc.ttl,
	)

	// this hash is testing `GenerateQueryHash`.
	//  note: the module gets hashed in the `GenerateQueryHash` function, so hashes will be unique to each module
	//  note: the queryID is a has of (module, callbackId, chainId, connectionId, queryType, and request)
	// .    meaning for a given query type, the ID will be identical across each epoch
	expectedId := "e97f7bdad3c4c521165321f78a8329c54f35db23ee9cec7bddf5c60703ac9ba7"
	s.Require().Equal(expectedId, actualQuery.Id)

	// RequestSent should be false
	s.Require().False(actualQuery.RequestSent)

	// all other arguments should be the same as the input
	s.Require().Equal(tc.connectionId, actualQuery.ConnectionId)
	s.Require().Equal(tc.chainId, actualQuery.ChainId)
	s.Require().Equal(tc.queryType, actualQuery.QueryType)
	s.Require().Equal(tc.request, actualQuery.Request)
	s.Require().Equal(tc.callbackId, actualQuery.CallbackId)
	s.Require().Equal(tc.ttl, actualQuery.Ttl)
}
