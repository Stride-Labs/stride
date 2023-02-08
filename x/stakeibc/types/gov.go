package types

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeAddValidator = "AddValidator"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidator)
}

var (
	_ govtypes.Content = &AddValidatorProposal{}
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
