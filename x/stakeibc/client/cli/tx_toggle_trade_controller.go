package cli

import (
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func CmdToggleTradeController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle-trade-controller [grant|revoke] [address]",
		Short: "Submits an ICA tx to grant or revoke permissions to trade on behalf of the trade ICA",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			permissionChangeString := args[0]
			address := args[1]

			permissionChangeInt, ok := types.AuthzPermissionChange_value[strings.ToUpper(permissionChangeString)]
			if !ok {
				return errors.New("invalid permission change, must be either 'grant' or 'revoke'")
			}
			permissionChange := types.AuthzPermissionChange(permissionChangeInt)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgToggleTradeController(
				clientCtx.GetFromAddress().String(),
				permissionChange,
				address,
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
