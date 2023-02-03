package types

import (
	"fmt"
	"strings"

	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewProposal(govProposal govtypesv1beta1.Proposal, hostZoneId string) (Proposal, error) {
	p := Proposal{
		GovProposal: govProposal,
		HostZoneId:  hostZoneId,
	}

	return p, nil
}

// String implements stringer interface
func (p Proposal) String() string {
	out := fmt.Sprintf("HostZone: %s", p.HostZoneId) + p.GovProposal.String()
	return strings.TrimSpace(out)
}

// Proposals is an array of proposal
type Proposals []Proposal
