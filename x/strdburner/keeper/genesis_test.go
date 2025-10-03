package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	acc1, acc2 := s.TestAccs[0], s.TestAccs[1]

	protocolBurned := sdkmath.NewInt(1000)
	totalUserBurned := sdkmath.NewInt(2000)
	totalBurned := sdkmath.NewInt(3000)

	userBurned1 := sdkmath.NewInt(4000)
	userBurned2 := sdkmath.NewInt(5000)

	genState := types.GenesisState{
		TotalUserUstrdBurned: totalUserBurned,
		ProtocolUstrdBurned:  protocolBurned,
		TotalUstrdBurned:     totalBurned,
		BurnedByAccount: []types.AddressBurnedAmount{
			{Address: acc1.String(), Amount: userBurned1},
			{Address: acc2.String(), Amount: userBurned2},
		},
	}

	// Initialize genesis
	s.App.StrdBurnerKeeper.InitGenesis(s.Ctx, genState)

	// Confirm new state
	s.Require().Equal(totalBurned, s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx))
	s.Require().Equal(protocolBurned, s.App.StrdBurnerKeeper.GetProtocolStrdBurned(s.Ctx))
	s.Require().Equal(totalUserBurned, s.App.StrdBurnerKeeper.GetTotalUserStrdBurned(s.Ctx))

	s.Require().Equal(userBurned1, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc1))
	s.Require().Equal(userBurned2, s.App.StrdBurnerKeeper.GetStrdBurnedByAddress(s.Ctx, acc2))

	// Confirm export
	s.Require().Equal(genState, *s.App.StrdBurnerKeeper.ExportGenesis(s.Ctx), "exported genesis")

	// Attempt to import with a dupe address, it should fail panic
	s.Panics(func() {
		genState.BurnedByAccount = append(genState.BurnedByAccount, types.AddressBurnedAmount{
			Address: acc2.String(),
		})
		s.App.StrdBurnerKeeper.InitGenesis(s.Ctx, genState)
	}, fmt.Sprintf("Duplicate address found: %s", acc2.String()))
}
