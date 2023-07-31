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
	DepositsKeyPrefix = KeyPrefix("Deposits/")
	VotesKeyPrefix = KeyPrefix("Votes/")	
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// ProposalKey returns the store key to retrieve a proposal from the hostZoneId + proposalID
func ProposalKey(hostZoneId string, proposalId uint64) []byte {
	var key []byte

	key = append(key, ProposalsKeyPrefix...)
	key = append(key, KeyPrefix(hostZoneId)...)
	key = append(key, govtypes.GetProposalIDBytes(proposalId)...)

	return key
}

// DepositKey returns the store key to retrieve a deposit amount from the creator address and denom fields
func DepositKey(creator string, hostZoneId string) []byte {
	var key []byte

	key = append(key, DepositsKeyPrefix...)
	key = append(key, KeyPrefix(hostZoneId)...)
	key = append(key, KeyPrefix(creator)...)

	return key
}

// VoteKey returns the store key to retrieve a specific vote from the creator given chain and proposal_id fields
func VoteKey(creator string, hostZoneId string, proposalId uint64) []byte {
	var key []byte

	key = append(key, VotesKeyPrefix...)
	key = append(key, KeyPrefix(hostZoneId)...)
	key = append(key, govtypes.GetProposalIDBytes(proposalId)...)
	key = append(key, KeyPrefix(creator)...)

	return key
}
