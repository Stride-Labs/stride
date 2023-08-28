package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type AddValidatorsTestCase struct {
	hostZone                 types.HostZone
	validMsg                 types.MsgAddValidators
	expectedValidators       []*types.Validator
	validatorQueryDataToName map[string]string
}

// Helper function to determine the validator's key in the staking store
// which is used as the request data in the ICQ
func (s *KeeperTestSuite) getSharesToTokensRateQueryData(validatorAddress string) []byte {
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	s.Require().NoError(err, "no error expected when decoding validator address")
	return stakingtypes.GetValidatorKey(validatorAddressBz)
}

func (s *KeeperTestSuite) SetupAddValidators() AddValidatorsTestCase {
	slashThreshold := uint64(10)
	params := types.DefaultParams()
	params.ValidatorSlashQueryThreshold = slashThreshold
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	totalDelegations := sdkmath.NewInt(100_000)
	expectedSlashCheckpoint := sdkmath.NewInt(10_000)

	hostZone := types.HostZone{
		ChainId:          "GAIA",
		ConnectionId:     ibctesting.FirstConnectionID,
		Validators:       []*types.Validator{},
		TotalDelegations: totalDelegations,
	}

	validatorAddresses := map[string]string{
		"val1": "stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm",
		"val2": "stridevaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rcfud5c",
		"val3": "stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f",
	}

	// mapping of query request data to validator name
	// serves as a reverse lookup to map sharesToTokens rate queries to validators
	validatorQueryDataToName := map[string]string{}
	for name, address := range validatorAddresses {
		queryData := s.getSharesToTokensRateQueryData(address)
		validatorQueryDataToName[string(queryData)] = name
	}

	validMsg := types.MsgAddValidators{
		Creator:  "stride_ADMIN",
		HostZone: HostChainId,
		Validators: []*types.Validator{
			{Name: "val1", Address: validatorAddresses["val1"], Weight: 1},
			{Name: "val2", Address: validatorAddresses["val2"], Weight: 2},
			{Name: "val3", Address: validatorAddresses["val3"], Weight: 3},
		},
	}

	expectedValidators := []*types.Validator{
		{Name: "val1", Address: validatorAddresses["val1"], Weight: 1},
		{Name: "val2", Address: validatorAddresses["val2"], Weight: 2},
		{Name: "val3", Address: validatorAddresses["val3"], Weight: 3},
	}
	for _, validator := range expectedValidators {
		validator.Delegation = sdkmath.ZeroInt()
		validator.SlashQueryProgressTracker = sdkmath.ZeroInt()
		validator.SharesToTokensRate = sdk.ZeroDec()
		validator.SlashQueryCheckpoint = expectedSlashCheckpoint
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the latest client height for the ICQ submission
	s.MockClientLatestHeight(1)

	return AddValidatorsTestCase{
		hostZone:                 hostZone,
		validMsg:                 validMsg,
		expectedValidators:       expectedValidators,
		validatorQueryDataToName: validatorQueryDataToName,
	}
}

func (s *KeeperTestSuite) TestAddValidators_Successful() {
	tc := s.SetupAddValidators()

	// Add validators
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(3, len(hostZone.Validators), "number of validators")

	for i := 0; i < 3; i++ {
		s.Require().Equal(*tc.expectedValidators[i], *hostZone.Validators[i], "validators %d", i)
	}

	// Confirm ICQs were submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 3)

	// Map the query responses to the validator names to get the names of the validators that
	// were queried
	queriedValidators := []string{}
	for i, query := range queries {
		validator, ok := tc.validatorQueryDataToName[string(query.RequestData)]
		s.Require().True(ok, "query from response %d does not match any expected requests", i)
		queriedValidators = append(queriedValidators, validator)
	}

	// Confirm the list of queried validators matches the full list of validators
	allValidatorNames := []string{}
	for _, expected := range tc.expectedValidators {
		allValidatorNames = append(allValidatorNames, expected.Name)
	}
	s.Require().ElementsMatch(allValidatorNames, queriedValidators, "queried validators")
}

func (s *KeeperTestSuite) TestAddValidators_HostZoneNotFound() {
	tc := s.SetupAddValidators()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsg
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, "Host Zone (gaia) not found: host zone not found")
}

func (s *KeeperTestSuite) TestAddValidators_AddressAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that the name val1 already exists
	hostZone := tc.hostZone
	duplicateAddress := tc.expectedValidators[0].Address
	duplicateVal := types.Validator{Name: "new_val", Address: duplicateAddress}
	hostZone.Validators = []*types.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator address to val1 so that the message errors
	expectedError := fmt.Sprintf("Validator address (%s) already exists on Host Zone (GAIA)", duplicateAddress)
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, expectedError)
}

func (s *KeeperTestSuite) TestAddValidators_NameAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that val1's address already exists
	hostZone := tc.hostZone
	duplicateName := tc.expectedValidators[0].Name
	duplicateVal := types.Validator{Name: duplicateName, Address: "new_address"}
	hostZone.Validators = []*types.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator name to val1 so that the message errors
	expectedError := fmt.Sprintf("Validator name (%s) already exists on Host Zone (GAIA)", duplicateName)
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, expectedError)
}
