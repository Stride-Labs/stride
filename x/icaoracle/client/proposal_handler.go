package client

import (
	"github.com/Stride-Labs/stride/v5/x/icaoracle/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	ToggleOracleProposalHandler = govclient.NewProposalHandler(cli.CmdToggleOracleProposal)
	RemoveOracleProposalHandler = govclient.NewProposalHandler(cli.CmdRemoveOracleProposal)
)
