package cli

import (
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

// string creator = 1;
//   string hostZoneId = 2;
//   uint64 epoch = 3;
//   string sender = 4;
// }

func CmdClaimUndelegatedTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-undelegated-tokens [host-zone] [epoch] [sender]",
		Short: "Broadcast message claimUndelegatedTokens",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argEpoch, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}
			argSender := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimUndelegatedTokens(
				clientCtx.GetFromAddress().String(),
				argHostZone,
				argEpoch,
				argSender,
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
