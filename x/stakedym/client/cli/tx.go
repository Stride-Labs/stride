package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v31/x/stakedym/types"
)

const (
	ArgIncrease          = "increase"
	ArgDecrease          = "decrease"
	RecordTypeDelegation = "delegation"
	RecordTypeUnbonding  = "unbonding"
	RecordTypeRedemption = "redemption"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdLiquidStake(),
		CmdRedeemStake(),
		CmdConfirmDelegation(),
		CmdConfirmUndelegation(),
		CmdConfirmUnbondedTokensSwept(),
		CmdAdjustDelegatedBalance(),
		CmdUpdateInnerRedemptionRateBounds(),
		CmdResumeHostZone(),
		CmdOverwriteRecord(),
		CmdRefreshRedemptionRate(),
		CmdSetOperatorAddress(),
	)

	return cmd
}

// User transaction to liquid stake native tokens into stTokens
func CmdLiquidStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-stake [amount]",
		Short: "Liquid stakes native tokens and receives stTokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid stakes native tokens and receives stTokens

Example:
  $ %[1]s tx %[2]s liquid-stake 10000
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, ok := sdkmath.NewIntFromString(args[0])
			if !ok {
				return errors.New("unable to parse amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidStake(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// User transaction to redeem stake stTokens into native tokens
func CmdRedeemStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-stake [amount]",
		Short: "Redeems stTokens tokens for native tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redeems stTokens tokens for native tokens. 
Native tokens will land in the redeeming address after they unbond

Example:
  $ %[1]s tx %[2]s redeem-stake 10000
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, ok := sdkmath.NewIntFromString(args[0])
			if !ok {
				return errors.New("unable to parse amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRedeemStake(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Operator transaction to confirm an delegation was submitted on the host chain
func CmdConfirmDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-delegation [record-id] [tx-hash]",
		Short: "Confirms that an delegation tx was submitted",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Confirms that a delegation tx was submitted on the host zone
The recordId corresponds with the delegation record, and the tx hash is the hash from the undelegation tx itself (used for logging purposes)

Example:
  $ %[1]s tx %[2]s confirm-delegation 100 XXXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			recordId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			txHash := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgConfirmDelegation(
				clientCtx.GetFromAddress().String(),
				recordId,
				txHash,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Operator transaction to confirm an undelegation was submitted on the host chain
func CmdConfirmUndelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-undelegation [record-id] [tx-hash]",
		Short: "Confirms that an undelegation tx was submitted",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Confirms that an undelegation tx was submitted on the host zone
The recordId corresponds with the unbonding record, and the tx hash is the hash from the undelegation tx itself (used for logging purposes)

Example:
  $ %[1]s tx %[2]s confirm-undelegation 100 XXXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			recordId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			txHash := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgConfirmUndelegation(
				clientCtx.GetFromAddress().String(),
				recordId,
				txHash,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Operator transaction to confirm unbonded tokens were transferred back to stride
func CmdConfirmUnbondedTokensSwept() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-sweep [record-id] [tx-hash]",
		Short: "Confirms that unbonded tokens were swept back to stride",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Confirms unbonded tokens were transferred back from the host zone to stride.
The recordId corresponds with the unbonding record, and the tx hash is the hash from the ibc-transfer tx itself (used for logging purposes)

Example:
  $ %[1]s tx %[2]s confirm-sweep 100 XXXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			recordId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			txHash := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgConfirmUnbondedTokenSweep(
				clientCtx.GetFromAddress().String(),
				recordId,
				txHash,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Operator transaction to adjust the delegated balance after a validator was slashed
func CmdAdjustDelegatedBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adjust-delegated-balance [increase|decrease] [delegation-offset] [validator]",
		Short: "Adjust the host zone delegated balance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Adjust the host zone's delegated balance and logs the validator in a slash record.
Note: You must specify whether the delegation should increase or decrease

Example:
  $ %[1]s tx %[2]s adjust-delegated-balance decrease 100000 XXXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			direction := args[0]
			delegationOffset, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return errors.New("unable to parse delegation offset")
			}
			validatorAddress := args[2]

			// Make the offset negative if the intention is to decrease the amount
			if direction == ArgDecrease {
				delegationOffset = delegationOffset.Neg()
			} else if direction != ArgIncrease {
				return fmt.Errorf("invalid direction specified, must be either %s or %s", ArgIncrease, ArgDecrease)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAdjustDelegatedBalance(
				clientCtx.GetFromAddress().String(),
				delegationOffset,
				validatorAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Adjusts the inner redemption rate bounds on the host zone
func CmdUpdateInnerRedemptionRateBounds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-redemption-rate-bounds [min-bound] [max-bound]",
		Short: "Sets the inner redemption rate bounds",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Sets the inner redemption rate bounds on a host zone

Example:
  $ %[1]s tx %[2]s set-redemption-rate-bounds 1.10 1.20
`, version.AppName, types.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			minInnerRedemptionRate := sdkmath.LegacyMustNewDecFromStr(args[0])
			maxInnerRedemptionRate := sdkmath.LegacyMustNewDecFromStr(args[1])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateInnerRedemptionRateBounds(
				clientCtx.GetFromAddress().String(),
				minInnerRedemptionRate,
				maxInnerRedemptionRate,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Unhalts the host zone if redemption rates were exceeded
func CmdResumeHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume-host-zone",
		Short: "Resumes a host zone after a halt",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Resumes a host zone after it was halted

Example:
  $ %[1]s tx %[2]s resume-host-zone
`, version.AppName, types.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgResumeHostZone(
				clientCtx.GetFromAddress().String(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// SAFE multisig overwrites record
func CmdOverwriteRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overwrite-record [delegation|unbonding|redemption] [json-file]",
		Short: "overwrites a record",
		Long: strings.TrimSpace(
			fmt.Sprint(`Submit an overwrite record tx. The record must be supplied via a JSON file.
			
Example:
$ tx stakedym overwrite-record [delegation|unbonding|redemption] <path/to/file.json> --from=<key_or_address>

Where file.json contains either...

Delegation Record (recordtype=delegation)
{
	"id": "4",
	"native_amount": "100",
	"status": "DELEGATION_QUEUE",
	"tx_hash": "C8C3CFF223CF4711E14F3E3918A3E82ED8BAA010445A4519BD0B2AFDB45897FE"
}

Unbonding Record (recordtype=unbonding)
{
	"id": "4",
	"native_amount": "100",
	"st_token_amount": "94",
	"UnbondingRecordStatus": "UNBONDING_QUEUE",
	"unbonding_completion_time": "1705802815"
	"undelegation_tx_hash": "C8C3CFF223CF4711E14F3E3918A3E82ED8BAA010445A4519BD0B2AFDB45897FE",
	"unbonding_token_swap_tx_hash": "C8C3CFF223CF4711E14F3E3918A3E82ED8BAA010445A4519BD0B2AFDB45897FE"
}

Redemption Record (recordtype=redemption)
{
	"unbonding_record_id": "4"
	"native_amount": "100",
	"st_token_amount": "107",
	"redeemer": "stride1zlu2l3lx5tqvzspvjwsw9u0e907kelhqae3yhk"
}
			
			`, version.AppName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			recordType := args[0]
			recordContents := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			switch recordType {
			case RecordTypeDelegation:
				return parseAndBroadcastOverwriteDelegation(clientCtx, cmd, recordContents)
			case RecordTypeUnbonding:
				return parseAndBroadcastOverwriteUnbonding(clientCtx, cmd, recordContents)
			case RecordTypeRedemption:
				return parseAndBroadcastOverwriteRedemption(clientCtx, cmd, recordContents)
			default:
				return fmt.Errorf("invalid record type specified, must be either %s, %s, or %s", RecordTypeDelegation, RecordTypeUnbonding, RecordTypeRedemption)
			}
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// triggers the redemption rate update
func CmdRefreshRedemptionRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-update-redemption-rate",
		Short: "triggers an update to the redemption rate",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Triggers an updated redemption rate calculation for the host zone
			
Example:
$ %[1]s tx %[2]s trigger-update-redemption-rate
			`, version.AppName, types.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRefreshRedemptionRate(
				clientCtx.GetFromAddress().String(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// triggers the redemption rate update
func CmdSetOperatorAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-operator-address [operator-address]",
		Short: "sets the operator address on the host zone record",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Triggers an updated redemption rate calculation for the host zone
			
Example:
$ %[1]s tx %[2]s setup-operator-address stride1265uqtckmd3kt7jek2pv0vrp04j0d74jj8ahq5
			`, version.AppName, types.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			operatorAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetOperatorAddress(
				clientCtx.GetFromAddress().String(),
				operatorAddress,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
