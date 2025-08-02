package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v28/x/stakeibc/client/cli"
)

func TestCmdLiquidStake(t *testing.T) {
	args := []string{
		"banana",
		"[lsm-token-denom]",
	}

	cmd := cli.CmdLiquidStake()
	ExecuteCLIExpectError(t, cmd, args, "can not convert string to int")
}

func TestCmdLSMLiquidStake(t *testing.T) {
	args := []string{
		"banana",
		"[lsm-token-denom]",
	}

	cmd := cli.CmdLSMLiquidStake()
	ExecuteCLIExpectError(t, cmd, args, "can not convert string to int")
}

func TestCmdRegisterHostZone(t *testing.T) {
	t.Run("unbonding-period not a number", func(t *testing.T) {
		args := []string{
			"[connection-id]",
			"[host-denom]",
			"[bech32prefix]",
			"[ibc-denom]",
			"[channel-id]",
			"[unbonding-period]",
			"1",
		}

		cmd := cli.CmdRegisterHostZone()
		ExecuteCLIExpectError(t, cmd, args, `strconv.ParseUint: parsing "[unbonding-period]": invalid syntax`)
	})

	t.Run("lsm-enabled not a boolean", func(t *testing.T) {
		args := []string{
			"[connection-id]",
			"[host-denom]",
			"[bech32prefix]",
			"[ibc-denom]",
			"[channel-id]",
			"0",
			"2",
		}

		cmd := cli.CmdRegisterHostZone()
		ExecuteCLIExpectError(t, cmd, args, `strconv.ParseBool: parsing "2": invalid syntax`)
	})
}

func TestCmdRedeemStake(t *testing.T) {
	args := []string{
		"[amount]",
		"[hostZoneID]",
		"[receiver]",
	}

	cmd := cli.CmdRedeemStake()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdClaimUndelegatedTokens(t *testing.T) {
	args := []string{
		"[host-zone]",
		"[epoch]",
		"[receiver]",
	}

	cmd := cli.CmdClaimUndelegatedTokens()
	ExecuteCLIExpectError(t, cmd, args, `unable to cast "[epoch]" of type string to uint64`)
}

func TestCmdRebalanceValidators(t *testing.T) {
	args := []string{
		"[host-zone]",
		"[num-to-rebalance]",
	}

	cmd := cli.CmdRebalanceValidators()
	ExecuteCLIExpectError(t, cmd, args, `strconv.ParseUint: parsing "[num-to-rebalance]": invalid syntax`)
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

func TestCmdClearBalance(t *testing.T) {
	args := []string{
		"[chain-id]",
		"[amount]",
		"[channel-id]",
	}

	cmd := cli.CmdClearBalance()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
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

func TestCmdSetCommunityPoolRebate(t *testing.T) {
	t.Run("invalid rebate-rate", func(t *testing.T) {
		args := []string{
			"[chain-id]",
			"[rebate-rate]",
			"[liquid-staked-sttoken-amount]",
		}

		cmd := cli.CmdSetCommunityPoolRebate()
		ExecuteCLIExpectError(t, cmd, args, `unable to parse rebate percentage: failed to set decimal string with base 10: [rebate-rate]000000000000000000`)
	})
	t.Run("invalid liquid-staked-sttoken-amount", func(t *testing.T) {
		args := []string{
			"[chain-id]",
			"0.123456789",
			"[liquid-staked-sttoken-amount]",
		}

		cmd := cli.CmdSetCommunityPoolRebate()
		ExecuteCLIExpectError(t, cmd, args, `unable to parse liquid stake amount`)
	})
}

func TestCmdToggleTradeController(t *testing.T) {
	args := []string{
		"[trade-chain-id]",
		"[grant|revoke]",
		"[address]",
	}

	cmd := cli.CmdToggleTradeController()
	ExecuteCLIExpectError(t, cmd, args, `invalid permission change, must be either 'grant' or 'revoke'`)
}
