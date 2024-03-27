package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	epochtypes "github.com/Stride-Labs/stride/v21/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v21/x/interchainquery/types"
	recordtypes "github.com/Stride-Labs/stride/v21/x/records/types"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

// -----------------------------
// Query Community Pool Balances
// -----------------------------

type QueryCommunityPoolBalanceTestCase struct {
	hostZone        types.HostZone
	timeoutDuration time.Duration
	expectedTimeout uint64
}

func (s *KeeperTestSuite) SetupQueryCommunityPoolBalance(icaAccountType types.ICAAccountType) QueryCommunityPoolBalanceTestCase {
	// We need to register the transfer channel to initialize the light client state
	s.CreateTransferChannel(HostChainId)

	// Create host zone
	// We must use valid addresses for each ICA since they're serialized for the query request
	depositAddress := s.TestAccs[0]
	returnAddress := s.TestAccs[1]
	hostZone := types.HostZone{
		ChainId:                        HostChainId,
		ConnectionId:                   ibctesting.FirstConnectionID,
		CommunityPoolDepositIcaAddress: depositAddress.String(),
		CommunityPoolReturnIcaAddress:  returnAddress.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create epoch tracker for timeout
	timeoutDuration := time.Second * 30
	epochEndTime := uint64(s.Ctx.BlockTime().Add(timeoutDuration).UnixNano())
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: epochEndTime,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	return QueryCommunityPoolBalanceTestCase{
		hostZone:        hostZone,
		timeoutDuration: timeoutDuration,
		expectedTimeout: epochEndTime,
	}
}

// Helper function to verify the query that was submitted from the community pool balance query
func (s *KeeperTestSuite) checkCommunityPoolQuerySubmission(
	tc QueryCommunityPoolBalanceTestCase,
	icaAccountType types.ICAAccountType,
) {
	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tc.hostZone.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tc.hostZone.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm query timeout info
	s.Require().Equal(tc.timeoutDuration, query.TimeoutDuration, "query callback id")
	s.Require().Equal(tc.expectedTimeout, query.TimeoutTimestamp, "query callback id")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_CommunityPoolIcaBalance, query.CallbackId, "query callback id")

	var actualCallbackData types.CommunityPoolBalanceQueryCallback
	err := proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	expectedCallbackData := types.CommunityPoolBalanceQueryCallback{
		IcaType: icaAccountType,
		Denom:   Atom,
	}
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callabck data")

	// Confirm query request info
	expectedIcaAddress := tc.hostZone.CommunityPoolDepositIcaAddress
	if icaAccountType == types.ICAAccountType_COMMUNITY_POOL_RETURN {
		expectedIcaAddress = tc.hostZone.CommunityPoolReturnIcaAddress
	}
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(expectedIcaAddress, actualAddress.String(), "query account address")
	s.Require().Equal(Atom, actualDenom, "query denom")
}

// Tests a community pool balance query to the deposit ICA account
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Successful_Deposit() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().NoError(err, "no error expected when querying pool balance")

	s.checkCommunityPoolQuerySubmission(tc, icaAccountType)
}

// Tests a community pool balance query to the return ICA account
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Successful_Return() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_RETURN
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().NoError(err, "no error expected when querying pool balance")

	s.checkCommunityPoolQuerySubmission(tc, icaAccountType)
}

// Tests a community pool balance query that fails due to an invalid account type
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_InvalidAccountType() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	invalidAccountType := types.ICAAccountType_DELEGATION
	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, tc.hostZone, invalidAccountType, Atom)
	s.Require().ErrorContains(err, "icaType must be either deposit or return!")
}

// Tests a community pool balance query that fails due to an invalid account address
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_InvalidAccountAddress() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Change the host zone account address to be invalid
	invalidHostZone := tc.hostZone
	invalidHostZone.CommunityPoolDepositIcaAddress = "invalid_address"

	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, invalidHostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "invalid COMMUNITY_POOL_DEPOSIT address, could not decode (invalid_address)")
}

