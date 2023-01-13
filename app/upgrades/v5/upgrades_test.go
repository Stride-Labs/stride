package v5_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	oldclaimtypes "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	recordkeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	oldrecordtypes "github.com/Stride-Labs/stride/v4/x/records/migrations/v2/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	oldstakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
	suite.Setup()

	codec := app.MakeEncodingConfig().Marshaler
	checkClaimStoreAfterMigration := suite.SetupOldClaimStore(codec)
	checkIcacallbackStoreAfterMigration := suite.SetupOldIcacallbackStore(codec)
	checkRecordStoreAfterMigration := suite.SetupOldRecordStore(codec)
	checkStakeibcStoreAfterMigration := suite.SetupOldStakeibcStore(codec)

	suite.ConfirmUpgradeSucceededs("v5", dummyUpgradeHeight)

	checkClaimStoreAfterMigration()
	checkIcacallbackStoreAfterMigration()
	checkRecordStoreAfterMigration()
	checkStakeibcStoreAfterMigration()
}

// Sets up the old claim store and returns a callback function that can be used to verify
// the store migration was successful
func (suite *UpgradeTestSuite) SetupOldClaimStore(codec codec.Codec) func() {
	claimStore := suite.Ctx.KVStore(suite.App.GetKey(claimtypes.StoreKey))

	airdropId := "id"
	params := oldclaimtypes.Params{
		Airdrops: []*oldclaimtypes.Airdrop{
			{
				AirdropIdentifier: airdropId,
				ClaimedSoFar:      1000000,
			},
		},
	}

	paramsBz, err := codec.MarshalJSON(&params)
	suite.Require().NoError(err)
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)

	// Callback to check claim store after migration
	return func() {
		claimParams, err := suite.App.ClaimKeeper.GetParams(suite.Ctx)
		suite.Require().NoError(err, "no error expected when getting claims")
		suite.Require().Equal(claimParams.Airdrops[0].AirdropIdentifier, airdropId, "airdrop identifier")
		suite.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdk.NewInt(1000000), "claimed so far")
	}
}

// Sets up the old icacallbacks store and returns a callback function that can be used to verify
// the store migration was successful
func (suite *UpgradeTestSuite) SetupOldIcacallbackStore(codec codec.Codec) func() {

	// Callback to check icacallback store after migration
	return func() {

	}
}

// Sets up the old records store and returns a callback function that can be used to verify
// the store migration was successful
func (suite *UpgradeTestSuite) SetupOldRecordStore(codec codec.Codec) func() {
	recordStore := suite.Ctx.KVStore(suite.App.GetKey(recordtypes.StoreKey))

	// set old deposit record
	depositRecordId := uint64(1)
	depositRecord := oldrecordtypes.DepositRecord{
		Id:     depositRecordId,
		Amount: int64(1000000),
		Status: oldrecordtypes.DepositRecord_DELEGATION_QUEUE,
		Source: oldrecordtypes.DepositRecord_WITHDRAWAL_ICA,
	}
	depositBz, err := codec.Marshal(&depositRecord)
	suite.Require().NoError(err)

	depositRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.DepositRecordKey))
	depositRecordStore.Set(recordkeeper.GetDepositRecordIDBytes(depositRecord.Id), depositBz)

	// set old user redemption record
	userRedemptionRecordId := "1"
	userRedemptionRecord := oldrecordtypes.UserRedemptionRecord{
		Id:     "1",
		Amount: uint64(1000000),
	}
	userRedemptionBz, err := codec.Marshal(&userRedemptionRecord)
	suite.Require().NoError(err)

	userRedemptionRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.UserRedemptionRecordKey))
	userRedemptionRecordStore.Set([]byte(userRedemptionRecord.Id), userRedemptionBz)

	// set old epoch unbongding record
	epochNumber := uint64(1)
	hostZoneId := "hz"
	epochUnbondingRecord := oldrecordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*oldrecordtypes.HostZoneUnbonding{
			{
				HostZoneId:        "hz",
				NativeTokenAmount: uint64(1000000),
				StTokenAmount:     uint64(2000000),
				Status:            oldrecordtypes.HostZoneUnbonding_CLAIMABLE,
			},
		},
	}
	epochUnbondingBz, err := codec.Marshal(&epochUnbondingRecord)
	suite.Require().NoError(err)

	epochUnbondingRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.EpochUnbondingRecordKey))
	epochUnbondingRecordStore.Set(recordkeeper.GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.EpochNumber), epochUnbondingBz)

	// Callback to check record store after migration
	return func() {
		depositRecord, found := suite.App.RecordsKeeper.GetDepositRecord(suite.Ctx, depositRecordId)
		suite.Require().True(found, "deposit record found")
		suite.Require().Equal(depositRecord.Id, depositRecordId, "deposit record id")
		suite.Require().Equal(depositRecord.Amount, sdk.NewInt(1000000), "deposit record amount")
		suite.Require().Equal(depositRecord.Status, recordtypes.DepositRecord_DELEGATION_QUEUE, "deposit record status")
		suite.Require().Equal(depositRecord.Source, recordtypes.DepositRecord_WITHDRAWAL_ICA, "deposit record source")

		userRedemptionRecord, found := suite.App.RecordsKeeper.GetUserRedemptionRecord(suite.Ctx, userRedemptionRecordId)
		suite.Require().True(found, "redemption record found")
		suite.Require().Equal(userRedemptionRecord.Id, userRedemptionRecordId, "redemption record id")
		suite.Require().Equal(userRedemptionRecord.Amount, sdk.NewInt(1000000), "redemption record amount")

		epochUnbondingRecord, found := suite.App.RecordsKeeper.GetEpochUnbondingRecord(suite.Ctx, epochNumber)
		suite.Require().True(found, "epoch unbonding record found")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].HostZoneId, hostZoneId, "host zone unbonding host zone id")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].NativeTokenAmount, sdk.NewInt(1000000), "host zone unbonding native token amount")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount, sdk.NewInt(2000000), "host zone unbonding sttoken amount")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].Status, recordtypes.HostZoneUnbonding_CLAIMABLE, "host zone unbonding status")
	}
}

