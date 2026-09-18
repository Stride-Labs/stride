package v34_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	cosmosHubOtherValidatorAddress = "cosmosvaloper1otherxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	cosmosHubHostDenom             = "uatom"
)

// setupCosmosHubHostZone seeds the cosmoshub-4 host zone with the stakewithus validator (holding
// trackedDelegation) plus one other validator, so tests can assert the other validator is left
// untouched while stakewithus and TotalDelegations move by exactly the closed amount.
func (s *UpgradeTestSuite) setupCosmosHubHostZone(trackedDelegation sdkmath.Int) (otherDelegation sdkmath.Int) {
	otherDelegation = sdkmath.NewInt(555_000_000)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:   v34.CosmosHubChainId,
		HostDenom: cosmosHubHostDenom,
		Validators: []*stakeibctypes.Validator{
			{
				Name:               "stakewithus",
				Address:            v34.CosmosHubStrandedLsmDeposit.ValidatorAddress,
				Delegation:         trackedDelegation,
				SharesToTokensRate: sdkmath.LegacyOneDec(),
			},
			{
				Name:               "other",
				Address:            cosmosHubOtherValidatorAddress,
				Delegation:         otherDelegation,
				SharesToTokensRate: sdkmath.LegacyOneDec(),
			},
		},
		TotalDelegations: trackedDelegation.Add(otherDelegation),
	})
	return otherDelegation
}

// seedCosmosHubLsmDeposit stores an LSM token deposit keyed at the constant's denom (so
// CloseCosmosHubLsmDeposit can find it by chain id + denom) with the given status, amount and
// validator, and returns the seeded deposit.
func (s *UpgradeTestSuite) seedCosmosHubLsmDeposit(
	status recordstypes.LSMTokenDeposit_Status,
	amount sdkmath.Int,
	validatorAddress string,
) recordstypes.LSMTokenDeposit {
	deposit := recordstypes.LSMTokenDeposit{
		DepositId:        "ba82f49a967dda66331221b18018d305198abd72fba55552a451f1012274a57d",
		ChainId:          v34.CosmosHubChainId,
		Denom:            v34.CosmosHubStrandedLsmDeposit.Denom,
		StakerAddress:    "stride1v6ll7lj9qeyf6g6at2c4vqetxg5877558cqpjj",
		ValidatorAddress: validatorAddress,
		Amount:           amount,
		Status:           status,
	}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)
	return deposit
}

// cosmosHubRateNumerator is the redemption rate numerator components a pure bucket move (LSM
// deposit -> native delegation) must leave unchanged to the uatom. Shared with the mainnet export
// suite, hence a free function over the common test helper.
func cosmosHubRateNumerator(s *apptesting.AppTestHelper) sdkmath.LegacyDec {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	s.Require().True(found)
	tokenized := s.App.StakeibcKeeper.GetTotalTokenizedDelegations(s.Ctx, hostZone)
	return tokenized.Add(sdkmath.LegacyNewDecFromInt(hostZone.TotalDelegations))
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit() {
	tracked := sdkmath.NewInt(1_000_000_000)
	otherDelegation := s.setupCosmosHubHostZone(tracked)
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED,
		v34.CosmosHubStrandedLsmDeposit.Amount, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)
	numeratorBefore := cosmosHubRateNumerator(&s.AppTestHelper)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().True(applied)

	_, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().False(found, "the LSM deposit record should be removed")

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	s.Require().True(found)
	validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)
	s.Require().True(found)
	s.Require().Equal(tracked.Add(v34.CosmosHubStrandedLsmDeposit.Amount).String(), validator.Delegation.String(),
		"stakewithus delegation up by exactly the closed amount")
	s.Require().Equal(tracked.Add(otherDelegation).Add(v34.CosmosHubStrandedLsmDeposit.Amount).String(), hostZone.TotalDelegations.String(),
		"TotalDelegations up by exactly the closed amount")

	other, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, cosmosHubOtherValidatorAddress)
	s.Require().True(found)
	s.Require().Equal(otherDelegation.String(), other.Delegation.String(), "other validator must be untouched")

	s.Require().Equal(numeratorBefore.String(), cosmosHubRateNumerator(&s.AppTestHelper).String(), "redemption rate components unchanged")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_WrongStatusSkips() {
	tracked := sdkmath.NewInt(1_000_000_000)
	s.setupCosmosHubHostZone(tracked)
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		v34.CosmosHubStrandedLsmDeposit.Amount, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	deposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().True(found, "the deposit must not be removed")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, deposit.Status)

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	validator, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)
	s.Require().Equal(tracked.String(), validator.Delegation.String(), "delegation must be untouched")
	s.Require().Equal(tracked.Add(sdkmath.NewInt(555_000_000)).String(), hostZone.TotalDelegations.String(), "TotalDelegations must be untouched")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_WrongAmountSkips() {
	tracked := sdkmath.NewInt(1_000_000_000)
	s.setupCosmosHubHostZone(tracked)
	wrongAmount := v34.CosmosHubStrandedLsmDeposit.Amount.AddRaw(1)
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, wrongAmount, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	deposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().True(found, "the deposit must not be removed")
	s.Require().Equal(wrongAmount.String(), deposit.Amount.String())

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	validator, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)
	s.Require().Equal(tracked.String(), validator.Delegation.String(), "delegation must be untouched")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_WrongValidatorOnRecordSkips() {
	tracked := sdkmath.NewInt(1_000_000_000)
	s.setupCosmosHubHostZone(tracked)
	// The record's validator address does not match the constant, even though that other
	// validator is present on the host zone
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED,
		v34.CosmosHubStrandedLsmDeposit.Amount, cosmosHubOtherValidatorAddress)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	_, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().True(found, "the deposit must not be removed")

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	other, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, cosmosHubOtherValidatorAddress)
	s.Require().Equal(sdkmath.NewInt(555_000_000).String(), other.Delegation.String(), "delegation must be untouched")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_MissingRecordSkips() {
	tracked := sdkmath.NewInt(1_000_000_000)
	s.setupCosmosHubHostZone(tracked)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	validator, _, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)
	s.Require().Equal(tracked.String(), validator.Delegation.String(), "delegation must be untouched")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_MissingValidatorSkips() {
	// Host zone exists but does not carry the stakewithus validator
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:   v34.CosmosHubChainId,
		HostDenom: cosmosHubHostDenom,
		Validators: []*stakeibctypes.Validator{
			{Name: "other", Address: cosmosHubOtherValidatorAddress, Delegation: sdkmath.NewInt(555_000_000)},
		},
		TotalDelegations: sdkmath.NewInt(555_000_000),
	})
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED,
		v34.CosmosHubStrandedLsmDeposit.Amount, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	_, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().True(found, "the deposit must not be removed")

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	s.Require().Equal(sdkmath.NewInt(555_000_000).String(), hostZone.TotalDelegations.String(), "TotalDelegations must be untouched")
}

func (s *UpgradeTestSuite) TestCloseCosmosHubLsmDeposit_MissingHostZoneSkips() {
	s.seedCosmosHubLsmDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED,
		v34.CosmosHubStrandedLsmDeposit.Amount, v34.CosmosHubStrandedLsmDeposit.ValidatorAddress)

	applied := v34.CloseCosmosHubLsmDeposit(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().False(applied)

	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.CosmosHubChainId)
	s.Require().False(found, "the close-out should not create the host zone")

	_, found = s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v34.CosmosHubChainId, v34.CosmosHubStrandedLsmDeposit.Denom)
	s.Require().True(found, "the deposit must not be removed")
}
