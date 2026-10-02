package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/client/cli"
)

func TestCmdClaimUndelegatedTokens(t *testing.T) {
	args := []string{
		"[host-zone]",
		"[epoch]",
		"[receiver]",
	}

	cmd := cli.CmdClaimUndelegatedTokens()
	ExecuteCLIExpectError(t, cmd, args, `unable to cast "[epoch]" of type string to uint64`)
}

func TestCmdAddValidators(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		args := []string{
			"[host-zone]",
			"[validator-list-file]",
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `open [validator-list-file]: no such file or directory`)
	})
	t.Run("empty file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `unexpected end of JSON input`)
	})
	t.Run("non json file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString("This is not JSON")
		require.NoError(t, err)

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `invalid character 'T' looking for beginning of value`)
	})
	t.Run("wrong json format", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{"blabla_validator_weights":[{"address":"cosmosXXX","weight":1}]}`)
		require.NoError(t, err)

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `invalid creator address (empty address string is not allowed): invalid address`)
	})
}

func TestCmdChangeValidatorWeight(t *testing.T) {
	args := []string{
		"[host-zone]",
		"[address]",
		"[weight]",
	}

	cmd := cli.CmdChangeValidatorWeight()
	ExecuteCLIExpectError(t, cmd, args, `unable to cast "[weight]" of type string to uint64`)
}

func TestCmdChangeMultipleValidatorWeight(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		args := []string{
			"[host-zone]",
			"[validator-list-file]",
		}

		cmd := cli.CmdChangeMultipleValidatorWeight()
		ExecuteCLIExpectError(t, cmd, args, `open [validator-list-file]: no such file or directory`)
	})
	t.Run("empty file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdChangeMultipleValidatorWeight()
		ExecuteCLIExpectError(t, cmd, args, `unexpected end of JSON input`)
	})
	t.Run("non json file", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString("This is not JSON")
		require.NoError(t, err)

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdChangeMultipleValidatorWeight()
		ExecuteCLIExpectError(t, cmd, args, `invalid character 'T' looking for beginning of value`)
	})
	t.Run("wrong json format", func(t *testing.T) {
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{"blabla_validator_weights":[{"address":"cosmosXXX","weight":1}]}`)
		require.NoError(t, err)

		args := []string{
			"[host-zone]",
			f.Name(),
		}

		cmd := cli.CmdChangeMultipleValidatorWeight()
		ExecuteCLIExpectError(t, cmd, args, `invalid creator address (empty address string is not allowed): invalid address`)
	})
}

func TestCmdUpdateInnerRedemptionRateBounds(t *testing.T) {
	t.Run("invalid min-bound", func(t *testing.T) {
		args := []string{
			"[chainid]",
			"[min-bound]",
			"[max-bound]",
		}

		cmd := cli.CmdUpdateInnerRedemptionRateBounds()
		assert.PanicsWithError(t, "failed to set decimal string with base 10: [min-bound]000000000000000000", func() {
			ExecuteCLIExpectError(t, cmd, args, "")
		})
	})
	t.Run("invalid max-bound", func(t *testing.T) {
		args := []string{
			"[chainid]",
			"0.123",
			"[max-bound]",
		}

		cmd := cli.CmdUpdateInnerRedemptionRateBounds()
		assert.PanicsWithError(t, "failed to set decimal string with base 10: [max-bound]000000000000000000", func() {
			ExecuteCLIExpectError(t, cmd, args, "")
		})
	})
}
