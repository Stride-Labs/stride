package v5_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/golang/protobuf/proto"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	oldclaimtypes "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordkeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	oldrecordtypes "github.com/Stride-Labs/stride/v4/x/records/migrations/v2/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
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
		suite.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdkmath.NewInt(1000000), "claimed so far")
	}
}

// Stores delegate callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (suite *UpgradeTestSuite) SetupOldDelegateCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	delegateValidator := "dval"
	delegateCallback := oldstakeibctypes.DelegateCallback{
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{
			{Validator: delegateValidator, Amount: uint64(1000000)},
		},
	}
	delegateCallbackArgs, err := proto.Marshal(&delegateCallback)
	suite.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	delegateKey := "delegate_key"
	delegateCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  delegateKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Delegate,
		CallbackArgs: delegateCallbackArgs,
	}
	delegateCallbackDataBz, err := codec.Marshal(&delegateCallbackData)
	suite.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(delegateKey), delegateCallbackDataBz)

	// Check delegate callback args after the migration
	return func() {
		delegateCallbackData, found := suite.App.IcacallbacksKeeper.GetCallbackData(suite.Ctx, delegateKey)
		suite.Require().True(found, "found delegate callback data")

		var delegateCallback stakeibctypes.DelegateCallback
		err := proto.Unmarshal(delegateCallbackData.CallbackArgs, &delegateCallback)
		suite.Require().NoError(err, "unmarshaling delegate callback args should not error")

		suite.Require().Equal(delegateValidator, delegateCallback.SplitDelegations[0].Validator, "delegate callback validator")
		suite.Require().Equal(sdkmath.NewInt(1000000), delegateCallback.SplitDelegations[0].Amount, "delegate callback amount")
	}
}

// Stores undelegate callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (suite *UpgradeTestSuite) SetupOldUndelegateCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	undelegateValidator := "uval"
	undelegateCallback := oldstakeibctypes.UndelegateCallback{
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{{Validator: undelegateValidator, Amount: uint64(3000000)}},
	}
	undelegateCallbackArgs, err := proto.Marshal(&undelegateCallback)
	suite.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	undelegateKey := "undelegate_key"
	undelegateCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  undelegateKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Undelegate,
		CallbackArgs: undelegateCallbackArgs,
	}
	undelegateCallbackDataBz, err := codec.Marshal(&undelegateCallbackData)
	suite.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(undelegateKey), undelegateCallbackDataBz)

	// Check undelegate callback args after the migration
	return func() {
		undelegateCallbackData, found := suite.App.IcacallbacksKeeper.GetCallbackData(suite.Ctx, undelegateKey)
		suite.Require().True(found, "found undelegate callback data")

		var undelegateCallback stakeibctypes.UndelegateCallback
		err = proto.Unmarshal(undelegateCallbackData.CallbackArgs, &undelegateCallback)
		suite.Require().NoError(err, "unmarshaling undelegate callback args should not error")

		suite.Require().Equal(undelegateValidator, undelegateCallback.SplitDelegations[0].Validator, "undelegate callback validator")
		suite.Require().Equal(sdkmath.NewInt(3000000), undelegateCallback.SplitDelegations[0].Amount, "undelegate callback amount")
	}
}

// Stores rebalance callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (suite *UpgradeTestSuite) SetupOldRebalanceCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	rebalanceValidator := "rval"
	rebalanceCallback := oldstakeibctypes.RebalanceCallback{
		Rebalancings: []*oldstakeibctypes.Rebalancing{
			{SrcValidator: rebalanceValidator, Amt: uint64(2000000)},
		},
	}
	rebalanceCallbackArgs, err := proto.Marshal(&rebalanceCallback)
	suite.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	rebalanceKey := "rebalance_key"
	rebalanceCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  rebalanceKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Rebalance,
		CallbackArgs: rebalanceCallbackArgs,
	}
	rebalanceCallbackDataBz, err := codec.Marshal(&rebalanceCallbackData)
	suite.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(rebalanceKey), rebalanceCallbackDataBz)

	// Check rebalance callback args after the migration
	return func() {
		rebalanceCallbackData, found := suite.App.IcacallbacksKeeper.GetCallbackData(suite.Ctx, rebalanceKey)
		suite.Require().True(found, "found rebalance callback data")

		var rebalanceCallback stakeibctypes.RebalanceCallback
		err = proto.Unmarshal(rebalanceCallbackData.CallbackArgs, &rebalanceCallback)
		suite.Require().NoError(err, "unmarshaling rebalance callback args should not error")

		suite.Require().Equal(rebalanceValidator, rebalanceCallback.Rebalancings[0].SrcValidator, "rebalance callback validator")
		suite.Require().Equal(sdkmath.NewInt(2000000), rebalanceCallback.Rebalancings[0].Amt, "rebalance callback amount")
	}
}

