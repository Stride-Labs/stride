package keeper_test

import (
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/stretchr/testify/require"
)

// TestICS23ReplaceDirectiveActive is a canary test that verifies Stride is
// built against the Stride-Labs ics23 fork rather than upstream cosmos/ics23.
//
// Upstream ics23 v0.11.0's LeafOp.Apply rejects empty-value leaves with
// "leaf op needs value" (ops.go:66). This breaks ICQ non-membership
// verification when the queried key's neighbor is one of the empty-value
// reverse-index entries written by SDK 0.50+ bank via collections'
// WithReversePairUncheckedValue — the failure mode tracked in cosmos/ics23#134.
//
// The fix is a one-line patch in the Stride-Labs fork (same module path
// swapped via `replace` in go.mod). This test ensures that swap is still in
// place — if someone drops the replace directive, bumps to a newer ics23
// version that hasn't merged the fix, or accidentally reverts the go.mod
// change, this test fails with a clear pointer to the cause.
//
// Do not remove. It runs in under a millisecond and catches a silent
// regression that would otherwise only surface as intermittent ICQ timeouts
// on host zones with zero balances.
func TestICS23ReplaceDirectiveActive(t *testing.T) {
	op := &ics23.LeafOp{
		Hash:         ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
		Prefix:       []byte{0},
	}
	proof := &ics23.ExistenceProof{
		Key:   []byte("uatom"),
		Value: []byte{},
		Leaf:  op,
	}

	_, err := proof.Calculate()
	require.NoError(t, err,
		"empty-value leaf was rejected — the ics23 replace directive in go.mod "+
			"may have been dropped or the fork may no longer carry the cosmos/ics23#134 fix. "+
			"Check `go list -m github.com/cosmos/ics23/go` for the active module.")
}
