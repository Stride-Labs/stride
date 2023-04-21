package v5_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	upgradev5 "github.com/Stride-Labs/stride/v9/app/upgrades/v5"
	oldclaimtypes "github.com/Stride-Labs/stride/v9/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	recordkeeper "github.com/Stride-Labs/stride/v9/x/records/keeper"
	oldrecordtypes "github.com/Stride-Labs/stride/v9/x/records/migrations/v2/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	oldstakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Setup stores for migrated modules
	codec := app.MakeEncodingConfig().Marshaler
	checkClaimStoreAfterMigration := s.SetupOldClaimStore(codec)
	checkIcacallbackStoreAfterMigration := s.SetupOldIcacallbackStore(codec)
	checkRecordStoreAfterMigration := s.SetupOldRecordStore(codec)
	checkStakeibcStoreAfterMigration := s.SetupOldStakeibcStore(codec)

	// Setup store for stale query and max slash percent param
	checkStaleQueryRemoval := s.SetupRemoveStaleQuery()
	checkMaxSlashParamAdded := s.SetupAddMaxSlashPercentParam()

	// Run upgrade
	s.ConfirmUpgradeSucceededs("v5", dummyUpgradeHeight)

	// Confirm store migrations were successful
	checkClaimStoreAfterMigration()
	checkIcacallbackStoreAfterMigration()
	checkRecordStoreAfterMigration()
	checkStakeibcStoreAfterMigration()

	// Confirm query was removed and max slash percent parameter was added
	checkStaleQueryRemoval()
	checkMaxSlashParamAdded()
}

// Sets up the old claim store and returns a callback function that can be used to verify
// the store migration was successful
func (s *UpgradeTestSuite) SetupOldClaimStore(codec codec.Codec) func() {
	claimStore := s.Ctx.KVStore(s.App.GetKey(claimtypes.StoreKey))

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
	s.Require().NoError(err)
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)

	// Callback to check claim store after migration
	return func() {
		claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
		s.Require().NoError(err, "no error expected when getting claims")
		s.Require().Equal(claimParams.Airdrops[0].AirdropIdentifier, airdropId, "airdrop identifier")
		s.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdkmath.NewInt(1000000), "claimed so far")
	}
}

// Stores delegate callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (s *UpgradeTestSuite) SetupOldDelegateCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	delegateValidator := "dval"
	delegateCallback := oldstakeibctypes.DelegateCallback{
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{
			{Validator: delegateValidator, Amount: uint64(1000000)},
		},
	}
	delegateCallbackArgs, err := proto.Marshal(&delegateCallback)
	s.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	delegateKey := "delegate_key"
	delegateCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  delegateKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Delegate,
		CallbackArgs: delegateCallbackArgs,
	}
	delegateCallbackDataBz, err := codec.Marshal(&delegateCallbackData)
	s.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(delegateKey), delegateCallbackDataBz)

	// Check delegate callback args after the migration
	return func() {
		delegateCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, delegateKey)
		s.Require().True(found, "found delegate callback data")

		var delegateCallback stakeibctypes.DelegateCallback
		err := proto.Unmarshal(delegateCallbackData.CallbackArgs, &delegateCallback)
		s.Require().NoError(err, "unmarshaling delegate callback args should not error")

		s.Require().Equal(delegateValidator, delegateCallback.SplitDelegations[0].Validator, "delegate callback validator")
		s.Require().Equal(sdkmath.NewInt(1000000), delegateCallback.SplitDelegations[0].Amount, "delegate callback amount")
	}
}

// Stores undelegate callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (s *UpgradeTestSuite) SetupOldUndelegateCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	undelegateValidator := "uval"
	undelegateCallback := oldstakeibctypes.UndelegateCallback{
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{{Validator: undelegateValidator, Amount: uint64(3000000)}},
	}
	undelegateCallbackArgs, err := proto.Marshal(&undelegateCallback)
	s.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	undelegateKey := "undelegate_key"
	undelegateCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  undelegateKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Undelegate,
		CallbackArgs: undelegateCallbackArgs,
	}
	undelegateCallbackDataBz, err := codec.Marshal(&undelegateCallbackData)
	s.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(undelegateKey), undelegateCallbackDataBz)

	// Check undelegate callback args after the migration
	return func() {
		undelegateCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, undelegateKey)
		s.Require().True(found, "found undelegate callback data")

		var undelegateCallback stakeibctypes.UndelegateCallback
		err = proto.Unmarshal(undelegateCallbackData.CallbackArgs, &undelegateCallback)
		s.Require().NoError(err, "unmarshaling undelegate callback args should not error")

		s.Require().Equal(undelegateValidator, undelegateCallback.SplitDelegations[0].Validator, "undelegate callback validator")
		s.Require().Equal(sdkmath.NewInt(3000000), undelegateCallback.SplitDelegations[0].Amount, "undelegate callback amount")
	}
}

