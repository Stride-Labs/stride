package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func CmdNextPacketSequence() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-packet-sequence [channel-id] [port-id]",
		Short: "returns the next packet sequence on a channel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			channelId := args[0]
			portId := args[1]

			params := &types.QueryGetNextPacketSequenceRequest{ChannelId: channelId, PortId: portId}

			res, err := queryClient.NextPacketSequence(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
