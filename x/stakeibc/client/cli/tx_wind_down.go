// x/stakeibc/client/cli/tx_wind_down.go
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Wind-down admin txs (spec §7). The helpers are shared by every command; each Cmd* function
// below is one command.

// validatorUndelegationInput is one entry of the validators file for undelegate-from-validators
type validatorUndelegationInput struct {
	Address string `json:"address"`
	Offset  string `json:"offset"` // base units, optional, default "0"
}

// ReadValidatorUndelegations parses the JSON validators file: [{"address": "...", "offset": "0"}, ...]
func ReadValidatorUndelegations(path string) ([]types.ValidatorUndelegation, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inputs []validatorUndelegationInput
	if err := json.Unmarshal(contents, &inputs); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to parse validators file: %s", err)
	}

	validators := make([]types.ValidatorUndelegation, 0, len(inputs))
	for _, input := range inputs {
		offset := sdkmath.ZeroInt()
		if input.Offset != "" {
			parsed, err := ParseBaseUnits(input.Offset)
			if err != nil {
				return nil, errorsmod.Wrapf(err, "offset for %s", input.Address)
			}
			offset = parsed
		}
		validators = append(validators, types.ValidatorUndelegation{Address: input.Address, Offset: offset})
	}
	return validators, nil
}

// ParseIcaType accepts DELEGATION, WITHDRAWAL, FEE or REDEMPTION in any case
func ParseIcaType(arg string) (types.ICAAccountType, error) {
	value, found := types.ICAAccountType_value[strings.ToUpper(arg)]
	if !found {
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unknown ica type %s", arg)
	}
	return types.ICAAccountType(value), nil
}

// ParseBaseUnits parses an integer amount of base units
func ParseBaseUnits(arg string) (sdkmath.Int, error) {
	parsed, found := sdkmath.NewIntFromString(arg)
	if !found {
		return sdkmath.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "can not convert string to int: %q", arg)
	}
	return parsed, nil
}

// broadcastWindDownTx builds the message with the --from address, validates it and broadcasts
func broadcastWindDownTx(cmd *cobra.Command, build func(creator string) sdk.Msg) error {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return err
	}
	msg := build(clientCtx.GetFromAddress().String())
	if err := msg.(interface{ ValidateBasic() error }).ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

const FlagUndelegateAll = "all"

// CmdUndelegateFromValidators: the wind-down drain.
func CmdUndelegateFromValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-from-validators [chain-id] [validators-file]",
		Short: "Wind-down: undelegate the recorded delegation from the validators in the file, or from every validator with --all",
		Long: `Submits MsgUndelegateFromValidators (admin only). With a file, only the listed validators
are drained, each for its recorded delegation minus the offset; the file is a JSON list:
  [{"address": "cosmosvaloper1...", "offset": "0"}, ...]
With --all and no file, every validator with a recorded delegation is drained in full. One of
the two is required; the full drain is never the default. An empty file is rejected.

Note for hand-written JSON txs (--generate-only, multisig): an omitted or empty "validators"
field means a full drain of every validator with a recorded delegation.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argChainId := args[0]
			drainAll, err := cmd.Flags().GetBool(FlagUndelegateAll)
			if err != nil {
				return err
			}
			hasFile := len(args) == 2
			if drainAll && hasFile {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pass a validators file or --all, not both")
			}
			if !drainAll && !hasFile {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pass a validators file or --all")
			}

			validators := []types.ValidatorUndelegation{}
			if hasFile {
				validators, err = ReadValidatorUndelegations(args[1])
				if err != nil {
					return err
				}
				// An empty list in the message means a full drain, so an empty file must not reach it
				if len(validators) == 0 {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validators file is empty; use --all for a full drain")
				}
			}

			return broadcastWindDownTx(cmd, func(creator string) sdk.Msg {
				return types.NewMsgUndelegateFromValidators(creator, argChainId, validators)
			})
		},
	}

	cmd.Flags().Bool(FlagUndelegateAll, false, "drain every validator with a recorded delegation (instead of a validators file)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTransferFromIca: ICA balance to the Osmosis vault.
func CmdTransferFromIca() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-from-ica [chain-id] [ica-type] [amount]",
		Short: "Wind-down: transfer an ICA balance to the Osmosis vault",
		Long: `Submits MsgTransferFromIca (admin only). ica-type is one of DELEGATION, WITHDRAWAL, FEE,
REDEMPTION; amount is a coin in the denom as it exists on the host (e.g. 1000000uatom). The
receiver and channel are hard-coded in the binary.

Note for hand-written JSON txs (--generate-only, multisig): DELEGATION is the enum's zero
value, so a message that omits ica_type moves the DELEGATION ICA. Always set ica_type
explicitly and check it in the unsigned tx before signing.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argChainId := args[0]
			icaType, err := ParseIcaType(args[1])
			if err != nil {
				return err
			}
			argAmount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount %q: %w", args[2], err)
			}

			return broadcastWindDownTx(cmd, func(creator string) sdk.Msg {
				return types.NewMsgTransferFromIca(creator, argChainId, icaType, argAmount)
			})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTransferStaketiaClaimBalance: claim-address TIA to the celestia delegation ICA.
func CmdTransferStaketiaClaimBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-staketia-claim-balance [amount]",
		Short: "Wind-down: move the staketia claim address's TIA to the celestia delegation ICA",
		Long: `Submits MsgTransferStaketiaClaimBalance (admin only). amount is in utia and is optional:
omitted or 0 moves the whole balance. Use a small amount first as the live test.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount := sdkmath.ZeroInt()
			if len(args) == 1 {
				parsed, err := ParseBaseUnits(args[0])
				if err != nil {
					return err
				}
				amount = parsed
			}

			return broadcastWindDownTx(cmd, func(creator string) sdk.Msg {
				return types.NewMsgTransferStaketiaClaimBalance(creator, amount)
			})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdSweepTokensOffStride submits one sweep batch: every listed denom, for every holder in the
// file (one bech32 address per line; blank lines and lines starting with # are ignored).
// The file is produced by scripts/wind-down/build_sweep_batches.py
func CmdSweepTokensOffStride() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep-tokens-off-stride [denoms] [addresses-file]",
		Short: "Sweep the listed denoms off Stride for every holder in the file (sweep operator only)",
		Long: `Sends each listed denom that each holder in the file owns to the holder's own address on the
destination chain: stTokens and ustrd to Osmosis over channel-5, IBC vouchers back one hop over the
channel they arrived on (whitelisted channels only). denoms is comma-separated. The file holds one
Stride address per line. There is no on-chain cap on the number of addresses: the tx is atomic, so
a batch too large for the block gas limit fails as a whole (build_sweep_batches.py defaults to 100).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denoms := parseCommaSeparated(args[0])
			if len(denoms) == 0 {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one denom is required")
			}
			addresses, err := readAddressesFile(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgSweepTokensOffStride(clientCtx.GetFromAddress().String(), denoms, addresses)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseCommaSeparated(raw string) []string {
	values := []string{}
	for _, value := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func readAddressesFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to read addresses file %s: %s", path, err)
	}
	defer file.Close()

	addresses := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		addresses = append(addresses, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to read addresses file %s: %s", path, err)
	}
	if len(addresses) == 0 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "addresses file is empty: %s", path)
	}
	return addresses, nil
}
