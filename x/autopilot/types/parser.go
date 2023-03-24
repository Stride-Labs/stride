package types

import (
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QUESTION: Should we just leave the address as a string in the parsed type
// and then combine these raw and parsed types?
type rawPacketMetadata struct {
	Autopilot *struct {
		Stakeibc *RawStakeibcPacketMetadata `json:"stakeibc"`
		Claim    *RawClaimPacketMetadata    `json:"claim"`
	} `json:"autopilot"`
}

type RawStakeibcPacketMetadata struct {
	Action        string `json:"action"`
	StrideAddress string `json:"stride_address"`
}

type RawClaimPacketMetadata struct {
	AirdropId     string `json:"airdrop_id"`
	StrideAddress string `json:"stride_address"`
}

type PacketMetadata struct {
	Reciever string
	Stakeibc StakeibcPacketMetadata
	Claim    ClaimPacketMetadata
}

type StakeibcPacketMetadata struct {
	Enabled       bool
	Action        string
	StrideAddress sdk.AccAddress
}

type ClaimPacketMetadata struct {
	Enabled       bool
	AirdropId     string
	StrideAddress sdk.AccAddress
}

// Validate stakeibc packet metadata fields
func (r *RawStakeibcPacketMetadata) ParseAndValidate() (*StakeibcPacketMetadata, error) {
	// If the stakeibc section of the memo field was empty, mark stakeibc as disabled
	if r == nil {
		return &StakeibcPacketMetadata{Enabled: false}, nil
	}

	// Validate the stride address and action
	address, err := sdk.AccAddressFromBech32(r.StrideAddress)
	if err != nil {
		return nil, err
	}
	if r.Action != "LiquidStake" {
		return nil, errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", r.Action)
	}

	return &StakeibcPacketMetadata{
		Action:        r.Action,
		StrideAddress: address,
		Enabled:       true,
	}, nil
}

// Validate claim packet metadata fields
func (r *RawClaimPacketMetadata) ParseAndValidate() (*ClaimPacketMetadata, error) {
	// If the claim section of the memo field was empty, mark claim as disabled
	if r == nil {
		return &ClaimPacketMetadata{Enabled: false}, nil
	}

	// Validate the stride address and airdrop ID
	address, err := sdk.AccAddressFromBech32(r.StrideAddress)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(r.AirdropId)) == 0 {
		return nil, ErrInvalidClaimAirdropId
	}

	return &ClaimPacketMetadata{
		AirdropId:     r.AirdropId,
		StrideAddress: address,
		Enabled:       true,
	}, nil
}

// Parse packet metadata intended for autopilot
// Returns nil if there was no metadata found
func ParsePacketMetadata(metadata string) (packetMetadata *PacketMetadata, err error) {
	// If we can't unmarshal the metadata into a PacketMetadata struct,
	// assume packet forwarding was no used and pass back nil so that autopilot is ignored
	var raw rawPacketMetadata
	if err := json.Unmarshal([]byte(metadata), &raw); err != nil {
		return nil, nil
	}

	// If no forwarding logic was used for autopilot, return the metadata with each disabled
	if raw.Autopilot == nil {
		return nil, nil
	}

	// Parse and validate the stakeibc module component from the receiver object
	parsedStakeibcPacketData, err := raw.Autopilot.Stakeibc.ParseAndValidate()
	if err != nil {
		return nil, errorsmod.Wrapf(err, ErrInvalidPacketMetadata.Error())
	}

	// Parse and validate the claim module component from the receiver object
	parsedClaimPacketData, err := raw.Autopilot.Claim.ParseAndValidate()
	if err != nil {
		return nil, errorsmod.Wrapf(err, ErrInvalidPacketMetadata.Error())
	}

	// Confirm only one module was enabled for autopilot
	if parsedStakeibcPacketData.Enabled && parsedClaimPacketData.Enabled {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrMulitpleAutopilotRoutesInTx.Error())
	}

	// Return the combined metadata struct with each module flagged as enabled/disabled
	return &PacketMetadata{
		Stakeibc: *parsedStakeibcPacketData,
		Claim:    *parsedClaimPacketData,
	}, nil
}
