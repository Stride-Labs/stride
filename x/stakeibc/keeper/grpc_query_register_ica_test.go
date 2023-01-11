package keeper_test

import (
	// "testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/stretchr/testify/require"
	// ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// testkeeper "github.com/Stride-Labs/stride/v4/testutil/keeper"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (suite *KeeperTestSuite) TestInterchainAccountFromAddress_Successful() {
	wctx := sdk.WrapSDKContext(suite.Ctx)
	address := "stride1mff2pry767gy7fwldrvdy243mr2ftzedr7gnr0"

	request := types.QueryInterchainAccountFromAddressRequest{
		Owner:        "GAIA.OWNER1",
		ConnectionId: "connection-1",
	}
	portID, _ := icatypes.NewControllerPortID(request.Owner)
	suite.App.StakeibcKeeper.ICAControllerKeeper.SetInterchainAccountAddress(suite.Ctx, request.ConnectionId, portID, address)
	response, err := suite.App.StakeibcKeeper.InterchainAccountFromAddress(wctx, &request)

	suite.Require().NoError(err)
	suite.Require().Equal(types.NewQueryInterchainAccountResponse(address), response)

}
func (suite *KeeperTestSuite) TestInterchainAccountFromAddress_NotFindAccount() {
	wctx := sdk.WrapSDKContext(suite.Ctx)

	request := types.QueryInterchainAccountFromAddressRequest{
		Owner:        "",
		ConnectionId: "connection-2",
	}
	_, err := suite.App.StakeibcKeeper.InterchainAccountFromAddress(wctx, &request)
	err_TestCase := status.Errorf(codes.InvalidArgument, "could not find account: %s", "owner address cannot be empty: invalid account address")
	
	suite.Require().EqualError(err, err_TestCase.Error())
}

func (suite *KeeperTestSuite) TestInterchainAccountFromAddress_NoAccountFoundForPortID() {
	wctx := sdk.WrapSDKContext(suite.Ctx)

	request := types.QueryInterchainAccountFromAddressRequest{
		Owner:        "GAIA.OWNER3",
		ConnectionId: "connection-3",
	}
	_, err := suite.App.StakeibcKeeper.InterchainAccountFromAddress(wctx, &request)
	portID, _ := icatypes.NewControllerPortID(request.Owner)
	err_TestCase := status.Errorf(codes.NotFound, "no account found for portID %s", portID)
	
	suite.Require().EqualError(err, err_TestCase.Error())
}
