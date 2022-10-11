package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddValidator defines the type for a CommunityPoolAddValidator
	ProposalTypeAddValidator = "AddValidator"
)

// Assert CommunityPoolSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &AddValidatorProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidator)
}

// NewCommunityPoolSpendProposal creates a new community pool spend proposal.
//nolint:interfacer
func NewAddValidatorProposal(title string, description string, hostZone string, validatorName string, validatorAddress string) *AddValidatorProposal {
	return &AddValidatorProposal{title, description, hostZone, validatorName, validatorAddress}
}

// GetDescription returns the routing key of a community pool spend proposal.
func (avp *AddValidatorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (avp *AddValidatorProposal) ProposalType() string { return ProposalTypeAddValidator }

// ValidateBasic runs basic stateless validity checks
func (avp *AddValidatorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(avp)
	if err != nil {
		return err
	}
	if len(avp.ValidatorAddress) == 0 {
		return ErrRequiredFieldEmpty
	}
	if len(avp.ValidatorName) == 0 {
		return ErrRequiredFieldEmpty
	}

	return nil
}
