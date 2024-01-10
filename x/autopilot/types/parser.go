package types

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	LiquidStake           = "liquid-stake"
	LiquidStakeAndForward = "liquid-stake-and-forward"
	RedeemStake           = "redeem-stake"
	Claim                 = "claim"

	// DEPRECATED: Remove in next release
	LegacyLiquidStake = "LiquidStake"
	LegacyRedeemStake = "RedeemStake"
)

// DEPRECATED: Remove in next release
// Packet metadata info specific to Stakeibc (e.g. 1-click liquid staking)
type StakeibcPacketMetadata struct {
	Action          string `json:"action"`
	IbcReceiver     string `json:"ibc_receiver,omitempty"`
	TransferChannel string `json:"transfer_channel,omitempty"`
}

// DEPRECATED: Remove in next release
// Packet metadata info specific to Claim (e.g. airdrops for non-118 coins)
type ClaimPacketMetadata struct{}

// DEPRECATED: Remove in next release
// Validate stakeibc packet metadata fields
// including the stride address and action type
func (m StakeibcPacketMetadata) Validate() error {
	switch m.Action {
	case LiquidStake:
	case RedeemStake:
	default:
		return errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", m.Action)
	}

	return nil
}

// DEPRECATED: Remove in next release
// Validate claim packet metadata includes the stride address
// TODO: remove this function
func (m ClaimPacketMetadata) Validate() error {
	return nil
}

// Metadata for a liquid stake and forward action
type LiquidStakeAndForwardMetadata struct {
	// Recipient of outbound IBC transfer
	TransferReceiver string
	// Channel for outbound transfer (optional)
	// Defaults to the host zone's channel if not specified
	TransferChannel string
}

// Builds and validates a new LiquidStakeAndForwardMetadata struct from the raw fields
// Only the receiver field is required
func NewLiquidStakeAndForwardMetadata(raw RawAutopilotMetadata) (metadata LiquidStakeAndForwardMetadata, err error) {
	if raw.TransferReceiver == "" {
		return metadata, errorsmod.Wrapf(ErrInvalidPacketMetadata, "transfer_receiver must be included in a liquid stake and forward")
	}
	metadata = LiquidStakeAndForwardMetadata{
		TransferReceiver: raw.TransferReceiver,
		TransferChannel:  raw.TransferChannel,
	}
	return metadata, nil
}

// Metadata for a redeem stake action
type RedeemStakeMetadata struct {
	RedemptionReceiver string
}

// Builds and validates a new RedeemStakeMetadata struct from the raw fields
func NewRedeemStakeMetadata(raw RawAutopilotMetadata) (metadata RedeemStakeMetadata, err error) {
	if raw.RedemptionReceiver == "" {
		return metadata, errorsmod.Wrapf(ErrInvalidPacketMetadata, "redemption_receiver must be included in a redeem stake")
	}
	metadata = RedeemStakeMetadata{
		RedemptionReceiver: raw.RedemptionReceiver,
	}
	return metadata, nil
}

// Validates the raw packet based on the action, and returns the relevant metadata
func (r RawAutopilotMetadata) ParseActionMetadata() (interface{}, error) {
	switch r.Action {
	case LiquidStake:
		return nil, nil
	case LiquidStakeAndForward:
		return NewLiquidStakeAndForwardMetadata(r)
	case RedeemStake:
		return NewRedeemStakeMetadata(r)
	case Claim:
		return nil, nil
	default:
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, "unrecognized action %s", r.Action)
	}
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

	// Packets cannot be used for both autopilot and PFM at the same time
	// If both fields were provided, reject the packet
	if raw.Autopilot != nil && raw.Forward != nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, "autopilot and pfm cannot both be used in the same packet")
	}

	// If no forwarding logic was used for autopilot, return nil to indicate that
	// there's no autopilot action needed
	if raw.Autopilot == nil {
		return nil, nil
	}

	// Confirm a receiver address was supplied
	if _, err := sdk.AccAddressFromBech32(raw.Autopilot.Receiver); err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrInvalidReceiverAddress.Error())
	}

	// Check if the legacy API was used, and if so, use the legacy parser
	isLegacyFormat := raw.Autopilot.Stakeibc != nil || raw.Autopilot.Claim != nil
	if isLegacyFormat {
		return LegacyParseAutopilotMetadata(raw)
	}

	// Otherwise, pull out the action and parse the action metadata
	action := raw.Autopilot.Action
	actionMetadata, err := raw.Autopilot.ParseActionMetadata()
	if err != nil {
		return nil, err
	}

	return &AutopilotMetadata{
		Receiver:    raw.Autopilot.Receiver,
		Action:      action,
		RoutingInfo: actionMetadata,
	}, nil
}

// DEPRECATED: Remove in next release
// Parse packet metadata intended for autopilot
// In the ICS-20 packet, the metadata can optionally indicate a module to route to (e.g. stakeibc)
// The AutopilotMetadata returned from this function contains attributes for each autopilot supported module
// It can only be forward to one module per packet
// Returns nil if there was no autopilot metadata found
func LegacyParseAutopilotMetadata(raw RawPacketMetadata) (*AutopilotMetadata, error) {
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

	action, actionMetadata, err := ConvertLegacyAutopilotMetadata(routingInfo)
	if err != nil {
		return nil, err
	}

	return &AutopilotMetadata{
		Receiver:    raw.Autopilot.Receiver,
		Action:      action,
		RoutingInfo: actionMetadata,
	}, nil
}

// DEPRECATED: Remove in next release
func ConvertLegacyAutopilotMetadata(routingInfoGeneric ModuleRoutingInfo) (action string, metadata interface{}, err error) {
	switch routingInfo := routingInfoGeneric.(type) {
	case StakeibcPacketMetadata:
		switch routingInfo.Action {

		// The legacy liquid stake action included both with and without forwarding
		case LegacyLiquidStake:
			// Liquid stakes without forwarding (identified by an empty receiver) only require an action in the new schema
			if routingInfo.IbcReceiver == "" {
				return LiquidStake, nil, nil
			}

			// Liquid stakes with forwarding have a custom action and metadata fields to designate the receiver
			metadata := LiquidStakeAndForwardMetadata{
				TransferReceiver: routingInfo.IbcReceiver,
				TransferChannel:  routingInfo.TransferChannel,
			}
			return LiquidStakeAndForward, metadata, nil

		// Redeem stakes have a custom action to specify the redemption receiver
		case LegacyRedeemStake:
			metadata := RedeemStakeMetadata{
				RedemptionReceiver: routingInfo.IbcReceiver,
			}
			return RedeemStake, metadata, nil

		default:
			return "", nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, "unrecognized legacy autopilot stakeibc metadata")
		}

	// The claim action had it's own metadata and the updated schema only requires the action name
	case ClaimPacketMetadata:
		return Claim, nil, nil

	default:
		return "", nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, "unrecognized legacy autopilot metadata")
	}
}
