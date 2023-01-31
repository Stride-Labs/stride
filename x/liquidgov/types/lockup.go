package types

import (
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewLockup creates a new lockup object
//
//nolint:interfacer
func NewLockup(creatorAddr sdk.AccAddress, amount sdk.Int, denom string) Lockup {
	return Lockup{
		Creator: creatorAddr.String(),
		Amount:  amount,
		Denom:   denom,
	}
}

// MustMarshalLockup returns the lockup bytes. Panics if fails
func MustMarshalLockup(cdc codec.BinaryCodec, lockup Lockup) []byte {
	return cdc.MustMarshal(&lockup)
}

// MustUnmarshalLockup return the unmarshaled lockup from bytes.
// Panics if fails.
func MustUnmarshalLockup(cdc codec.BinaryCodec, value []byte) Lockup {
	lockup, err := UnmarshalLockup(cdc, value)
	if err != nil {
		panic(err)
	}

	return lockup
}

// return the lockup
func UnmarshalLockup(cdc codec.BinaryCodec, value []byte) (lockup Lockup, err error) {
	err = cdc.Unmarshal(value, &lockup)
	return lockup, err
}

func (l Lockup) GetCreator() sdk.AccAddress {
	crAddr := sdk.MustAccAddressFromBech32(l.Creator)

	return crAddr
}

func (l Lockup) GetAmount() sdk.Int { return l.Amount }

func (l Lockup) GetDenom() string { return l.Denom }

// String returns a human readable string representation of a Delegation.
func (l Lockup) String() string {
	out, _ := yaml.Marshal(l)
	return string(out)
}

// Lockups is a collection of lockups
type Lockups []Lockup

func (l Lockups) String() (out string) {
	for _, lock := range l {
		out += lock.String() + "\n"
	}

	return strings.TrimSpace(out)
}

func NewUnlockingRecordEntry(creationHeight int64, completionTime time.Time, balance math.Int) UnlockingRecordEntry {
	return UnlockingRecordEntry{
		CreationHeight: creationHeight,
		CompletionTime: completionTime,
		Balance:        balance,
	}
}

// String implements the stringer interface for a UnlockingRecordEntry.
func (e UnlockingRecordEntry) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

// IsMature - is the current entry mature
func (e UnlockingRecordEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// NewUnlockingRecord - create a new unlocking record object
//
//nolint:interfacer
func NewUnlockingRecord(
	creator sdk.AccAddress, denom string,
	creationHeight int64, minTime time.Time, balance math.Int,
) UnlockingRecord {
	return UnlockingRecord{
		Creator: creator.String(),
		Denom:   denom,
		Entries: []UnlockingRecordEntry{
			NewUnlockingRecordEntry(creationHeight, minTime, balance),
		},
	}
}

// AddEntry - append entry to the unlocking record
func (ur *UnlockingRecord) AddEntry(creationHeight int64, minTime time.Time, balance math.Int) {
	entry := NewUnlockingRecordEntry(creationHeight, minTime, balance)
	ur.Entries = append(ur.Entries, entry)
}

// RemoveEntry - remove entry at index i to the unlocking record
func (ur *UnlockingRecord) RemoveEntry(i int64) {
	ur.Entries = append(ur.Entries[:i], ur.Entries[i+1:]...)
}

// return the unlocking record
func MustMarshalUR(cdc codec.BinaryCodec, ur UnlockingRecord) []byte {
	return cdc.MustMarshal(&ur)
}

// unmarshal a unlocking record from a store value
func MustUnmarshalUR(cdc codec.BinaryCodec, value []byte) UnlockingRecord {
	ur, err := UnmarshalUR(cdc, value)
	if err != nil {
		panic(err)
	}

	return ur
}

// unmarshal a unlocking record from a store value
func UnmarshalUR(cdc codec.BinaryCodec, value []byte) (ur UnlockingRecord, err error) {
	err = cdc.Unmarshal(value, &ur)
	return ur, err
}

// String returns a human readable string representation of an UnlockingRecord.
func (ur UnlockingRecord) String() string {
	out := fmt.Sprintf(`Unlocking Tokens for:
  Creator:                 %s
	Entries:`, ur.Creator)
	for i, entry := range ur.Entries {
		out += fmt.Sprintf(`    Unlocking Record %d:
      Creation Height:           %v
      Min time to unbond (unix): %v
      Expected balance:          %s`, i, entry.CreationHeight,
			entry.CompletionTime, entry.Balance)
	}

	return out
}

// UnlockingRecords is a collection of UnlockingRecord
type UnlockingRecords []UnlockingRecord

func (urs UnlockingRecords) String() (out string) {
	for _, u := range urs {
		out += u.String() + "\n"
	}

	return strings.TrimSpace(out)
}
