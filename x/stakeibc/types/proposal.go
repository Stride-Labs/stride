package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddValidator defines the type for a CommunityPoolAddValidator
	ProposalTypeAddValidator = "AddValidator"
	// ProposalTypeDeleteValidator defines the type for a CommunityPoolDeleteValidator
	ProposalTypeDeleteValidator = "DeleteValidator"
)

// Assert AddValidatorProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &AddValidatorProposal{}
var _ govtypes.Content = &DeleteValidatorProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidator)
	govtypes.RegisterProposalType(ProposalTypeDeleteValidator)
}

// Add Validator Proposal

// NewAddValidatorProposal creates a new add validator spend proposal.
//nolint:interfacer
func NewAddValidatorProposal(title string, description string, hostZone string, validatorName string, validatorAddress string) *AddValidatorProposal {
	return &AddValidatorProposal{title, description, hostZone, validatorName, validatorAddress}
}

// GetDescription returns the routing key of a add validator spend proposal.
func (avp *AddValidatorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a add validator spend proposal.
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

// DeleteValidator Proposal

// NewDeleteValidatorProposal creates a new delete validator spend proposal.
//nolint:interfacer
func NewDeleteValidatorProposal(title string, description string, hostZone string, validatorAddress string) *DeleteValidatorProposal {
	return &DeleteValidatorProposal{title, description, hostZone, validatorAddress}
}

// GetDescription returns the routing key of a delete validator spend proposal.
func (avp *DeleteValidatorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a delete validator spend proposal.
func (avp *DeleteValidatorProposal) ProposalType() string { return ProposalTypeDeleteValidator }

// ValidateBasic runs basic stateless validity checks
func (avp *DeleteValidatorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(avp)
	if err != nil {
		return err
	}
	if len(avp.ValidatorAddress) == 0 {
		return ErrRequiredFieldEmpty
	}

	return nil
}
