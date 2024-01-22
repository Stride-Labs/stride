package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// ----------------------------------------------
//                MsgLiquidStake
// ----------------------------------------------

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

// ----------------------------------------------
//             MsgResumeHostZone
// ----------------------------------------------

// Test cases
// - Zone is not halted
// - Zone is halted - unhalt it
func (s *KeeperTestSuite) TestResumeHostZone() {
	// TODO [sttia]: verify denom blacklisting removal works
	// TODO [sttia]: verify this fails if issues by non-admin

	adminAddress := "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh" // admin address
	zone := types.HostZone{
		Halted: false,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)
	msg := types.MsgResumeHostZone{
		Creator: adminAddress,
	}

	// TEST 1: Zone is not halted
	// Try to unhalt the unhalted zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &msg)
	s.Require().ErrorContains(err, "zone is not halted")

	// Confirm the zone is not halted
	zone, err = s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")
	s.Require().False(zone.Halted, "zone should not be halted")

	// TEST 2: Zone is halted
	// Halt the zone
	zone.Halted = true
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	// Try to unhalt the halted zone
	_, err = s.GetMsgServer().ResumeHostZone(s.Ctx, &msg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the zone is not halted
	zone, err = s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")
	s.Require().False(zone.Halted, "zone should not be halted")
}

// ----------------------------------------------
//      MsgUpdateInnerRedemptionRateBounds
// ----------------------------------------------

func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds() {
	// TODO [sttia]: verify this fails if issues by non-admin

	// Register a host zone
	adminAddress := "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh" // admin address
	zone := types.HostZone{
		ChainId: HostChainId,
		// Upper bound 1.5
		MaxRedemptionRate: sdk.NewDec(3).Quo(sdk.NewDec(2)),
		// Lower bound 0.9
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
	}

	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	initialMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                adminAddress,
		MinInnerRedemptionRate: sdk.NewDec(90).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(105).Quo(sdk.NewDec(100)),
	}

	updateMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                adminAddress,
		MinInnerRedemptionRate: sdk.NewDec(95).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(11).Quo(sdk.NewDec(10)),
	}

	invalidMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                adminAddress,
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

// ----------------------------------------------
//            MsgSetOperatorAddress
// ----------------------------------------------

// Verify that operator address can be set successfully
func (s *KeeperTestSuite) TestSetOperatorAddress() {

	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	nonAdminAddress := s.TestAccs[2].String()

	// set the host zone
	zone := types.HostZone{
		SafeAddress:     safeAddress,
		OperatorAddress: operatorAddress,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	// Set the operator address, signed by the SAFE address
	msgSetOperatorAddress := types.MsgSetOperatorAddress{
		Signer:   safeAddress,
		Operator: nonAdminAddress,
	}

	_, err := s.GetMsgServer().SetOperatorAddress(s.Ctx, &msgSetOperatorAddress)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the operator address was updated
	zone, err = s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")
	s.Require().Equal(s.TestAccs[2].String(), zone.OperatorAddress, "operator address should be set")

	// Confirm the operator address cannot be set by a non-safe address
	msgSetOperatorAddressWrongSafe := types.MsgSetOperatorAddress{
		Signer:   operatorAddress,
		Operator: nonAdminAddress,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)
	_, err = s.GetMsgServer().SetOperatorAddress(s.Ctx, &msgSetOperatorAddressWrongSafe)
	s.Require().Error(err, "invalid safe address")
}

// ----------------------------------------------
//         MsgConfirmUnbondingTokensSweep
// ----------------------------------------------

func (s *KeeperTestSuite) SetupUnbondingRecordsAndHostZone() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	hostZone := s.MustGetHostZone()
	hostZone.OperatorAddress = operatorAddress
	hostZone.SafeAddress = safeAddress
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
}

// Verify that ConfirmUnbondingTokenSweep succeeds, and non-admins cannot call it
func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep() {
	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	nonAdminAddress := s.TestAccs[2].String()

	// Confirm that ConfirmDelegation can be called by the operator address
	s.SetupUnbondingRecordsAndHostZone()
	MsgConfirmUnbondedTokenSweepOperator := types.MsgConfirmUnbondedTokenSweep{
		Operator: operatorAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err := s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &MsgConfirmUnbondedTokenSweepOperator)
	s.Require().NoError(err, "operator should be able to confirm unbonded token sweep")

	// Confirm that ConfirmDelegation can be called by the safe address
	s.SetupUnbondingRecordsAndHostZone()
	msgConfirmUnbondedTokenSweepSafe := types.MsgConfirmUnbondedTokenSweep{
		Operator: safeAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &msgConfirmUnbondedTokenSweepSafe)
	s.Require().NoError(err, "safe should be able to confirm unbonded token sweep")

	// Confirm that ConfirmDelegation cannot be called by a non-admin address
	s.SetupUnbondingRecordsAndHostZone()
	msgConfirmUnbondedTokenSweepNonAdmin := types.MsgConfirmUnbondedTokenSweep{
		Operator: nonAdminAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &msgConfirmUnbondedTokenSweepNonAdmin)
	s.Require().ErrorIs(err, types.ErrInvalidAdmin, "non-admin should not be able to confirm unbonded token sweep")
}
