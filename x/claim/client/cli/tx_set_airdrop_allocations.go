package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

func CmdSetAirdropAllocations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-airdrop-allocations [airdrop-identifier] [user-addresses] [user-weights]",
		Short: "Broadcast message set-airdrop-allocations",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddresses := strings.Split(args[1], ",")
			argWeights := strings.Split(args[2], ",")
			weights := []sdk.Dec{}

			for _, weight := range argWeights {
				weightDec, err := sdk.NewDecFromStr(weight)
				if err != nil {
					return err
				}
				weights = append(weights, weightDec)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetAirdropAllocations(
				clientCtx.GetFromAddress().String(),
				args[0],
				argAddresses,
				weights,
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
