package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

// Setup LSM deposits in keeper

func (s *KeeperTestSuite) TestLSMDeposit() {
	// setup expected deposit in stakeibckeeper
	initToken := types.LSMTokenDeposit{ChainId: "1", Denom: "validator70027"}
	s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, initToken)

	// input nil, no chain-id, no denom --> error invalid input
	invalidInputs := []*types.QueryLSMDepositRequest{}
	noChainInput := types.QueryLSMDepositRequest{ChainId: "", Denom: "validator93740"}
	noDenomInput := types.QueryLSMDepositRequest{ChainId: "42", Denom: ""}
	invalidInputs = append(invalidInputs, nil, &noChainInput, &noDenomInput)

	_, err1 := s.App.StakeibcKeeper.LSMDeposit(s.Ctx, &noChainInput)
	s.Require().ErrorContains(err1, "invalid request")

	// no matching deposit found --> error not found
	missingInput := types.QueryLSMDepositRequest{ChainId: "2", Denom: "validator9374999"}
	_, err2 := s.App.StakeibcKeeper.LSMDeposit(s.Ctx, &missingInput)
	s.Require().ErrorContains(err2, "LSM deposit not found")

	// found the deposit --> no error, deposit returned with matching chain-id and denom
	expectedInput := types.QueryLSMDepositRequest{ChainId: "1", Denom: "validator70027"}
	response, err3 := s.App.StakeibcKeeper.LSMDeposit(s.Ctx, &expectedInput)
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
				status := types.LSMDepositStatus(types.LSMDepositStatus_value[statusStr])
				initToken := types.LSMTokenDeposit{ChainId: chainId, ValidatorAddress: validator, Status: status, Denom: denom}
				s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, initToken)
			}
		}
	}

	// input nil --> error invalid input
	_, err1 := s.App.StakeibcKeeper.LSMDeposit(s.Ctx, nil)
	s.Require().ErrorContains(err1, "invalid request")

	// Adding case where string is empty "" meaning match all
	// Adding case where string is "missing_X" which is an example which is not found
	chainIds = append(chainIds, "", "missing_chain")
	validators = append(validators, "", "missing_validator")
	statuses = append(statuses, "", "missing_status")
	for _, chainId := range chainIds {
		chainMatchNum := 1
		if chainId == "" {
			chainMatchNum = len(chainIds) - 2
		}
		if chainId == "missing_chain" {
			chainMatchNum = 0
		}
		for _, validator := range validators {
			validatorMatchNum := 1
			if validator == "" {
				validatorMatchNum = len(validators) - 2
			}
			if validator == "missing_validator" {
				validatorMatchNum = 0
			}
			for _, status := range statuses {
				statusMatchNum := 1
				if status == "" {
					statusMatchNum = len(statuses) - 2
				}
				if status == "missing_status" {
					statusMatchNum = 0
				}

				expectedNumDeposits := chainMatchNum * validatorMatchNum * statusMatchNum
				params := types.QueryLSMDepositsRequest{ChainId: chainId, ValidatorAddress: validator, Status: status}
				response, err := s.App.StakeibcKeeper.LSMDeposits(s.Ctx, &params)
				s.Require().NoError(err)
				actualDeposits := response.Deposits
				s.Require().Equal(expectedNumDeposits, len(actualDeposits), "unexpected number of deposits returned")
				for _, actualDeposit := range actualDeposits {
					if chainId != "" {
						s.Require().Equal(chainId, actualDeposit.ChainId)
					}
					if validator != "" {
						s.Require().Equal(validator, actualDeposit.ValidatorAddress)
					}
					if status != "" {
						s.Require().Equal(status, actualDeposit.Status.String())
					}
				}
			}
		}
	}

}
