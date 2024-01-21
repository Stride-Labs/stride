package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

type LiquidStakeTestCase struct {
	liquidStakeAmount sdkmath.Int
	expectedStAmount  sdkmath.Int
	stakerAddress     sdk.AccAddress
	depositAddress    sdk.AccAddress
}

// ----------------------------------------------------
//                LiquidStake
// ----------------------------------------------------

// Helper function to mock relevant state before testing a liquid stake
func (s *KeeperTestSuite) SetupTestLiquidStake(
	redemptionRate sdk.Dec,
	liquidStakeAmount,
	expectedStAmount sdkmath.Int,
) LiquidStakeTestCase {
	// Create relevant addresses
	stakerAddress := s.TestAccs[0]
	depositAddress := s.TestAccs[1]

	// Create a host zone with relevant denom's and addresses
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:                HostChainId,
		NativeTokenDenom:       HostNativeDenom,
		NativeTokenIbcDenom:    HostIBCDenom,
		DepositAddress:         depositAddress.String(),
		RedemptionRate:         redemptionRate,
		MinRedemptionRate:      redemptionRate.Sub(sdk.MustNewDecFromStr("0.2")),
		MinInnerRedemptionRate: redemptionRate.Sub(sdk.MustNewDecFromStr("0.1")),
		MaxInnerRedemptionRate: redemptionRate.Add(sdk.MustNewDecFromStr("0.1")),
		MaxRedemptionRate:      redemptionRate.Add(sdk.MustNewDecFromStr("0.2")),
	})

	// Fund the staker
	liquidStakeToken := sdk.NewCoin(HostIBCDenom, liquidStakeAmount)
	s.FundAccount(stakerAddress, liquidStakeToken)

	return LiquidStakeTestCase{
		liquidStakeAmount: liquidStakeAmount,
		expectedStAmount:  expectedStAmount,
		stakerAddress:     stakerAddress,
		depositAddress:    depositAddress,
	}
}

// Helper function to setup the state with default values
// (useful when testing error cases)
func (s *KeeperTestSuite) DefaultSetupTestLiquidStake() LiquidStakeTestCase {
	redemptionRate := sdk.MustNewDecFromStr("1.0")
	liquidStakeAmount := sdkmath.NewInt(1000)
	stAmount := sdkmath.NewInt(1000)
	return s.SetupTestLiquidStake(redemptionRate, liquidStakeAmount, stAmount)
}

// Helper function to confirm balances after a successful liquid stake
func (s *KeeperTestSuite) ConfirmLiquidStakeTokenTransfer(tc LiquidStakeTestCase) {
	zeroNativeTokens := sdk.NewCoin(HostIBCDenom, sdk.ZeroInt())
	liquidStakedNativeTokens := sdk.NewCoin(HostIBCDenom, tc.liquidStakeAmount)

	zeroStTokens := sdk.NewCoin(StDenom, sdk.ZeroInt())
	liquidStakedStTokens := sdk.NewCoin(StDenom, tc.expectedStAmount)

	// Confirm native tokens were escrowed
	// Staker balance should have decreased to zero
	// Deposit balance should have increased by liquid stake amount
	stakerNativeBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.stakerAddress, HostIBCDenom)
	s.CompareCoins(zeroNativeTokens, stakerNativeBalance, "staker native balance")

	depositNativeBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.depositAddress, HostIBCDenom)
	s.CompareCoins(liquidStakedNativeTokens, depositNativeBalance, "deposit native balance")

	// Confirm stTokens were minted to the user
	// Staker balance should increase by the liquid stake amount
	// Deposit balance should still be zero
	stakerStBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.stakerAddress, StDenom)
	s.CompareCoins(liquidStakedStTokens, stakerStBalance, "staker stToken balance")

	depositStBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.depositAddress, StDenom)
	s.CompareCoins(zeroStTokens, depositStBalance, "deposit native balance")
}

func (s *KeeperTestSuite) TestLiquidStake_Successful() {
	// Test liquid stake across different redemption rates
	testCases := []struct {
		name              string
		redemptionRate    sdk.Dec
		liquidStakeAmount sdkmath.Int
		expectedStAmount  sdkmath.Int
	}{
		{
			// Redemption Rate of 1:
			// 1000 native -> 1000 stTokens
			name:              "redemption rate of 1",
			redemptionRate:    sdk.MustNewDecFromStr("1.0"),
			liquidStakeAmount: sdkmath.NewInt(1000),
			expectedStAmount:  sdkmath.NewInt(1000),
		},
		{
			// Redemption Rate of 2:
			// 1000 native -> 500 stTokens
			name:              "redemption rate of 2",
			redemptionRate:    sdk.MustNewDecFromStr("2.0"),
			liquidStakeAmount: sdkmath.NewInt(1000),
			expectedStAmount:  sdkmath.NewInt(500),
		},
		{
			// Redemption Rate of 0.5:
			// 1000 native -> 2000 stTokens
			name:              "redemption rate of 0.5",
			redemptionRate:    sdk.MustNewDecFromStr("0.5"),
			liquidStakeAmount: sdkmath.NewInt(1000),
			expectedStAmount:  sdkmath.NewInt(2000),
		},
		{
			// Redemption Rate of 1.1:
			// 333 native -> 302.72 (302) stTokens
			name:              "int truncation",
			redemptionRate:    sdk.MustNewDecFromStr("1.1"),
			liquidStakeAmount: sdkmath.NewInt(333),
			expectedStAmount:  sdkmath.NewInt(302),
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			s.SetupTest() // reset state
			tc := s.SetupTestLiquidStake(testCase.redemptionRate, testCase.liquidStakeAmount, testCase.expectedStAmount)

			// Confirm liquid stake succeeded
			stTokenResponse, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
			s.Require().NoError(err, "no error expected during liquid stake")

			// Confirm the stToken from the response matches expectations
			s.Require().Equal(StDenom, stTokenResponse.Denom, "st token denom in liquid stake response")
			s.Require().Equal(tc.expectedStAmount.Int64(), stTokenResponse.Amount.Int64(),
				"st token amount in liquid stake response")

			// Confirm the native token escrow and stToken mint succeeded
			s.ConfirmLiquidStakeTokenTransfer(tc)
		})
	}
}

