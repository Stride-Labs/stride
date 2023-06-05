package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func NewZoneAddress(chainId string) sdk.AccAddress {
	key := append([]byte("zone"), []byte(chainId)...)
	return address.Module(ModuleName, key)
}
