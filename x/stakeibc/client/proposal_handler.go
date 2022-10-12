package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/Stride-Labs/stride/x/stakeibc/client/cli"
)

// ProposalHandler is the add validator proposal handler.
var (
	AddValidatorProposalHandler    = govclient.NewProposalHandler(cli.GetCmdAddValidatorProposal, nil)    // TODO(riley-stride) we set the REST entrypoint to nil for now. Check: does it need to be exposed?
	DeleteValidatorProposalHandler = govclient.NewProposalHandler(cli.GetCmdDeleteValidatorProposal, nil) // TODO(riley-stride) we set the REST entrypoint to nil for now. Check: does it need to be exposed?
)