// Stores rebalance callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should get migrated
func (s *UpgradeTestSuite) SetupOldRebalanceCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the marshalled callback args
	rebalanceValidator := "rval"
	rebalanceCallback := oldstakeibctypes.RebalanceCallback{
		Rebalancings: []*oldstakeibctypes.Rebalancing{
			{SrcValidator: rebalanceValidator, Amt: uint64(2000000)},
		},
	}
	rebalanceCallbackArgs, err := proto.Marshal(&rebalanceCallback)
	s.Require().NoError(err)

	// Create the callback data (which has callback args as an attribute)
	rebalanceKey := "rebalance_key"
	rebalanceCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  rebalanceKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Rebalance,
		CallbackArgs: rebalanceCallbackArgs,
	}
	rebalanceCallbackDataBz, err := codec.Marshal(&rebalanceCallbackData)
	s.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(rebalanceKey), rebalanceCallbackDataBz)

	// Check rebalance callback args after the migration
	return func() {
		rebalanceCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, rebalanceKey)
		s.Require().True(found, "found rebalance callback data")

		var rebalanceCallback stakeibctypes.RebalanceCallback
		err = proto.Unmarshal(rebalanceCallbackData.CallbackArgs, &rebalanceCallback)
		s.Require().NoError(err, "unmarshaling rebalance callback args should not error")

		s.Require().Equal(rebalanceValidator, rebalanceCallback.Rebalancings[0].SrcValidator, "rebalance callback validator")
		s.Require().Equal(sdkmath.NewInt(2000000), rebalanceCallback.Rebalancings[0].Amt, "rebalance callback amount")
	}
}

// Stores claim callback args in the icacallback store and returns a function used to verify
// the store was migrated successfully
// The callback args should NOT get migrated
func (s *UpgradeTestSuite) SetupOldClaimCallback(codec codec.Codec, callbackDataStore sdk.KVStore) func() {
	// Create the callback data for the claim callback
	claimKey := "claim_key"
	oldClaimCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  claimKey,
		CallbackId:   stakeibckeeper.ICACallbackID_Claim,
		CallbackArgs: []byte{1, 2, 3},
	}
	claimCallbackDataBz, err := codec.Marshal(&oldClaimCallbackData)
	s.Require().NoError(err)

	// Store the callback data
	callbackDataStore.Set(icacallbacktypes.CallbackDataKey(claimKey), claimCallbackDataBz)

	// Check rebalance callback args after the migration
	// The callback data should not have been modified
	return func() {
		newClaimCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, claimKey)
		s.Require().True(found, "found claim callback data")
		s.Require().Equal(oldClaimCallbackData, newClaimCallbackData, "claim callback data")
	}
}

// Sets up the old icacallbacks store and returns a callback function that can be used to verify
// the store migration was successful
func (s *UpgradeTestSuite) SetupOldIcacallbackStore(codec codec.Codec) func() {
	icacallbackStore := s.Ctx.KVStore(s.App.GetKey(icacallbacktypes.StoreKey))
	callbackDataStore := prefix.NewStore(icacallbackStore, icacallbacktypes.KeyPrefix(icacallbacktypes.CallbackDataKeyPrefix))

	checkDelegateCallbackAfterMigration := s.SetupOldDelegateCallback(codec, callbackDataStore)
	checkUndelegateCallbackAfterMigration := s.SetupOldUndelegateCallback(codec, callbackDataStore)
	checkRebalanceCallbackAfterMigration := s.SetupOldRebalanceCallback(codec, callbackDataStore)
	checkClaimCallbackAfterMigration := s.SetupOldClaimCallback(codec, callbackDataStore)

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
func (s *UpgradeTestSuite) SetupOldRecordStore(codec codec.Codec) func() {
	recordStore := s.Ctx.KVStore(s.App.GetKey(recordtypes.StoreKey))

	// set old deposit record
	depositRecordId := uint64(1)
	depositRecord := oldrecordtypes.DepositRecord{
		Id:     depositRecordId,
		Amount: int64(1000000),
		Status: oldrecordtypes.DepositRecord_DELEGATION_QUEUE,
		Source: oldrecordtypes.DepositRecord_WITHDRAWAL_ICA,
	}
	depositBz, err := codec.Marshal(&depositRecord)
	s.Require().NoError(err)

	depositRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.DepositRecordKey))
	depositRecordStore.Set(recordkeeper.GetDepositRecordIDBytes(depositRecord.Id), depositBz)

	// set old user redemption record
	userRedemptionRecordId := "1"
	userRedemptionRecord := oldrecordtypes.UserRedemptionRecord{
		Id:     "1",
		Amount: uint64(1000000),
	}
	userRedemptionBz, err := codec.Marshal(&userRedemptionRecord)
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	epochUnbondingRecordStore := prefix.NewStore(recordStore, recordtypes.KeyPrefix(recordtypes.EpochUnbondingRecordKey))
	epochUnbondingRecordStore.Set(recordkeeper.GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.EpochNumber), epochUnbondingBz)

	// Callback to check record store after migration
	return func() {
		depositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, depositRecordId)
		s.Require().True(found, "deposit record found")
		s.Require().Equal(depositRecord.Id, depositRecordId, "deposit record id")
		s.Require().Equal(depositRecord.Amount, sdkmath.NewInt(1000000), "deposit record amount")
		s.Require().Equal(depositRecord.Status, recordtypes.DepositRecord_DELEGATION_QUEUE, "deposit record status")
		s.Require().Equal(depositRecord.Source, recordtypes.DepositRecord_WITHDRAWAL_ICA, "deposit record source")

		userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, userRedemptionRecordId)
		s.Require().True(found, "redemption record found")
		s.Require().Equal(userRedemptionRecord.Id, userRedemptionRecordId, "redemption record id")
		s.Require().Equal(userRedemptionRecord.Amount, sdkmath.NewInt(1000000), "redemption record amount")

		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, epochNumber)
		s.Require().True(found, "epoch unbonding record found")
		s.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].HostZoneId, hostZoneId, "host zone unbonding host zone id")
		s.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].NativeTokenAmount, sdkmath.NewInt(1000000), "host zone unbonding native token amount")
		s.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount, sdkmath.NewInt(2000000), "host zone unbonding sttoken amount")
		s.Require().Equal(epochUnbondingRecord.HostZoneUnbondings[0].Status, recordtypes.HostZoneUnbonding_CLAIMABLE, "host zone unbonding status")
	}
}

