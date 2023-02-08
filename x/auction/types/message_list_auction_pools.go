package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgListAuctionPools = "list_auction_pools"

var _ sdk.Msg = &MsgListAuctionPools{}

func NewMsgListAuctionPools(creator string) *MsgListAuctionPools {
  return &MsgListAuctionPools{
		Creator: creator,
	}
}

func (msg *MsgListAuctionPools) Route() string {
  return RouterKey
}

func (msg *MsgListAuctionPools) Type() string {
  return TypeMsgListAuctionPools
}

func (msg *MsgListAuctionPools) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgListAuctionPools) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgListAuctionPools) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  	}
  return nil
}

