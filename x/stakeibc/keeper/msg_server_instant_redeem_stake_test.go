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

type InstantRedeemStakeState struct {
	epochNumber         uint64
	depositRecordAmount sdkmath.Int
	hostZone            stakeibctypes.HostZone
}

type InstantRedeemStakeTestCase struct {
	user         Account
	zoneAccount  Account
	initialState InstantRedeemStakeState
	validMsg     stakeibctypes.MsgInstantRedeemStake
}

func (s *KeeperTestSuite) SetupInstantRedeemStake() InstantRedeemStakeTestCase {
	unbondAmount := sdkmath.NewInt(1_000_000)
	initialDepositAmount := sdkmath.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 10_000_000),
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
		StakedBal:      initialDepositAmount,
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

	return InstantRedeemStakeTestCase{
		user:        user,
		zoneAccount: zoneAccount,
		initialState: InstantRedeemStakeState{
			epochNumber:         epochTracker.EpochNumber,
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibctypes.MsgInstantRedeemStake{
			Creator:  user.acc.String(),
			HostZone: HostChainId,
			Amount:   unbondAmount,
		},
	}
}

// TODO: Need to add tests for at least multiple deposit records, non 1.0 Redemption Rates, and probably some other basic scenarios.
func (s *KeeperTestSuite) TestInstantRedeemStake_Successful() {
	tc := s.SetupInstantRedeemStake()
	user := tc.user
	zoneAccount := tc.zoneAccount
	initialStAtomSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)
	msg := tc.validMsg

	// Validate Instant Redeem Stake
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// User STUATOM balance should have DECREASED by the amount unbonded
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(msg.Amount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom)
	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	// User IBC/UATOM balance should have INCREASED by the amount unbonded
	expectedUserAtomBalance := user.atomBalance.AddAmount(msg.Amount)
	actualUserAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom)
	s.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	// zoneAccount IBC/UATOM balance should have DECREASED by the size of the stake
	expectedzoneAccountAtomBalance := zoneAccount.atomBalance.SubAmount(msg.Amount)
	actualzoneAccountAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount.acc, IbcAtom)
	s.CompareCoins(expectedzoneAccountAtomBalance, actualzoneAccountAtomBalance, "zoneAccount ibc/uatom balance")
	// Bank supply of STUATOM should have DECREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.SubAmount(msg.Amount)
	actualBankSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)
	s.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")
}

func (s *KeeperTestSuite) TestInstantRedeemStake_InvalidCreatorAddress() {
	tc := s.SetupInstantRedeemStake()
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_HostZoneNotFound() {
	tc := s.SetupInstantRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}

func (s *KeeperTestSuite) TestInstantRedeemStake_RateAboveMaxThreshold() {
	tc := s.SetupInstantRedeemStake()

	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.NewDec(100)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestInstantRedeemStake_RedeemMoreThanStaked() {
	tc := s.SetupInstantRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_000_000_000_000)
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("cannot unstake an amount g.t. staked balance on host zone: %v: invalid amount", invalidMsg.Amount))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_InvalidHostAddress() {
	tc := s.SetupInstantRedeemStake()

	// Update hostzone with invalid address
	badHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	badHostZone.Address = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}
