package cli

// import (
// 	"strconv"

// 	"github.com/Stride-Labs/stride/x/interchainquery/types"
// 	"github.com/cosmos/cosmos-sdk/client"
// 	"github.com/cosmos/cosmos-sdk/client/flags"
// 	"github.com/cosmos/cosmos-sdk/client/tx"
// 	"github.com/spf13/cobra"
// )

// var _ = strconv.Itoa(0)

// // TODO(TEST-78) rename from query-XXX => update-XXX
// func CmdQueryExchangerate() *cobra.Command {
// 	cmd := &cobra.Command{
// 		// TODO(TEST-77) camel case
// 		Use:   "query-exchangerate [chain-id]",
// 		Short: "Broadcast message query-exchangerate",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) (err error) {
// 			argChainID := args[0]

// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			msg := types.NewMsgQueryExchangerate(
// 				clientCtx.GetFromAddress().String(),
// 				argChainID,
// 			)
// 			if err := msg.ValidateBasic(); err != nil {
// 				return err
// 			}
// 			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
// 		},
// 	}

// 	flags.AddTxFlagsToCmd(cmd)

// 	return cmd
// }
