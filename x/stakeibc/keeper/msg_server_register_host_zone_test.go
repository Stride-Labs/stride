package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RegisterHostZoneTestCase struct {
	validMsg                   stakeibc.MsgRegisterHostZone
	epochUnbondingRecordNumber uint64
	unbondingFrequency         uint64
	defaultRedemptionRate      sdk.Dec
	atomHostZoneChainId        string
}

func (s *KeeperTestSuite) SetupRegisterHostZone() RegisterHostZoneTestCase {
	epochUnbondingRecordNumber := uint64(3)
	unbondingFrequency := uint64(3)
	defaultRedemptionRate := sdk.NewDec(1)
	atomHostZoneChainId := "GAIA"

	s.CreateTransferChannel(HostChainId)

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), stakeibc.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     epochUnbondingRecordNumber,
	})

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochUnbondingRecordNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)

	defaultMsg := stakeibc.MsgRegisterHostZone{
		ConnectionId:       ibctesting.FirstConnectionID,
		Bech32Prefix:       GaiaPrefix,
		HostDenom:          Atom,
		IbcDenom:           IbcAtom,
		TransferChannelId:  ibctesting.FirstChannelID,
		UnbondingFrequency: unbondingFrequency,
	}

	return RegisterHostZoneTestCase{
		validMsg:                   defaultMsg,
		epochUnbondingRecordNumber: epochUnbondingRecordNumber,
		unbondingFrequency:         unbondingFrequency,
		defaultRedemptionRate:      defaultRedemptionRate,
		atomHostZoneChainId:        atomHostZoneChainId,
	}
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success() {
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// Register host zone
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "able to successfully register host zone")

	// Confirm host zone unbonding was added
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.defaultRedemptionRate, hostZone.RedemptionRate, "redemption rate set to default: 1")
	s.Require().Equal(tc.defaultRedemptionRate, hostZone.LastRedemptionRate, "last redemption rate set to default: 1")
	s.Require().Equal(tc.unbondingFrequency, hostZone.UnbondingFrequency, "unbonding frequency set to default: 3")

	// Confirm host zone unbonding record was created
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), tc.epochUnbondingRecordNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Len(epochUnbondingRecord.HostZoneUnbondings, 1, "host zone unbonding record has one entry")

	// Confirm host zone unbonding was added
	hostZoneUnbonding := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(HostChainId, hostZoneUnbonding.HostZoneId, "host zone unbonding set for this host zone")
	s.Require().Equal(uint64(0), hostZoneUnbonding.NativeTokenAmount, "host zone unbonding set to 0 tokens")
	s.Require().Equal(recordstypes.HostZoneUnbonding_BONDED, hostZoneUnbonding.Status, "host zone unbonding set to bonded")

}

func (s *KeeperTestSuite) TestRegisterHostZone_WrongConnectionid() {
	// tests for a failure if we register with an invalid connection id
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg
	msg.ConnectionId = "connection-10" // an invalid connection ID

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := fmt.Sprintf("unable to obtain chain id: invalid connection id, %s not found", msg.ConnectionId)
	s.Require().EqualError(err, expectedErrMsg, "expected error when registering with an invalid connection id")
}

func (s *KeeperTestSuite) TestRegisterHostZone_RegisterTwiceFails() {
	// tests for a failure if we register the same host zone twice
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "able to successfully register host zone once")

	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := fmt.Sprintf("invalid chain id, zone for %s already registered", tc.atomHostZoneChainId)
	s.Require().EqualError(err, expectedErrMsg, "registering host zone twice should fail")
}
