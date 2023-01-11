package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupSubmitHostZoneUnbondingMsg(hostZoneUnbonding recordtypes.HostZoneUnbonding) {
	s.CreateICAChannel("GAIA.DELEGATION")

	gaiaValAddr := "cosmos_VALIDATOR"
	gaiaDelegationAddr := "cosmos_DELEGATION"
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       gaiaValAddr,
			DelegationAmt: sdk.NewInt(5_000_000),
			Weight:        uint64(10),
		},
		{
			Address:       gaiaValAddr + "2",
			DelegationAmt: sdk.NewInt(3_000_000),
			Weight:        uint64(6),
		},
	}
	gaiaDelegationAccount := stakeibc.ICAAccount{
		Address: gaiaDelegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:            HostChainId,
			HostDenom:          Atom,
			Bech32Prefix:       GaiaPrefix,
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
			StakedBal:          sdk.NewInt(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	// list of epoch unbonding records
	default_unbonding := []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding}

	for _, epochNumber := range []uint64{5} {
		epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
			EpochNumber:        epochNumber,
			HostZoneUnbondings: default_unbonding,
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibc.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        12,
		NextEpochStartTime: uint64(2661750006000000000), // arbitrary time in the future, year 2056 I believe
		Duration:           uint64(1000000000000),       // 16 min 40 sec
	})

}

func (s *KeeperTestSuite) TestSubmitHostZoneUnbondingMsg_Successful() {
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:        HostChainId,
		StTokenAmount:     sdk.NewInt(1_900_000),
		NativeTokenAmount: sdk.NewInt(1_000_000),
		Denom:             Atom,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	s.SetupSubmitHostZoneUnbondingMsg(hostZoneUnbonding)
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, hostZoneUnbonding.HostZoneId)
	s.Require().True(found)
	msgs, totalAmtToUnbond, marshalledCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, hostZone)
	s.Require().NoError(err)
	err = s.App.StakeibcKeeper.SubmitHostZoneUnbondingMsg(s.Ctx, msgs, totalAmtToUnbond, marshalledCallbackArgs, hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestSubmitHostZoneUnbondingMsg_NoMsgsToSubmit() {
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:        HostChainId,
		StTokenAmount:     sdk.NewInt(1_900_000),
		NativeTokenAmount: sdk.ZeroInt(),
		Denom:             Atom,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	s.SetupSubmitHostZoneUnbondingMsg(hostZoneUnbonding)
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, hostZoneUnbonding.HostZoneId)
	s.Require().True(found)
	msgs, totalAmtToUnbond, marshalledCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, hostZone)
	s.Require().NoError(err)
	err = s.App.StakeibcKeeper.SubmitHostZoneUnbondingMsg(s.Ctx, msgs, totalAmtToUnbond, marshalledCallbackArgs, hostZone)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSubmitHostZoneUnbondingMsg_ErrorSubmittingUnbonding() {
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:        HostChainId,
		StTokenAmount:     sdk.NewInt(1_900_000),
		NativeTokenAmount: sdk.NewInt(1_000_000),
		Denom:             Atom,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	s.SetupSubmitHostZoneUnbondingMsg(hostZoneUnbonding)
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, hostZoneUnbonding.HostZoneId)
	s.Require().True(found)
	msgs, totalAmtToUnbond, marshalledCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, hostZone)
	s.Require().NoError(err)
	hostZone.ConnectionId = "InvalidConnectionId"
	err = s.App.StakeibcKeeper.SubmitHostZoneUnbondingMsg(s.Ctx, msgs, totalAmtToUnbond, marshalledCallbackArgs, hostZone)
	s.Require().Error(err)
}
