package types

import (
	"github.com/cosmos/cosmos-sdk/types/address"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	LockupKey          = []byte{0x10}
	UnlockingRecordKey = []byte{0x11}
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

// GetLockupKey creates the key for lockup of denom tokens
// VALUE: liquidgov/Lockup
func GetLockupKey(creatorAddr sdk.AccAddress, denom string) []byte {
	denomBytes := []byte(denom)
	return append(GetLockupsKey(creatorAddr), denomBytes...)
}

// GetLockupsKey creates the prefix for a lockup creator for all denoms
func GetLockupsKey(creatorAddr sdk.AccAddress) []byte {
	return append(LockupKey, address.MustLengthPrefix(creatorAddr)...)
}

// GetURKey creates the key for an unlocking record by creator addr and denom
// VALUE: liquidgov/UnlockingRecord
func GetURKey(creator sdk.AccAddress, denom string) []byte {
	denomBytes := []byte(denom)
	return append(GetURsKey(creator.Bytes()), address.MustLengthPrefix(denomBytes)...)
}

// GetURsKey creates the prefix for all unlocking recorde from a creator
func GetURsKey(creator sdk.AccAddress) []byte {
	return append(UnlockingRecordKey, address.MustLengthPrefix(creator)...)
}
