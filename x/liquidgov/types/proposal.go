package types

import (
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewProposal(govProposal govtypesv1beta1.Proposal, hostZoneId string) (Proposal, error) {
	p := Proposal{
		GovProposal: govProposal,
		HostZoneId:  hostZoneId,
	}

	return p, nil
}
