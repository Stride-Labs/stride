package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v27/x/strdburner/types"
)

func (s *KeeperTestSuite) TestEndBlocker() {
	t := s.T()
	burnerAddress := s.App.StrdBurnerKeeper.GetStrdBurnerAddress()

	tests := []struct {
		name           string
		initialBalance sdk.Coin
		shouldBurn     bool
	}{
		{
			name:           "burn non-zero balance",
			initialBalance: sdk.NewCoin("ustrd", sdkmath.NewInt(1000)),
			shouldBurn:     true,
		},
		{
			name:           "zero balance - no burn",
			initialBalance: sdk.NewCoin("ustrd", sdkmath.NewInt(0)),
			shouldBurn:     false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Setup test state
			s.SetupTest()

			// Set initial balance if non-zero
			if !tc.initialBalance.IsZero() {
				s.FundModuleAccount(types.ModuleName, tc.initialBalance)
			}

			// Verify initial state
			initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnerAddress, "ustrd")
			require.Equal(t, tc.initialBalance, initialBalance)

			initialTotalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
			require.Equal(t, sdkmath.ZeroInt(), initialTotalBurned)

			// Run EndBlocker
			s.App.StrdBurnerKeeper.EndBlocker(s.Ctx)

			// Verify final state
			finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnerAddress, "ustrd")
			require.Equal(t, sdk.NewCoin("ustrd", sdkmath.ZeroInt()), finalBalance)

			finalTotalBurned := s.App.StrdBurnerKeeper.GetTotalStrdBurned(s.Ctx)
			if tc.shouldBurn {
				require.Equal(t, tc.initialBalance.Amount, finalTotalBurned)

				// Verify event was emitted
				events := s.Ctx.EventManager().Events()
				var found bool
				for _, event := range events {
					if event.Type == types.EventTypeBurn {
						found = true
						for _, attr := range event.Attributes {
							if string(attr.Key) == types.AttributeAmount {
								require.Equal(t, tc.initialBalance.String(), string(attr.Value))
							}
						}
					}
				}
				require.True(t, found, "burn event should have been emitted")
			} else {
				require.Equal(t, sdkmath.ZeroInt(), finalTotalBurned)
			}
		})
	}
}
