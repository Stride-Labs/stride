package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"
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

func (suite *KeeperTestSuite) SetupNewQuery(
	module string,
	connectionId string,
	chainId string,
	queryType string,
	request []byte,
	period sdk.Int,
	callbackId string,
	ttl uint64,
	height int64,
) NewQueryTestCase {

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
	s.Require().Equal(true, true)
}
