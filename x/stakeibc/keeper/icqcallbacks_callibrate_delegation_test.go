package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestCalibrateDelegation() {
	valIndexQueried := 1
	tokensBeforeSlash := sdkmath.NewInt(10000)
	sharesToTokensRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := sdk.NewDec(5000)

	// 5000 shares * 0.5 sharesToTokens rate = 2500 tokens
	// 10000 tokens - 2500 token = 7500 tokens slashed
	// 7500 slash tokens / 10000 initial tokens = 75% slash
	expectedTokensAfterSlash := sdkmath.NewInt(2500)
	expectedSlashAmount := tokensBeforeSlash.Sub(expectedTokensAfterSlash)
	weightBeforeSlash := uint64(20)
	totalDelegation := sdkmath.NewInt(100000)

	s.Require().Equal(numShares, sdk.NewDecFromInt(expectedTokensAfterSlash.Mul(sdkmath.NewInt(2))), "tokens, shares, and sharesToTokens rate aligned")

	hostZone := types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: totalDelegation,
		Validators: []*types.Validator{
			// This validator isn't being queried
			{
				Name:       "val1",
				Address:    "valoper1",
				Delegation: sdkmath.ZeroInt(),
			},
			// This is the validator being queried
			{
				Name:               "val2",
				Address:            ValAddress,
				SharesToTokensRate: sharesToTokensRate,
				Delegation:         tokensBeforeSlash,
			},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock out the query and query response
	queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, numShares)
	query := icqtypes.Query{ChainId: HostChainId}

	delegationAmountBeforeSlash := hostZone.Validators[valIndexQueried].Delegation

	//////////

	// Callback
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
	// nil return indicates that the callback was not successful
	s.Require().Nil(err)

	// Confirm the staked balance was not decreased on the host
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(sdk.ZeroInt().Int64(), hostZone.TotalDelegations.Sub(hostZone.TotalDelegations).Int64(), "staked bal not slashed")

	// Confirm the validator's weight and delegation amount were not reduced
	validator := hostZone.Validators[valIndexQueried]
	s.Require().Equal(weightBeforeSlash, validator.Weight, "validator weight unchanged")
	s.Require().Equal(delegationAmountBeforeSlash.Int64(), validator.Delegation.Int64(), "validator delegation amount")

	// Confirm the validator query is still in progress (calibration callback does not set it false)
	s.Require().True(validator.SlashQueryInProgress, "slash query in progress")
}