func (s *UpgradeTestSuite) SetupOldStakeibcStore(codec codec.Codec) func() {
	stakeibcStore := s.Ctx.KVStore(s.App.GetKey(stakeibctypes.StoreKey))

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
	s.Require().NoError(err)

	hostzoneStore := prefix.NewStore(stakeibcStore, stakeibctypes.KeyPrefix(stakeibctypes.HostZoneKey))
	hostzoneStore.Set([]byte(hostZone.ChainId), hostZoneBz)

	// Callback to check stakeibc store after migration
	return func() {
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, hostZoneId)
		s.Require().True(found, "host zone found")
		s.Require().Equal(hostZone.ChainId, hostZoneId, "host zone chain id")

		s.Require().Equal(hostZone.DelegationAccount.Address, delegationAddress, "delegation address")
		s.Require().Equal(hostZone.RedemptionAccount.Address, redemptionAddress, "redemption address")

		s.Require().Equal(hostZone.DelegationAccount.Target, stakeibctypes.ICAAccountType_DELEGATION, "delegation target")
		s.Require().Equal(hostZone.RedemptionAccount.Target, stakeibctypes.ICAAccountType_REDEMPTION, "redemption target")

		s.Require().Nil(hostZone.FeeAccount, "fee account")
		s.Require().Nil(hostZone.WithdrawalAccount, "withdrawal account")

		s.Require().Equal(hostZone.Validators[0].DelegationAmt, sdkmath.NewInt(1000000), "host zone validators delegation amount")
		s.Require().Equal(hostZone.BlacklistedValidators[0].DelegationAmt, sdkmath.NewInt(2000000), "host zone blacklisted validators delegation amount")
		s.Require().Equal(hostZone.StakedBal, sdkmath.NewInt(3000000), "host zone staked balance")
	}
}

// Adds the stale query to the store and returns a callback to check
// that it was successfully removed after the upgrade
func (s *UpgradeTestSuite) SetupRemoveStaleQuery() func() {
	// Add the stale query
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{Id: upgradev5.StaleQueryId})
	query, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, upgradev5.StaleQueryId)

	// Confirm it was added successfully
	s.Require().True(found, "stale query successfully added to store")
	s.Require().Equal(upgradev5.StaleQueryId, query.Id, "query id")

	// Callback to check that the query was successfully removed
	return func() {
		_, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, upgradev5.StaleQueryId)
		s.Require().False(found)
	}
}

// Changes the SafetyMaxSlashPercent parameter to 0 and returns a callback to check
// the the parameter was successfully updated back to it's default value after the upgrade
func (s *UpgradeTestSuite) SetupAddMaxSlashPercentParam() func() {
	// Set the max slash percent to 0
	stakeibcParamStore := s.App.GetSubspace(stakeibctypes.ModuleName)
	stakeibcParamStore.Set(s.Ctx, stakeibctypes.KeySafetyMaxSlashPercent, uint64(0))

	// Confirm it was updated
	maxSlashPercent := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeySafetyMaxSlashPercent)
	s.Require().Equal(uint64(0), maxSlashPercent, "max slash percent should be 0")

	// Callback to check that the parameter was added to the store
	return func() {
		// Confirm MaxSlashPercent was added with the default value
		maxSlashPercent := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeySafetyMaxSlashPercent)
		s.Require().Equal(stakeibctypes.DefaultSafetyMaxSlashPercent, maxSlashPercent, "max slash percent should be default")
	}
}
