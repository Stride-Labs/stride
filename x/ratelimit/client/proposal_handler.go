package client

import (
	"github.com/Stride-Labs/stride/v9/x/ratelimit/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddRateLimitProposalHandler    = govclient.NewProposalHandler(cli.CmdAddRateLimitProposal)
	UpdateRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdUpdateRateLimitProposal)
	RemoveRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdRemoveRateLimitProposal)
	ResetRateLimitProposalHandler  = govclient.NewProposalHandler(cli.CmdResetRateLimitProposal)
)
