package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v6/app/apptesting"

	strideapp "github.com/Stride-Labs/stride/v6/app"
	"github.com/Stride-Labs/stride/v6/app/ante"
)

var (
	minCoins          = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000000))
	insufficientCoins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	addrs             = []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
)

type GovAnteHandlerTestSuite struct {
	suite.Suite

	app       *strideapp.StrideApp
	ctx       sdk.Context
	clientCtx client.Context
	apptesting.AppTestHelper
}

func (s *GovAnteHandlerTestSuite) SetupTest() {
	s.Setup()
}

func TestGovSpamPreventionSuite(t *testing.T) {
	suite.Run(t, new(GovAnteHandlerTestSuite))
}

func (s *GovAnteHandlerTestSuite) TestGlobalFeeMinimumGasFeeAnteHandler() {
	// setup test
	s.SetupTest()
	tests := []struct {
		title, description string
		proposalType       string
		proposerAddr       sdk.AccAddress
		initialDeposit     sdk.Coins
		expectPass         bool
	}{
		{"Passing proposal", "the purpose of this proposal is to pass", govtypes.ProposalTypeText, addrs[0], minCoins, true},
		{"Failing proposal", "the purpose of this proposal is to fail", govtypes.ProposalTypeText, addrs[0], insufficientCoins, false},
	}

	decorator := ante.NewGovPreventSpamDecorator(s.App.AppCodec(), &s.App.GovKeeper)

	for _, tc := range tests {
		content, ok := govtypes.ContentFromProposalType(tc.title, tc.description, tc.proposalType)
		s.Require().True(ok)

		msg, err := govtypes.NewMsgSubmitProposal(
			content,
			tc.initialDeposit,
			tc.proposerAddr,
		)

		s.Require().NoError(err)

		if tc.expectPass {
			err := decorator.CheckSpamSubmitProposalMsg(s.Ctx, []sdk.Msg{msg})
			s.Require().NoError(err, "expected %v to pass", tc.title)
		} else {
			s.Require().NoError(err, "expected %v to fail", tc.title)
		}
	}
}