// Tests a community pool balance query that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_MissingEpoch() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a community pool balance query that fails to submit the query
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_FailedQuerySubmission() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Set an invalid connection ID for the host zone so that the query submission fails
	invalidHostZone := tc.hostZone
	invalidHostZone.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.QueryCommunityPoolIcaBalance(s.Ctx, invalidHostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "Error submitting query for pool ica balance")
}

// ----------------------------------
// Liquid Stake Community Pool Tokens
// ----------------------------------

type LiquidStakeCommunityPoolTokensTestCase struct {
	hostZone            types.HostZone
	initialNativeTokens sdkmath.Int
	initialDummyTokens  sdkmath.Int
}

func (s *KeeperTestSuite) SetupLiquidStakeCommunityPoolTokens() LiquidStakeCommunityPoolTokensTestCase {
	s.CreateTransferChannel(HostChainId)

	// Create relevant module and ICA accounts
	depositAddress := s.TestAccs[0]
	communityPoolHoldingAddress := s.TestAccs[1]
	communityPoolReturnICAAddress := s.TestAccs[2]

	// Create a host zone with valid addresses to perform the liquid stake
	hostZone := types.HostZone{
		ChainId:                          HostChainId,
		HostDenom:                        Atom,
		IbcDenom:                         IbcAtom,
		TransferChannelId:                ibctesting.FirstChannelID,
		CommunityPoolStakeHoldingAddress: communityPoolHoldingAddress.String(),
		CommunityPoolReturnIcaAddress:    communityPoolReturnICAAddress.String(),
		DepositAddress:                   depositAddress.String(),
		RedemptionRate:                   sdk.OneDec(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create the epoch tracker and deposit records so the liquid stake succeeds
	epochNumber := uint64(1)
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        epochNumber,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000), // dictates transfer timeout
	}
	depositRecord := recordtypes.DepositRecord{
		Id:                 epochNumber,
		DepositEpochNumber: epochNumber,
		HostZoneId:         HostChainId,
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)

	// Fund the holding address with native tokens (in IBC form) and
	// some dummy tokens that should not get touched by these functions
	initialNativeTokens := sdk.NewInt(1000)
	initialDummyTokens := sdk.NewInt(999)
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(IbcAtom, initialNativeTokens))
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(Atom, initialDummyTokens))   // dummy token
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(StAtom, initialDummyTokens)) // dummy token

	return LiquidStakeCommunityPoolTokensTestCase{
		hostZone:            hostZone,
		initialNativeTokens: initialNativeTokens,
		initialDummyTokens:  initialDummyTokens,
	}
}

func (s *KeeperTestSuite) TestLiquidStakeCommunityPoolTokens_Success() {
	tc := s.SetupLiquidStakeCommunityPoolTokens()

	transferPortId := transfertypes.PortID
	transferChannelId := ibctesting.FirstChannelID
	communityPoolHoldingAddress := sdk.MustAccAddressFromBech32(tc.hostZone.CommunityPoolStakeHoldingAddress)

	// Call liquid stake which should convert the whole native tokens amount to stTokens and transfer it
	err := s.App.StakeibcKeeper.LiquidStakeCommunityPoolTokens(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected during liquid stake")

	// Confirm there are no longer native tokens in the holding address
	ibcAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, communityPoolHoldingAddress, IbcAtom)
	s.Require().Zero(ibcAtomBalance.Amount.Int64(), "balance of holding address should be zero")

	// Confirm the dummy tokens are still present
	dummyTokenBalance1 := s.App.BankKeeper.GetBalance(s.Ctx, communityPoolHoldingAddress, Atom)
	dummyTokenBalance2 := s.App.BankKeeper.GetBalance(s.Ctx, communityPoolHoldingAddress, StAtom)
	s.Require().Equal(tc.initialDummyTokens, dummyTokenBalance2.Amount, "dummy token 2 was not touched")
	s.Require().Equal(tc.initialDummyTokens, dummyTokenBalance1.Amount, "dummy token 1 was not touched")

	// Confirm the stTokens have been escrowed as a result of the transfer
	escrowAddress := transfertypes.GetEscrowAddress(transferPortId, transferChannelId)
	stTokenEscrowBalance := s.App.BankKeeper.GetBalance(s.Ctx, escrowAddress, StAtom)
	s.Require().Equal(tc.initialNativeTokens.Int64(), stTokenEscrowBalance.Amount.Int64(), "st token escrow balance")

	// Check that if we run the liquid stake function again, nothing should get transferred
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, transferPortId, transferChannelId)
	s.Require().True(found, "sequence number not found before liquid stake")

	err = s.App.StakeibcKeeper.LiquidStakeCommunityPoolTokens(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected during second liquid stake")

	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, transferPortId, transferChannelId)
	s.Require().True(found, "sequence number not found after after liquid stake")

	s.Require().Equal(startSequence, endSequence, "no transfer should have been initiated")
}

