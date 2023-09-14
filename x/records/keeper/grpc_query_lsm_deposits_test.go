package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/x/records/types"
)

func (s *KeeperTestSuite) TestLSMDeposit() {
	// setup expected deposit in stakeibckeeper
	initToken := types.LSMTokenDeposit{ChainId: "1", Denom: "validator70027"}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, initToken)

	// input nil, no chain-id, no denom --> error invalid input
	invalidInputs := []*types.QueryLSMDepositRequest{}
	noChainInput := types.QueryLSMDepositRequest{ChainId: "", Denom: "validator93740"}
	noDenomInput := types.QueryLSMDepositRequest{ChainId: "42", Denom: ""}
	invalidInputs = append(invalidInputs, nil, &noChainInput, &noDenomInput)

	for _, invalidInput := range invalidInputs {
		_, err1 := s.App.RecordsKeeper.LSMDeposit(s.Ctx, invalidInput)
		s.Require().ErrorContains(err1, "invalid request")
	}

	// no matching deposit found --> error not found
	missingInput := types.QueryLSMDepositRequest{ChainId: "2", Denom: "validator9374999"}
	_, err2 := s.App.RecordsKeeper.LSMDeposit(s.Ctx, &missingInput)
	s.Require().ErrorContains(err2, "LSM deposit not found")

	// found the deposit --> no error, deposit returned with matching chain-id and denom
	expectedInput := types.QueryLSMDepositRequest{ChainId: "1", Denom: "validator70027"}
	response, err3 := s.App.RecordsKeeper.LSMDeposit(s.Ctx, &expectedInput)
	s.Require().NoError(err3)
	s.Require().Equal("1", response.Deposit.ChainId)
	s.Require().Equal("validator70027", response.Deposit.Denom)
}

func (s *KeeperTestSuite) TestLSMDeposits() {
	// setup expected desposits in stakeibckeeper
	chainIds := []string{"1"}
	validators := []string{"validator22313", "validator30472"}
	statuses := []string{"TRANSFER_IN_PROGRESS", "TRANSFER_FAILED", "DETOKENIZATION_QUEUE"}
	for _, chainId := range chainIds {
		for _, validator := range validators {
			for _, statusStr := range statuses {
				denom := chainId + validator + statusStr // has to be present and unique for each token
				status := types.LSMTokenDeposit_Status(types.LSMTokenDeposit_Status_value[statusStr])
				initToken := types.LSMTokenDeposit{ChainId: chainId, ValidatorAddress: validator, Status: status, Denom: denom}
				s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, initToken)
			}
		}
	}

	// input nil --> error invalid input
	_, err1 := s.App.RecordsKeeper.LSMDeposit(s.Ctx, nil)
	s.Require().ErrorContains(err1, "invalid request")

	// Adding case where string is empty "" meaning match all
	// Adding case where string is "missing_X" which is an example which is not found
	chainIds = append(chainIds, "", "missing_chain")
	validators = append(validators, "", "missing_validator")
	statuses = append(statuses, "", "missing_status")
	for _, chainId := range chainIds {
		chainMatchNum := 1 // case: chain-id is a specific value with matching deposits
		if chainId == "" { // case: chain-id filter not applied, all len-2 chain-ids match
			chainMatchNum = len(chainIds) - 2
		}
		if chainId == "missing_chain" { // case: chain-id is specific value matching 0 deposits
			chainMatchNum = 0
		}
		for _, validator := range validators {
			validatorMatchNum := 1 // case: validator is a specific value with matching deposits
			if validator == "" {   // case: validator filter not applied, all len-2 validators match
				validatorMatchNum = len(validators) - 2
			}
			if validator == "missing_validator" { // case: validator is specific value matching 0 deposits
				validatorMatchNum = 0
			}
			for _, status := range statuses {
				statusMatchNum := 1 // case: status is a specific value with matching deposits
				if status == "" {   // case: status filter not applied, all len-2 statuses match
					statusMatchNum = len(statuses) - 2
				}
				if status == "missing_status" { // case: status is specific value matching 0 deposits
					statusMatchNum = 0
				}

				expectedNumDeposits := chainMatchNum * validatorMatchNum * statusMatchNum
				params := types.QueryLSMDepositsRequest{ChainId: chainId, ValidatorAddress: validator, Status: status}
				response, err := s.App.RecordsKeeper.LSMDeposits(s.Ctx, &params)
				// Verify no errors in general, it can b empty but should be no errors
				s.Require().NoError(err)
				// Verify that all the deposits expected were found by matching the number set in the keeper
				actualDeposits := response.Deposits
				s.Require().Equal(expectedNumDeposits, len(actualDeposits), "unexpected number of deposits returned")
				testCaseMsg := fmt.Sprintf(" Test Case ChainId: %s, Validator: %s, Status: %s", chainId, validator, status)
				for _, actualDeposit := range actualDeposits {
					if chainId != "" { // Check that every returned deposit matches, if given specific chain-id value
						errMsg := "chain-id on returned deposit does not match requested chain-id filter! %s"
						s.Require().Equal(chainId, actualDeposit.ChainId, errMsg, testCaseMsg)
					}
					if validator != "" { // Check that every returned deposit matches, if given specific validator value
						errMsg := "validator on returned deposit does not match requested validator filter! %s"
						s.Require().Equal(validator, actualDeposit.ValidatorAddress, errMsg, testCaseMsg)
					}
					if status != "" { // Check that every returned deposit matches, if given specific status value
						errMsg := "status on returned deposit does not match requested status filter! %s"
						s.Require().Equal(status, actualDeposit.Status.String(), errMsg, testCaseMsg)
					}
				}
			}
		}
	}

}
