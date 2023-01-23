package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	epochtypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v5/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"
)

type FastUnbondState struct {
	epochNumber         uint64
	depositRecordAmount sdkmath.Int
	hostZone            stakeibctypes.HostZone
}

type FastUnbondTestCase struct {
	user         Account
	zoneAccount  Account
	initialState FastUnbondState
	validMsg     stakeibctypes.MsgFastUnbond
}

func (s *KeeperTestSuite) SetupFastUnbond() FastUnbondTestCase {
	stakeAmount := sdkmath.NewInt(1_000_000)
	initialDepositAmount := sdkmath.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 1_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 1_000_000),
	}
	s.FundAccount(user.acc, user.atomBalance)
	s.FundAccount(user.acc, user.stAtomBalance)

	zoneAddress := stakeibctypes.NewZoneAddress(HostChainId)

	zoneAccount := Account{
		acc:           zoneAddress,
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Address:        zoneAddress.String(),
		StakedBal:      stakeAmount,
	}

	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             initialDepositAmount,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord)

	return FastUnbondTestCase{
		user:        user,
		zoneAccount: zoneAccount,
		initialState: FastUnbondState{
			epochNumber:         epochTracker.EpochNumber,
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibctypes.MsgFastUnbond{
			Creator:  user.acc.String(),
			HostZone: HostChainId,
			Amount:   stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestFastUnbond_Successful() {
	tc := s.SetupFastUnbond()
	msg := tc.validMsg

	// Validate Fast Unbonding
	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// TODO: Check math
}

func (s *KeeperTestSuite) TestFastUnbond_InvalidCreatorAddress() {
	tc := s.SetupFastUnbond()
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (s *KeeperTestSuite) TestFastUnbond_HostZoneNotFound() {
	tc := s.SetupFastUnbond()

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}

func (s *KeeperTestSuite) TestFastUnbond_RateAboveMaxThreshold() {
	tc := s.SetupFastUnbond()

	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.NewDec(100)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestFastUnbond_RedeemMoreThanStaked() {
	tc := s.SetupFastUnbond()

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_000_000_000_000)
	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("cannot unstake an amount g.t. staked balance on host zone: %v: invalid amount", invalidMsg.Amount))
}

func (s *KeeperTestSuite) TestFastUnbond_NoEpochTrackerDay() {
	tc := s.SetupFastUnbond()

	invalidMsg := tc.validMsg
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)
	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("no epoch number for epoch (%s): not found", epochtypes.STRIDE_EPOCH))
}

func (s *KeeperTestSuite) TestFastUnbond_InvalidHostAddress() {
	tc := s.SetupFastUnbond()

	// Update hostzone with invalid address
	badHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	badHostZone.Address = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().FastUnbond(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}
