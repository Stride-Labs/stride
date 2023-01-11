package keeper_test

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
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

func (s *KeeperTestSuite) SetupSweepAllUnbondedTokensForHostZone() SweepUnbondedTokensTestCase {
	s.CreateICAChannel("GAIA.DELEGATION")
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	gaiaDelegationAccount := stakeibc.ICAAccount{
		Address: "cosmos_DELEGATION",
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	gaiaRedemptionAccount := stakeibc.ICAAccount{
		Address: "cosmos_REDEMPTION",
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:       "osmo_VALIDATOR",
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	osmoDelegationAccount := stakeibc.ICAAccount{
		Address: "osmo_DELEGATION",
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	osmoRedemptionAccount := stakeibc.ICAAccount{
		Address: "osmo_REDEMPTION",
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:            HostChainId,
			HostDenom:          Atom,
			Bech32Prefix:       GaiaPrefix,
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
			RedemptionAccount:  &gaiaRedemptionAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
		{
			ChainId:            OsmoChainId,
			HostDenom:          Osmo,
			Bech32Prefix:       OsmoPrefix,
			UnbondingFrequency: 4,
			Validators:         osmoValidators,
			DelegationAccount:  &osmoDelegationAccount,
			RedemptionAccount:  &osmoRedemptionAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	dayEpochTracker := stakeibc.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(10)
	lightClientTime := unbondingTime + 1
	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: 2_000_000,
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: 2_000_000,
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: 5_000_000,
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: 5_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	return SweepUnbondedTokensTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
		lightClientTime:       lightClientTime,
	}
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_success() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[0], tc.epochUnbondingRecords)
	s.Require().True(success, "sweep all tokens for hostzone GAIA success")
	s.Require().Equal(int64(2_000_000), sweepAmount, "sweep all unbonded tokens (with status EXIT_TRANSFER_QUEUE) for hostone GAIA success")
	success, sweepAmount = s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().True(success, "sweep all tokens for hostzone OSMO success")
	s.Require().Equal(int64(3_000_000), sweepAmount, "sweep all unbonded tokens (with status EXIT_TRANSFER_QUEUE) for hostone OSMO success")
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_HostZoneUnbondingNotFound() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	for _, epochUnbonding := range epochUnbondingRecords {
		epochUnbonding.HostZoneUnbondings = []*recordtypes.HostZoneUnbonding{
			epochUnbonding.HostZoneUnbondings[0],
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbonding)
	}
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().True(success, "sweep all tokens for hostzone OSMO still success (even when overflow happended)")
	s.Require().Equal(int64(0), sweepAmount, "No Unbonded tokens for hostzone OSMO is sweeped (because we removed the hostzone earlier)")
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_blockTimeForHostZoneNotFound() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	tc.hostZones[1].ConnectionId = "random-connection"
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().True(success, "sweep all tokens for hostzone OSMO still success (even when failed to get blockTime)")
	s.Require().Equal(int64(0), sweepAmount, "No Unbonded tokens for hostzone OSMO is sweeped")
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_overflowSweepAmount() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	for _, epochUnbonding := range epochUnbondingRecords {
		epochUnbonding.HostZoneUnbondings[1].NativeTokenAmount = math.MaxInt64 + 1
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbonding)
	}
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().True(success, "sweep all tokens for hostzone OSMO still success (even when overflow happended)")
	s.Require().Equal(int64(0), sweepAmount, "No Unbonded tokens for hostzone OSMO is sweeped")
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_DelegationAddressNotFound() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	tc.hostZones[1].DelegationAccount = nil
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().False(success, "sweep all tokens for hostzone OSMO fail (when delegationAccount not found)")
	s.Require().Equal(int64(0), sweepAmount, "No Unbonded tokens for hostzone OSMO is sweeped")
}

func (s *KeeperTestSuite) TestSweepAllUnbondedTokensForHostZone_RedemptionAddressNotFound() {
	tc := s.SetupSweepAllUnbondedTokensForHostZone()
	tc.hostZones[1].RedemptionAccount = nil
	success, sweepAmount := s.App.StakeibcKeeper.SweepAllUnbondedTokensForHostZone(s.Ctx, tc.hostZones[1], tc.epochUnbondingRecords)
	s.Require().False(success, "sweep all tokens for hostzone OSMO fail (when redemptionAccount not found)")
	s.Require().Equal(int64(0), sweepAmount, "No Unbonded tokens for hostzone OSMO is sweeped")
}
