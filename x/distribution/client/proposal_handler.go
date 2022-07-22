package client

import (
	"github.com/Stride-Labs/stride/x/distribution/client/cli"
	"github.com/Stride-Labs/stride/x/distribution/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// ProposalHandler is the community spend proposal handler.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
