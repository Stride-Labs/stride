package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/Stride-Labs/stride/v33/x/stakeibc/client/cli"
)

var (
	AddValidatorsProposalHandler = govclient.NewProposalHandler(cli.CmdAddValidatorsProposal)
	ToggleLSMProposalHandler     = govclient.NewProposalHandler(cli.CmdToggleLSMProposal)
)