// Test liquid stake with an invalid stake holding address
func (s *KeeperTestSuite) TestLiquidStakeCommunityPoolTokens_Failure_InvalidAddress() {
	tc := s.SetupLiquidStakeCommunityPoolTokens()

	invalidHostZone := tc.hostZone
	invalidHostZone.CommunityPoolStakeHoldingAddress = "invalid"

	err := s.App.StakeibcKeeper.LiquidStakeCommunityPoolTokens(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "decoding bech32 failed")
}

// Test liquid stake with an invalid host denom, which should cause the liquid stake to fail
func (s *KeeperTestSuite) TestLiquidStakeCommunityPoolTokens_LiquidStakeFailure() {
	tc := s.SetupLiquidStakeCommunityPoolTokens()

	invalidHostZone := tc.hostZone
	invalidHostZone.HostDenom = "invalid"

	err := s.App.StakeibcKeeper.LiquidStakeCommunityPoolTokens(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "Failed to liquid stake")
}

// Set an invalid transfer channel on the host so that the transfer fails
func (s *KeeperTestSuite) TestLiquidStakeCommunityPoolTokens_TransferFailure() {
	tc := s.SetupLiquidStakeCommunityPoolTokens()

	invalidHostZone := tc.hostZone
	invalidHostZone.TransferChannelId = "channel-X"

	err := s.App.StakeibcKeeper.LiquidStakeCommunityPoolTokens(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "Error submitting ibc transfer")
}

// ----------------------------------
// Redeem Community Pool Tokens
// ----------------------------------

type RedeemCommunityPoolTokensTestCase struct {
	hostZone                    types.HostZone
	initialDummyTokens          sdkmath.Int
	communityPoolHoldingAddress sdk.AccAddress
}