// Stores claim callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should NOT get migrated
func (suite *UpgradeTestSuite) SetupOldClaimCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the callback data for the claim callback
	claimKey := "claim_key"
	oldClaimCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  claimKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Claim,
		CallbackArgs: []byte{1, 2, 3},
	}
	claimCallbackDataBz, err := codec.Marshal(&oldClaimCallbackData)
	suite.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(claimKey), claimCallbackDataBz)

	// Check rebalance callback args after the migration
	// The callback data should not have been modified
	return func() {
		newClaimCallbackData, found := suite.App.IcacallbacksKeeper.GetCallbackData(suite.Ctx, claimKey)
		suite.Require().True(found, "found claim callback data")
		suite.Require().Equal(oldClaimCallbackData, newClaimCallbackData, "claim callback data")
	}
}

// Sets up the old icacallbacks store and returns a callback function that can be used to verify
// the store migration was successful
func (suite *UpgradeTestSuite) SetupOldIcacallbackStore(codec codec.Codec) func() {
	icacallbackStore := suite.Ctx.KVStore(suite.App.GetKey(icacallbacktypes.StoreKey))
	callbackDataStore := prefix.NewStore(icacallbackStore, icacallbacktypes.KeyPrefix(icacallbacktypes.CallbackDataKeyPrefix))

	checkDelegateCallbackAfterMigration := suite.SetupOldDelegateCallback(codec, callbackDataStore)
	checkUndelegateCallbackAfterMigration := suite.SetupOldUndelegateCallback(codec, callbackDataStore)
	checkRebalanceCallbackAfterMigration := suite.SetupOldRebalanceCallback(codec, callbackDataStore)
	checkClaimCallbackAfterMigration := suite.SetupOldClaimCallback(codec, callbackDataStore)

	// Callback to check store after migration
	return func() {
		checkDelegateCallbackAfterMigration()
		checkUndelegateCallbackAfterMigration()
		checkRebalanceCallbackAfterMigration()
		checkClaimCallbackAfterMigration()
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
		suite.Require().Equal(depositRecord.Amount, sdkmath.NewInt(1000000), "deposit record amount")
		suite.Require().Equal(depositRecord.Status, recordtypes.DepositRecord_DELEGATION_QUEUE, "deposit record status")
		suite.Require().Equal(depositRecord.Source, recordtypes.DepositRecord_WITHDRAWAL_ICA, "deposit record source")

		userRedemptionRecord, found := suite.App.RecordsKeeper.GetUserRedemptionRecord(suite.Ctx, userRedemptionRecordId)
		suite.Require().True(found, "redemption record found")
		suite.Require().Equal(userRedemptionRecord.Id, userRedemptionRecordId, "redemption record id")
		suite.Require().Equal(userRedemptionRecord.Amount, sdkmath.NewInt(1000000), "redemption record amount")

		epochUnbondingRecord, found := suite.App.RecordsKeeper.GetEpochUnbondingRecord(suite.Ctx, epochNumber)
		suite.Require().True(found, "epoch unbonding record found")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].HostZoneId, hostZoneId, "host zone unbonding host zone id")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].NativeTokenAmount, sdkmath.NewInt(1000000), "host zone unbonding native token amount")
		suite.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount, sdkmath.NewInt(2000000), "host zone unbonding sttoken amount")
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

		suite.Require().Equal(hostZone.Validators[0].DelegationAmt, sdkmath.NewInt(1000000), "host zone validators delegation amount")
		suite.Require().Equal(hostZone.BlacklistedValidators[0].DelegationAmt, sdkmath.NewInt(2000000), "host zone blacklisted validators delegation amount")
		suite.Require().Equal(hostZone.StakedBal, sdkmath.NewInt(3000000), "host zone staked balance")
	}
}
