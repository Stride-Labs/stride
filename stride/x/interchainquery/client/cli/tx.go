package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"

	// "github.com/cosmos/interchain-accounts/x/inter-tx/types"
	"github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/spf13/cobra"
)

// GetTxCmd creates and returns the intertx tx command
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryBalanceCmd(),
		SubmitQueryResponse(),
	)

	return cmd
}

func QueryBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-balance [chain_id] [address] [denom]",
		Short: `Query the balance on a chain.`,
		Long: `query a specified account's balance of a specified denomination on a specified chain
		e.g. "GAIA_1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 uatom"`,
		Example: `query-balance GAIA_1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 uatom`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			chain_id := args[0]
			address := args[1]
			denom := args[2]
			caller := clientCtx.GetFromAddress().String()

			// TODO cleanup
			if len(caller) < 1 {
				return fmt.Errorf("Error: empty --from address.")
			}
			fmt.Println(caller)

			// TODO(TEST-50) create message based on parsed json
			// msg, err := types.NewMsgSubmitTx(txMsg, viper.GetString(FlagConnectionID), clientCtx.GetFromAddress().String())
			msg := types.NewQueryBalance(chain_id, address, denom, viper.GetString(FlagConnectionID), caller)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// TODO what do these do? require a connection flag when submitting the command?
	cmd.Flags().AddFlagSet(fsConnectionID)
	_ = cmd.MarkFlagRequired(FlagConnectionID)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func SubmitQueryResponse() *cobra.Command {
	cmd := &cobra.Command{
		// "@type": "/stride.interchainquery.MsgSubmitQueryResponse",
		// "chain_id": "GAIA_1",
		// "query_id": "8e3451aea1ca8438f4ba9292a3add814d50d45d163ff28f859836cbb074584c0",
		// "result": "ChUKBXVhdG9tEgw0OTkwMDAwMDAwMDASAhAB",
		// "height": "40",
		// "from_address": "stride1wlgadk2gndm96tvf0v6207jckqu8e2huyfhsp5"
		Use:   "submitqueryresponse [chain_id] [query_id] [result] [height] [from_address]",
		Short: `Submit Query Response.`,
		Long: `
		e.g. "submitqueryresponse GAIA_1 8e3451aea1ca8438f4ba9292a3add814d50d45d163ff28f859836cbb074584c0 ChUKBXVhdG9tEgw0OTkwMDAwMDAwMDASAhAB 40 stride1wlgadk2gndm96tvf0v6207jckqu8e2huyfhsp5"`,
		Example: `submitqueryresponse GAIA_1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 uatom`,
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			chain_id := args[0]
			// query_id := args[1]
			result := args[2]
			// height := args[3]
			from_address := args[4]
			from_addr_sdk, _ := sdk.AccAddressFromBech32(from_address)
			caller := clientCtx.GetFromAddress().String()

			// TODO cleanup
			if len(caller) < 1 {
				return fmt.Errorf("Error: empty --from address.")
			}
			fmt.Println(caller)

			// TODO(TEST-50) create message based on parsed json
			// msg, err := types.NewMsgSubmitTx(txMsg, viper.GetString(FlagConnectionID), clientCtx.GetFromAddress().String())
			msg := types.NewMsgSubmitQueryResponse(chain_id, result, from_addr_sdk)
			//  viper.GetString(FlagConnectionID), caller)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
