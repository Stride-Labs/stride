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
	ProposalIDPrefix   = KeyPrefix("ProposalID/")
	DepositsKeyPrefix = KeyPrefix("Deposits/")
	VotesKeyPrefix = KeyPrefix("Votes/")		
)


func KeyPrefix(p string) []byte {
	return []byte(p)
}

// ProposalKey returns the store key to retrieve a proposal from the proposalID and chain_id fields
func ProposalKey(proposal_id uint64, chain_id string) []byte {
	var key []byte

	key = append(key, ProposalsKeyPrefix...)
	key = append(key, KeyPrefix(chain_id)...)
	key = append(key, govtypes.GetProposalIDBytes(proposal_id)...)

	return key
}

// ProposalIDKey returns the store key to retrieve the current proposalID
func ProposalIDKey(chain_id string) []byte {
	var key []byte

	key = append(key, ProposalIDPrefix...)
	key = append(key, KeyPrefix(chain_id)...)

	return key
}

// DepositKey returns the store key to retrieve a deposit amount from the creator address and denom fields
func DepositKey(creator string, denom string) []byte {
	var key []byte

	key = append(key, DepositsKeyPrefix...)
	key = append(key, KeyPrefix(denom)...)
	key = append(key, KeyPrefix(creator)...)

	return key
}

// VoteKey returns the store key to retrieve a specific vote from the creator given chain and proposal_id fields
func VoteKey(creator string, chain_id string, proposal_id uint64) []byte {
	var key []byte

	key = append(key, VotesKeyPrefix...)
	key = append(key, KeyPrefix(creator)...)
	key = append(key, KeyPrefix(chain_id)...)
	key = append(key, govtypes.GetProposalIDBytes(proposal_id)...)

	return key
}
