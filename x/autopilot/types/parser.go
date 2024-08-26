package types

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const LiquidStake = "LiquidStake"
const RedeemStake = "RedeemStake"

// Packet metadata info specific to Stakeibc (e.g. 1-click liquid staking)
type StakeibcPacketMetadata struct {
	Action string `json:"action"`
	// TODO [cleanup]: Rename to FallbackAddress
	StrideAddress   string
	IbcReceiver     string `json:"ibc_receiver,omitempty"`
	TransferChannel string `json:"transfer_channel,omitempty"`
}

// Packet metadata info specific to Claim (e.g. airdrops for non-118 coins)
// TODO: remove this struct
type ClaimPacketMetadata struct {
	StrideAddress string
}

// Validate stakeibc packet metadata fields
// including the stride address and action type
func (m StakeibcPacketMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.StrideAddress)
	if err != nil {
		return err
	}
	switch m.Action {
	case LiquidStake:
	case RedeemStake:
	default:
		return errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", m.Action)
	}

	return nil
}

// Validate claim packet metadata includes the stride address
// TODO: remove this function
func (m ClaimPacketMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.StrideAddress)
	if err != nil {
		return err
	}

	return nil
}

// Parse packet metadata intended for autopilot
// In the ICS-20 packet, the metadata can optionally indicate a module to route to (e.g. stakeibc)
// The AutopilotMetadata returned from this function contains attributes for each autopilot supported module
// It can only be forward to one module per packet
// Returns nil if there was no autopilot metadata found
func ParseAutopilotMetadata(metadata string) (*AutopilotMetadata, error) {
	// If we can't unmarshal the metadata into a PacketMetadata struct,
	// assume packet forwarding was no used and pass back nil so that autopilot is ignored
	var raw RawPacketMetadata
	if err := json.Unmarshal([]byte(metadata), &raw); err != nil {
		return nil, nil
	}

	// Packets cannot be used for more than one of autopilot, pfm, or wasmhooks at the same time
	// If more than one module key was provided, reject the packet
	middlewareModulesEnabled := 0
	if raw.Autopilot != nil {
		middlewareModulesEnabled++
	}
	if raw.Forward != nil {
		middlewareModulesEnabled++
	}
	if raw.Wasm != nil {
		middlewareModulesEnabled++
	}
	if middlewareModulesEnabled > 1 {
		return nil, errorsmod.Wrap(ErrInvalidPacketMetadata,
			"only one of autopilot, pfm, and wasm can both be used in the same packet")
	}

	// If no forwarding logic was used for autopilot, return nil to indicate that
	// there's no autopilot action needed
	if raw.Autopilot == nil {
		return nil, nil
	}

	// Confirm a receiver address was supplied
	if _, err := sdk.AccAddressFromBech32(raw.Autopilot.Receiver); err != nil {
		return nil, errorsmod.Wrap(ErrInvalidPacketMetadata, ErrInvalidReceiverAddress.Error())
	}

	// Parse the packet info into the specific module type
	// We increment the module count to ensure only one module type was provided
	moduleCount := 0
	var routingInfo ModuleRoutingInfo
	if raw.Autopilot.Stakeibc != nil {
		// override the stride address with the receiver address
		raw.Autopilot.Stakeibc.StrideAddress = raw.Autopilot.Receiver
		moduleCount++
		routingInfo = *raw.Autopilot.Stakeibc
	}
	if raw.Autopilot.Claim != nil {
		// override the stride address with the receiver address
		raw.Autopilot.Claim.StrideAddress = raw.Autopilot.Receiver
		moduleCount++
		routingInfo = *raw.Autopilot.Claim
	}
	if moduleCount != 1 {
		return nil, errorsmod.Wrap(ErrInvalidPacketMetadata, ErrInvalidModuleRoutes.Error())
	}

	// Validate the packet info according to the specific module type
	if err := routingInfo.Validate(); err != nil {
		return nil, errorsmod.Wrap(err, ErrInvalidPacketMetadata.Error())
	}

	return &AutopilotMetadata{
		Receiver:    raw.Autopilot.Receiver,
		RoutingInfo: routingInfo,
	}, nil
}
