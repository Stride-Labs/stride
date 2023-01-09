package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	_ "github.com/stretchr/testify/suite"
)

type DelegateOnHostTestcase struct {
	hostZone      stakeibctypes.HostZone
	amt           sdk.Coin
	depositRecord []recordstypes.DepositRecord
}

func (s *KeeperTestSuite) SetupDelegationOnHost() DelegateOnHostTestcase {
	//Set Deposit Records of type Delegate
	stakedBal := sdk.NewInt(5_000)
	AddressDelegateAccount := "1"
	DelegateAccount := &types.ICAAccount{Address: AddressDelegateAccount, Target: types.ICAAccountType_DELEGATION}
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	DepositRecordDelegate := []recordstypes.DepositRecord{
		{
			Id:         1,
			Amount:     sdk.NewInt(1000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
		{
			Id:         2,
			Amount:     sdk.NewInt(3000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
	}
	stakeAmount := sdk.NewInt(1_000_000)
	stakeCoin := sdk.NewCoin(Atom, stakeAmount)

	//  define the host zone with stakedBal and validators with staked amounts
	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	amtVal1 := sdk.NewInt(1_000_000)
	amtVal2 := sdk.NewInt(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)

	validators := []*stakeibctypes.Validator{
		{
			Address:       hostVal1Addr,
			DelegationAmt: amtVal1,
			Weight:        wgtVal1,
		},
		{
			Address:       hostVal2Addr,
			DelegationAmt: amtVal2,
			Weight:        wgtVal2,
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		StakedBal:         stakedBal,
		Validators:        validators,
		DelegationAccount: DelegateAccount,
		ConnectionId:      ibctesting.FirstConnectionID,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	//set Stride epoch
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().UnixNano()),
		Duration:           uint64(2),
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	return DelegateOnHostTestcase{
		hostZone:      hostZone,
		amt:           stakeCoin,
		depositRecord: DepositRecordDelegate,
	}
}

func (s *KeeperTestSuite) TestDelegateOnHost_Successful() {
	tc := s.SetupDelegationOnHost()
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])
	DelegateDepositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.depositRecord[0].Id)
	s.Require().NoError(err)
	s.Require().Equal(DelegateDepositRecord.Status, recordstypes.DepositRecord_DELEGATION_IN_PROGRESS, "record must be status of Delegation in progress")
	s.Require().Equal(found, true, "record must be found")
}

func (s *KeeperTestSuite) TestDelegateOnHost_ConnectionIDNotFound() {
	tc := s.SetupDelegationOnHost()
	//change Hostzone's chainId so no connection should be found
	tc.hostZone.ChainId = "superGAIA"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `icacontroller-superGAIA.DELEGATION has no associated connection: `
	error += `invalid chain-id`
	s.EqualError(err, error, "Hostzone's chainId has been changed so this should fail")
}

func (s *KeeperTestSuite) TestDelegateOnHost_InvalidDelegationAccount() {
	tc := s.SetupDelegationOnHost()
	//change Hostzone's Delegation accounts so no accounts should be found
	AddressDelegateAccount := ""
	DelegateAccount := &types.ICAAccount{Address: AddressDelegateAccount, Target: types.ICAAccountType_DELEGATION}
	tc.hostZone.DelegationAccount = DelegateAccount
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `Invalid delegation account: `
	error += `invalid address`
	s.EqualError(err, error, "Hostzone's Delegation accounts has been changed so this should fail")
}

func (s *KeeperTestSuite) TestDelegationOnHost_ErrorOnValidators() {
	tc := s.SetupDelegationOnHost()
	//change Hostzone's Validators to empty so error should be returnd
	validators := []*stakeibctypes.Validator{}
	tc.hostZone.Validators = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err := s.App.StakeibcKeeper.DelegateOnHost(s.Ctx, tc.hostZone, tc.amt, tc.depositRecord[0])

	error := `no non-zero validator weights`
	s.EqualError(err, error, "Hostzone's Validators has been cleared so this should fail")
}
