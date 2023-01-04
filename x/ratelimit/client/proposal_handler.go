package client

import (
	"github.com/Stride-Labs/stride/v4/x/ratelimit/client/cli"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddRateLimitProposalHandler    = govclient.NewProposalHandler(cli.CmdAddRateLimitProposal, rest.ProposalAddRateLimitRESTHandler)
	UpdateRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdUpdateRateLimitProposal, rest.ProposalUpdateRateLimitRESTHandler)
	RemoveRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdRemoveRateLimitProposal, rest.ProposalRemoveRateLimitRESTHandler)
	ResetRateLimitProposalHandler  = govclient.NewProposalHandler(cli.CmdResetRateLimitProposal, rest.ProposalResetRateLimitRESTHandler)
)
