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
	ProposalTypeUpdateRateLimit = "UpdateRateLimit"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateRateLimit)
}

var (
	_ govtypes.Content = &UpdateRateLimitProposal{}
)

func NewUpdateRateLimitProposal(title, description, denom, channelId string, maxPercentSend sdkmath.Int, maxPercentRecv sdkmath.Int, durationHours uint64) govtypes.Content {
	return &UpdateRateLimitProposal{
		Title:          title,
		Description:    description,
		Denom:          denom,
		ChannelId:      channelId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
	}
}

func (p *UpdateRateLimitProposal) GetTitle() string { return p.Title }

func (p *UpdateRateLimitProposal) GetDescription() string { return p.Description }

func (p *UpdateRateLimitProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateRateLimitProposal) ProposalType() string {
	return ProposalTypeUpdateRateLimit
}

func (p *UpdateRateLimitProposal) ValidateBasic() error {
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

func (p UpdateRateLimitProposal) String() string {
	return fmt.Sprintf(`Update Rate Limit Proposal:
	Title:           %s
	Description:     %s
	Denom:           %s
	ChannelId:      %s
	MaxPercentSend: %v
	MaxPercentRecv: %v
	DurationHours:  %d
  `, p.Title, p.Description, p.Denom, p.ChannelId, p.MaxPercentSend, p.MaxPercentRecv, p.DurationHours)
}
