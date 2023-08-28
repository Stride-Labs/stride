package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type ValidatorICQCallbackState struct {
	hostZone         types.HostZone
	validator        types.Validator
	delegator        string
	lsmTokenIBCDenom string
	stakerBalance    sdkmath.Int
}

type ValidatorICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type ValidatorICQCallbackTestCase struct {
	initialState                ValidatorICQCallbackState
	validArgs                   ValidatorICQCallbackArgs
	sharesToTokensRateIfSlashed sdk.Dec
}

func (s *KeeperTestSuite) CreateValidatorQueryResponse(address string, tokens int64, shares int64) []byte {
	validator := stakingtypes.Validator{
		OperatorAddress: address,
		Tokens:          sdkmath.NewInt(tokens),
		DelegatorShares: sdk.NewDec(shares),
	}
	validatorBz := s.App.RecordsKeeper.Cdc.MustMarshal(&validator)
	return validatorBz
}

func (s *KeeperTestSuite) SetupValidatorICQCallback(validatorSlashed, liquidStakeCallback bool) ValidatorICQCallbackTestCase {
	// The transfer channel is required in the event that we're testing an LSMCallback and have to transfer the LSM Token
	s.CreateTransferChannel(HostChainId)

	// These must be valid delegation account address, otherwise the bech decoding will fail
	delegatorAddress := "cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm"
	depositAddress := types.NewHostZoneDepositAddress(HostChainId).String()

	// In this example, the validator has 2000 shares, originally had 2000 tokens,
	// and now has 1000 tokens (after being slashed)
	numShares := int64(2000)
	sharesToTokensRate := sdk.NewDec(1)
	sharesToTokensRateIfSlashed := sdk.MustNewDecFromStr("0.5")

	// The validator we'll query the sharesToTokens rate for
	queriedValidator := types.Validator{
		Name:               "val1",
		Address:            ValAddress,
		SharesToTokensRate: sharesToTokensRate,
	}

	// Mocked state is required for (optional) delegator shares ICQ submission
	// and (optional) LSM Liquid Stake completion
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		TransferChannelId:    ibctesting.FirstChannelID,
		DelegationIcaAddress: delegatorAddress,
		DepositAddress:       depositAddress,
		RedemptionRate:       sdk.NewDec(1),
		Validators: []*types.Validator{
			&queriedValidator,
			{Name: "val2"}, // This validator isn't being queried
		},
		LsmLiquidStakeEnabled: true,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock out the query response (which will be a validator object)
	// If we're testing that a slash occurred, cut the number tokens in half (to represent the slash)
	// Otherwise, use the same number of tokens as shares
	numTokens := numShares
	if validatorSlashed {
		numTokens /= 2
	}
	queryResponse := s.CreateValidatorQueryResponse(ValAddress, numTokens, numShares)

	// If we're testing a liquid stake callback, mock out the callback data
	var err error
	lsmTokenIBCDenom := ""
	callbackDataBz := []byte{}
	stakeAmount := sdkmath.NewInt(1_000_000)
	if liquidStakeCallback {
		// Need valid IBC denom here to test parsing
		lsmTokenIBCDenom = s.getLSMTokenIBCDenom()

		// Fund the user's account with the LSM token
		liquidStaker := s.TestAccs[0]
		s.FundAccount(liquidStaker, sdk.NewCoin(lsmTokenIBCDenom, stakeAmount))

		// The callback data consists of the LSMTokenDeposit wrapped in additional state context
		lsmTokenDeposit := recordstypes.LSMTokenDeposit{
			ChainId:       HostChainId,
			Denom:         LSMTokenBaseDenom,
			StakerAddress: liquidStaker.String(),
			Amount:        stakeAmount,
			IbcDenom:      lsmTokenIBCDenom,
			StToken:       sdk.NewCoin(StAtom, stakeAmount),
		}
		lsmLiquidStake := types.LSMLiquidStake{
			HostZone:  &hostZone,
			Validator: &queriedValidator,
			Deposit:   &lsmTokenDeposit,
		}
		callbackDataBz, err = proto.Marshal(&types.ValidatorSharesToTokensQueryCallback{
			LsmLiquidStake: &lsmLiquidStake,
		})
		s.Require().NoError(err, "no error expected when marshalling callback data")
	}

	return ValidatorICQCallbackTestCase{
		initialState: ValidatorICQCallbackState{
			hostZone:         hostZone,
			validator:        queriedValidator,
			delegator:        delegatorAddress,
			lsmTokenIBCDenom: lsmTokenIBCDenom,
			stakerBalance:    stakeAmount,
		},
		sharesToTokensRateIfSlashed: sharesToTokensRateIfSlashed,
		validArgs: ValidatorICQCallbackArgs{
			query: icqtypes.Query{
				ChainId:          HostChainId,
				CallbackData:     callbackDataBz,
				TimeoutTimestamp: uint64(s.Ctx.BlockTime().Add(time.Minute).UnixNano()),
			},
			callbackArgs: queryResponse,
		},
	}
}

