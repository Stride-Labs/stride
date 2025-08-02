package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

func TestFormatHostZoneICAOwner(t *testing.T) {
	chainId := "chain-0"

	testCases := []struct {
		accountType types.ICAAccountType
		owner       string
	}{
		{
			accountType: types.ICAAccountType_DELEGATION,
			owner:       "chain-0.DELEGATION",
		},
		{
			accountType: types.ICAAccountType_WITHDRAWAL,
			owner:       "chain-0.WITHDRAWAL",
		},
		{
			accountType: types.ICAAccountType_REDEMPTION,
			owner:       "chain-0.REDEMPTION",
		},
		{
			accountType: types.ICAAccountType_FEE,
			owner:       "chain-0.FEE",
		},
		{
			accountType: types.ICAAccountType_COMMUNITY_POOL_DEPOSIT,
			owner:       "chain-0.COMMUNITY_POOL_DEPOSIT",
		},
		{
			accountType: types.ICAAccountType_COMMUNITY_POOL_RETURN,
			owner:       "chain-0.COMMUNITY_POOL_RETURN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.accountType.String(), func(t *testing.T) {
			actual := types.FormatHostZoneICAOwner(chainId, tc.accountType)
			require.Equal(t, tc.owner, actual)
		})
	}
}

func TestFormatTradeRouteICAOwner(t *testing.T) {
	chainId := "chain-0"
	rewardDenom := "ureward"
	hostDenom := "uhost"

	testCases := []struct {
		accountType types.ICAAccountType
		owner       string
	}{
		{
			accountType: types.ICAAccountType_CONVERTER_UNWIND,
			owner:       "chain-0.ureward-uhost.CONVERTER_UNWIND",
		},
		{
			accountType: types.ICAAccountType_CONVERTER_TRADE,
			owner:       "chain-0.ureward-uhost.CONVERTER_TRADE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.accountType.String(), func(t *testing.T) {
			actual := types.FormatTradeRouteICAOwner(chainId, rewardDenom, hostDenom, tc.accountType)
			require.Equal(t, actual, tc.owner, "format trade route ICA owner")

			tradeRouteId := "ureward-uhost"
			actual = types.FormatTradeRouteICAOwnerFromRouteId(chainId, tradeRouteId, tc.accountType)
			require.Equal(t, actual, tc.owner, "format trade route ICA owner by account")
		})
	}
}
