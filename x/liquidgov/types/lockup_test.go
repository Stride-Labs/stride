package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pk1      = ed25519.GenPrivKey().PubKey()
	addr1, _ = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())

	denom1 = "ustrd"
	denom2 = "uatom"
)

func TestLockupEqual(t *testing.T) {
	l1 := NewLockup(sdk.AccAddress(addr1), sdk.NewInt(100), denom1)
	l2 := l1

	ok := l1.String() == l2.String()
	require.True(t, ok)

	l2.Denom = denom2
	l2.Amount = sdk.NewInt(200)

	ok = l1.String() == l2.String()
	require.False(t, ok)
}

func TestLockupString(t *testing.T) {
	l := NewLockup(sdk.AccAddress(addr1), sdk.NewInt(100), denom1)
	require.NotEmpty(t, l.String())
}

func TestUnlockingRecordEqual(t *testing.T) {
	ur1 := NewUnlockingRecord(sdk.AccAddress(addr1), denom1, 0,
		time.Unix(0, 0), sdk.NewInt(0))
	ur2 := ur1

	ok := ur1.String() == ur2.String()
	require.True(t, ok)

	ur2.Denom = denom2

	ur2.Entries[0].CompletionTime = time.Unix(20*20*2, 0)
	s1 := ur1.String()
	s2 := ur2.String()
	fmt.Println(s1)
	fmt.Println(s2)
	ok = (ur1.String() == ur2.String())
	require.False(t, ok)
}

func TestUnlockingRecordString(t *testing.T) {
	ubd := NewUnlockingRecord(sdk.AccAddress(addr1), denom1, 0,
		time.Unix(0, 0), sdk.NewInt(0))

	require.NotEmpty(t, ubd.String())
}