// Helper function to check the validator's shares to tokens rate after the query
func (s *KeeperTestSuite) checkValidatorSharesToTokensRate(expectedSharesToTokensRate sdk.Dec, tc ValidatorICQCallbackTestCase) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(expectedSharesToTokensRate.String(), hostZone.Validators[0].SharesToTokensRate.String(),
		"validator shares to tokens rate")
}

// Check that the LSMLiquidStake callback succeeded by looking for a successful event emission
func (s *KeeperTestSuite) checkLSMLiquidStakeSuccess() {
	s.CheckEventValueEmitted(
		types.EventTypeLSMLiquidStakeRequest,
		types.AttributeKeyTransactionStatus,
		types.AttributeValueTransactionSucceeded,
	)
}

// Check that the LSMLiquidStake callback failed by looking for a failed event emission
func (s *KeeperTestSuite) checkLSMLiquidStakeFailed() {
	// Confirm failure was emitted
	s.CheckEventValueEmitted(
		types.EventTypeLSMLiquidStakeRequest,
		types.AttributeKeyTransactionStatus,
		types.AttributeValueTransactionFailed,
	)
	// Confirm success was NOT emitted (to confirm short circuiting)
	s.CheckEventValueNotEmitted(
		types.EventTypeLSMLiquidStakeRequest,
		types.AttributeKeyTransactionStatus,
		types.AttributeValueTransactionSucceeded,
	)
}

// Check that the liquid stake code was not called
func (s *KeeperTestSuite) checkLSMLiquidStakeNotCalled() {
	s.CheckEventTypeNotEmitted(types.EventTypeLSMLiquidStakeRequest)
}

// Helper function to check that the delegator shares query was submitted by checking
// that the query object was stored
func (s *KeeperTestSuite) checkDelegatorSharesQuerySubmitted(liquidStakeCallback bool, tc ValidatorICQCallbackTestCase) {
	// Check that this is only one query in the store
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should be one new query submitted for delegator shares")

	// Confirm the query metadata matches expectations
	query := queries[0]
	s.Require().Equal(HostChainId, query.ChainId, "query chain-id")
	s.Require().Equal(ibctesting.FirstConnectionID, query.ConnectionId, "query connnection-id")
	s.Require().Equal(icqtypes.STAKING_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_Delegation, query.CallbackId, "query callback-id")
	s.Require().Equal(false, query.RequestSent, "query sent")

	// Confirm validator and delegator are in query data
	_, validatorAddressBz, err := bech32.DecodeAndConvert(ValAddress)
	s.Require().NoError(err, "no error expected when decoding validator address")

	_, delegatorAddressBz, err := bech32.DecodeAndConvert(tc.initialState.delegator)
	s.Require().NoError(err, "no error expected when decoding delegation address")

	expectedQueryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)
	s.Require().Equal(expectedQueryData, query.RequestData, "query request-data")

	// Confirm timeout based on the type of query (LSM or manual)
	timeoutDuration := time.Hour
	expectedTimeout := s.Ctx.BlockTime().UnixNano() + (timeoutDuration).Nanoseconds()
	s.Require().Equal(timeoutDuration, query.TimeoutDuration, "query timeout duration")
	s.Require().Equal(expectedTimeout, int64(query.TimeoutTimestamp), "query timeout timestamp")

	// Confirm query callback data
	var callbackData types.DelegatorSharesQueryCallback
	err = proto.Unmarshal(query.CallbackData, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	expectedInitialDelegation := tc.initialState.validator.Delegation
	s.Require().Equal(expectedInitialDelegation.Int64(), callbackData.InitialValidatorDelegation.Int64(),
		"query callback-data initial delegation")

	// Confirm the validator's flagged as having a query in progress
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().True(hostZone.Validators[0].SlashQueryInProgress, "slash query in progress")
}

