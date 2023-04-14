package keeper_test

import (

	//nolint:staticcheck

	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/gogo/protobuf/proto"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestDetokenizeLSMDeposit() {
	// Create the delegation ICA
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	s.CreateICAChannel(owner)
	portId, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "no error expected when formatting portId")

	// Get the ica address that was just created
	delegationICAAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portId)
	s.Require().True(found, "ICA account should have been created")
	s.Require().NotEmpty(delegationICAAddress, "ICA Address should not be empty")

	// Build the host zone and deposit (which are arguments to detokenize)
	hostZone := types.HostZone{
		ChainId: HostChainId,
		DelegationAccount: &types.ICAAccount{
			Address: delegationICAAddress,
		},
		ConnectionId: ibctesting.FirstConnectionID,
	}

	denom := "cosmosvalXXX/42"
	initalDeposit := types.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   denom,
		Amount:  sdk.NewInt(1000),
		Status:  types.DETOKENIZATION_QUEUE,
	}
	s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, initalDeposit)

	// Successfully Detokenize
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZone, initalDeposit)
	s.Require().NoError(err, "no error expected when detokenizing")

	// Confirm deposit status was updated
	finalDeposit, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, denom)
	s.Require().True(found, "deposit should have been found")
	s.Require().Equal(types.DETOKENIZATION_IN_PROGRESS.String(), finalDeposit.Status.String(), "deposit status")

	// Check callback data was stored
	allCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(allCallbackData, 1, "length of callback data")

	var callbackData types.DetokenizeSharesCallback
	err = proto.Unmarshal(allCallbackData[0].CallbackArgs, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	s.Require().Equal(initalDeposit, *callbackData.Deposit, "callback data LSM deposit")

	// Remove connection ID and re-submit - should fail
	hostZoneWithoutConnectionId := hostZone
	hostZoneWithoutConnectionId.ConnectionId = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZoneWithoutConnectionId, initalDeposit)
	s.Require().ErrorContains(err, "unable to submit detokenization ICA")

	// Remove delegation account and re-submit - should also fail
	hostZoneWithoutDelegationAccount := hostZone
	hostZoneWithoutDelegationAccount.DelegationAccount.Address = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZoneWithoutDelegationAccount, initalDeposit)
	s.Require().ErrorContains(err, "no delegation account found")
}

func (s *KeeperTestSuite) TestDetokenizeAllLSMDeposits() {
	// Create an open delegation ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	s.CreateICAChannel(owner)
	portId, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "no error expected when formatting portId")

	// Get the ica address that was just created
	delegationICAAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portId)
	s.Require().True(found, "ICA account should have been created")
	s.Require().NotEmpty(delegationICAAddress, "ICA Address should not be empty")

	// Store two host zones - one with an open Delegation channel, and one without
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		DelegationAccount: &types.ICAAccount{
			Address: delegationICAAddress,
		},
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      OsmoChainId,
		ConnectionId: "connection-2",
	})

	// For each host chain store 4 deposits
	// 2 of which are ready to be detokenized, and 2 of which are not
	expectedDepositStatus := map[string]types.LSMDepositStatus{}
	for _, chainId := range []string{HostChainId, OsmoChainId} {
		for _, startingStatus := range []types.LSMDepositStatus{types.DETOKENIZATION_QUEUE, types.TRANSFER_IN_PROGRESS} {

			for i := 0; i < 2; i++ {
				denom := fmt.Sprintf("denom-starting-in-status-%s-%d", startingStatus.String(), i)
				depositKey := fmt.Sprintf("%s-%s", chainId, denom)
				deposit := types.LSMTokenDeposit{
					ChainId: chainId,
					Denom:   denom,
					Status:  startingStatus,
				}
				s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

				// The status is only expected to change for the QUEUED records on the
				// host chain with the open delegation channel
				expectedStatus := startingStatus
				if chainId == HostChainId && startingStatus == types.DETOKENIZATION_QUEUE {
					expectedStatus = types.DETOKENIZATION_IN_PROGRESS
				}
				expectedDepositStatus[depositKey] = expectedStatus
			}
		}
	}

	// Call detokenization across all hosts
	s.App.StakeibcKeeper.DetokenizeAllLSMDeposits(s.Ctx)

	// Check that the status of the relevant records was updated
	allDeposits := s.App.StakeibcKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Len(allDeposits, 8) // 2 host zones, 2 statuses, 2 deposits = 2 * 2 * 2 = 8

	for _, deposit := range allDeposits {
		depositKey := fmt.Sprintf("%s-%s", deposit.ChainId, deposit.Denom)
		s.Require().Equal(expectedDepositStatus[depositKey].String(), deposit.Status.String(), "deposit status for %s", depositKey)
	}
}
