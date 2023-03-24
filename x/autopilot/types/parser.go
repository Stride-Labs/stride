package types

import (
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RawReceiver struct {
	Stakeibc *RawStakeibcReceiver `json:"stakeibc"`
	Claim    *RawClaimReceiver    `json:"claim"`
}

type RawStakeibcReceiver struct {
	Action        string `json:"action"`
	StrideAddress string `json:"stride_address"`
}

type RawClaimReceiver struct {
	AirdropId     string `json:"airdrop_id"`
	StrideAddress string `json:"stride_address"`
}

type ParsedReceiver struct {
	Stakeibc ParsedStakeibcReceiver
	Claim    ParsedClaimReceiver
}

type ParsedStakeibcReceiver struct {
	Enabled       bool
	Action        string
	StrideAddress sdk.AccAddress
}

type ParsedClaimReceiver struct {
	Enabled       bool
	AirdropId     string
	StrideAddress sdk.AccAddress
}

// Validate stakeibc receiver fields
func (r *RawStakeibcReceiver) ParseAndValidate() (*ParsedStakeibcReceiver, error) {
	// If the stakeibc section of the memo field was empty, mark stakeibc as disabled
	if r == nil {
		return &ParsedStakeibcReceiver{Enabled: false}, nil
	}

	address, err := sdk.AccAddressFromBech32(r.StrideAddress)
	if err != nil {
		return nil, err
	}
	if r.Action != "LiquidStake" {
		return nil, errorsmod.Wrapf(ErrUnsupportedStakeibcAction, "action %s is not supported", r.Action)
	}
	return &ParsedStakeibcReceiver{
		Action:        r.Action,
		StrideAddress: address,
		Enabled:       true,
	}, nil
}

// Validate claim receiver fields
func (r *RawClaimReceiver) ParseAndValidate() (*ParsedClaimReceiver, error) {
	// If the claim section of the memo field was empty, mark claim as disabled
	if r == nil {
		return &ParsedClaimReceiver{Enabled: false}, nil
	}

	// Validate the stride address and airdrop ID
	address, err := sdk.AccAddressFromBech32(r.StrideAddress)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(r.AirdropId)) == 0 {
		return nil, ErrInvalidClaimAirdropId
	}

	return &ParsedClaimReceiver{
		AirdropId:     r.AirdropId,
		StrideAddress: address,
		Enabled:       true,
	}, nil
}

func ParseReceiverData(receiverData string) (receiver *ParsedReceiver, err error) {
	// If an empty string is provided, add brackets so it gets unmarshalled as an empty JSON
	if receiverData == "" {
		receiverData = "{}"
	}

	// Unmarshal the memo field in to the raw receiver object
	var raw RawReceiver
	if err := json.Unmarshal([]byte(receiverData), &raw); err != nil {
		return receiver, errorsmod.Wrapf(err, ErrInvalidReceiverData.Error())
	}

	// Parse and validate the stakeibc module component from the receiver object
	parsedStakeibcReceiver, err := raw.Stakeibc.ParseAndValidate()
	if err != nil {
		return receiver, errorsmod.Wrapf(err, ErrInvalidReceiverData.Error())
	}

	// Parse and validate the claim module component from the receiver object
	parsedClaimReceiver, err := raw.Claim.ParseAndValidate()
	if err != nil {
		return receiver, errorsmod.Wrapf(err, ErrInvalidReceiverData.Error())
	}

	return &ParsedReceiver{
		Stakeibc: *parsedStakeibcReceiver,
		Claim:    *parsedClaimReceiver,
	}, nil
}
