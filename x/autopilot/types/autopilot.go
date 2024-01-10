package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// RawPacketMetadata defines the raw JSON memo that's used in an autopilot transfer
// The PFM forward key is also used here to validate that the packet is not trying
// to use autopilot and PFM at the same time
// As a result, only the forward key is needed. The actual parsing of the PFM section
// will occur in the PFM module
type RawPacketMetadata struct {
	Autopilot *RawAutopilotMetadata `json:"autopilot"`
	Forward   *interface{}          `json:"forward"`
}

// This struct has all the possible autopilot fields across all actions
type RawAutopilotMetadata struct {
	// Address on stride that should receiver the transfer
	Receiver string `json:"receiver"`
	// The autopilot action
	Action string `json:"action"`

	// Fields for LiquidStakeAndForward - dictates the recipient of the outbound IBC transfer
	TransferReceiver string `json:"transfer_receiver"`
	TransferChannel  string `json:"transfer_channel"`

	// Address on
	RedemptionReceiver string `json:"redemption_receiver"`

	// Legacy fields - will be deprecated
	Stakeibc *StakeibcPacketMetadata `json:"stakeibc,omitempty"`
	Claim    *ClaimPacketMetadata    `json:"claim,omitempty"`
}

// TODO [v17]: Rename to AutopilotActionMetadata
// AutopilotActionMetadata stores the metadata that's specific to the autopilot action
// e.g. Fields required for LiquidStake
type AutopilotMetadata struct {
	Receiver string
	Action   string
	// TODO [v17]: Rename to ActionMetadata
	RoutingInfo interface{}
}

// DEPRECATED: Remove in future release
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
