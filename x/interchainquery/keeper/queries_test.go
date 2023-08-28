package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetQueryId() {
	// chain Id of the target chain you're querying
	chainId := "GAIA"
	// connectionId of the target chain you're querying
	connectionId := "connection-0"
	// QueryType is a string that is used to identify the store you'd like to query, as well as whether you'd like a proof returned alongside the result
	// use "staking" store to access validator which lives in the staking module
	// use "key" suffix to retrieve a proof alongside the query result
	queryType := types.STAKING_STORE_QUERY_WITH_PROOF
	// requestData is a byte array that points to the entry in the target store you'd like to query on the host zone
	//	 e.g. for querying a validator, `data := stakingtypes.GetValidatorKey(valAddr)`
	requestData := []byte{0x01, 0x02, 0x03}
	// module is the name of the module invoking the query, used to find the callback upon response
	module := "stakeibc"
	// callbackId is a string that is used to identify the callback you'd like to execute upon receiving the result of the query
	callbackId := "validator"
	// timeout is the expiry time of the query, in absolute units of time, unix nanos
	timeoutTimestamp := uint64(100) // timeout

	//  note: the queryID is a has of (module, callbackId, chainId, connectionId, queryType, and request)
	// .    meaning for a given query type, the ID will be identical across each epoch
	expectedQueryId := "e97f7bdad3c4c521165321f78a8329c54f35db23ee9cec7bddf5c60703ac9ba7"
	expectedUniqueQueryId := "cd2662154e6d76b2b2b92e70c0cac3ccf534f9b74eb5b89819ec509083d00a50"

	query := types.Query{
		ChainId:          chainId,
		ConnectionId:     connectionId,
		QueryType:        queryType,
		RequestData:      requestData,
		CallbackModule:   module,
		CallbackId:       callbackId,
		TimeoutTimestamp: timeoutTimestamp,
	}

	queryId := s.App.InterchainqueryKeeper.GetQueryId(s.Ctx, query, false)
	uniqueQueryId := s.App.InterchainqueryKeeper.GetQueryId(s.Ctx, query, true)

	s.Require().Equal(expectedQueryId, queryId, "query ID")
	s.Require().Equal(expectedUniqueQueryId, uniqueQueryId, "unique query ID")
}

func (s *KeeperTestSuite) TestValidateQuery() {
	validChainId := "chain-0"
	validConnectionId := "connection-0"
	validQueryType := "store/key/query"
	validTimeout := time.Duration(10)

	s.Ctx = s.Ctx.WithBlockTime(time.Unix(0, 0)) // unix 0

	// We'll borrow a callback from stakeibc since it's should be already registered in the App
	validCallbackModule := stakeibctypes.ModuleName
	validCallbackId := stakeibckeeper.ICQCallbackID_Delegation

	testCases := []struct {
		name          string
		query         types.Query
		expectedError string
	}{
		{
			name: "valid query",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
		},
		{
			name: "missing chain id",
			query: types.Query{
				ChainId:         "",
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "chain-id cannot be empty",
		},
		{
			name: "missing connection id",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    "",
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "connection-id cannot be empty",
		},
		{
			name: "invalid connection id",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    "connection",
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "invalid connection-id (connection)",
		},
		{
			name: "invalid query type",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       "",
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "query type cannot be empty",
		},
		{
			name: "missing callback module",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  "",
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "callback module must be specified",
		},
		{
			name: "missing callback-id",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      "",
				TimeoutDuration: validTimeout,
			},
			expectedError: "callback-id cannot be empty",
		},
		{
			name: "invalid timeout duration",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      validCallbackId,
				TimeoutDuration: time.Duration(0),
			},
			expectedError: "timeout duration must be set",
		},
		{
			name: "module not registered",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  "fake-module",
				CallbackId:      validCallbackId,
				TimeoutDuration: validTimeout,
			},
			expectedError: "no callback handler registered for module (fake-module)",
		},
		{
			name: "callback not registered",
			query: types.Query{
				ChainId:         validChainId,
				ConnectionId:    validConnectionId,
				QueryType:       validQueryType,
				CallbackModule:  validCallbackModule,
				CallbackId:      "fake-callback",
				TimeoutDuration: validTimeout,
			},
			expectedError: "callback-id (fake-callback) is not registered for module (stakeibc)",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actualError := s.App.InterchainqueryKeeper.ValidateQuery(s.Ctx, tc.query)
			if tc.expectedError == "" {
				s.Require().NoError(actualError)
			} else {
				s.Require().ErrorContains(actualError, tc.expectedError)
			}
		})
	}
}

