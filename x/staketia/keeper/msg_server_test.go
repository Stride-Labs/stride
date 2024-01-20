package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// More granular testing of liquid stake is done in the keeper function
// This just tests the msg server wrapper
func (s *KeeperTestSuite) TestMsgServerLiquidStake() {
	tc := s.DefaultSetupTestLiquidStake()

	// Attempt a successful liquid stake
	validMsg := types.MsgLiquidStake{
		Staker:       tc.stakerAddress.String(),
		NativeAmount: tc.liquidStakeAmount,
	}
	resp, err := s.GetMsgServer().LiquidStake(sdk.UnwrapSDKContext(s.Ctx), &validMsg)
	s.Require().NoError(err, "no error expected during liquid stake")
	s.Require().Equal(tc.expectedStAmount.Int64(), resp.StToken.Amount.Int64(), "stToken amount")

	s.ConfirmLiquidStakeTokenTransfer(tc)

	// Attempt a liquid stake again, it should fail now that the staker is out of funds
	_, err = s.GetMsgServer().LiquidStake(sdk.UnwrapSDKContext(s.Ctx), &validMsg)
	s.Require().ErrorContains(err, "insufficient funds")
}

func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds() {
	// Register a host zone
	zone := types.HostZone{
		ChainId: HostChainId,
		// Upper bound 1.5
		MaxRedemptionRate: sdk.NewDec(3).Quo(sdk.NewDec(2)),
		// Lower bound 0.9
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
	}

	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	initialMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(90).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(105).Quo(sdk.NewDec(100)),
	}

	updateMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(95).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(11).Quo(sdk.NewDec(10)),
	}

	invalidMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(0),
		MaxInnerRedemptionRate: sdk.NewDec(2),
	}

	// Set the inner bounds on the host zone for the first time
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &initialMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone = s.MustGetHostZone()
	s.Require().Equal(initialMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(initialMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")

	// Update the inner bounds on the host zone
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &updateMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone = s.MustGetHostZone()
	s.Require().Equal(updateMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(updateMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")

	// Set the inner bounds on the host zone for the first time
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "invalid host zone redemption rate inner bounds")
}
