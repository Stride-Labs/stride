package cli

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

type ValidatorWeightList struct {
	ValidatorWeights []*types.ValidatorWeight `json:"validator_weights,omitempty"`
}

// Parse a JSON with a list of validators in the format
//
//	{
//		  "validator_weights": [
//		     {"address": "cosmosXXX", "weight": 1},
//			 {"address": "cosmosXXX", "weight": 2}
//	   ]
//	}
func parseChangeValidatorWeightsFile(validatorsFile string) (weights []*types.ValidatorWeight, err error) {
	fileContents, err := os.ReadFile(validatorsFile)
	if err != nil {
		return weights, err
	}

	var weightsList ValidatorWeightList
	if err = json.Unmarshal(fileContents, &weightsList); err != nil {
		return weights, err
	}

	return weightsList.ValidatorWeights, nil
}

// Updates the weight for a single validator
func CmdChangeValidatorWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-validator-weight [host-zone] [address] [weight]",
		Short: "Broadcast message change-validator-weight to update the weight for a single validator",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			valAddress := args[1]
			weight, err := cast.ToUint64E(args[2])
			if err != nil {
				return err
			}
			weights := []*types.ValidatorWeight{
				{
					Address: valAddress,
					Weight:  weight,
				},
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeValidatorWeights(
				clientCtx.GetFromAddress().String(),
				hostZone,
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

// Updates the weight for multiple validators
//
// Accepts a file in the following format:
//
//	{
//		"validator_weights": [
//		     {"address": "cosmosXXX", "weight": 1},
//			 {"address": "cosmosXXX", "weight": 2}
//	    ]
//	}
func CmdChangeMultipleValidatorWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-validator-weights [host-zone] [validator-weight-file]",
		Short: "Broadcast message change-validator-weights to update the weights for multiple validators",
		Long: strings.TrimSpace(
			`Changes multiple validator weights at once, using a JSON file in the following format
	{
		"validator_weights": [
			{"address": "cosmosXXX", "weight": 1},
			{"address": "cosmosXXX", "weight": 2}
		]
	}	
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			validatorWeightChangeFile := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			weights, err := parseChangeValidatorWeightsFile(validatorWeightChangeFile)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeValidatorWeights(
				clientCtx.GetFromAddress().String(),
				hostZone,
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
