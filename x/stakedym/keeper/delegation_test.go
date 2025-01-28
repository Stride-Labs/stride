package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v25/x/stakedym/types"
)

var InitialDelegation = sdkmath.NewInt(1_000_000)

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
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
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
			stTokenResponse, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
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
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "host zone is halted")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidAddresse() {
	tc := s.DefaultSetupTestLiquidStake()

	// Pass an invalid staker address and confirm it fails
	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, "invalid_address", tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "user's address is invalid")

	// Set an invalid deposit address and confirm it fails
	hostZone := s.MustGetHostZone()
	hostZone.DepositAddress = "invalid_address"
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	_, err = s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "host zone deposit address is invalid")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidRedemptionRate() {
	tc := s.DefaultSetupTestLiquidStake()

	// Update the redemption rate so it exceeds the bounds
	hostZone := s.MustGetHostZone()
	hostZone.RedemptionRate = hostZone.MaxInnerRedemptionRate.Add(sdk.MustNewDecFromStr("0.01"))
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "redemption rate outside inner safety bounds")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidIBCDenom() {
	tc := s.DefaultSetupTestLiquidStake()

	// Set an invalid IBC denom on the host so the liquid stake fails
	hostZone := s.MustGetHostZone()
	hostZone.NativeTokenIbcDenom = "non-ibc-denom"
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "denom is not an IBC token")
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientLiquidStake() {
	// Adjust redemption rate so that a small liquid stake will result in 0 stTokens
	// stTokens = 1(amount) / 1.1(RR) = rounds down to 0
	redemptionRate := sdk.MustNewDecFromStr("1.1")
	liquidStakeAmount := sdkmath.NewInt(1)
	expectedStAmount := sdkmath.ZeroInt()
	tc := s.SetupTestLiquidStake(redemptionRate, liquidStakeAmount, expectedStAmount)

	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), tc.liquidStakeAmount)
	s.Require().ErrorContains(err, "Liquid staked amount is too small")
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientFunds() {
	// Attempt to liquid stake more tokens than the staker has available
	tc := s.DefaultSetupTestLiquidStake()

	excessiveLiquidStakeAmount := sdkmath.NewInt(10000000000)
	_, err := s.App.StakedymKeeper.LiquidStake(s.Ctx, tc.stakerAddress.String(), excessiveLiquidStakeAmount)
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
	delegationAddress := "dymXXX"

	// We must use a valid IBC denom for this test
	nativeIbcDenom := s.CreateAndStoreIBCDenom(HostNativeDenom)

	// Create the host zone with relevant addresses and an IBC denom
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
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
	err := s.App.StakedymKeeper.PrepareDelegation(s.Ctx, epochNumber, epochDuration)
	s.Require().NoError(err, "no error expected when preparing delegation")

	// check that a delegation record was created
	delegationRecords := s.App.StakedymKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(1, len(delegationRecords), "number of delegation records")

	// check that the delegation record has the correct id, status, and amount
	delegationRecord := delegationRecords[0]
	s.Require().Equal(epochNumber, delegationRecord.Id, "delegation record epoch number")
	s.Require().Equal(types.TRANSFER_IN_PROGRESS, delegationRecord.Status, "delegation record status")
	s.Require().Equal(depositAccountBalance, delegationRecord.NativeAmount, "delegation record amount")

	// check that the transfer in progress record was created
	transferInProgressRecordId, found := s.App.StakedymKeeper.GetTransferInProgressRecordId(s.Ctx, ibctesting.FirstChannelID, startSequence)
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

	// Check that if we ran this again immediately, it would error because there is a transfer record in progress already
	err = s.App.StakedymKeeper.PrepareDelegation(s.Ctx, epochNumber+1, epochDuration)
	s.Require().ErrorContains(err, "cannot prepare delegation while a transfer is in progress")

	// Remove the record and try to run it again
	s.App.StakedymKeeper.ArchiveDelegationRecord(s.Ctx, delegationRecord)
	err = s.App.StakedymKeeper.PrepareDelegation(s.Ctx, epochNumber+1, epochDuration)
	s.Require().NoError(err, "no error expected when preparing delegation again")

	// It should not create a new record since there is nothing to delegate
	delegationRecords = s.App.StakedymKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(0, len(delegationRecords), "there should be no delegation records")

	// Halt zone
	s.App.StakedymKeeper.HaltZone(s.Ctx)
	err = s.App.StakedymKeeper.PrepareDelegation(s.Ctx, epochNumber, epochDuration)
	s.Require().ErrorContains(err, "host zone is halted")
}

// ----------------------------------------------------
//	               ConfirmDelegation
// ----------------------------------------------------