func (suite *UpgradeTestSuite) SetupOldStakeibcStore(codec codec.Codec) func() {
	stakeibcStore := suite.Ctx.KVStore(suite.App.GetKey(stakeibctypes.StoreKey))

	// set old hostzone
	hostZoneId := "hz"
	delegationAddress := "delegation"
	redemptionAddress := "redemption"
	hostZone := oldstakeibctypes.HostZone{
		ChainId:           hostZoneId,
		DelegationAccount: &oldstakeibctypes.ICAAccount{Address: delegationAddress, Target: oldstakeibctypes.ICAAccountType_DELEGATION},
		RedemptionAccount: &oldstakeibctypes.ICAAccount{Address: redemptionAddress, Target: oldstakeibctypes.ICAAccountType_REDEMPTION},
		Validators: []*oldstakeibctypes.Validator{
			{
				DelegationAmt: uint64(1000000),
			},
		},
		BlacklistedValidators: []*oldstakeibctypes.Validator{
			{
				DelegationAmt: uint64(2000000),
			},
		},
		StakedBal:          uint64(3000000),
		LastRedemptionRate: sdk.OneDec(),
		RedemptionRate:     sdk.OneDec(),
	}
	hostZoneBz, err := codec.Marshal(&hostZone)
	suite.Require().NoError(err)

	hostzoneStore := prefix.NewStore(stakeibcStore, stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))
	hostzoneStore.Set([]byte(hostZone.ChainId), hostZoneBz)

	// Callback to check stakeibc store after migration
	return func() {
		hostZone, found := suite.App.StakeibcKeeper.GetHostZone(suite.Ctx, hostZoneId)
		suite.Require().True(found, "host zone found")
		suite.Require().Equal(hostZone.ChainId, hostZoneId, "host zone chain id")

		suite.Require().Equal(hostZone.DelegationAccount.Address, delegationAddress, "delegation address")
		suite.Require().Equal(hostZone.RedemptionAccount.Address, redemptionAddress, "redemption address")

		suite.Require().Equal(hostZone.DelegationAccount.Target, stakeibctypes.ICAAccountType_DELEGATION, "delegation target")
		suite.Require().Equal(hostZone.RedemptionAccount.Target, stakeibctypes.ICAAccountType_REDEMPTION, "redemption target")

		suite.Require().Nil(hostZone.FeeAccount, "fee account")
		suite.Require().Nil(hostZone.WithdrawalAccount, "withdrawal account")

		suite.Require().Equal(hostZone.Validators[0].DelegationAmt, sdk.NewInt(1000000), "host zone validators delegation amount")
		suite.Require().Equal(hostZone.BlacklistedValidators[0].DelegationAmt, sdk.NewInt(2000000), "host zone blacklisted validators delegation amount")
		suite.Require().Equal(hostZone.StakedBal, sdk.NewInt(3000000), "host zone staked balance")
	}
}
