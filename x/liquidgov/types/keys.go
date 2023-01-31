package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "liquidgov"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_liquidgov"
)

var (
	ProposalsKeyPrefix = []byte{0x00}
	ProposalIDPrefix   = []byte{0x01}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// ProposalKey returns the store key to retrieve a proposal from the proposalID and chain_id fields
func ProposalKey(proposalID uint64, chain_id string) []byte {
	var key []byte

	epochIdentifierBytes := []byte(chain_id)
	key = append(key, ProposalsKeyPrefix...)
	key = append(key, epochIdentifierBytes...)
	key = append(key, govtypes.GetProposalIDBytes(proposalID)...)

	return key
}

// ProposalKey returns the store key to retrieve a proposal from the proposalID and chain_id fields
func ProposalIDKey(chain_id string) []byte {
	var key []byte

	epochIdentifierBytes := []byte(chain_id)
	key = append(key, ProposalIDPrefix...)
	key = append(key, epochIdentifierBytes...)

	return key
}
