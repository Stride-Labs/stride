package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// ---------------------------
// AddValidatorProposal
// ---------------------------
const (
	ProposalTypeAddValidators = "AddValidators"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddValidators)
}

var (
	_ govtypes.Content = &AddValidatorsProposal{}
)

func NewAddValidatorsProposal(title, description, hostZone string, validators []*Validator) govtypes.Content {
	return &AddValidatorsProposal{
		Title:       title,
		Description: description,
		HostZone:    hostZone,
		Validators:  validators,
	}
}

func (p *AddValidatorsProposal) GetTitle() string { return p.Title }

func (p *AddValidatorsProposal) GetDescription() string { return p.Description }

func (p *AddValidatorsProposal) ProposalRoute() string { return RouterKey }

func (p *AddValidatorsProposal) ProposalType() string {
	return ProposalTypeAddValidators
}

func (p *AddValidatorsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if len(p.Validators) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one validator must be provided")
	}

	for i, validator := range p.Validators {
		if len(strings.TrimSpace(validator.Name)) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator name is required (index %d)", i)
		}
		if len(strings.TrimSpace(validator.Address)) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address is required (index %d)", i)
		}
	}

	return nil
}

func (p AddValidatorsProposal) String() string {
	return fmt.Sprintf(`Add Validators Proposal:
	Title:            %s
	Description:      %s
	HostZone:         %s
	Validators:       %+v
  `, p.Title, p.Description, p.HostZone, p.Validators)
}

func (v *Validator) Equal(other *Validator) bool {
	if v == nil || other == nil {
		return false
	}
	if v.Address != other.Address {
		return false
	}
	if v.Name != other.Name {
		return false
	}
	return true
}

// ---------------------------
// ToggleLSMProposal
// ---------------------------

const (
	ProposalTypeToggleLSMProposal = "ToggleLSMProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeToggleLSMProposal)
}

var (
	_ govtypes.Content = &ToggleLSMProposal{}
)

func NewToggleLSMProposal(title, description, hostZone string, enabled bool) govtypes.Content {
	return &ToggleLSMProposal{
		Title:       title,
		Description: description,
		HostZone:    hostZone,
		Enabled:     enabled,
	}
}

func (p *ToggleLSMProposal) GetTitle() string { return p.Title }

func (p *ToggleLSMProposal) GetDescription() string { return p.Description }

func (p *ToggleLSMProposal) ProposalRoute() string { return RouterKey }

func (p *ToggleLSMProposal) ProposalType() string {
	return ProposalTypeAddValidators
}

func (p *ToggleLSMProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p ToggleLSMProposal) String() string {
	return fmt.Sprintf(`Add Validators Proposal:
	Title:            %s
	Description:      %s
	HostZone:         %s
	Enabled:          %v
  `, p.Title, p.Description, p.HostZone, p.Enabled)
}
