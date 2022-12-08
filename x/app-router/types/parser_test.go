package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const Bech32Prefix = "stride"

func init() {
	config := sdk.GetConfig()
	valoper := sdk.PrefixValidator + sdk.PrefixOperator
	valoperpub := sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	config.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32Prefix+valoper, Bech32Prefix+valoperpub)
}

func TestParseReceiverDataTransfer(t *testing.T) {
	data := "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl|stakeibc/LiquidStake"
	pt, err := ParseReceiverData(data)

	require.NoError(t, err)
	require.True(t, pt.ShouldLiquidStake)
	require.Equal(t, pt.StrideAccAddress.String(), "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl")
}

func TestParseReceiverDataNoTransfer(t *testing.T) {
	data := "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
	pt, err := ParseReceiverData(data)

	require.NoError(t, err)
	require.False(t, pt.ShouldLiquidStake)
}

func TestParseReceiverDataErrors(t *testing.T) {
	// empty transfer field
	pt, err := ParseReceiverData("")
	require.NoError(t, err)
	require.False(t, pt.ShouldLiquidStake)

	// invalid string
	pt, err = ParseReceiverData("abc:def:")
	require.NoError(t, err)
	require.False(t, pt.ShouldLiquidStake)

	// invalid function
	pt, err = ParseReceiverData("stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl|stakeibc/xxx")
	require.NoError(t, err)
	require.False(t, pt.ShouldLiquidStake)

	// invalid address
	pt, err = ParseReceiverData("xxx|stakeibc/LiquidStake")
	require.EqualError(t, err, "decoding bech32 failed: invalid bech32 string length 3")
	require.Nil(t, pt)
}
