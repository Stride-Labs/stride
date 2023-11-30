package types

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const TypeMsgCreateTradeRoute = "create_trade_route"

const (
	ConnectionIdRegex = `^connection-\d+$`
	ChannelIdRegex    = `^channel-\d+$`
	IBCPrefix         = "ibc/"
)

var (
	_ sdk.Msg            = &MsgCreateTradeRoute{}
	_ legacytx.LegacyMsg = &MsgCreateTradeRoute{}
)

func (msg *MsgCreateTradeRoute) Type() string {
	return TypeMsgCreateTradeRoute
}

func (msg *MsgCreateTradeRoute) Route() string {
	return RouterKey
}

func (msg *MsgCreateTradeRoute) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateTradeRoute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgCreateTradeRoute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if msg.HostChainId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "host chain ID cannot be empty")
	}

	if err := ValidateConnectionId(msg.StrideToRewardConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid stride to reward connection ID")
	}
	if err := ValidateConnectionId(msg.StrideToTradeConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid stride to trade connection ID")
	}

	if err := ValidateChannelId(msg.HostToRewardTransferChannelId); err != nil {
		return errorsmod.Wrap(err, "invalid host to reward channel ID")
	}
	if err := ValidateChannelId(msg.RewardToTradeTransferChannelId); err != nil {
		return errorsmod.Wrap(err, "invalid reward to trade channel ID")
	}
	if err := ValidateChannelId(msg.TradeToHostTransferChannelId); err != nil {
		return errorsmod.Wrap(err, "invalid trade to host channel ID")
	}

	if err := ValidateDenom(msg.RewardDenomOnHost, true); err != nil {
		return errorsmod.Wrap(err, "invalid reward denom on host")
	}
	if err := ValidateDenom(msg.RewardDenomOnReward, false); err != nil {
		return errorsmod.Wrap(err, "invalid reward denom on reward")
	}
	if err := ValidateDenom(msg.RewardDenomOnTrade, true); err != nil {
		return errorsmod.Wrap(err, "invalid reward denom on trade")
	}
	if err := ValidateDenom(msg.HostDenomOnTrade, true); err != nil {
		return errorsmod.Wrap(err, "invalid host denom on trade")
	}
	if err := ValidateDenom(msg.HostDenomOnHost, false); err != nil {
		return errorsmod.Wrap(err, "invalid host denom on host")
	}

	if msg.PoolId < 1 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid pool id")
	}
	if msg.MaxSwapAmount.GT(sdkmath.ZeroInt()) && msg.MinSwapAmount.GT(msg.MaxSwapAmount) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min swap amount cannot be greater than max swap amount")
	}

	maxAllowedSwapLossRate, err := sdk.NewDecFromStr(msg.MaxAllowedSwapLossRate)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to cast max allowed swap loss rate to a decimal")
	}
	if maxAllowedSwapLossRate.LT(sdk.ZeroDec()) || maxAllowedSwapLossRate.GT(sdk.OneDec()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max allowed swap loss rate must be between 0 and 1")
	}

	return nil
}

// Helper function to validate a connection Id
func ValidateConnectionId(connectionId string) error {
	matched, err := regexp.MatchString(ConnectionIdRegex, connectionId)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to verify connnection-id (%s)", connectionId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid connection-id (%s), must be of the format 'connection-{N}'", connectionId)
	}
	return nil
}

// Helper function to validate a channel Id
func ValidateChannelId(channelId string) error {
	matched, err := regexp.MatchString(ChannelIdRegex, channelId)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to verify channel-id (%s)", channelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid channel-id (%s), must be of the format 'channel-{N}'", channelId)
	}
	return nil
}

// Helper function to validate a denom
func ValidateDenom(denom string, ibc bool) error {
	if denom == "" {
		return errorsmod.Wrap(ErrInvalidDenom, "denom is empty")
	}
	if ibc && !strings.HasPrefix(denom, IBCPrefix) {
		return errorsmod.Wrapf(ErrInvalidDenom, "denom (%s) should have ibc prefix", denom)
	}
	return nil
}
