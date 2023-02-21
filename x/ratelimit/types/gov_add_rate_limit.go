package types

import (
	"fmt"

	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeAddRateLimit = "AddRateLimit"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddRateLimit)
}

var (
	_ govtypes.Content = &AddRateLimitProposal{}
)

func NewAddRateLimitProposal(title, description, denom, channelId string, maxPercentSend sdkmath.Int, maxPercentRecv sdkmath.Int, durationHours uint64) govtypes.Content {
	return &AddRateLimitProposal{
		Title:          title,
		Description:    description,
		Denom:          denom,
		ChannelId:      channelId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
	}
}

func (p *AddRateLimitProposal) GetTitle() string { return p.Title }

func (p *AddRateLimitProposal) GetDescription() string { return p.Description }

func (p *AddRateLimitProposal) ProposalRoute() string { return RouterKey }

func (p *AddRateLimitProposal) ProposalType() string {
	return ProposalTypeAddRateLimit
}

func (p *AddRateLimitProposal) ValidateBasic() error {
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

	if p.MaxPercentSend.GT(sdkmath.NewInt(100)) || p.MaxPercentSend.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max-percent-send percent must be between 0 and 100 (inclusively), Provided: %v", p.MaxPercentSend)
	}

	if p.MaxPercentRecv.GT(sdkmath.NewInt(100)) || p.MaxPercentRecv.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max-percent-recv percent must be between 0 and 100 (inclusively), Provided: %v", p.MaxPercentRecv)
	}

	if p.MaxPercentRecv.IsZero() && p.MaxPercentSend.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "either the max send or max receive threshold must be greater than 0")
	}

	if p.DurationHours == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}

	return nil
}

func (p AddRateLimitProposal) String() string {
	return fmt.Sprintf(`Add Rate Limit Proposal:
	Title:           %s
	Description:     %s
	Denom:           %s
	ChannelId:      %s
	MaxPercentSend: %v
	MaxPercentRecv: %v
	DurationHours:  %d
  `, p.Title, p.Description, p.Denom, p.ChannelId, p.MaxPercentSend, p.MaxPercentRecv, p.DurationHours)
}
