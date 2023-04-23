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

// Packet metadata info specific to Stakeibc (e.g. 1-click liquid staking)
type StakeibcPacketMetadata struct {
	Action        string `json:"action"`
	StrideAddress string `json:"stride_address"`
}

// Packet metadata info specific to Claim (e.g. airdrops for non-118 coins)
type ClaimPacketMetadata struct {
	StrideAddress string `json:"stride_address"`
}

// Validate stakeibc packet metadata fields
// including the stride address and action type
func (m StakeibcPacketMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.StrideAddress)
	if err != nil {
		return err
	}
	if m.Action != "LiquidStake" {
		return errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", m.Action)
	}

	return nil
}

// Validate claim packet metadata includes the stride address
func (m ClaimPacketMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.StrideAddress)
	if err != nil {
		return err
	}

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
		moduleCount++
		routingInfo = *raw.Autopilot.Stakeibc
	}
	if raw.Autopilot.Claim != nil {
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
