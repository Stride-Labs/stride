package client

import (
	"github.com/Stride-Labs/stride/v9/x/stakeibc/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddValidatorsProposalHandler          = govclient.NewProposalHandler(cli.CmdAddValidatorsProposal)
	DeleteValidatorsProposalHandler       = govclient.NewProposalHandler(cli.CmdDeleteValidatorsProposal)
	ChangeValidatorWeightsProposalHandler = govclient.NewProposalHandler(cli.CmdChangeValidatorWeightsProposal)
	RegisterHostZoneProposalHandler       = govclient.NewProposalHandler(cli.CmdRegisterHostZoneProposal)
)
