package types

import (
	"fmt"

	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeRemoveRateLimit = "RemoveRateLimit"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRemoveRateLimit)
}

var (
	_ govtypes.Content = &RemoveRateLimitProposal{}
)

func NewRemoveRateLimitProposal(title, description, denom, channelId string) govtypes.Content {
	return &RemoveRateLimitProposal{
		Title:       title,
		Description: description,
		Denom:       denom,
		ChannelId:   channelId,
	}
}

func (p *RemoveRateLimitProposal) GetTitle() string { return p.Title }

func (p *RemoveRateLimitProposal) GetDescription() string { return p.Description }

func (p *RemoveRateLimitProposal) ProposalRoute() string { return RouterKey }

func (p *RemoveRateLimitProposal) ProposalType() string {
	return ProposalTypeRemoveRateLimit
}

func (p *RemoveRateLimitProposal) ValidateBasic() error {
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

func (p RemoveRateLimitProposal) String() string {
	return fmt.Sprintf(`Remove Rate Limit Proposal:
	Title:           %s
	Description:     %s
	Denom:           %s
	ChannelId:      %s
  `, p.Title, p.Description, p.Denom, p.ChannelId)
}
