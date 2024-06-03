package cli

import (
	"encoding/json"
	"os"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
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
