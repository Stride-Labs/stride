package v35_test

import (
	"time"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	"github.com/Stride-Labs/stride/v35/utils"
)

func (s *UpgradeTestSuite) seedPendingGovProposal(expedited, voting bool) govv1.Proposal {
	s.T().Helper()
	params, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	params.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 100))
	params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 200))
	s.Require().NoError(s.App.GovKeeper.Params.Set(s.Ctx, params))

	message := &banktypes.MsgSend{
		FromAddress: v35.GovModuleAddress().String(), ToAddress: s.TestAccs[3].String(),
		Amount: sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 1)),
	}
	proposal, err := s.App.GovKeeper.SubmitProposal(s.Ctx, []sdk.Msg{message}, "", "pending", "pending",
		s.TestAccs[0], expedited)
	s.Require().NoError(err)
	deposit := proposal.GetMinDepositFromParams(params)[0]
	halfDeposit := sdk.NewCoin(deposit.Denom, deposit.Amount.QuoRaw(2))
	s.FundAccount(s.TestAccs[0], halfDeposit)
	activated, err := s.App.GovKeeper.AddDeposit(s.Ctx, proposal.Id, s.TestAccs[0], sdk.NewCoins(halfDeposit))
	s.Require().NoError(err)
	s.Require().False(activated)
	if voting {
		s.FundAccount(s.TestAccs[1], halfDeposit)
		activated, err = s.App.GovKeeper.AddDeposit(s.Ctx, proposal.Id, s.TestAccs[1], sdk.NewCoins(halfDeposit))
		s.Require().NoError(err)
		s.Require().True(activated)
		s.Require().NoError(s.App.GovKeeper.AddVote(s.Ctx, proposal.Id, s.TestAccs[0],
			govv1.WeightedVoteOptions{govv1.NewWeightedVoteOption(govv1.OptionYes, sdkmath.LegacyOneDec())}, ""))
	}
	stored, err := s.App.GovKeeper.Proposals.Get(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	return stored
}

func (s *UpgradeTestSuite) assertGovProposalRejected(proposal govv1.Proposal) {
	s.T().Helper()
	after, err := s.App.GovKeeper.Proposals.Get(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	s.Require().Equal(govv1.StatusRejected, after.Status)
	s.Require().Equal("Governance closed by v35 wind-down", after.FailedReason)
	expected := proposal
	expected.Status = after.Status
	expected.FailedReason = after.FailedReason
	s.Require().Equal(expected.String(), after.String(), "preserve proposal history")

	inactive, err := s.App.GovKeeper.InactiveProposalsQueue.Has(s.Ctx, collections.Join(*proposal.DepositEndTime, proposal.Id))
	s.Require().NoError(err)
	s.Require().False(inactive)
	if proposal.VotingEndTime != nil {
		active, err := s.App.GovKeeper.ActiveProposalsQueue.Has(s.Ctx, collections.Join(*proposal.VotingEndTime, proposal.Id))
		s.Require().NoError(err)
		s.Require().False(active)
	}
	voting, err := s.App.GovKeeper.VotingPeriodProposals.Has(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	s.Require().False(voting)
	deposits, err := s.App.GovKeeper.GetDeposits(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	s.Require().Empty(deposits)
	iterator, err := s.App.GovKeeper.Votes.Iterate(s.Ctx, collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id))
	s.Require().NoError(err)
	votes, err := iterator.Values()
	s.Require().NoError(err)
	s.Require().Empty(votes)
}

func (s *UpgradeTestSuite) TestUpgradeRejectsPendingGovProposals() {
	for _, testCase := range []struct {
		name      string
		expedited bool
	}{{name: "regular"}, {name: "expedited", expedited: true}} {
		s.Run(testCase.name, func() {
			s.Setup()
			validator, _ := s.seedValidator(81, stakingtypes.Bonded, 1)
			operator := sdk.AccAddress(validator)
			s.TestAccs[0] = operator
			s.delegate(operator, validator, 5_000_000_000)
			proposal := s.seedPendingGovProposal(testCase.expedited, true)
			s.FundModuleAccount(govtypes.ModuleName, sdk.NewInt64Coin(utils.BaseStrideDenom, 1))

			// A failed operator undelegation leaves enough voting power to pass this proposal.
			s.Require().NoError(s.App.DistrKeeper.DeleteDelegatorStartingInfo(s.Ctx, validator, operator))
			s.ConfirmUpgradeSucceeded(v35.UpgradeName)
			s.Require().True(s.hasDelegation(operator, validator))
			s.assertGovProposalRejected(proposal)
			refund := int64(50)
			if testCase.expedited {
				refund = 100
			}
			for _, depositor := range s.TestAccs[:2] {
				s.Require().Equal(sdkmath.NewInt(refund), s.App.BankKeeper.GetBalance(s.Ctx, depositor, utils.BaseStrideDenom).Amount)
			}

			// Expiry must neither execute the bank transfer nor convert an expedited proposal.
			s.Ctx = s.Ctx.WithBlockTime(proposal.VotingStartTime.Add(14 * 24 * time.Hour))
			s.Require().NoError(gov.EndBlocker(s.Ctx, &s.App.GovKeeper))
			s.assertGovProposalRejected(proposal)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[3], utils.BaseStrideDenom).IsZero())
			s.Require().ErrorIs(s.App.GovKeeper.AddVote(s.Ctx, proposal.Id, operator,
				govv1.WeightedVoteOptions{govv1.NewWeightedVoteOption(govv1.OptionYes, sdkmath.LegacyOneDec())}, ""), govtypes.ErrInactiveProposal)
			_, err := s.App.GovKeeper.AddDeposit(s.Ctx, proposal.Id, operator, sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 1)))
			s.Require().ErrorIs(err, govtypes.ErrInactiveProposal)
		})
	}
}

