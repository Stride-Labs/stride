package keeper_test

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type LSMLiquidStakeTestCase struct {
	hostZone             types.HostZone
	liquidStakerAddress  sdk.AccAddress
	depositAddress       sdk.AccAddress
	initialBalance       sdkmath.Int
	initialQueryProgress sdkmath.Int
	queryCheckpoint      sdkmath.Int
	lsmTokenIBCDenom     string
	validMsg             *types.MsgLSMLiquidStake
}

// Helper function to add the port and channel onto the LSMTokenBaseDenom,
// hash it, and then store the trace in the IBC store
// Returns the ibc hash
func (s *KeeperTestSuite) getLSMTokenIBCDenom() string {
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, ibctesting.FirstChannelID)
	prefixedDenom := sourcePrefix + LSMTokenBaseDenom
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)
	return lsmTokenDenomTrace.IBCDenom()
}

func (s *KeeperTestSuite) SetupTestLSMLiquidStake() LSMLiquidStakeTestCase {
	initialBalance := sdkmath.NewInt(3000)
	stakeAmount := sdkmath.NewInt(1000)
	userAddress := s.TestAccs[0]
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)

	// Need valid IBC denom here to test parsing
	lsmTokenIBCDenom := s.getLSMTokenIBCDenom()

	// Fund the user's account with the LSM token
	s.FundAccount(userAddress, sdk.NewCoin(lsmTokenIBCDenom, initialBalance))

	// Add the slash interval
	// TVL: 100k, Checkpoint: 1% of 1M = 10k
	// Progress towards query: 8000
	// => Liquid Stake of 2k will trip query
	totalHostZoneStake := sdkmath.NewInt(1_000_000)
	queryCheckpoint := sdkmath.NewInt(10_000)
	progressTowardsQuery := sdkmath.NewInt(8000)
	params := types.DefaultParams()
	params.ValidatorSlashQueryThreshold = 1 // 1 %
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// Sanity check
	onePercent := sdk.MustNewDecFromStr("0.01")
	s.Require().Equal(queryCheckpoint.Int64(), onePercent.Mul(sdk.NewDecFromInt(totalHostZoneStake)).TruncateInt64(),
		"setup failed - query checkpoint must be 1% of total host zone stake")

	// Add the host zone with a valid zone address as the LSM custodian
	hostZone := types.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		RedemptionRate:    sdk.NewDec(1.0),
		DepositAddress:    depositAddress.String(),
		TransferChannelId: ibctesting.FirstChannelID,
		ConnectionId:      ibctesting.FirstConnectionID,
		TotalDelegations:  totalHostZoneStake,
		Validators: []*types.Validator{{
			Address:                   ValAddress,
			SlashQueryProgressTracker: progressTowardsQuery,
			SlashQueryCheckpoint:      queryCheckpoint,
			SharesToTokensRate:        sdk.OneDec(),
		}},
		DelegationIcaAddress:  "cosmos_DELEGATION",
		LsmLiquidStakeEnabled: true,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the latest client height for the ICQ submission
	s.MockClientLatestHeight(1)

	return LSMLiquidStakeTestCase{
		hostZone:             hostZone,
		liquidStakerAddress:  userAddress,
		depositAddress:       depositAddress,
		initialBalance:       initialBalance,
		initialQueryProgress: progressTowardsQuery,
		queryCheckpoint:      queryCheckpoint,
		lsmTokenIBCDenom:     lsmTokenIBCDenom,
		validMsg: &types.MsgLSMLiquidStake{
			Creator:          userAddress.String(),
			LsmTokenIbcDenom: lsmTokenIBCDenom,
			Amount:           stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestLSMLiquidStake_Successful_NoSharesToTokensRateQuery() {
	tc := s.SetupTestLSMLiquidStake()

	// Call LSM Liquid stake with a valid message
	msgResponse, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().NoError(err, "no error expected when calling lsm liquid stake")
	s.Require().True(msgResponse.TransactionComplete, "transaction should be complete")

	// Confirm the LSM token was sent to the protocol
	userLsmBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, tc.lsmTokenIBCDenom)
	s.Require().Equal(tc.initialBalance.Sub(tc.validMsg.Amount).Int64(), userLsmBalance.Amount.Int64(),
		"lsm token balance of user account")

	// Confirm stToken was sent to the user
	userStTokenBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom)
	s.Require().Equal(tc.validMsg.Amount.Int64(), userStTokenBalance.Amount.Int64(), "user stToken balance")

	// Confirm an LSMDeposit was created
	expectedDepositId := keeper.GetLSMTokenDepositId(s.Ctx.BlockHeight(), HostChainId, tc.validMsg.Creator, LSMTokenBaseDenom)
	expectedDeposit := recordstypes.LSMTokenDeposit{
		DepositId:        expectedDepositId,
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		StakerAddress:    s.TestAccs[0].String(),
		IbcDenom:         tc.lsmTokenIBCDenom,
		ValidatorAddress: ValAddress,
		Amount:           tc.validMsg.Amount,
		Status:           recordstypes.LSMTokenDeposit_TRANSFER_QUEUE,
		StToken:          sdk.NewCoin(StAtom, tc.validMsg.Amount),
	}
	actualDeposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "lsm token deposit should have been found after LSM liquid stake")
	s.Require().Equal(expectedDeposit, actualDeposit)

	// Confirm slash query progress was incremented
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	expectedQueryProgress := tc.initialQueryProgress.Add(tc.validMsg.Amount)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(expectedQueryProgress.Int64(), hostZone.Validators[0].SlashQueryProgressTracker.Int64(), "slash query progress")
}

