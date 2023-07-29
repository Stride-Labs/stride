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

	// QuerierRoute is the querier route for the liquidgov store.
	QuerierRoute = StoreKey

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_liquidgov"
)

var (
	ProposalsKeyPrefix = KeyPrefix("Proposals/")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// ProposalKey returns the store key to retrieve a proposal from the chainId + proposalID
func ProposalKey(chainId string, proposalId uint64) []byte {
	var key []byte

	key = append(key, ProposalsKeyPrefix...)
	key = append(key, KeyPrefix(chainId)...)
	key = append(key, govtypes.GetProposalIDBytes(proposalId)...)

	return key
}