func (s *KeeperTestSuite) TestLiquidStake_HostZoneHalted() {
	tc := s.DefaultSetupTestLiquidStake()

	// Halt the host zone so the liquid stake fails
	hostZone := s.MustGetHostZone()
	hostZone.Halted = true
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "host zone is halted")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidAddresse() {
	tc := s.DefaultSetupTestLiquidStake()

	// Pass an invalid staker address and confirm it fails
	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, "invalid_address", tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "user's address is invalid")

	// Set an invalid deposit address and confirm it fails
	hostZone := s.MustGetHostZone()
	hostZone.DepositAddress = "invalid_address"
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	_, err = s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "host zone deposit address is invalid")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidRedemptionRate() {
	tc := s.DefaultSetupTestLiquidStake()

	// Update the redemption rate so it exceeds the bounds
	hostZone := s.MustGetHostZone()
	hostZone.RedemptionRate = hostZone.MaxInnerRedemptionRate.Add(sdk.MustNewDecFromStr("0.01"))
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "redemption rate outside inner safety bounds")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidIBCDenom() {
	tc := s.DefaultSetupTestLiquidStake()

	// Set an invalid IBC denom on the host so the liquid stake fails
	hostZone := s.MustGetHostZone()
	hostZone.NativeTokenIbcDenom = "non-ibc-denom"
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "denom is not an IBC token")
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientLiquidStake() {
	// Adjust redemption rate so that a small liquid stake will result in 0 stTokens
	// stTokens = 1(amount) / 1.1(RR) = rounds down to 0
	redemptionRate := sdk.MustNewDecFromStr("1.1")
	liquidStakeAmount := sdkmath.NewInt(1)
	expectedStAmount := sdkmath.ZeroInt()
	tc := s.SetupTestLiquidStake(redemptionRate, liquidStakeAmount, expectedStAmount)

	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "Liquid staked amount is too small")
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientFunds() {
	// Attempt to liquid stake more tokens than the staker has available
	tc := s.DefaultSetupTestLiquidStake()

	excessiveLiquidStakeAmount := sdkmath.NewInt(10000000000)
	_, err := s.App.StaketiaKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), excessiveLiquidStakeAmount)
	s.Require().ErrorContains(err, "failed to send tokens from liquid staker")
	s.Require().ErrorContains(err, "insufficient funds")
}

// ----------------------------------------------------
//	               PrepareDelegation
// ----------------------------------------------------

func (s *KeeperTestSuite) TestPrepareDelegation() {
	s.CreateTransferChannel(HostChainId)

	// Only the deposit address must be valid
	depositAddress := s.TestAccs[0]
	delegationAddress := "celestiaXXX"

	// We must use a valid IBC denom for this test
	nativeIbcDenom := s.CreateAndStoreIBCDenom(HostNativeDenom)

	// Create the host zone with relevant addresses and an IBC denom
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		DepositAddress:      depositAddress.String(),
		DelegationAddress:   delegationAddress,
		NativeTokenIbcDenom: nativeIbcDenom,
		TransferChannelId:   ibctesting.FirstChannelID,
	})

	// Fund the deposit account with tokens that will be transferred
	depositAccountBalance := sdkmath.NewInt(1_000_000)
	nativeTokensInDeposit := sdk.NewCoin(nativeIbcDenom, depositAccountBalance)
	s.FundAccount(depositAddress, nativeTokensInDeposit)

	// Get next sequence number to confirm IBC transfer
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	// submit prepare delegation
	epochNumber := uint64(1)
	epochDuration := time.Hour * 24
	err := s.App.StaketiaKeeper.PrepareDelegation(s.Ctx, epochNumber, epochDuration)
	s.Require().NoError(err, "no error expected when preparing delegation")

	// check that a delegation record was created
	delegationRecords := s.App.StaketiaKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(1, len(delegationRecords), "number of delegation records")

	// check that the delegation record has the correct id, status, and amount
	delegationRecord := delegationRecords[0]
	s.Require().Equal(epochNumber, delegationRecord.Id, "delegation record epoch number")
	s.Require().Equal(types.TRANSFER_IN_PROGRESS, delegationRecord.Status, "delegation record status")
	s.Require().Equal(depositAccountBalance, delegationRecord.NativeAmount, "delegation record amount")

	// check that the transfer in progress record was created
	transferInProgressRecordId, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, ibctesting.FirstChannelID, startSequence)
	s.Require().True(found, "transfer in progress record should have been found")
	s.Require().Equal(epochNumber, transferInProgressRecordId, "transfer in progress record ID")

	// check that the tokens were burned and the sequence number was incremented
	// (indicating that the transfer was submitted)
	endSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
	s.Require().Equal(startSequence+1, endSequence, "sequence number should have incremented")

	nativeTokenSupply := s.App.BankKeeper.GetSupply(s.Ctx, nativeIbcDenom)
	s.Require().Zero(nativeTokenSupply.Amount.Int64(), "ibc tokens should have been burned")

	// Check that the deposit account is empty
	depositAccountBalance = s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, nativeIbcDenom).Amount
	s.Require().Zero(depositAccountBalance.Int64(), "deposit account balance should be empty")
}