func (s *KeeperTestSuite) SetupRedeemCommunityPoolTokens() RedeemCommunityPoolTokensTestCase {
	s.CreateTransferChannel(HostChainId)

	// Create relevant module and ICA accounts
	depositAddress := s.TestAccs[0]
	communityPoolHoldingAddress := s.TestAccs[1]
	communityPoolReturnICAAddress := HostICAAddress // need an address on HostChain (starts cosmos)

	// stTokens which will be redeemed, dummy tokens which should not be touched
	initialStTokens := sdk.NewInt(1000)
	initialDummyTokens := sdk.NewInt(999)

	// Fund the redeem holding address with stTokens and
	// some dummy tokens that should not get touched while redeeming
	stDenom := types.StAssetDenomFromHostZoneDenom(Atom)
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(stDenom, initialStTokens))
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(Atom, initialDummyTokens))    // dummy token
	s.FundAccount(communityPoolHoldingAddress, sdk.NewCoin(IbcAtom, initialDummyTokens)) // dummy token

	// Create a host zone with valid addresses to perform the liquid stake
	hostZone := types.HostZone{
		ChainId:                           HostChainId,  //GAIA
		Bech32Prefix:                      Bech32Prefix, //cosmos
		HostDenom:                         Atom,
		IbcDenom:                          IbcAtom,
		TransferChannelId:                 ibctesting.FirstChannelID,
		CommunityPoolRedeemHoldingAddress: communityPoolHoldingAddress.String(),
		CommunityPoolReturnIcaAddress:     communityPoolReturnICAAddress,
		DepositAddress:                    depositAddress.String(),
		TotalDelegations:                  initialStTokens, // at least as much as we are trying to redeem
		RedemptionRate:                    sdk.OneDec(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create the epoch tracker and deposit records so the liquid stake succeeds
	epochNumber := uint64(1)
	epochTracker := types.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     epochNumber,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	// Setup the epoch unbonding record with a HostZoneUnbonding for the hostZone
	var unbondings []*recordtypes.HostZoneUnbonding
	unbonding := &recordtypes.HostZoneUnbonding{
		HostZoneId: HostChainId,
	}
	unbondings = append(unbondings, unbonding)
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: unbondings,
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	return RedeemCommunityPoolTokensTestCase{
		hostZone:                    hostZone,
		initialDummyTokens:          initialDummyTokens,
		communityPoolHoldingAddress: communityPoolHoldingAddress,
	}
}

func (s *KeeperTestSuite) TestRedeemCommunityPoolTokens_Success() {
	tc := s.SetupRedeemCommunityPoolTokens()

	// Verify that no user redemption records exist yet
	userRedemptionRecords := s.App.RecordsKeeper.GetAllUserRedemptionRecord(s.Ctx)
	s.Require().Zero(len(userRedemptionRecords), "No user redemption records expected yet")

	// Call redeem stake which should start the unbonding for the stToken amount
	err := s.App.StakeibcKeeper.RedeemCommunityPoolTokens(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected during redeem stake")

	// Check that a new user redemption record was created from the call to redeem
	userRedemptionRecords = s.App.RecordsKeeper.GetAllUserRedemptionRecord(s.Ctx)
	s.Require().Equal(1, len(userRedemptionRecords), "New user redemption records should be created")

	// Confirm there are no longer staked tokens in the holding address after redeem
	stDenom := types.StAssetDenomFromHostZoneDenom(tc.hostZone.HostDenom)
	stAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.communityPoolHoldingAddress, stDenom)
	s.Require().Zero(stAtomBalance.Amount.Int64(), "balance of redeem holidng address should be zero")

	// Confirm the dummy tokens were untouched by the redeem call
	atomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.communityPoolHoldingAddress, Atom)
	ibcAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.communityPoolHoldingAddress, IbcAtom)
	s.Require().Equal(tc.initialDummyTokens.Int64(), atomBalance.Amount.Int64(), "Atom tokens should not be touched")
	s.Require().Equal(tc.initialDummyTokens.Int64(), ibcAtomBalance.Amount.Int64(), "IbcAtom tokens should not be touched")

	// Call redeem stake again but now there is no more stTokens to be redeemed.
	err = s.App.StakeibcKeeper.RedeemCommunityPoolTokens(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected during redeem stake")

	// Check that no new user redemption records were created from the second call to redeem
	userRedemptionRecords = s.App.RecordsKeeper.GetAllUserRedemptionRecord(s.Ctx)
	s.Require().Equal(1, len(userRedemptionRecords), "New user redemption records should be created")
}

// Test redeem stake with an invalid redeem holding address
func (s *KeeperTestSuite) TestRedeemCommunityPoolTokens_Failure_InvalidAddress() {
	tc := s.SetupRedeemCommunityPoolTokens()

	invalidHostZone := tc.hostZone
	invalidHostZone.CommunityPoolRedeemHoldingAddress = "invalid"

	err := s.App.StakeibcKeeper.RedeemCommunityPoolTokens(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "decoding bech32 failed")
}

// Test redeem stake with an invalid redeem holding address
func (s *KeeperTestSuite) TestRedeemCommunityPoolTokens_Failure_NotEnoughDelegations() {
	tc := s.SetupRedeemCommunityPoolTokens()

	invalidHostZone := tc.hostZone
	invalidHostZone.TotalDelegations = sdk.ZeroInt()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	err := s.App.StakeibcKeeper.RedeemCommunityPoolTokens(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "invalid amount")
}