func (s *KeeperTestSuite) GetDefaultDelegationRecords() []types.DelegationRecord {
	delegationRecords := []types.DelegationRecord{
		{
			Id:           1,
			NativeAmount: sdk.NewInt(1000),
			Status:       types.TRANSFER_IN_PROGRESS,
			TxHash:       "",
		},
		{
			Id:           6, // out of order to make sure this won't break anything
			NativeAmount: sdk.NewInt(6000),
			Status:       types.DELEGATION_QUEUE, // to be set
			TxHash:       "",
		},
		{
			Id:           5, // out of order to make sure this won't break anything
			NativeAmount: sdk.NewInt(5000),
			Status:       types.TRANSFER_IN_PROGRESS,
			TxHash:       "",
		},
		{
			Id:           3,
			NativeAmount: sdk.NewInt(3000),
			Status:       types.TRANSFER_FAILED,
			TxHash:       "",
		},
		{
			Id:           2,
			NativeAmount: sdk.NewInt(2000),
			Status:       types.DELEGATION_QUEUE, // to be set
			TxHash:       "",
		},
		{
			Id:           7,
			NativeAmount: sdk.NewInt(7000),
			Status:       types.TRANSFER_FAILED,
			TxHash:       ValidTxHashDefault,
		},
	}

	return delegationRecords
}

// Helper function to setup delegation records, returns a list of records
func (s *KeeperTestSuite) SetupDelegationRecords() {
	// Set Delegation Records
	delegationRecords := s.GetDefaultDelegationRecords()
	// loop through and set each record
	for _, delegationRecord := range delegationRecords {
		s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	}

	// Set HostZone
	hostZone := s.initializeHostZone()
	hostZone.DelegatedBalance = InitialDelegation
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)
}

func (s *KeeperTestSuite) VerifyDelegationRecords(verifyIdentical bool, archiveIds ...uint64) {
	defaultDelegationRecords := s.GetDefaultDelegationRecords()

	hostZone := s.MustGetHostZone()

	for _, defaultDelegationRecord := range defaultDelegationRecords {
		// check if record should be archived
		shouldBeArchived := false
		for _, archiveId := range archiveIds {
			if defaultDelegationRecord.Id == archiveId {
				shouldBeArchived = true
				break
			}
		}

		// grab relevant record in store
		loadedDelegationRecord := types.DelegationRecord{}
		found := false
		if shouldBeArchived {
			loadedDelegationRecord, found = s.App.StakedymKeeper.GetArchivedDelegationRecord(s.Ctx, defaultDelegationRecord.Id)
		} else {
			loadedDelegationRecord, found = s.App.StakedymKeeper.GetDelegationRecord(s.Ctx, defaultDelegationRecord.Id)
		}
		s.Require().True(found)
		// verify record is correct
		s.Require().Equal(defaultDelegationRecord.Id, loadedDelegationRecord.Id)
		s.Require().Equal(defaultDelegationRecord.NativeAmount, loadedDelegationRecord.NativeAmount)

		// Verify status and txHash are correct, if needed
		if (defaultDelegationRecord.Status == types.TRANSFER_FAILED) ||
			(defaultDelegationRecord.Status == types.TRANSFER_IN_PROGRESS) ||
			verifyIdentical {
			s.Require().Equal(defaultDelegationRecord.Status, loadedDelegationRecord.Status)
			s.Require().Equal(defaultDelegationRecord.TxHash, loadedDelegationRecord.TxHash)
		}

		// if nothing should have changed, verify that host zone balance is unmodified
		if verifyIdentical {
			// verify hostZone delegated balance is same as initial delegation
			s.Require().Equal(InitialDelegation.Int64(), hostZone.DelegatedBalance.Int64(), "hostZone delegated balance should not have changed")
		}
	}
}

func (s *KeeperTestSuite) TestConfirmDelegation_Successful() {
	s.SetupDelegationRecords()

	// we're halting the zone to test that the tx works even when the host zone is halted
	s.App.StakedymKeeper.HaltZone(s.Ctx)

	// try setting valid delegation queue
	err := s.App.StakedymKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyDelegationRecords(false, 6)

	// verify record 6 modified
	loadedDelegationRecord, found := s.App.StakedymKeeper.GetArchivedDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	s.Require().Equal(types.DELEGATION_COMPLETE, loadedDelegationRecord.Status, "delegation record should be updated to status DELEGATION_ARCHIVE")
	s.Require().Equal(ValidTxHashNew, loadedDelegationRecord.TxHash, "delegation record should be updated with txHash")

	// verify hostZone delegated balance is same as initial delegation + 6000
	hostZone := s.MustGetHostZone()
	s.Require().Equal(InitialDelegation.Int64()+6000, hostZone.DelegatedBalance.Int64(), "hostZone delegated balance should have increased by 6000")
}

