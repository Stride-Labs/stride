package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RegisterHostZoneTestCase struct {
	validMsg stakeibc.MsgRegisterHostZone
}

func (s *KeeperTestSuite) SetupRegisterHostZone() RegisterHostZoneTestCase {
	s.CreateTransferChannel(HostChainId)

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), stakeibc.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        3,
		NextEpochStartTime: uint64(2661750006000000000), // arbitrary time in the future, year 2056 I believe
		Duration:           uint64(1000000000000),       // 16 min 40 sec
	})

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        3,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)

	defaultMsg := stakeibc.MsgRegisterHostZone{
		ConnectionId:       ibctesting.FirstConnectionID,
		Bech32Prefix:       GaiaPrefix,
		HostDenom:          Atom,
		IbcDenom:           IbcAtom,
		Creator:            "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq",
		TransferChannelId:  ibctesting.FirstChannelID,
		UnbondingFrequency: 3,
	}

	return RegisterHostZoneTestCase{
		validMsg: defaultMsg,
	}
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success() {
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "able to successfully register host zone")
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), 3)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Len(epochUnbondingRecord.HostZoneUnbondings, 1, "host zone unbonding record has one entry")
	hostZoneUnbonding := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(HostChainId, hostZoneUnbonding.HostZoneId, "host zone unbonding set for this host zone")
	s.Require().Equal(uint64(0), hostZoneUnbonding.NativeTokenAmount, "host zone unbonding set to 0 tokens")
	s.Require().Equal(recordstypes.HostZoneUnbonding_BONDED, hostZoneUnbonding.Status, "host zone unbonding set to bonded")
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(sdk.NewDec(1), hostZone.RedemptionRate, "redemption rate set to 1")
	s.Require().Equal(sdk.NewDec(1), hostZone.LastRedemptionRate, "last redemption rate set to 1")
	s.Require().Equal(uint64(3), hostZone.UnbondingFrequency, "unbonding frequency set to 3")
}

func (s *KeeperTestSuite) TestRegisterHostZone_WrongConnectionid() {
	// tests for a failure if we register with an invalid connection id
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg
	msg.ConnectionId = "connection-10" // an invalid connection ID

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().Error(err, "expected error when registering with an invalid connection id")
}

func (s *KeeperTestSuite) TestRegisterHostZone_RegisterTwiceFails() {
	// tests for a failure if we register the same host zone twice
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "able to successfully register host zone once")
	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().Error(err, "registering host zone twice should fail")
}
