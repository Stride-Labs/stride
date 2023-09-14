package keeper_test

import (
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestGovToggleOracle() {
	oracles := s.CreateTestOracles()

	oracleIndexToToggle := 1
	oracleToToggle := oracles[oracleIndexToToggle]

	// Set the oracle to inactive
	_, err := s.GetMsgServer().ToggleOracle(s.Ctx, &types.MsgToggleOracle{
		Authority:     s.App.ICAOracleKeeper.GetAuthority(),
		OracleChainId: oracleToToggle.ChainId,
		Active:        false,
	})
	s.Require().NoError(err)

	// Confirm it's the only oracle inactive
	for i, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracle.ChainId)
		s.Require().True(found, "oracle %s does not exist", oracle.ChainId)

		if i == oracleIndexToToggle {
			s.Require().False(oracle.Active, "oracle %s should have been toggled to inactive", oracle.ChainId)
		} else {
			s.Require().True(oracle.Active, "oracle %s should still be active", oracle.ChainId)
		}
	}

	// Set it back to active
	_, err = s.GetMsgServer().ToggleOracle(s.Ctx, &types.MsgToggleOracle{
		Authority:     s.App.ICAOracleKeeper.GetAuthority(),
		OracleChainId: oracleToToggle.ChainId,
		Active:        true,
	})
	s.Require().NoError(err)

	// Confirm all oracles are active again
	for _, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracle.ChainId)
		s.Require().True(found, "oracle %s does not exist", oracle.ChainId)
		s.Require().True(oracle.Active, "oracle %s should still be active", oracle.ChainId)
	}
}
