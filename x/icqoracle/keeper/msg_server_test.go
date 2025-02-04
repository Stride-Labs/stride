package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

func (s *KeeperTestSuite) TestRegisterTokenPriceQuery() {
	// Create a new token price query
	msg := types.MsgRegisterTokenPriceQuery{
		BaseDenom:         "uatom",
		QuoteDenom:        "uusdc",
		OsmosisPoolId:     "1",
		OsmosisBaseDenom:  "ibc/uatom",
		OsmosisQuoteDenom: "uusdc",
	}
	_, err := s.GetMsgServer().RegisterTokenPriceQuery(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when registering token price query")

	// Confirm the token price was created
	tokenPrice := s.MustGetTokenPrice(msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)

	s.Require().Equal(msg.BaseDenom, tokenPrice.BaseDenom, "base denom")
	s.Require().Equal(msg.QuoteDenom, tokenPrice.QuoteDenom, "quote denom")
	s.Require().Equal(msg.OsmosisPoolId, tokenPrice.OsmosisPoolId, "osmosis pool id")
	s.Require().Equal(msg.OsmosisBaseDenom, tokenPrice.OsmosisBaseDenom, "osmosis base denom")
	s.Require().Equal(msg.OsmosisQuoteDenom, tokenPrice.OsmosisQuoteDenom, "osmosis quote denom")
	s.Require().Equal(sdkmath.LegacyZeroDec(), tokenPrice.SpotPrice, "spot price")
	s.Require().Equal(false, tokenPrice.QueryInProgress, "query in progress")
	s.Require().Equal(time.Time{}, tokenPrice.LastQueryTime, "updated at")

	// Attempt to register it again, it should fail
	_, err = s.GetMsgServer().RegisterTokenPriceQuery(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrTokenPriceAlreadyExists)
}

func (s *KeeperTestSuite) TestRemoveTokenPriceQuery() {
	// Create a token price
	tokenPrice := types.TokenPrice{
		BaseDenom:         "uatom",
		QuoteDenom:        "uusdc",
		OsmosisPoolId:     "1",
		OsmosisBaseDenom:  "ibc/uatom",
		OsmosisQuoteDenom: "uusdc",
		SpotPrice:         sdkmath.LegacyNewDec(1),
		LastQueryTime:     time.Now().UTC(),
		QueryInProgress:   false,
	}
	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err, "no error expected when setting token price")

	// Remove the token price
	msg := types.MsgRemoveTokenPriceQuery{
		BaseDenom:     tokenPrice.BaseDenom,
		QuoteDenom:    tokenPrice.QuoteDenom,
		OsmosisPoolId: tokenPrice.OsmosisPoolId,
	}
	_, err = s.GetMsgServer().RemoveTokenPriceQuery(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when removing token price query")

	// Confirm the token price was removed
	tp, err := s.App.ICQOracleKeeper.GetTokenPrice(s.Ctx, msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)
	s.Require().Error(err, "token price %+v should have been removed", tp)

	// Try to remove it again, it should still succeed
	_, err = s.GetMsgServer().RemoveTokenPriceQuery(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when removing non-existent token price query")
}
