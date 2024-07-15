package keeper_test

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func (s *KeeperTestSuite) TestAddUserLink() {
	testCases := []struct {
		name              string
		initialLinks      []string
		addedLink         string
		expectedLinks     []string
		initialClaimType  types.ClaimType
		expectedClaimType types.ClaimType
	}{
		{
			name:          "add first link",
			initialLinks:  []string{},
			addedLink:     "dym",
			expectedLinks: []string{"dym"},
		},
		{
			name:          "add second",
			initialLinks:  []string{"dym"},
			addedLink:     "agoric",
			expectedLinks: []string{"dym", "agoric"},
		},
		{
			name:          "already exists",
			initialLinks:  []string{"dym"},
			addedLink:     "dym",
			expectedLinks: []string{"dym"},
		},
		{
			name:              "reset claim type early",
			initialLinks:      []string{"dym"},
			addedLink:         "agoric",
			expectedLinks:     []string{"dym", "agoric"},
			initialClaimType:  types.CLAIM_EARLY,
			expectedClaimType: types.CLAIM_DAILY,
		},
		{
			name:              "reset claim type stake",
			initialLinks:      []string{"dym"},
			addedLink:         "agoric",
			expectedLinks:     []string{"dym", "agoric"},
			initialClaimType:  types.CLAIM_AND_STAKE,
			expectedClaimType: types.CLAIM_DAILY,
		},
		// TODO add link to a non existing allocation?
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset state

			claimer := s.TestAccs[0]
			strideAddress := claimer.String()

			distributor := s.TestAccs[1]

			// Fund the distributor
			initialDistributorBalance := sdk.NewInt(1000)
			s.FundAccount(distributor, sdk.NewCoin(RewardDenom, initialDistributorBalance))

			// Create the initial airdrop config
			s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
				Id:                    AirdropId,
				RewardDenom:           RewardDenom,
				DistributionAddress:   distributor.String(),
				DistributionStartDate: &DistributionStartDate,
				DistributionEndDate:   &DistributionEndDate,
			})

			// Set the block time to be inside the distribution window
			blockTime := DistributionStartDate.Add(time.Nanosecond)
			s.Ctx = s.Ctx.WithBlockTime(blockTime)

			// Create the initial users and allocations
			dymAddress, err := bech32.ConvertAndEncode("dym", claimer)
			s.Require().NoError(err, "bech32 dym")

			agoricAddress, err := bech32.ConvertAndEncode("agoric", claimer)
			s.Require().NoError(err, "bech32 agoric")

			// Populate tc with the actual dym and agoric addresses
			for i := range tc.initialLinks {
				tc.initialLinks[i] = strings.Replace(tc.initialLinks[i], "dym", dymAddress, 1)
				tc.initialLinks[i] = strings.Replace(tc.initialLinks[i], "agoric", agoricAddress, 1)
			}

			tc.addedLink = strings.Replace(tc.addedLink, "dym", dymAddress, 1)
			tc.addedLink = strings.Replace(tc.addedLink, "agoric", agoricAddress, 1)

			for i := range tc.expectedLinks {
				tc.expectedLinks[i] = strings.Replace(tc.expectedLinks[i], "dym", dymAddress, 1)
				tc.expectedLinks[i] = strings.Replace(tc.expectedLinks[i], "agoric", agoricAddress, 1)
			}

			// Set initial links in state
			// This has to happen before SetUserAllocation because
			// AddUserLink is supposed to reset claim type and we don't want that to happen on init
			for i := range tc.initialLinks {
				s.App.AirdropKeeper.AddUserLink(s.Ctx, AirdropId, strideAddress, tc.initialLinks[i])
			}

			// Set user allocation for stride, dym and agoric addresses
			for _, address := range []string{strideAddress, dymAddress, agoricAddress} {
				s.App.AirdropKeeper.SetUserAllocation(s.Ctx, types.UserAllocation{
					AirdropId:   AirdropId,
					Address:     address,
					ClaimType:   tc.initialClaimType,
					Claimed:     sdkmath.ZeroInt(),
					Forfeited:   sdkmath.ZeroInt(),
					Allocations: allocationsToSdkInt([]int64{10, 10, 10}),
				})
			}

			// Call add link
			s.App.AirdropKeeper.AddUserLink(s.Ctx, AirdropId, strideAddress, tc.addedLink)

			// Check that the link was updated
			userLinks := s.MustGetUserLinks(AirdropId, strideAddress)
			s.Require().Equal(strideAddress, userLinks.StrideAddress, "stride address")
			s.Require().Equal(tc.expectedLinks, userLinks.HostAddresses, "host addresses")

			for _, address := range []string{strideAddress, dymAddress, agoricAddress} {
				userAllocation := s.MustGetUserAllocation(AirdropId, address)
				s.Require().Equal(tc.expectedClaimType, userAllocation.ClaimType, "claim type")
			}
		})
	}
}
