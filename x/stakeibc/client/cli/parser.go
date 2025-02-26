package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

type ValidatorsList struct {
	Validators []*types.Validator `json:"validators,omitempty"`
}

type ValidatorWeightList struct {
	ValidatorWeights []*types.ValidatorWeight `json:"validator_weights,omitempty"`
}

// Parse a JSON with a list of validators in the format
//
//	{
//		  "validators": [
//		     {"name": "val1", "address": "cosmosXXX", "weight": 1},
//			 {"name": "val2", "address": "cosmosXXX", "weight": 2}
//	   ]
//	}
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

func parseAddValidatorsProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.AddValidatorsProposal, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	proposal.Title = fmt.Sprintf("Add validators to %s", proposal.HostZone)

	return proposal, nil
}

func parseToggleLSMProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.ToggleLSMProposal, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	action := "Disable"
	if proposal.Enabled {
		action = "Enable"
	}
	proposal.Title = fmt.Sprintf("%s LSMLiquidStakes for %s", action, proposal.HostZone)

	return proposal, nil
}
