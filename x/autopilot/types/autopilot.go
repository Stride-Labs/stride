package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// GenerateHashedReceiver returns the receiver address for a given channel and original sender.
// It overrides the receiver address to be a hash of the channel/origSender so that
// the receiver address is deterministic and can be used to identify the sender on the
// initial chain.

// GenerateHashedReceiver generates a new receiver address for a packet, by hashing
// the channel and original sender.
// This makes the receiver address deterministic and can used to identify the sender
// on the initial chain.
// Additionally, this prevents a forwarded packet from inpersonating a different account
// when moving to the next hop (i.e. receiver of this hop, becomes sender of next)
//
// This function was borrowed from PFM
func GenerateHashedReceiver(channelId, originalSender string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channelId, originalSender)
	senderHash32 := address.Hash(ModuleName, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}
