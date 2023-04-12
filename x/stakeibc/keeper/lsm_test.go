package keeper_test

import (

	//nolint:staticcheck

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
	hostZone.ConnectionId = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZone, initalDeposit)
	s.Require().ErrorContains(err, "unable to submit detokenization ICA")

	// Remove delegation account and re-submit - should also fail
	hostZone.DelegationAccount.Address = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZone, initalDeposit)
	s.Require().ErrorContains(err, "no delegation account found")
}
