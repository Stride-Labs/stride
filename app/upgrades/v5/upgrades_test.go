package v5_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v4/x/claim/types/v1"
	recordkeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	recordv1types "github.com/Stride-Labs/stride/v4/x/records/types/v1"
	// stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	// stakeibcv1types "github.com/Stride-Labs/stride/v4/x/stakeibc/types/v1"
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
				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v5", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
				suite.Require().NoError(err)
				plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
				suite.Require().True(exists)

				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(suite.Ctx, beginBlockRequest)
				})
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
	codec := simapp.MakeTestEncodingConfig().Marshaler

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

}

func (suite *UpgradeTestSuite) CheckStoreMigration() {
	claimParams, err := suite.App.ClaimKeeper.GetParams(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdk.NewInt(1000000))

	depositRecord, bool := suite.App.RecordsKeeper.GetDepositRecord(suite.Ctx, uint64(1))
	suite.Require().True(bool)
	suite.Require().Equal(depositRecord.Amount, sdk.NewInt(1000000))
}