// Helper function to check that the delegator shares query was not submitted
// by confirming there are no queries in the store
func (s *KeeperTestSuite) checkDelegatorSharesQueryNotSubmitted() {
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "the delegator shares query should not have been submitted")
}

// Test case where the callback was successful, there was no slash, and this was not a liquid stake callback
// Here the sharesToTokens rate should not update and there should be no delegator shares query submitted
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_NoSlash_NoLiquidStake() {
	validatorSlashed := false
	lsmCallback := false
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID NOT update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm the delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()

	// Confirm the liquid stake flow as not touched
	s.checkLSMLiquidStakeNotCalled()
}

// Test case where the callback was successful and there was a slash, but the query was issued manually
// Here the sharesToTokens rate should update and the delegator shares query should be submitted
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_Slash_NoLiquidStake() {
	validatorSlashed := true
	lsmCallback := false
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID update
	s.checkValidatorSharesToTokensRate(tc.sharesToTokensRateIfSlashed, tc)

	// Confirm delegator shares query WAS submitted
	s.checkDelegatorSharesQuerySubmitted(lsmCallback, tc)

	// Confirm the liquid stake flow as not touched
	s.checkLSMLiquidStakeNotCalled()
}

// Test case where the callback was successful, there was no slash, and this query was from a liquid stake
// Here the sharesToTokens rate should not update, the delegator shares query should not be submitted, and
// the liquid stake should have succeeded
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_NoSlash_LiquidStake() {
	validatorSlashed := false
	lsmCallback := true
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID NOT update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm the delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()

	// Confirm the liquid stake was a success
	s.checkLSMLiquidStakeSuccess()
}

// Test case where the callback was successful and this query was from a liquid stake,
// but the finishing of the liquid stake failed
// Any state changes from the finish liquid stake should be discarded, including
// the transfer of LSM tokens to the deposit account
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_NoSlash_LiquidStakeFailed() {
	validatorSlashed := false
	lsmCallback := true
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	// Remove the host zone's delegation account - this should cause the finishing of the LSM liquid stake to fail
	var callbackData types.ValidatorSharesToTokensQueryCallback
	err := proto.Unmarshal(tc.validArgs.query.CallbackData, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshaling query args")

	callbackData.LsmLiquidStake.HostZone.DelegationIcaAddress = ""
	invalidCallbackData, err := proto.Marshal(&callbackData)
	s.Require().NoError(err, "no error expected when marshaling query args")

	invalidQuery := tc.validArgs.query
	invalidQuery.CallbackData = invalidCallbackData

	// When the callback runs, the finishing of the LSM liquid stake should make partial state changes, including
	// the sending of LSM tokens to the module account
	// However, that change should be discarded since the liquid stake failed
	err = keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID NOT update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()

	// Confirm the liquid stake failed
	// We'll check both that the failed event was emitted, and the success event was not emitted
	// (to confirm short circuiting)
	s.checkLSMLiquidStakeFailed()

	// Confirm the tokens were not sent to the module account since the state changes were discarded
	stakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], tc.initialState.lsmTokenIBCDenom)
	s.Require().Equal(tc.initialState.stakerBalance.Int64(), stakerBalance.Amount.Int64(),
		"staker balance after failed liquid stake")
}

// Test case where the callback was successful, there was a slash, and this query was from a liquid stake
// Here the sharesToTokens rate should update, the delegator shares query should be submitted,
// and the liquid stake should be rejected
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_Slash_LiquidStake() {
	validatorSlashed := true
	lsmCallback := true
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID update
	s.checkValidatorSharesToTokensRate(tc.sharesToTokensRateIfSlashed, tc)

	// Confirm delegator shares query WAS submitted
	s.checkDelegatorSharesQuerySubmitted(lsmCallback, tc)

	// Confirm the liquid stake failed
	s.checkLSMLiquidStakeFailed()
}

// Test case where the callback was successful, but there was not previous sharesToTokens rate to determine if a slash occured
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_Successful_NoPreviousSharesToTokensRate() {
	validatorSlashed := false
	tc := s.SetupValidatorICQCallback(validatorSlashed, false)

	// The sharesToTokens rate should update to the initial sharesToTokens rate from the test setup
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate

	// Set the sharesToTokens rate to zero
	hostZone := tc.initialState.hostZone
	hostZone.Validators[0].SharesToTokensRate = sdk.ZeroDec()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID update
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()
}

