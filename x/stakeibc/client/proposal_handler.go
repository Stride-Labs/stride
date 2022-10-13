package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/Stride-Labs/stride/x/stakeibc/client/cli"
	"github.com/Stride-Labs/stride/x/stakeibc/client/rest"
)

// ProposalHandler is the add validator proposal handler.
var (
	AddValidatorProposalHandler    = govclient.NewProposalHandler(cli.GetCmdAddValidatorProposal, rest.ProposalAddValidatorRESTHandler)       // TODO(riley-stride) we set the REST entrypoint to nil for now. Check: does it need to be exposed?
	DeleteValidatorProposalHandler = govclient.NewProposalHandler(cli.GetCmdDeleteValidatorProposal, rest.ProposalDeleteValidatorRESTHandler) // TODO(riley-stride) we set the REST entrypoint to nil for now. Check: does it need to be exposed?
)
