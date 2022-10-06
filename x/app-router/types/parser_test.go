package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseReceiverDataTransfer(t *testing.T) {
	data := "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl|stakeibc/liquidstake"
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
	testCases := []struct {
		name          string
		data          string
		errStartsWith string
	}{
		{
			"unparsable transfer field",
			"",
			"unparsable receiver",
		},
		{
			"unparsable transfer field",
			"abc:def:",
			"unparsable receiver",
		},
		{
			"missing pipe",
			"transfer/channel-0:cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
			"formatting incorrect",
		},
		{
			"invalid this chain address",
			"somm16plylpsgxechajltx9yeseqexzdzut9g8vla4k|transfer/channel-0:cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
			"decoding bech32 failed",
		},
		{
			"missing slash",
			"cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k|transfer\\channel-0:cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
			"formatting incorrect",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseReceiverData(tc.data)
			require.Error(t, err)
			require.Equal(t, err.Error()[:len(tc.errStartsWith)], tc.errStartsWith)
		})
	}
}
