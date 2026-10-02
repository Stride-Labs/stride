package keeper_test

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cometbft/cometbft/proto/tendermint/crypto"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Use committed host state and real client updates so an empty result can only
// zero the recorded delegation after the production proof verifier accepts it.
func (s *KeeperTestSuite) TestMsgSubmitQueryResponse_CalibrationProofs() {
	testCases := []struct {
		name               string
		delegationExists   bool
		mutation           string
		expectInvalidProof bool
	}{
		{name: "valid absence reaches calibration"},
		{name: "valid membership reaches calibration", delegationExists: true},
		{name: "missing proof", mutation: "nil proof", expectInvalidProof: true},
		{name: "empty proof", mutation: "empty proof", expectInvalidProof: true},
		{name: "malformed proof", mutation: "malformed proof", expectInvalidProof: true},
		{name: "membership cannot prove absence", delegationExists: true, mutation: "erase result", expectInvalidProof: true},
		{name: "another absent key cannot erase a delegation", delegationExists: true, mutation: "wrong key", expectInvalidProof: true},
		{name: "bank absence cannot prove staking absence", mutation: "wrong store", expectInvalidProof: true},
		{name: "tampered proof", mutation: "proof bytes", expectInvalidProof: true},
		{name: "tampered delegation value", delegationExists: true, mutation: "value", expectInvalidProof: true},
		{name: "unknown proof height", mutation: "height", expectInvalidProof: true},
		{name: "proof no newer than submission", mutation: "stale", expectInvalidProof: true},
		{name: "sender chain label cannot redirect callback", mutation: "chain label"},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.CreateTransferChannel(HostChainId)
			validatorAddress := sdk.ValAddress(s.TestAccs[1]).String()
			hostZone := stakeibctypes.HostZone{
				ChainId:              HostChainId,
				ConnectionId:         s.TransferPath.EndpointA.ConnectionID,
				DelegationIcaAddress: s.TestAccs[0].String(),
				TotalDelegations:     sdkmath.NewInt(10_000),
				Validators: []*stakeibctypes.Validator{{
					Address:            validatorAddress,
					Delegation:         sdkmath.NewInt(7_500),
					SharesToTokensRate: sdkmath.LegacyOneDec(),
				}},
			}
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
			s.Require().NoError(s.App.StakeibcKeeper.SubmitCalibrationICQ(s.Ctx, hostZone, validatorAddress))
			queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
			s.Require().Len(queries, 1)
			query := queries[0]
			s.Require().Equal(types.STAKING_STORE_QUERY_WITH_PROOF, query.QueryType)
			s.Require().True(query.InvokeCallbackOnEmptyResponse)

			delegation := stakingtypes.Delegation{
				DelegatorAddress: hostZone.DelegationIcaAddress,
				ValidatorAddress: validatorAddress,
				Shares:           sdkmath.LegacyNewDec(6_000),
			}
			if tc.delegationExists {
				encoded := s.HostApp.AppCodec().MustMarshal(&delegation)
				s.HostChain.GetContext().KVStore(s.HostApp.GetKey(stakingtypes.StoreKey)).Set(query.RequestData, encoded)
			}

			// The queried store version must be newer than the query's submission height,
			// and its following header must be trusted by Stride's light client.
			s.Coordinator.CommitBlock(s.HostChain)
			s.Coordinator.CommitBlock(s.HostChain)
			s.Require().NoError(s.TransferPath.EndpointA.UpdateClient())
			s.Ctx = s.StrideChain.GetContext()

			proofKey := query.RequestData
			proofStore := stakingtypes.StoreKey
			if tc.mutation == "wrong key" {
				proofKey = stakingtypes.GetDelegationKey(s.TestAccs[2], sdk.ValAddress(s.TestAccs[3]))
			}
			if tc.mutation == "wrong store" {
				proofStore = banktypes.StoreKey
			}
			response, err := s.HostChain.App.Query(s.HostChain.GetContext().Context(), &abci.RequestQuery{
				Path:   fmt.Sprintf("store/%s/key", proofStore),
				Height: s.HostChain.App.LastBlockHeight() - 1,
				Data:   proofKey,
				Prove:  true,
			})
			s.Require().NoError(err)
			s.Require().Zero(response.Code, response.Log)
			s.Require().Greater(uint64(response.Height), query.SubmissionHeight)
			msg := &types.MsgSubmitQueryResponse{
				ChainId:     HostChainId,
				QueryId:     query.Id,
				Result:      response.Value,
				ProofOps:    response.ProofOps,
				Height:      response.Height,
				FromAddress: s.TestAccs[3].String(),
			}
			switch tc.mutation {
			case "nil proof":
				msg.ProofOps = nil
			case "empty proof":
				msg.ProofOps = &cryptotypes.ProofOps{}
			case "malformed proof":
				msg.ProofOps = &cryptotypes.ProofOps{Ops: []cryptotypes.ProofOp{{Data: []byte{0xff}}}}
			case "erase result":
				msg.Result = nil
			case "proof bytes":
				ops := msg.ProofOps.Ops
				lastProofData := ops[len(ops)-1].Data
				lastProofData[len(lastProofData)-1] ^= 1
			case "value":
				delegation.Shares = sdkmath.LegacyNewDec(9_000)
				msg.Result = s.HostApp.AppCodec().MustMarshal(&delegation)
			case "height":
				msg.Height += 10_000
			case "stale":
				msg.Height = int64(query.SubmissionHeight)
			case "chain label":
				msg.ChainId = "attacker-chosen-chain"
			}
			s.Require().NoError(msg.ValidateBasic())
			beforeResponse, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, query.Id)
			s.Require().True(found)
			_, err = s.GetMsgServer().SubmitQueryResponse(s.Ctx, msg)
			after, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
			s.Require().True(found)
			pendingQuery, pending := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, query.Id)
			if tc.expectInvalidProof {
				s.Require().ErrorIs(err, types.ErrInvalidICQProof)
				s.Require().Equal(hostZone.TotalDelegations, after.TotalDelegations)
				s.Require().Equal(hostZone.Validators[0].Delegation, after.Validators[0].Delegation)
				s.Require().True(pending, "forged response cannot consume query")
				s.Require().Equal(beforeResponse, pendingQuery, "forged response cannot alter query")
				return
			}
			s.Require().NoError(err)
			expectedDelegation := sdkmath.ZeroInt()
			if tc.delegationExists {
				expectedDelegation = sdkmath.NewInt(6_000)
			}
			s.Require().Equal(expectedDelegation, after.Validators[0].Delegation)
			s.Require().Equal(sdkmath.NewInt(2_500).Add(expectedDelegation), after.TotalDelegations)
			s.Require().False(pending)

			// Change the accounting before replay so a second callback would be observable.
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
			_, err = s.GetMsgServer().SubmitQueryResponse(s.Ctx, msg)
			s.Require().NoError(err)
			replayed, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
			s.Require().True(found)
			s.Require().Equal(hostZone.TotalDelegations, replayed.TotalDelegations)
			s.Require().Equal(hostZone.Validators[0].Delegation, replayed.Validators[0].Delegation)
		})
	}
}
