package v5_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v4/x/claim/types/v1"
	recordkeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	recordv1types "github.com/Stride-Labs/stride/v4/x/records/types/v1"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibcv1types "github.com/Stride-Labs/stride/v4/x/stakeibc/types/v1"
)

const dummyUpgradeHeight = 5

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		msg        string
		preUpdate  func()
		update     func()
		postUpdate func()
		expPass    bool
	}{
		{
			"Test that upgrade does not panic and store migrate successfully",
			func() {
				suite.Setup()
				suite.SetUpOldStore()
			},
			func() {
				suite.ConfirmUpgradeSucceededs("v5", dummyUpgradeHeight)
			},
			func() {
				suite.CheckStoreMigration()
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc.preUpdate()
		tc.update()
		tc.postUpdate()
	}
}

func (suite *UpgradeTestSuite) SetUpOldStore() {
	codec := simapp.MakeTestEncodingConfig().Codec

	// set up claim store
	claimStore := suite.Ctx.KVStore(suite.App.GetKey(claimtypes.StoreKey))

	params := claimv1types.Params{
		Airdrops: []*claimv1types.Airdrop{
			{
				AirdropStartTime: time.Now(),
				ClaimedSoFar: 1000000,
				AirdropDuration: time.Hour,
			},
		},
	}
	paramsBz, err := codec.MarshalJSON(&params)
	suite.Require().NoError(err)
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)

	// set up record store
	recordStore := suite.Ctx.KVStore(suite.App.GetKey(recordtypes.StoreKey))

	// set old deposit record
	depositRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.DepositRecordKey))
	depositRecord := recordv1types.DepositRecord{
		Id: uint64(1),
		Amount: int64(1000000),
		Denom: "ATOM",
		HostZoneId: "GAIA",
		DepositEpochNumber: uint64(1),
		Status: recordv1types.DepositRecord_DELEGATION_QUEUE,
		Source: recordv1types.DepositRecord_STRIDE,
	}
	depositBz, err := codec.Marshal(&depositRecord)
	suite.Require().NoError(err)
	depositRecordStore.Set(recordkeeper.GetDepositRecordIDBytes(depositRecord.Id), depositBz)

	// set old user redemption record
	userRedemptionRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.UserRedemptionRecordKey))
	userRedemptionRecord := recordv1types.UserRedemptionRecord{
		Id: "1",
		Amount: uint64(1000000),
		Denom: "ATOM",
		HostZoneId: "GAIA",
	}
	userRedemptionBz, err := codec.Marshal(&userRedemptionRecord)
	suite.Require().NoError(err)
	userRedemptionRecordStore.Set([]byte(userRedemptionRecord.Id), userRedemptionBz)

	// set old epoch unbongding record
	epochUnbondingRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.EpochUnbondingRecordKey))
	epochUnbondingRecord := recordv1types.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordv1types.HostZoneUnbonding{
			{
				HostZoneId: "GAIA",
				NativeTokenAmount: uint64(1000000),
				StTokenAmount: uint64(2000000),
				Status: recordv1types.HostZoneUnbonding_CLAIMABLE,
			},
		},
	}
	epochUnbondingBz, err := codec.Marshal(&epochUnbondingRecord)
	suite.Require().NoError(err)
	epochUnbondingRecordStore.Set(recordkeeper.GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.EpochNumber), epochUnbondingBz)

	// set up stakeibc module store
	stakeIbcStore := suite.Ctx.KVStore(suite.App.GetKey(stakeibctypes.StoreKey))

	// set old hostzone
	hostzoneStore := prefix.NewStore(stakeIbcStore, recordtypes.KeyPrefix(stakeibctypes.HostZoneKey))
	hz := stakeibcv1types.HostZone{
		ChainId: "GAIA",
		Validators: []*stakeibcv1types.Validator{
			{
				DelegationAmt: uint64(1000000),
			},
		},
		BlacklistedValidators: []*stakeibcv1types.Validator{
			{
				DelegationAmt: uint64(2000000),
			},
		},
		StakedBal: uint64(3000000),
		LastRedemptionRate: sdk.OneDec(),
		RedemptionRate: sdk.OneDec(),
	}
	hzBz, err := codec.Marshal(&hz)
	suite.Require().NoError(err)
	hostzoneStore.Set([]byte(hz.ChainId), hzBz)	
}

func (suite *UpgradeTestSuite) CheckStoreMigration() {
	claimParams, err := suite.App.ClaimKeeper.GetParams(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdk.NewInt(1000000))

	depositRecord, bool := suite.App.RecordsKeeper.GetDepositRecord(suite.Ctx, uint64(1))
	suite.Require().True(bool)
	suite.Require().Equal(depositRecord.Amount, sdk.NewInt(1000000))

	userRedeemRecord, bool := suite.App.RecordsKeeper.GetUserRedemptionRecord(suite.Ctx, "1")
	suite.Require().True(bool)
	suite.Require().Equal(userRedeemRecord.Amount, sdk.NewInt(1000000))

	epochUnbondingRecord, bool := suite.App.RecordsKeeper.GetEpochUnbondingRecord(suite.Ctx, 1)
	suite.Require().True(bool)
	suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount, sdk.NewInt(2000000))
	suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].NativeTokenAmount, sdk.NewInt(1000000))

	hz, bool := suite.App.StakeibcKeeper.GetHostZone(suite.Ctx, "GAIA")
	suite.Require().True(bool)
	suite.Require().Equal(hz.StakedBal, sdk.NewInt(3000000))
	suite.Require().Equal(hz.Validators[0].DelegationAmt, sdk.NewInt(1000000))
	suite.Require().Equal(hz.BlacklistedValidators[0].DelegationAmt, sdk.NewInt(2000000))
}
