package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewProposal(govProposal govtypes.Proposal, hostZoneId string) (Proposal, error) {
	p := Proposal{
		HostZoneId:  hostZoneId,		
		GovProposal: govProposal,
	}

	return p, nil
}

// Proposals is an array of proposal
type Proposals []Proposal