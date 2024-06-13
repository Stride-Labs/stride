package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/client/cli"
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
			"[connection-id]", "[host-denom]", "[bech32prefix]", "[ibc-denom]", "[channel-id]", "[unbonding-period]", "1",
		}

		cmd := cli.CmdRegisterHostZone()
		ExecuteCLIExpectError(t, cmd, args, `strconv.ParseUint: parsing "[unbonding-period]": invalid syntax`)
	})

	t.Run("lsm-enabled not a boolean", func(t *testing.T) {
		args := []string{
			"[connection-id]", "[host-denom]", "[bech32prefix]", "[ibc-denom]", "[channel-id]", "0", "2",
		}

		cmd := cli.CmdRegisterHostZone()
		ExecuteCLIExpectError(t, cmd, args, `strconv.ParseBool: parsing "2": invalid syntax`)
	})
}

func TestCmdRedeemStake(t *testing.T) {
	args := []string{
		"[amount]", "[hostZoneID]", "[receiver]",
	}

	cmd := cli.CmdRedeemStake()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdClaimUndelegatedTokens(t *testing.T) {
	args := []string{
		"[host-zone]", "[epoch]", "[receiver]",
	}

	cmd := cli.CmdClaimUndelegatedTokens()
	ExecuteCLIExpectError(t, cmd, args, `unable to cast "[epoch]" of type string to uint64`)
}

func TestCmdRebalanceValidators(t *testing.T) {
	args := []string{
		"[host-zone]", "[num-to-rebalance]",
	}

	cmd := cli.CmdRebalanceValidators()
	ExecuteCLIExpectError(t, cmd, args, `strconv.ParseUint: parsing "[num-to-rebalance]": invalid syntax`)
}

func TestCmdAddValidators(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		args := []string{
			"[host-zone]", "[validator-list-file]",
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `open [validator-list-file]: no such file or directory`)
	})
	t.Run("empty file", func(t *testing.T) {
		// pass an temp file
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()

		args := []string{
			"[host-zone]", f.Name(),
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `unexpected end of JSON input`)
	})
	t.Run("non-JSON file", func(t *testing.T) {
		// pass a temp file with non-JSON content
		f, err := os.CreateTemp("", "")
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString("This is not JSON")
		require.NoError(t, err)

		args := []string{
			"[host-zone]", f.Name(),
		}

		cmd := cli.CmdAddValidators()
		ExecuteCLIExpectError(t, cmd, args, `invalid character 'T' looking for beginning of value`)
	})
}

func TestCmdChangeValidatorWeight(t *testing.T) {
	args := []string{
		"[host-zone]", "[address]", "[weight]",
	}

	cmd := cli.CmdChangeValidatorWeight()
	ExecuteCLIExpectError(t, cmd, args, `unable to cast "[weight]" of type string to uint64`)
}

func TestCmdChangeMultipleValidatorWeight(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdChangeMultipleValidatorWeight()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdDeleteValidator(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdDeleteValidator()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdRestoreInterchainAccount(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdRestoreInterchainAccount()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdUpdateValidatorSharesExchRate(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdUpdateValidatorSharesExchRate()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdCalibrateDelegation(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdCalibrateDelegation()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdClearBalance(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdClearBalance()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdUpdateInnerRedemptionRateBounds(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdUpdateInnerRedemptionRateBounds()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdResumeHostZone(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdResumeHostZone()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdSetCommunityPoolRebate(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdSetCommunityPoolRebate()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}

func TestCmdToggleTradeController(t *testing.T) {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdToggleTradeController()
	ExecuteCLIExpectError(t, cmd, args, `can not convert string to int: invalid type`)
}
