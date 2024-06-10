package cli_test

import (
	"testing"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/client/cli"
)

func (s *ClientTestSuite) TestCmdLiquidStake() {
	args := []string{
		"banana",
		"[lsm-token-denom]",
	}

	cmd := cli.CmdLiquidStake()
	s.ExecuteCLIExpectError(cmd, args, "can not convert string to int")
}

func (s *ClientTestSuite) TestCmdLSMLiquidStake() {
	args := []string{
		"banana",
		"[lsm-token-denom]",
	}

	cmd := cli.CmdLSMLiquidStake()
	s.ExecuteCLIExpectError(cmd, args, "can not convert string to int")
}

func (s *ClientTestSuite) TestCmdRegisterHostZone() {
	s.T().Run("unbonding-period not a number", func(t *testing.T) {
		args := []string{
			"[connection-id]", "[host-denom]", "[bech32prefix]", "[ibc-denom]", "[channel-id]", "banana", "1",
		}

		cmd := cli.CmdRegisterHostZone()
		s.ExecuteCLIExpectError(cmd, args, `strconv.ParseUint: parsing "banana": invalid syntax`)
	})

	s.T().Run("lsm-enabled not a boolean", func(t *testing.T) {
		args := []string{
			"[connection-id]", "[host-denom]", "[bech32prefix]", "[ibc-denom]", "[channel-id]", "0", "2",
		}

		cmd := cli.CmdRegisterHostZone()
		s.ExecuteCLIExpectError(cmd, args, `strconv.ParseBool: parsing "2": invalid syntax`)
	})
}

func (s *ClientTestSuite) TestCmdRedeemStake() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdRedeemStake()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdClaimUndelegatedTokens() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdClaimUndelegatedTokens()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdRebalanceValidators() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdRebalanceValidators()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdAddValidators() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdAddValidators()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdChangeValidatorWeight() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdChangeValidatorWeight()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdChangeMultipleValidatorWeight() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdChangeMultipleValidatorWeight()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdDeleteValidator() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdDeleteValidator()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdRestoreInterchainAccount() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdRestoreInterchainAccount()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdUpdateValidatorSharesExchRate() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdUpdateValidatorSharesExchRate()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdCalibrateDelegation() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdCalibrateDelegation()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdClearBalance() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdClearBalance()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdUpdateInnerRedemptionRateBounds() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdUpdateInnerRedemptionRateBounds()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdResumeHostZone() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdResumeHostZone()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdSetCommunityPoolRebate() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdSetCommunityPoolRebate()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}

func (s *ClientTestSuite) TestCmdToggleTradeController() {
	args := []string{
		"1",
		"utia",
	}

	cmd := cli.CmdToggleTradeController()
	s.ExecuteTxAndCheckSuccessful(cmd, args)
}
