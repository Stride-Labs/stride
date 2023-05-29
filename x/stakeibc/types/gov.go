package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

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

var (
	_ govtypes.Content = &RegisterHostZoneProposal{}
)

func NewRegisterHostZoneProposal(
	title, description,
	creator string,
	connectionId string,
	bech32prefix string,
	hostDenom string,
	ibcDenom string,
	transferChannelId string,
	unbondingFrequency uint64,
	minRedemptionRate,
	maxRedemptionRate sdk.Dec,
	deposit string) govtypes.Content {
	return &RegisterHostZoneProposal{
		Title:              title,
		Description:        description,
		ConnectionId:       connectionId,
		Bech32Prefix:       bech32prefix,
		HostDenom:          hostDenom,
		IbcDenom:           ibcDenom,
		TransferChannelId:  transferChannelId,
		UnbondingFrequency: unbondingFrequency,
		MinRedemptionRate:  minRedemptionRate,
		MaxRedemptionRate:  maxRedemptionRate,
		Deposit:            deposit,
	}
}

func (p *RegisterHostZoneProposal) GetTitle() string { return p.Title }

func (p *RegisterHostZoneProposal) GetDescription() string { return p.Description }

func (p *RegisterHostZoneProposal) ProposalRoute() string { return RouterKey }

func (p *RegisterHostZoneProposal) ProposalType() string {
	return ProposalTypeAddValidators
}

func (p *RegisterHostZoneProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// VALIDATE DENOMS
	// host denom cannot be empty
	if p.HostDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(p.HostDenom); err != nil {
		return err
	}

	// ibc denom cannot be empty and must begin with "ibc"
	if p.IbcDenom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ibc denom cannot be empty")
	}
	if !strings.HasPrefix(p.IbcDenom, "ibc") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ibc denom must begin with 'ibc'")
	}
	// ibc denom must be valid
	err = ibctransfertypes.ValidateIBCDenom(p.IbcDenom)
	if err != nil {
		return err
	}
	// bech32 prefix must be non-empty (we validate it fully in msg_server)
	if strings.TrimSpace(p.Bech32Prefix) == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "bech32 prefix must be non-empty")
	}
	// connection id cannot be empty and must begin with "connection"
	if p.ConnectionId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "connection id cannot be empty")
	}
	if !strings.HasPrefix(p.ConnectionId, "connection") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "connection id must begin with 'connection'")
	}
	// transfer channel id cannot be empty
	if p.TransferChannelId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "transfer channel id cannot be empty")
	}
	// transfer channel id must begin with "channel"
	if !strings.HasPrefix(p.TransferChannelId, "channel") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "transfer channel id must begin with 'channel'")
	}
	// unbonding frequency must be positive nonzero
	if p.UnbondingFrequency < 1 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unbonding frequency must be greater than zero")
	}
	// min/max redemption rate check
	if !p.MinRedemptionRate.IsNil() && p.MinRedemptionRate.IsNegative() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "min redemption rate should not be negative")
	}
	if !p.MaxRedemptionRate.IsNil() && p.MaxRedemptionRate.IsNegative() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max redemption rate should not be negative")
	}
	if !p.MinRedemptionRate.IsNil() &&
		!p.MaxRedemptionRate.IsNil() &&
		!p.MinRedemptionRate.IsZero() &&
		!p.MaxRedemptionRate.IsZero() &&
		p.MinRedemptionRate.GTE(p.MaxRedemptionRate) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "min redemption rate should be lower than max redemption rate")
	}

	return nil
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
