package cli

import (
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

func CmdLiquidVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-vote [host-zone-id] [proposal-id] [amount] [vote-option]",
		Short: "Submit a liquid vote through stride to another hub",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			host_zone_id := args[0]
			proposal_id, _ := strconv.ParseUint(args[1], 10, 64)
			amount, _ := sdk.NewIntFromString(args[2])
			vote_option := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vote := govtypes.OptionAbstain;
			if strings.ToUpper(vote_option) == "YES" {
				vote = govtypes.OptionYes
			} else if strings.ToUpper(vote_option) == "NO" {
				vote = govtypes.OptionNo
			} else if strings.ToUpper(vote_option) == "VETO" {
				vote = govtypes.OptionNoWithVeto
			}
	
			msg := types.NewMsgLiquidVote(
				clientCtx.GetFromAddress().String(),
				host_zone_id,
				proposal_id,
				amount,
				vote,
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
