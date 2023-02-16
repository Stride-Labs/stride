package types

import (
	"fmt"

	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeResetRateLimit = "ResetRateLimit"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeResetRateLimit)
}

var (
	_ govtypes.Content = &ResetRateLimitProposal{}
)

func NewResetRateLimitProposal(title, description, denom, channelId string) govtypes.Content {
	return &ResetRateLimitProposal{
		Title:       title,
		Description: description,
		Denom:       denom,
		ChannelId:   channelId,
	}
}

func (p *ResetRateLimitProposal) GetTitle() string { return p.Title }

func (p *ResetRateLimitProposal) GetDescription() string { return p.Description }

func (p *ResetRateLimitProposal) ProposalRoute() string { return RouterKey }

func (p *ResetRateLimitProposal) ProposalType() string {
	return ProposalTypeResetRateLimit
}

func (p *ResetRateLimitProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom (%s)", p.Denom)
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, p.ChannelId)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", p.ChannelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid channel-id (%s), must be of the format 'channel-{N}'", p.ChannelId)
	}

	return nil
}

func (p ResetRateLimitProposal) String() string {
	return fmt.Sprintf(`Reset Rate Limit Proposal:
	Title:           %s
	Description:     %s
	Denom:           %s
	ChannelId:      %s
  `, p.Title, p.Description, p.Denom, p.ChannelId)
}