func (s *KeeperTestSuite) TestLSMLiquidStake_Successful_WithSharesToTokensRateQuery() {
	tc := s.SetupTestLSMLiquidStake()

	// Increase the liquid stake size so that it breaks the query checkpoint
	// queryProgressSlack is the remaining amount that can be staked in one message before a slash query is issued
	queryProgressSlack := tc.queryCheckpoint.Sub(tc.initialQueryProgress)
	tc.validMsg.Amount = queryProgressSlack.Add(sdk.NewInt(1000))

	// Call LSM Liquid stake
	msgResponse, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().NoError(err, "no error expected when calling lsm liquid stake")
	s.Require().False(msgResponse.TransactionComplete, "transaction should still be pending")

	// Confirm stToken was NOT sent to the user
	userStTokenBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom)
	s.Require().True(userStTokenBalance.Amount.IsZero(), "user stToken balance")

	// Confirm query was submitted
	allQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(allQueries, 1)

	// Confirm query metadata
	actualQuery := allQueries[0]
	s.Require().Equal(HostChainId, actualQuery.ChainId, "query chain-id")
	s.Require().Equal(ibctesting.FirstConnectionID, actualQuery.ConnectionId, "query connection-id")
	s.Require().Equal(icqtypes.STAKING_STORE_QUERY_WITH_PROOF, actualQuery.QueryType, "query types")

	s.Require().Equal(types.ModuleName, actualQuery.CallbackModule, "callback module")
	s.Require().Equal(keeper.ICQCallbackID_Validator, actualQuery.CallbackId, "callback-id")

	expectedTimeout := uint64(s.Ctx.BlockTime().UnixNano() + (keeper.LSMSlashQueryTimeout).Nanoseconds())
	s.Require().Equal(keeper.LSMSlashQueryTimeout, actualQuery.TimeoutDuration, "timeout duration")
	s.Require().Equal(int64(expectedTimeout), int64(actualQuery.TimeoutTimestamp), "timeout timestamp")

	// Confirm query callback data
	s.Require().True(len(actualQuery.CallbackData) > 0, "callback data exists")

	expectedStToken := sdk.NewCoin(StAtom, tc.validMsg.Amount)
	expectedDepositId := keeper.GetLSMTokenDepositId(s.Ctx.BlockHeight(), HostChainId, tc.validMsg.Creator, LSMTokenBaseDenom)
	expectedLSMTokenDeposit := recordstypes.LSMTokenDeposit{
		DepositId:        expectedDepositId,
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		IbcDenom:         tc.lsmTokenIBCDenom,
		StakerAddress:    tc.validMsg.Creator,
		ValidatorAddress: ValAddress,
		Amount:           tc.validMsg.Amount,
		StToken:          expectedStToken,
		Status:           recordstypes.LSMTokenDeposit_DEPOSIT_PENDING,
	}

	var actualCallbackData types.ValidatorSharesToTokensQueryCallback
	err = proto.Unmarshal(actualQuery.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling query callback data")

	lsmLiquidStake := actualCallbackData.LsmLiquidStake
	s.Require().Equal(HostChainId, lsmLiquidStake.HostZone.ChainId, "callback data - host zone")
	s.Require().Equal(ValAddress, lsmLiquidStake.Validator.Address, "callback data - validator")

	s.Require().Equal(expectedLSMTokenDeposit, *lsmLiquidStake.Deposit, "callback data - deposit")
}

