package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v20/x/stakedym/types"
)

func (s *KeeperTestSuite) TestUpdateRedemptionRate() {
	depositAddress := s.TestAccs[0]

	testCases := []struct {
		expectedRedemptionRate sdk.Dec
		depositBalance         sdkmath.Int
		delegatedBalance       sdkmath.Int
		stTokenSupply          sdkmath.Int
		delegationRecords      []types.DelegationRecord
	}{
		{
			// Deposit: 250, Undelegated: 500, Delegated: 250, StTokens: 1000
			// (250 + 500 + 250 / 1000) = 1000 / 1000 = 1.0
			expectedRedemptionRate: sdk.MustNewDecFromStr("1.0"),
			depositBalance:         sdkmath.NewInt(250),
			delegatedBalance:       sdkmath.NewInt(250),
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 2, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
			},
			stTokenSupply: sdkmath.NewInt(1000),
		},
		{
			// Deposit: 500, Undelegated: 500, Delegated: 250, StTokens: 1000
			// (500 + 500 + 250 / 1000) = 1250 / 1000 = 1.25
			expectedRedemptionRate: sdk.MustNewDecFromStr("1.25"),
			depositBalance:         sdkmath.NewInt(500),
			delegatedBalance:       sdkmath.NewInt(250),
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 2, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
			},
			stTokenSupply: sdkmath.NewInt(1000),
		},
		{
			// Deposit: 250, Undelegated: 500, Delegated: 500, StTokens: 1000
			// (500 + 500 + 250 / 1000) = 1250 / 1000 = 1.250
			expectedRedemptionRate: sdk.MustNewDecFromStr("1.25"),
			depositBalance:         sdkmath.NewInt(250),
			delegatedBalance:       sdkmath.NewInt(500),
			delegationRecords: []types.DelegationRecord{
				{Id: 2, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 3, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
			},
			stTokenSupply: sdkmath.NewInt(1000),
		},
		{
			// Deposit: 250, Undelegated: 1000, Delegated: 250, StTokens: 1000
			// (250 + 1000 + 250 / 1000) = 1500 / 1000 = 1.5
			expectedRedemptionRate: sdk.MustNewDecFromStr("1.5"),
			depositBalance:         sdkmath.NewInt(250),
			delegatedBalance:       sdkmath.NewInt(250),
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 2, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
				{Id: 4, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 6, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
			},
			stTokenSupply: sdkmath.NewInt(1000),
		},
		{
			// Deposit: 250, Undelegated: 500, Delegated: 250, StTokens: 2000
			// (250 + 500 + 250 / 2000) = 1000 / 2000 = 0.5
			expectedRedemptionRate: sdk.MustNewDecFromStr("0.5"),
			depositBalance:         sdkmath.NewInt(250),
			delegatedBalance:       sdkmath.NewInt(250),
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(250), Status: types.TRANSFER_IN_PROGRESS},
				{Id: 2, NativeAmount: sdkmath.NewInt(250), Status: types.DELEGATION_QUEUE},
			},
			stTokenSupply: sdkmath.NewInt(2000),
		},
	}

	for i, tc := range testCases {
		s.Run(fmt.Sprintf("test-%d", i), func() {
			s.SetupTest() // reset state

			// Fund the deposit balance
			s.FundAccount(depositAddress, sdk.NewCoin(HostIBCDenom, tc.depositBalance))

			// Create the host zone with the delegated balance and deposit address
			initialRedemptionRate := sdk.MustNewDecFromStr("0.999")
			s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
				NativeTokenDenom:    HostNativeDenom,
				NativeTokenIbcDenom: HostIBCDenom,
				DepositAddress:      depositAddress.String(),
				DelegatedBalance:    tc.delegatedBalance,
				RedemptionRate:      initialRedemptionRate,
			})

			// Set each delegation record
			for _, delegationRecord := range tc.delegationRecords {
				s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
			}

			// Add some archive delegation records that should be excluded
			// We'll create these by first creating normal records and then removing them
			for i := 0; i <= 5; i++ {
				id := uint64(i * 1000)
				s.App.StakedymKeeper.SetArchivedDelegationRecord(s.Ctx, types.DelegationRecord{Id: id})
			}

			// Mint sttokens for the supply (fund account calls mint)
			s.FundAccount(s.TestAccs[1], sdk.NewCoin(StDenom, tc.stTokenSupply))

			// Update the redemption rate and check that it matches
			err := s.App.StakedymKeeper.UpdateRedemptionRate(s.Ctx)
			s.Require().NoError(err, "no error expected when calculating redemption rate")

			hostZone := s.MustGetHostZone()
			s.Require().Equal(tc.expectedRedemptionRate, hostZone.RedemptionRate, "redemption rate")

			// Check that the last redemption rate was set
			s.Require().Equal(initialRedemptionRate, hostZone.LastRedemptionRate, "redemption rate")
		})

	}
}

func (s *KeeperTestSuite) TestUpdateRedemptionRate_NoTokens() {
	depositAddress := s.TestAccs[0]

	// Create the host zone with no delegated balance
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
		NativeTokenDenom:    HostNativeDenom,
		NativeTokenIbcDenom: HostIBCDenom,
		DepositAddress:      depositAddress.String(),
		DelegatedBalance:    sdkmath.ZeroInt(),
		RedemptionRate:      sdk.OneDec(),
	})

	// Check that the update funtion returns nil, since there are no stTokens
	err := s.App.StakedymKeeper.UpdateRedemptionRate(s.Ctx)
	s.Require().NoError(err, "no error when there are no stTokens")

	// Check that the redemption rate was not updated
	hostZone := s.MustGetHostZone()
	s.Require().Equal(sdk.OneDec(), hostZone.RedemptionRate, "redemption rate should not have been updated")

	// Mint stTokens
	s.FundAccount(s.TestAccs[1], sdk.NewCoin(StDenom, sdkmath.NewInt(1000)))

	// Try to update again, now it should error since there's stTokens but no native tokens
	err = s.App.StakedymKeeper.UpdateRedemptionRate(s.Ctx)
	s.Require().ErrorContains(err, "Non-zero stToken supply, yet the zero delegated and undelegated balance")
}

func (s *KeeperTestSuite) TestCheckRedemptionRateExceedsBounds() {
	testCases := []struct {
		name          string
		hostZone      types.HostZone
		exceedsBounds bool
	}{
		{
			name: "valid bounds",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"), // <--
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: false,
		},
		{
			name: "outside min inner",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				RedemptionRate:         sdk.MustNewDecFromStr("0.9"), // <--
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max inner",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.1"), // <--
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside min outer",
			hostZone: types.HostZone{
				RedemptionRate:         sdk.MustNewDecFromStr("0.8"), // <--
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.9"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max outer",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.1"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.2"), // <--
			},
			exceedsBounds: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.App.StakedymKeeper.SetHostZone(s.Ctx, tc.hostZone)
			err := s.App.StakedymKeeper.CheckRedemptionRateExceedsBounds(s.Ctx)
			if tc.exceedsBounds {
				s.Require().ErrorIs(err, types.ErrRedemptionRateOutsideSafetyBounds)
			} else {
				s.Require().NoError(err, "no error expected")
			}
		})
	}
}