// ----------------------------------------------------------
//               BuildFundCommunityPoolMsg
// ----------------------------------------------------------

func (s *KeeperTestSuite) TestBuildFundCommunityPoolMsg() {
	withdrawalICA := "withdrawal_ica"
	communityPoolReturnICA := "community_pool_return_ica"
	communityPoolTreasuryAddress := "community_pool_treasury"

	testCases := []struct {
		name             string
		senderAccoutType types.ICAAccountType
		sendToTreasury   bool
		expectedSender   string
		expectedReceiver string
		expectedError    string
	}{
		{
			name:             "community pool return ICA to main community pool",
			senderAccoutType: types.ICAAccountType_COMMUNITY_POOL_RETURN,
			sendToTreasury:   false,
			expectedSender:   communityPoolReturnICA,
		},
		{
			name:             "community pool return ICA to treasury",
			senderAccoutType: types.ICAAccountType_COMMUNITY_POOL_RETURN,
			sendToTreasury:   true,
			expectedSender:   communityPoolReturnICA,
			expectedReceiver: communityPoolTreasuryAddress,
		},
		{
			name:             "withdrawal ICA to main community pool",
			senderAccoutType: types.ICAAccountType_WITHDRAWAL,
			sendToTreasury:   false,
			expectedSender:   withdrawalICA,
		},
		{
			name:             "withdrawal ICA to treasury",
			senderAccoutType: types.ICAAccountType_WITHDRAWAL,
			sendToTreasury:   true,
			expectedSender:   withdrawalICA,
			expectedReceiver: communityPoolTreasuryAddress,
		},
		{
			name:             "invalid sender",
			senderAccoutType: types.ICAAccountType_DELEGATION,
			expectedError:    "fund community pool ICA can only be initiated from",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Define the sending tokens and input host zone struct
			// If the test case sends to the treasury, we have to set the community pool treasury
			// address to be non-empty
			tokens := sdk.NewCoins(sdk.NewCoin(HostDenom, sdk.NewInt(1000)))
			hostZone := types.HostZone{
				CommunityPoolReturnIcaAddress: communityPoolReturnICA,
				WithdrawalIcaAddress:          withdrawalICA,
			}
			if tc.sendToTreasury {
				hostZone.CommunityPoolTreasuryAddress = communityPoolTreasuryAddress
			}

			// Build the fund msg
			actualMsg, actualErr := s.App.StakeibcKeeper.BuildFundCommunityPoolMsg(s.Ctx, hostZone, tokens, tc.senderAccoutType)

			// If there's not error expected, validate the underlying message
			if tc.expectedError == "" {
				s.Require().Len(actualMsg, 1, "there should be one message")

				// If the recipient was the treasury, confirm it was a valid bank send
				if tc.sendToTreasury {
					bankSendMsg, ok := actualMsg[0].(*banktypes.MsgSend)
					s.Require().True(ok, "ICA message should have been a bank send")
					s.Require().Equal(tokens, bankSendMsg.Amount, "bank send amount")
					s.Require().Equal(tc.expectedSender, bankSendMsg.FromAddress, "bank send from address")
					s.Require().Equal(tc.expectedReceiver, bankSendMsg.ToAddress, "bank send to address")
				} else {
					fundCommunityPoolMsg, ok := actualMsg[0].(*disttypes.MsgFundCommunityPool)
					s.Require().True(ok, "ICA message should have been a fund community pool message")
					s.Require().Equal(tokens, fundCommunityPoolMsg.Amount, "fund community pool amount")
					s.Require().Equal(tc.expectedSender, fundCommunityPoolMsg.Depositor, "bank send from address")
				}
			} else {
				// If there was an expected error, confirm the error message
				s.Require().ErrorContains(actualErr, tc.expectedError)
			}
		})
	}
}
