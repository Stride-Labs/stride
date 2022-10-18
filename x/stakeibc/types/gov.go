package types

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeAddValidator    = "AddValidator"
	ProposalTypeDeleteValidator = "DeleteValidator"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidator)
	govtypes.RegisterProposalTypeCodec(&AddValidatorProposal{}, "stakeibc/AddValidatorProposal")
	govtypes.RegisterProposalType(ProposalTypeDeleteValidator)
	govtypes.RegisterProposalTypeCodec(&DeleteValidatorProposal{}, "stakeibc/DeleteValidatorProposal")
}

var (
	_ govtypes.Content = &AddValidatorProposal{}
	_ govtypes.Content = &DeleteValidatorProposal{}
)

func NewAddValidatorProposal(title, description, hostZone, name, address string) govtypes.Content {
	return &AddValidatorProposal{
		Title:            title,
		Description:      description,
		HostZone:         hostZone,
		ValidatorName:    name,
		ValidatorAddress: address,
	}
}

func (p *AddValidatorProposal) GetTitle() string { return p.Title }

func (p *AddValidatorProposal) GetDescription() string { return p.Description }

func (p *AddValidatorProposal) ProposalRoute() string { return RouterKey }

func (p *AddValidatorProposal) ProposalType() string {
	return ProposalTypeAddValidator
}

func (p *AddValidatorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// TODO: Add more checks?
	if len(p.ValidatorAddress) == 0 {
		return ErrRequiredFieldEmpty
	}
	if len(p.ValidatorName) == 0 {
		return ErrRequiredFieldEmpty
	}

	return nil
}

func (p AddValidatorProposal) String() string {
	return fmt.Sprintf(`Add Validator Proposal:
	Title:            %s
	Description:      %s
	HostZone:         %s
	ValidatorName:    %s
	ValidatorAddress: %s
  `, p.Title, p.Description, p.HostZone, p.ValidatorName, p.ValidatorAddress)
}

func NewDeleteValidatorProposal(title, description, hostZone, address string) govtypes.Content {
	return &DeleteValidatorProposal{
		Title:            title,
		Description:      description,
		HostZone:         hostZone,
		ValidatorAddress: address,
	}
}

func (p *DeleteValidatorProposal) GetTitle() string { return p.Title }

func (p *DeleteValidatorProposal) GetDescription() string { return p.Description }

func (p *DeleteValidatorProposal) ProposalRoute() string { return RouterKey }

func (p *DeleteValidatorProposal) ProposalType() string {
	return ProposalTypeDeleteValidator
}

func (p *DeleteValidatorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// TODO: Add more checks?
	if len(p.ValidatorAddress) == 0 {
		return ErrRequiredFieldEmpty
	}

	return nil
}

func (p DeleteValidatorProposal) String() string {
	return fmt.Sprintf(`Delete Validator Proposal:
	Title:            %s
	Description:      %s
	HostZone:         %s
	ValidatorAddress: %s
  `, p.Title, p.Description, p.HostZone, p.ValidatorAddress)
}