func (s *KeeperTestSuite) TestConfirmDelegation_DelegationZero() {
	s.SetupDelegationRecords()

	// try setting delegation queue with zero delegation
	delegationRecord, found := s.App.StakedymKeeper.GetDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	delegationRecord.NativeAmount = sdk.NewInt(0)
	s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	err := s.App.StakedymKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState, "not allowed to confirm zero delegation")
}

func (s *KeeperTestSuite) TestConfirmDelegation_DelegationNegative() {
	s.SetupDelegationRecords()

	// try setting delegation queue with negative delegation
	delegationRecord, found := s.App.StakedymKeeper.GetDelegationRecord(s.Ctx, 6)
	s.Require().True(found)
	delegationRecord.NativeAmount = sdk.NewInt(-10)
	s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	err := s.App.StakedymKeeper.ConfirmDelegation(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState, "not allowed to confirm negative delegation")
}

func (s *KeeperTestSuite) TestConfirmDelegation_RecordDoesntExist() {
	s.SetupDelegationRecords()

	// try setting invalid record id
	err := s.App.StakedymKeeper.ConfirmDelegation(s.Ctx, 15, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrDelegationRecordNotFound)

	// verify delegation records haven't changed
	s.VerifyDelegationRecords(true)
}

func (s *KeeperTestSuite) TestConfirmDelegation_RecordIncorrectState() {
	s.SetupDelegationRecords()

	// first verify records in wrong status
	ids := []uint64{1, 3, 5, 7}
	for _, id := range ids {
		err := s.App.StakedymKeeper.ConfirmDelegation(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrDelegationRecordInvalidState)
		// verify delegation records haven't changed
		s.VerifyDelegationRecords(true)
	}
}

// ----------------------------------------------------
//	          LiquidStakeAndDistributeFees
// ----------------------------------------------------

func (s *KeeperTestSuite) TestLiquidStakeAndDistributeFees() {
	// Create relevant addresses
	depositAddress := s.TestAccs[0]
	feeAddress := s.App.AccountKeeper.GetModuleAddress(types.FeeAddress)

	// Liquid stake 1000 with a RR of 2, should return 500 tokens
	liquidStakeAmount := sdkmath.NewInt(1000)
	redemptionRate := sdk.NewDec(2)
	expectedStTokens := sdkmath.NewInt(500)

	// Create a host zone with relevant denom's and addresses
	hostZone := types.HostZone{
		ChainId:                HostChainId,
		NativeTokenDenom:       HostNativeDenom,
		NativeTokenIbcDenom:    HostIBCDenom,
		DepositAddress:         depositAddress.String(),
		RedemptionRate:         redemptionRate,
		MinRedemptionRate:      redemptionRate.Sub(sdk.MustNewDecFromStr("0.2")),
		MinInnerRedemptionRate: redemptionRate.Sub(sdk.MustNewDecFromStr("0.1")),
		MaxInnerRedemptionRate: redemptionRate.Add(sdk.MustNewDecFromStr("0.1")),
		MaxRedemptionRate:      redemptionRate.Add(sdk.MustNewDecFromStr("0.2")),
	}
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	// Fund the fee address with native tokens
	liquidStakeToken := sdk.NewCoin(HostIBCDenom, liquidStakeAmount)
	s.FundAccount(feeAddress, liquidStakeToken)

	// Call liquid stake and distribute
	err := s.App.StakedymKeeper.LiquidStakeAndDistributeFees(s.Ctx)
	s.Require().NoError(err, "no error expected when liquid staking fee tokens")

	// Confirm stTokens were sent to the fee collector
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	feeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddress, StDenom)
	s.Require().Equal(expectedStTokens.Int64(), feeCollectorBalance.Amount.Int64(),
		"fee collector should have received sttokens")

	// Attempt to liquid stake again when there are no more rewards, it should succeed but do nothing
	err = s.App.StakedymKeeper.LiquidStakeAndDistributeFees(s.Ctx)
	s.Require().NoError(err, "no error expected when liquid staking again")

	feeCollectorBalance = s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddress, StDenom)
	s.Require().Equal(expectedStTokens.Int64(), feeCollectorBalance.Amount.Int64(),
		"fee collector should not have changed")

	// Test that if the host zone is halted, it will error
	haltedHostZone := hostZone
	haltedHostZone.Halted = true
	s.App.StakedymKeeper.SetHostZone(s.Ctx, haltedHostZone)

	err = s.App.StakedymKeeper.LiquidStakeAndDistributeFees(s.Ctx)
	s.Require().ErrorContains(err, "host zone is halted")
}
