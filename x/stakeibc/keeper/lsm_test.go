package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetLSMTokenDenomTrace() {
	baseDenom := "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/48"
	path := "transfer/channel-0"
	ibcDenom := transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s", path, baseDenom)).IBCDenom()

	// Store denom trace so the transfer keeper can look it up
	expectedDenomTrace := transfertypes.DenomTrace{
		BaseDenom: baseDenom,
		Path:      path,
	}
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, expectedDenomTrace)

	// Test parsing of IBC Denom
	actualDenomTrace, err := s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, ibcDenom)
	s.Require().NoError(err, "no error expected with successful parse")
	s.Require().Equal(expectedDenomTrace, actualDenomTrace, "denom trace")

	// Attempt to parse with a non-ibc denom - it should fail
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, "non-ibc-denom")
	s.Require().ErrorContains(err, "lsm token is not an IBC token (non-ibc-denom)")

	// Attempt to parse with an invalid ibc-denom - it should fail
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, "ibc/xxx")
	s.Require().ErrorContains(err, "unable to get ibc hex hash from denom ibc/xxx")

	// Attempt to parse with a valid ibc denom that is not registered - it should fail
	notRegisteredIBCDenom := transfertypes.ParseDenomTrace("transfer/channel-0/cosmosXXX").IBCDenom()
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, notRegisteredIBCDenom)
	s.Require().ErrorContains(err, "denom trace not found")
}

func (s *KeeperTestSuite) TestIsValidIBCPath() {
	validIBCPaths := []string{
		"transfer/channel-0",
		"transfer/channel-10",
		"transfer/channel-99999",
	}
	invalidIBCPaths := []string{
		"transferx/channel-0",
		"transfer/channel-X",
		"transfer/channel-0/transfer/channel-1",
	}

	for _, validPath := range validIBCPaths {
		s.Require().True(keeper.IsValidIBCPath(validPath), "should be valid")
	}
	for _, validPath := range invalidIBCPaths {
		s.Require().False(keeper.IsValidIBCPath(validPath), "should be invalid")
	}
}

func (s *KeeperTestSuite) TestGetHostZoneFromLSMTokenPath() {
	// Set a host zone in the store with channel-0
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:           HostChainId,
		TransferChannelId: ibctesting.FirstChannelID,
	})

	// Successful lookup
	validPath := fmt.Sprintf("%s/%s", transfertypes.PortID, ibctesting.FirstChannelID)
	hostZone, err := s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, validPath)
	s.Require().NoError(err, "no error expected from valid path")
	s.Require().Equal(HostChainId, hostZone.ChainId, "host zone")

	// Invalid IBC path should fail
	_, err = s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, "transfer/channel-0/transfer/channel-1")
	s.Require().ErrorContains(err, "ibc path of LSM token (transfer/channel-0/transfer/channel-1) cannot be more than 1 hop away")

	// Passing an unregistered channel-id should cause it to fail
	_, err = s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, "transfer/channel-1")
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-1) does not match any registered host zone")
}

func (s *KeeperTestSuite) TestGetValidatorFromLSMTokenDenom() {
	valAddress := "cosmosvaloperXXX"
	denom := valAddress + "/42" // add record ID
	validators := []*types.Validator{{Address: valAddress}}

	// Successful lookup
	validator, err := s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom(denom, validators)
	s.Require().NoError(err, "no error expected from valid lsm denom")
	s.Require().Equal(valAddress, validator.Address, "host zone")

	// Invalid LSM denoms - should fail
	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom("invalid_denom", validators)
	s.Require().ErrorContains(err, "lsm token base denom is not of the format {val-address}/{record-id} (invalid_denom)")

	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom("cosmosvaloperXXX/42/1", validators)
	s.Require().ErrorContains(err, "lsm token base denom is not of the format {val-address}/{record-id} (cosmosvaloperXXX/42/1)")

	// Validator does not exist - should fail
	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom(denom, []*types.Validator{})
	s.Require().ErrorContains(err, "validator (cosmosvaloperXXX) is not registered in the Stride validator set")
}

func (s *KeeperTestSuite) TestShouldQueryValidatorExchangeRate() {
	testCases := []struct {
		name                string
		queryInterval       uint64
		progress            sdk.Int
		stakeAmount         sdk.Int
		expectedShouldQuery bool
	}{
		{
			name:                "interval 1 - short of checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(99),
			expectedShouldQuery: false,
		},
		{
			name:                "interval 1 - at checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(100),
			expectedShouldQuery: true,
		},
		{
			name:                "interval 1 - past checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(101),
			expectedShouldQuery: true,
		},
		{
			name:                "interval 2 - short of checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(133),  // 4,133
			expectedShouldQuery: false,
		},
		{
			name:                "interval 2 - at checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(134),  // 4,134
			expectedShouldQuery: true,
		},
		{
			name:                "interval 2 - past checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(135),  // 4,135
			expectedShouldQuery: true,
		},
	}

	for _, tc := range testCases {
		// Store query interval param
		params := types.DefaultParams()
		params.ValidatorExchangeRateQueryInterval = tc.queryInterval
		s.App.StakeibcKeeper.SetParams(s.Ctx, params)

		validator := types.Validator{ProgressTowardsExchangeRateQuery: tc.progress}
		actualShouldQuery := s.App.StakeibcKeeper.ShouldQueryValidatorExchangeRate(s.Ctx, validator, tc.stakeAmount)
		s.Require().Equal(tc.expectedShouldQuery, actualShouldQuery, tc.name)
	}
}