func (s *UpgradeTestSuite) TestUpgradeRejectsAllPendingProposalsAndPreservesFinishedHistory() {
	proposal := s.seedPendingGovProposal(false, false)
	regular := s.seedPendingGovProposal(false, true)
	expedited := s.seedPendingGovProposal(true, true)
	for _, status := range []govv1.ProposalStatus{govv1.StatusPassed, govv1.StatusRejected, govv1.StatusFailed} {
		finished := proposal
		finished.Id = 100 + uint64(status)
		finished.Status = status
		finished.FailedReason = "historical result"
		s.Require().NoError(s.App.GovKeeper.SetProposal(s.Ctx, finished))
	}

	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
	for _, pending := range []govv1.Proposal{proposal, regular, expedited} {
		s.assertGovProposalRejected(pending)
	}
	s.Require().NoError(v35.RejectPendingGovProposals(s.Ctx, s.App.GovKeeper), "repeated cleanup must not refund twice")
	s.Require().Equal(sdkmath.NewInt(200), s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], utils.BaseStrideDenom).Amount)
	s.Require().Equal(sdkmath.NewInt(150), s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], utils.BaseStrideDenom).Amount)
	for _, status := range []govv1.ProposalStatus{govv1.StatusPassed, govv1.StatusRejected, govv1.StatusFailed} {
		finished, err := s.App.GovKeeper.Proposals.Get(s.Ctx, 100+uint64(status))
		s.Require().NoError(err)
		expected := proposal
		expected.Id = finished.Id
		expected.Status = status
		expected.FailedReason = "historical result"
		s.Require().Equal(expected.String(), finished.String())
	}
	s.Ctx = s.Ctx.WithBlockTime(proposal.DepositEndTime.Add(time.Second))
	s.Require().NoError(gov.EndBlocker(s.Ctx, &s.App.GovKeeper))
	s.assertGovProposalRejected(proposal)
}

func (s *UpgradeTestSuite) TestRejectPendingGovProposals_RefundFailure() {
	proposal := s.seedPendingGovProposal(false, true)
	s.Require().NoError(s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, govtypes.ModuleName, s.TestAccs[3], proposal.TotalDeposit))

	err := v35.RejectPendingGovProposals(s.Ctx, s.App.GovKeeper)
	s.Require().ErrorContains(err, "refund proposal")
	after, err := s.App.GovKeeper.Proposals.Get(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	s.Require().Equal(proposal.String(), after.String(), "a failed refund must not be reported as rejection")
	active, err := s.App.GovKeeper.ActiveProposalsQueue.Has(s.Ctx, collections.Join(*proposal.VotingEndTime, proposal.Id))
	s.Require().NoError(err)
	s.Require().True(active)
	deposits, err := s.App.GovKeeper.GetDeposits(s.Ctx, proposal.Id)
	s.Require().NoError(err)
	s.Require().Len(deposits, 2)
}
