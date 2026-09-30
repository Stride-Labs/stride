// x/stakeibc/client/cli/tx_wind_down_test.go
package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/client/cli"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestReadValidatorUndelegations(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := cli.ReadValidatorUndelegations("/does/not/exist.json")
		require.ErrorContains(t, err, "no such file")
	})

	t.Run("bad offset", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "validators.json")
		require.NoError(t, os.WriteFile(path, []byte(`[{"address":"cosmosvaloper1abc","offset":"banana"}]`), 0o600))
		_, err := cli.ReadValidatorUndelegations(path)
		require.ErrorContains(t, err, "can not convert string to int")
	})

	t.Run("offset defaults to zero", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "validators.json")
		require.NoError(t, os.WriteFile(path, []byte(`[{"address":"cosmosvaloper1abc"},{"address":"cosmosvaloper1def","offset":"7"}]`), 0o600))
		validators, err := cli.ReadValidatorUndelegations(path)
		require.NoError(t, err)
		require.Equal(t, []types.ValidatorUndelegation{
			{Address: "cosmosvaloper1abc", Offset: sdkmath.ZeroInt()},
			{Address: "cosmosvaloper1def", Offset: sdkmath.NewInt(7)},
		}, validators)
	})
}

func TestParseIcaType(t *testing.T) {
	for _, name := range []string{"DELEGATION", "withdrawal", "Fee", "REDEMPTION"} {
		icaType, err := cli.ParseIcaType(name)
		require.NoError(t, err, name)
		require.Equal(t, strings.ToUpper(name), icaType.String())
	}
	_, err := cli.ParseIcaType("TREASURY")
	require.ErrorContains(t, err, "unknown ica type")
}

func TestParseBaseUnits(t *testing.T) {
	amount, err := cli.ParseBaseUnits("1000000")
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000_000), amount)

	_, err = cli.ParseBaseUnits("banana")
	require.ErrorContains(t, err, "can not convert string to int")
}

func TestCmdSweepTokensOffStride(t *testing.T) {
	t.Run("addresses file missing", func(t *testing.T) {
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"stuatom,ustrd", "/nonexistent/addresses.txt"}, "unable to read addresses file")
	})

	t.Run("empty denoms", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "addresses.txt")
		require.NoError(t, os.WriteFile(file, []byte("stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7\n"), 0o600))
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"", file}, "at least one denom is required")
	})

	t.Run("empty addresses file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "addresses.txt")
		require.NoError(t, os.WriteFile(file, []byte("\n\n"), 0o600))
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"stuatom", file}, "addresses file is empty")
	})
}
