// x/stakeibc/client/cli/tx_wind_down.go
package cli

import (
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
// below is a stub that its task replaces (Task 3: undelegate, Task 4: transfer-from-ica,
// Task 5: transfer-staketia-claim-balance; PR 5 adds sweep-tokens-off-stride).

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

// notWiredYet is the RunE of every stub below
func notWiredYet(name string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("%s is wired in a later task of this PR", name)
	}
}

// CmdUndelegateFromValidators: the wind-down drain (Task 3 replaces this function).
func CmdUndelegateFromValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-from-validators [chain-id]",
		Short: "Wind-down: undelegate the recorded delegation from validators",
		RunE:  notWiredYet("undelegate-from-validators"),
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTransferFromIca: ICA balance to the Osmosis vault (Task 4 replaces this function).
func CmdTransferFromIca() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-from-ica [chain-id] [ica-type] [amount]",
		Short: "Wind-down: transfer an ICA balance to the Osmosis vault",
		RunE:  notWiredYet("transfer-from-ica"),
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTransferStaketiaClaimBalance: claim-address TIA to the celestia delegation ICA (Task 5
// replaces this function).
func CmdTransferStaketiaClaimBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-staketia-claim-balance [amount]",
		Short: "Wind-down: move the staketia claim address's TIA to the celestia delegation ICA",
		RunE:  notWiredYet("transfer-staketia-claim-balance"),
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
