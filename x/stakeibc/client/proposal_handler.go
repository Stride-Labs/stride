package client

import (
	"github.com/Stride-Labs/stride/x/stakeibc/client/cli"
	"github.com/Stride-Labs/stride/x/stakeibc/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddValidatorProposalHandler = govclient.NewProposalHandler(cli.CmdAddValidatorProposal, rest.ProposalAddValidatorRESTHandler)
)
