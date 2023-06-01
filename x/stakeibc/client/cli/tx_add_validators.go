package cli

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type ValidatorsList struct {
	Validators []*types.Validator `json:"validators,omitempty"`
}

// Parse a JSON with a list of validators in the format
// {
//	  "validators": [
//	     {"name": "val1", "address": "cosmosXXX", "weight": 1},
//		 {"name": "val2", "address": "cosmosXXX", "weight": 2}
//    ]
// }
func parseAddValidatorsFile(validatorsFile string) (validators ValidatorsList, err error) {
	fileContents, err := os.ReadFile(validatorsFile)
	if err != nil {
		return validators, err
	}

	if err = json.Unmarshal(fileContents, &validators); err != nil {
		return validators, err
	}

	return validators, nil
}

func CmdAddValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validators [host-zone] [validator-list-file]",
		Short: "Broadcast message add-validators",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			validatorListProposalFile := args[1]

			validators, err := parseAddValidatorsFile(validatorListProposalFile)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddValidators(
				clientCtx.GetFromAddress().String(),
				hostZone,
				validators.Validators,
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
