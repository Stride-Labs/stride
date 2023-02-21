package v6_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	//nolint:staticcheck
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v5/app"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	oldclaimtypes "github.com/Stride-Labs/stride/v5/x/claim/migrations/v2/types"
	claimtypes "github.com/Stride-Labs/stride/v5/x/claim/types"
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

	airdropId := "id"
	params := oldclaimtypes.Params{
		Airdrops: []*oldclaimtypes.Airdrop{
			{
				AirdropIdentifier: airdropId,
				ClaimedSoFar:      1000000,
			},
		},
	}

	paramsBz, err := codec.MarshalJSON(&params)
	s.Require().NoError(err)
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)

	// Callback to check claim store after migration
	return func() {
		claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
		s.Require().NoError(err, "no error expected when getting claims")
		s.Require().Equal(claimParams.Airdrops[0].AirdropIdentifier, airdropId, "airdrop identifier")
		s.Require().Equal(claimParams.Airdrops[0].ClaimedSoFar, sdkmath.NewInt(1000000), "claimed so far")
	}
}