// Test case where the there was no slash, but the liquid stake callback failed
func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_NoSlash_LiqudStakeFailed() {
	validatorSlashed := false
	lsmCallback := true
	tc := s.SetupValidatorICQCallback(validatorSlashed, lsmCallback)

	// Remove the LSM tokens from the user account so that they have insufficient funds to finish the liquid stake
	liquidStaker := s.TestAccs[0]
	recipient := s.TestAccs[1]
	balance := s.App.BankKeeper.GetBalance(s.Ctx, liquidStaker, tc.initialState.lsmTokenIBCDenom)
	err := s.App.BankKeeper.SendCoins(s.Ctx, liquidStaker, recipient, sdk.NewCoins(balance))
	s.Require().NoError(err, "no error expected when sending liquid staker's LSM tokens")

	// Now when we call the callback, the callback itself should succeed, but the finishing of the liquid stake should fail
	err = keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator sharesToTokens rate callback error")

	// Confirm validator's sharesToTokens rate DID update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()

	// Confirm the liquid stake failed
	s.checkLSMLiquidStakeFailed()
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_NoLiquidStake_QueryTimeout() {
	lsmCallback := false
	tc := s.SetupValidatorICQCallback(false, lsmCallback)

	// Update the query so that it timed out
	badQuery := tc.validArgs.query
	badQuery.TimeoutTimestamp = 0

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().NoError(err, "validator shares to tokens rate callback error")

	// Confirm validator's shares to tokens rate DID NOT update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_LiquidStake_QueryTimeout() {
	lsmCallback := true
	tc := s.SetupValidatorICQCallback(false, lsmCallback)

	// Update the query so that it timed out
	badQuery := tc.validArgs.query
	badQuery.TimeoutTimestamp = 0

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().NoError(err, "validator shares to tokens rate callback error")

	// Confirm validator's shares to tokens rate DID NOT update
	expectedSharesToTokensRate := tc.initialState.validator.SharesToTokensRate
	s.checkValidatorSharesToTokensRate(expectedSharesToTokensRate, tc)

	// Confirm delegator shares query WAS NOT submitted
	s.checkDelegatorSharesQueryNotSubmitted()

	// Confirm the liquid stake failed
	s.checkLSMLiquidStakeFailed()
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_HostZoneNotFound() {
	tc := s.SetupValidatorICQCallback(false, false)

	// Set an incorrect host zone in the query
	badQuery := tc.validArgs.query
	badQuery.ChainId = "fake_host_zone"

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_InvalidCallbackArgs() {
	tc := s.SetupValidatorICQCallback(false, false)

	// Submit callback with invalid callback args (so that it can't unmarshal into a validator)
	invalidArgs := []byte("random bytes")
	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "unable to unmarshal query response into Validator type")
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_InvalidCallbackData() {
	tc := s.SetupValidatorICQCallback(false, false)

	// Submit callback with invalid callback args (so that it can't unmarshal into a validator)
	invalidQuery := tc.validArgs.query
	invalidQuery.CallbackData = []byte("random bytes")

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to unmarshal validator sharesToTokens rate callback data")
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_ValidatorNotFound() {
	tc := s.SetupValidatorICQCallback(false, false)

	// Update the callback args to contain a validator address that doesn't exist
	badCallbackArgs := s.CreateValidatorQueryResponse("fake_val", 1, 1)
	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no registered validator for address (fake_val): validator not found")
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_DelegatorSharesZero() {
	tc := s.SetupValidatorICQCallback(false, false)

	// Set the delegator shares to 0, which cause division by zero in `validator.TokensFromShares`
	valAddress := tc.initialState.validator.Address
	badCallbackArgs := s.CreateValidatorQueryResponse(valAddress, 1000, 0) // the 1000 is arbitrary, the zero here is what matters
	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)

	expectedErrMsg := "can't calculate validator internal sharesToTokens rate because delegation amount is 0 "
	expectedErrMsg += fmt.Sprintf("(validator: %s): division by zero", valAddress)
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestValidatorSharesToTokensRateCallback_DelegationQueryFailed() {
	tc := s.SetupValidatorICQCallback(true, false)

	// Remove host zone delegation address so delegation query fails
	badHostZone := tc.initialState.hostZone
	badHostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := keeper.ValidatorSharesToTokensRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "Failed to submit ICQ validator delegations")
}
