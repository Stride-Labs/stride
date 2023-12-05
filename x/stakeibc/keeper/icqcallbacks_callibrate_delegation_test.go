package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupDelegatorSharesICQCallbackLargeSlash() DelegatorSharesICQCallbackTestCase {
	// Setting this up to initialize the coordinator for the block time
	s.CreateTransferChannel(HostChainId)

	valIndexQueried := 1
	tokensBeforeSlash := sdkmath.NewInt(10000)
	sharesToTokensRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := sdk.NewDec(5000)

	// 5000 shares * 0.5 sharesToTokens rate = 2500 tokens
	// 10000 tokens - 2500 token = 7500 tokens slashed
	// 7500 slash tokens / 10000 initial tokens = 75% slash
	expectedTokensAfterSlash := sdkmath.NewInt(2500)
	expectedSlashAmount := tokensBeforeSlash.Sub(expectedTokensAfterSlash)
	slashPercentage := sdk.MustNewDecFromStr("0.75")
	weightBeforeSlash := uint64(20)
	expectedWeightAfterSlash := uint64(5)
	totalDelegation := sdkmath.NewInt(100000)

	s.Require().Equal(numShares, sdk.NewDecFromInt(expectedTokensAfterSlash.Mul(sdkmath.NewInt(2))), "tokens, shares, and sharesToTokens rate aligned")
	s.Require().Equal(slashPercentage, sdk.NewDecFromInt(expectedSlashAmount).Quo(sdk.NewDecFromInt(tokensBeforeSlash)), "expected slash percentage")
	s.Require().Equal(slashPercentage, sdk.NewDec(int64(weightBeforeSlash-expectedWeightAfterSlash)).Quo(sdk.NewDec(int64(weightBeforeSlash))), "weight reduction")

	hostZone := types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: totalDelegation,
		Validators: []*types.Validator{
			// This validator isn't being queried
			{
				Name:       "val1",
				Address:    "valoper1",
				Weight:     1,
				Delegation: sdkmath.ZeroInt(),
			},
			// This is the validator in question
			{
				Name:                        "val2",
				Address:                     ValAddress,
				SharesToTokensRate:          sharesToTokensRate,
				Delegation:                  tokensBeforeSlash,
				Weight:                      weightBeforeSlash,
				SlashQueryInProgress:        true,
				DelegationChangesInProgress: 0,
			},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, numShares)

	// Create callback data
	callbackDataBz, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: tokensBeforeSlash,
	})
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Set the timeout timestamp to be 1 minute after the block time, and
	// the timeout duration to be 5 minutes
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Ctx = s.Ctx.WithBlockTime(blockTime)
	timeoutDuration := time.Minute
	timeoutTimestamp := uint64(blockTime.Add(timeoutDuration).UnixNano())

	// Create the query that represents the ICQ in flight
	query := icqtypes.Query{
		Id:               "query-1",
		ChainId:          HostChainId,
		ConnectionId:     ibctesting.FirstConnectionID,
		QueryType:        icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		CallbackData:     callbackDataBz,
		CallbackId:       keeper.ICQCallbackID_Delegation,
		CallbackModule:   types.ModuleName,
		TimeoutDuration:  timeoutDuration,
		TimeoutTimestamp: timeoutTimestamp,
		RequestSent:      true,
		TimeoutPolicy:    icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)

	return DelegatorSharesICQCallbackTestCase{
		valIndexQueried: valIndexQueried,
		validArgs: DelegatorSharesICQCallbackArgs{
			query:        query,
			callbackArgs: queryResponse,
		},
		hostZone:                 hostZone,
		numShares:                numShares,
		slashPercentage:          slashPercentage,
		expectedDelegationAmount: expectedTokensAfterSlash,
		expectedSlashAmount:      expectedSlashAmount,
		expectedWeight:           expectedWeightAfterSlash,
		sharesToTokensRate:       sharesToTokensRate,
		retryTimeoutDuration:     timeoutDuration,
	}
}

func (s *KeeperTestSuite) TestCalibrateDelegationCallback_SlashBeyondThreshold() {
	tc := s.SetupDelegatorSharesICQCallbackLargeSlash()
	weightBeforeSlash := tc.hostZone.Validators[tc.valIndexQueried].Weight
	delegationAmountBeforeSlash := tc.hostZone.Validators[tc.valIndexQueried].Delegation
	// Callback
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	// nil return indicates that the callback was not successful
	s.Require().Nil(err)

	// Confirm the staked balance was not decreased on the host
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	// println(tc.hostZone.TotalDelegations.Sub(hostZone.TotalDelegations).Int64())
	// println(tc.hostZone.TotalDelegations.Int64())
	s.Require().Equal(sdk.ZeroInt().Int64(), tc.hostZone.TotalDelegations.Sub(hostZone.TotalDelegations).Int64(), "staked bal not slashed")

	// Confirm the validator's weight and delegation amount were not reduced
	validator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(weightBeforeSlash, validator.Weight, "validator weight unchanged")
	s.Require().Equal(delegationAmountBeforeSlash.Int64(), validator.Delegation.Int64(), "validator delegation amount")

	// Confirm the validator query is still in progress (calibration callback does not set it false)
	s.Require().True(validator.SlashQueryInProgress, "slash query in progress")
}
