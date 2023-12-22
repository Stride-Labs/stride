package types

import (
	"errors"
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// RawPacketMetadata defines the raw JSON memo that's used in an autopilot transfer
// The PFM forward key is also used here to validate that the packet is not trying
// to use autopilot and PFM at the same time
// As a result, only the forward key is needed, cause the actual parsing of the PFM
// packet will occur in the PFM module
type RawPacketMetadata struct {
	Autopilot *struct {
		Receiver string                  `json:"receiver"`
		Stakeibc *StakeibcPacketMetadata `json:"stakeibc,omitempty"`
		Claim    *ClaimPacketMetadata    `json:"claim,omitempty"`
	} `json:"autopilot"`
	Forward *interface{} `json:"forward"`
}

// TokenPacketMetadata is meant to replicate transfertypes.FungibleTokenPacketData
// but it drops the original sender (who is untrusted) and adds a hashed receiver
// that can be used for any forwarding
type AutopilotTransferMetadata struct {
	Sender           string
	OriginalReceiver string
	HashedReceiver   string
	Amount           sdkmath.Int
	Denom            string
}

// AutopilotActionMetadata stores the metadata that's specific to the autopilot action
// e.g. Fields required for LiquidStake
type AutopilotActionMetadata struct {
	Receiver    string
	RoutingInfo ModuleRoutingInfo
}

// ModuleRoutingInfo defines the interface required for each autopilot action
type ModuleRoutingInfo interface {
	Validate() error
}

// Builds a AutopilotTransferMetadata object using the fields of a FungibleTokenPacketData
// and adding a hashed receiver
func NewAutopilotTransferMetadata(channelId string, data transfertypes.FungibleTokenPacketData) (AutopilotTransferMetadata, error) {
	hashedReceiver, err := GenerateHashedReceiver(channelId, data.Sender)
	if err != nil {
		return AutopilotTransferMetadata{}, err
	}

	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return AutopilotTransferMetadata{}, errors.New("not a parsable amount field")
	}

	return AutopilotTransferMetadata{
		Sender:           data.Sender,
		OriginalReceiver: data.Receiver,
		HashedReceiver:   hashedReceiver,
		Amount:           amount,
		Denom:            data.Denom,
	}, nil
}

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
