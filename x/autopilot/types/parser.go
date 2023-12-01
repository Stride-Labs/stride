package types

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RawPacketMetadata struct {
	Autopilot *struct {
		Receiver string                  `json:"receiver"`
		Stakeibc *StakeibcPacketMetadata `json:"stakeibc,omitempty"`
		Claim    *ClaimPacketMetadata    `json:"claim,omitempty"`
	} `json:"autopilot"`
}

type PacketForwardMetadata struct {
	Receiver    string
	RoutingInfo ModuleRoutingInfo
}

type ModuleRoutingInfo interface {
	Validate() error
}

// NOTE: I removed StakeibcPacketMetadata from StakeibcPacketMetadata/ClaimPacketMetadata because it was
// just being set to the receiver.
// I'm not sure this was necessary for airdrop linking, but it was necessary for liquid stakes, because the
// address that liquid stakes _must_ be the address that received tokens.
// A cleaner design here would be to do something like PFM, which is
// liquid_staking_address = hash(channel, sender)
// This removes the need for a receiving address.
// Tokens can be liquid staked via this protocol-owned address, which more clearly separates
// concerns (random addresses can't receive tokens, then be forced to liquid stake, which is the case today).
// The above isn't a big deal at face value - if you want to send someone an LST you can also do it using the bank module.
// For liquid stake, this is straightforward - we can just use a bank send after the tokens have been liquid staked.

// For liquid stake and forward, it is slightly more challenging, because if the IBC transfer fails, a fallback address on Stride
// is required (and this might not happen immediately - it could happen in a timeout). Unfortunately, liquid stake and forward
// re-introduces the PFM bug, because senders can no longer be trusted. Consider the following case:
// - A on Evmos sends 10 EVMOS to C on Stride, to be forwarded to B on Osmosis
// - 10 EVMOS -> 10 stEVMOS via B on Stride, IBC transferred to C on Osmosis with sender B
// - solution: enforce B is the same address as A, so A = B = C and the sender is trusted
// - drawback: doesn't work for chains with a different address derivition - the fallback address B
// 				might not be accessible (low confidence)

// Possible solutions
// (1) Just make the sender hash(channel, sender) so it's unusable on the destination chain
// 		pros: PFM-like solution, proven in prod
// 		cons: more complex since we need a fallback if the IBC transfer fails
// (2) Constrain the forwarding logic
// 		(1) tokens can only be sent to/from Evmos on the canonical channel
// 		(2) the receiver on Stride is overridden and set to the mechanical Stride address
// 			- what if the IBC times out and the user needs to manually claim their tokens?
// 			- we could design a retry mechanism (tx on Stride that anyone can call)

// Packet metadata info specific to Stakeibc (e.g. 1-click liquid staking)
type StakeibcPacketMetadata struct {
	Action string `json:"action"`
}

// Packet metadata info specific to Claim (e.g. airdrops for non-118 coins)
type ClaimPacketMetadata struct {
}

// Validate stakeibc packet metadata fields
// including the stride address and action type
func (m StakeibcPacketMetadata) Validate() error {
	if m.Action != "LiquidStake" {
		return errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", m.Action)
	}

	return nil
}

// Should we remove this struct?
func (m ClaimPacketMetadata) Validate() error {
	return nil
}

// Parse packet metadata intended for autopilot
// In the ICS-20 packet, the metadata can optionally indicate a module to route to (e.g. stakeibc)
// The PacketForwardMetadata returned from this function contains attributes for each autopilot supported module
// It can only be forward to one module per packet
// Returns nil if there was no metadata found
func ParsePacketMetadata(metadata string) (*PacketForwardMetadata, error) {
	// If we can't unmarshal the metadata into a PacketMetadata struct,
	// assume packet forwarding was no used and pass back nil so that autopilot is ignored
	var raw RawPacketMetadata
	if err := json.Unmarshal([]byte(metadata), &raw); err != nil {
		return nil, nil
	}

	// If no forwarding logic was used for autopilot, return the metadata with each disabled
	if raw.Autopilot == nil {
		return nil, nil
	}

	// Confirm a receiver address was supplied
	if _, err := sdk.AccAddressFromBech32(raw.Autopilot.Receiver); err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrInvalidReceiverAddress.Error())
	}

	// Parse the packet info into the specific module type
	// We increment the module count to ensure only one module type was provided
	moduleCount := 0
	var routingInfo ModuleRoutingInfo
	if raw.Autopilot.Stakeibc != nil {
		// override the stride address with the receiver address
		moduleCount++
		routingInfo = *raw.Autopilot.Stakeibc
	}
	if raw.Autopilot.Claim != nil {
		// override the stride address with the receiver address
		moduleCount++
		routingInfo = *raw.Autopilot.Claim
	}
	if moduleCount != 1 {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrInvalidModuleRoutes.Error())
	}

	// Validate the packet info according to the specific module type
	if err := routingInfo.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, ErrInvalidPacketMetadata.Error())
	}

	return &PacketForwardMetadata{
		Receiver:    raw.Autopilot.Receiver,
		RoutingInfo: routingInfo,
	}, nil
}