func (s *KeeperTestSuite) TestLSMLiquidStake_DifferentRedemptionRates() {
	tc := s.SetupTestLSMLiquidStake()
	tc.validMsg.Amount = sdk.NewInt(100) // reduce the stake amount to prevent insufficient balance error

	// Loop over sharesToTokens rates: {0.92, 0.94, ..., 1.2}
	interval := sdk.MustNewDecFromStr("0.01")
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := interval.Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.02
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate

		// Update rate in host zone
		hz := tc.hostZone
		hz.RedemptionRate = newRedemptionRate
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom).Amount
		_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
		s.Require().NoError(err)
		endingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom).Amount
		actualStAtomMinted := endingStAtomBalance.Sub(startingStAtomBalance)

		expectedStAtomMinted := sdk.NewDecFromInt(tc.validMsg.Amount).Quo(redemptionRateFloat).TruncateInt()
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		s.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)

		// Cleanup the LSMTokenDeposit record to prevent an error on the next run
		s.App.RecordsKeeper.RemoveLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	}
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_NotIBCDenom() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the denom is not an IBC token
	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = "fake_ibc_denom"

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "lsm token is not an IBC token (fake_ibc_denom)")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_HostZoneNotFound() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the denom is an IBC denom from a channel that is not supported
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, "channel-1")
	prefixedDenom := sourcePrefix + LSMTokenBaseDenom
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)

	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = lsmTokenDenomTrace.IBCDenom()

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-1) does not match any registered host zone")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_ValidatorNotFound() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the base denom is from a non-existent validator
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, ibctesting.FirstChannelID)
	prefixedDenom := sourcePrefix + "cosmosvaloperXXX/42"
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)

	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = lsmTokenDenomTrace.IBCDenom()

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "validator (cosmosvaloperXXX) is not registered in the Stride validator set")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_DepositAlreadyExists() {
	tc := s.SetupTestLSMLiquidStake()

	// Set a deposit with the same chainID and denom in the store
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordstypes.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   LSMTokenBaseDenom,
	})

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "there is already a previous record with this denom being processed")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_InvalidDepositAddress() {
	tc := s.SetupTestLSMLiquidStake()

	// Remove the host zones address from the store
	invalidHostZone := tc.hostZone
	invalidHostZone.DepositAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "host zone address is invalid")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_InsufficientBalance() {
	tc := s.SetupTestLSMLiquidStake()

	// Send out all the user's coins so that they have an insufficient balance of LSM tokens
	initialBalanceCoin := sdk.NewCoins(sdk.NewCoin(tc.lsmTokenIBCDenom, tc.initialBalance))
	err := s.App.BankKeeper.SendCoins(s.Ctx, tc.liquidStakerAddress, s.TestAccs[1], initialBalanceCoin)
	s.Require().NoError(err)

	_, err = s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "insufficient funds")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_ZeroStTokens() {
	tc := s.SetupTestLSMLiquidStake()

	// Adjust redemption rate and liquid stake amount so that the number of stTokens would be zero
	// stTokens = 1(amount) / 1.1(RR) = rounds down to 0
	hostZone := tc.hostZone
	hostZone.RedemptionRate = sdk.NewDecWithPrec(11, 1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	tc.validMsg.Amount = sdkmath.NewInt(1)

	// The liquid stake should fail
	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().EqualError(err, "Liquid stake of 1uatom would return 0 stTokens: Liquid staked amount is too small")
}
