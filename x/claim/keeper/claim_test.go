package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:         addr1.String(),
			Weight:          sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted: []bool{false, false, false},
		},
	}

	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, claimRecords[0].Weight, coins1.String())

	coins3, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Equal(coins3, sdk.Coins{})

	// get rewards amount for free
	coins4, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionFree)
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).String())

	// get completed activities
	claimedCoins, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionFree)
	suite.Require().NoError(err)
	suite.Require().Equal(claimedCoins.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).String())

	// get balance after free claim
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).String())

	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, sdk.DefaultBondDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 9000000))
}
