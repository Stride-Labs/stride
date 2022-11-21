package client

import (
	"github.com/Stride-Labs/stride/v3/x/stakeibc/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddValidatorProposalHandler = govclient.NewProposalHandler(cli.CmdAddValidatorProposal)
)
