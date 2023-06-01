package v6_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	//nolint:staticcheck
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
)

const dummyUpgradeHeight = 5

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Setup stores for migrated modules
	codec := app.MakeEncodingConfig().Marshaler
	checkClaimStoreAfterMigration := s.SetupClaimStore(codec)

	// Run upgrade
	s.ConfirmUpgradeSucceededs("v6", dummyUpgradeHeight)

	// Confirm store migrations were successful
	checkClaimStoreAfterMigration()
}

// Sets up the old claim store and returns a callback function that can be used to verify
// the store migration was successful
func (s *UpgradeTestSuite) SetupClaimStore(codec codec.Codec) func() {
	claimStore := s.Ctx.KVStore(s.App.GetKey(claimtypes.StoreKey))

	airdropId := "osmosis"
	params := claimtypes.Params{
		Airdrops: []*claimtypes.Airdrop{
			{
				AirdropIdentifier: airdropId,
				ClaimedSoFar:      sdkmath.NewInt(1000000),
			},
		},
	}
	paramsBz, err := codec.MarshalJSON(&params)
	s.Require().NoError(err)
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)

	claimRecords := []types.ClaimRecord{}
	addr1 := "stride12a06af3mm5j653446xr4dguacuxfkj293ey2vh"
	addr2 := "stride1udf2vyj5wyjckl7nzqn5a2vh8fpmmcffey92y8"
	addr3 := "stride1uc8ccxy5s2hw55fn8963ukfdycaamq95jqcfnr"
	// Add some claim records
	claimRecord1 := claimtypes.ClaimRecord{
		AirdropIdentifier: airdropId,
		Address:           addr1,
		Weight:            sdk.NewDec(1000),
		ActionCompleted:   []bool{true, true, false},
	}
	claimRecords = append(claimRecords, claimRecord1)
	claimRecord2 := claimtypes.ClaimRecord{
		AirdropIdentifier: airdropId,
		Address:           addr2,
		Weight:            sdk.NewDec(50),
		ActionCompleted:   []bool{true, true, true},
	}
	claimRecords = append(claimRecords, claimRecord2)
	claimRecord3 := claimtypes.ClaimRecord{
		AirdropIdentifier: airdropId,
		Address:           addr3,
		Weight:            sdk.NewDec(100),
		ActionCompleted:   []bool{false, false, false},
	}
	claimRecords = append(claimRecords, claimRecord3)
	err = s.App.ClaimKeeper.SetClaimRecords(s.Ctx, claimRecords)
	s.Require().NoError(err, "no error expected when setting claim records")

	types.DefaultVestingInitialPeriod, err = time.ParseDuration("0s")
	s.Require().NoError(err, "no error expected when setting vesting initial period")

	// Callback to check claim store after migration
	return func() {
		claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
		claimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, airdropId)

		s.Require().NoError(err, "no error expected when getting claims")
		s.Require().Equal(claimParams.Airdrops[0].AirdropIdentifier, airdropId, "airdrop identifier")
		s.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdkmath.NewInt(0), "claimed so far")
		s.Require().Equal(len(claimRecords), 3, "claim records length")
		fully_reset_action := []bool{false, false, false}
		s.Require().Equal(claimRecords[0].ActionCompleted, fully_reset_action, "record 1 reset")
		s.Require().Equal(claimRecords[0].Address, addr1, "record 1 address")
		s.Require().Equal(claimRecords[0].Weight, sdk.NewDec(1000), "record 1 weight")
		s.Require().Equal(claimRecords[1].ActionCompleted, fully_reset_action, "record 2 reset")
		s.Require().Equal(claimRecords[1].Address, addr2, "record 2 address")
		s.Require().Equal(claimRecords[1].Weight, sdk.NewDec(50), "record 2 weight")
		s.Require().Equal(claimRecords[2].ActionCompleted, fully_reset_action, "record 3 reset")
		s.Require().Equal(claimRecords[2].Address, addr3, "record 3 address")
		s.Require().Equal(claimRecords[2].Weight, sdk.NewDec(100), "record 3 weight")
	}
}