func (s *KeeperTestSuite) GetQueryUID() {
	// Helper function to get the next uid
	getUniqueSuffix := func() int {
		return int(binary.BigEndian.Uint64(s.App.InterchainqueryKeeper.GetQueryUID(s.Ctx)))
	}

	// Grabbing the uid for the first time should return 1
	s.Require().Equal(1, getUniqueSuffix())

	// Grabbing it a second time should return 2
	s.Require().Equal(2, getUniqueSuffix())

	// call it 1000 more times
	var suffix int
	for i := 0; i < 1000; i++ {
		suffix = getUniqueSuffix()
	}
	s.Require().Equal(1002, suffix)
}

func TestUnmarshalAmountFromBalanceQuery(t *testing.T) {
	type InputType int64
	const (
		rawBytes InputType = iota
		coinType
		intType
	)

	testCases := []struct {
		name           string
		inputType      InputType
		raw            []byte
		coin           sdk.Coin
		integer        sdkmath.Int
		expectedAmount sdkmath.Int
		expectedError  string
	}{
		{
			name:           "full_coin",
			inputType:      coinType,
			coin:           sdk.Coin{Denom: "denom", Amount: sdkmath.NewInt(50)},
			expectedAmount: sdkmath.NewInt(50),
		},
		{
			name:           "coin_no_denom",
			inputType:      coinType,
			coin:           sdk.Coin{Amount: sdkmath.NewInt(60)},
			expectedAmount: sdkmath.NewInt(60),
		},
		{
			name:           "coin_no_amount",
			inputType:      coinType,
			coin:           sdk.Coin{Denom: "denom"},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "zero_coin",
			inputType:      coinType,
			coin:           sdk.Coin{Amount: sdkmath.NewInt(0)},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_coin",
			inputType:      coinType,
			coin:           sdk.Coin{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "positive_int",
			inputType:      intType,
			integer:        sdkmath.NewInt(20),
			expectedAmount: sdkmath.NewInt(20),
		},
		{
			name:           "zero_int",
			inputType:      intType,
			integer:        sdkmath.NewInt(0),
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_int",
			inputType:      intType,
			integer:        sdkmath.Int{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_bytes",
			inputType:      rawBytes,
			raw:            []byte{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:          "invalid_bytes",
			inputType:     rawBytes,
			raw:           []byte{1, 2},
			expectedError: "unable to unmarshal balance query response",
		},
		{
			name:          "nil_bytes",
			inputType:     rawBytes,
			raw:           nil,
			expectedError: "query response is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var args []byte
			var err error
			switch tc.inputType {
			case rawBytes:
				args = tc.raw
			case coinType:
				args, err = tc.coin.Marshal()
			case intType:
				args, err = tc.integer.Marshal()
			}
			require.NoError(t, err)

			if tc.expectedError == "" {
				actualAmount, err := keeper.UnmarshalAmountFromBalanceQuery(types.ModuleCdc, args)
				require.NoError(t, err)
				require.Equal(t, tc.expectedAmount.Int64(), actualAmount.Int64())
			} else {
				_, err := keeper.UnmarshalAmountFromBalanceQuery(types.ModuleCdc, args)
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
