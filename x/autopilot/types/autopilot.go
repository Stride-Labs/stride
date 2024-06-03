package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// RawPacketMetadata defines the raw JSON memo that's used in an autopilot transfer
// The PFM forward key and wasm keys are also used here to validate that the packet
// was not trying to use more than one of autopilot PFM, and wasm at the same time
// As a result, only the key is needed, cause the actual parsing of the PFM/wasm
// packet will occur in the respective module
type RawPacketMetadata struct {
	Autopilot *struct {
		Receiver string                  `json:"receiver"`
		Stakeibc *StakeibcPacketMetadata `json:"stakeibc,omitempty"`
		Claim    *ClaimPacketMetadata    `json:"claim,omitempty"`
	} `json:"autopilot"`
	Forward *interface{} `json:"forward"`
	Wasm    *interface{} `json:"wasm"`
}

// AutopilotActionMetadata stores the metadata that's specific to the autopilot action
// e.g. Fields required for LiquidStake
type AutopilotMetadata struct {
	Receiver    string
	RoutingInfo ModuleRoutingInfo
}

// ModuleRoutingInfo defines the interface required for each autopilot action
type ModuleRoutingInfo interface {
	Validate() error
}

// GenerateHashedSender generates a new  address for a packet, by hashing
// the channel and original sender.
// This makes the address deterministic and can used to identify the sender
// from the preivous hop
// Additionally, this prevents a forwarded packet from impersonating a different account
// when moving to the next hop (i.e. receiver of one hop, becomes sender of next)
//
// This function was borrowed from PFM
func GenerateHashedAddress(channelId, originalSender string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channelId, originalSender)
	senderHash32 := address.Hash(ModuleName, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}
