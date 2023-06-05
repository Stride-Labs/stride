package v9_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app"
	"github.com/Stride-Labs/stride/v9/app/apptesting"
	v9 "github.com/Stride-Labs/stride/v9/app/upgrades/v9"
	"github.com/Stride-Labs/stride/v9/utils"

	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"

	// This isn't the exact type host zone schema as the one that's will be in the store
	// before the upgrade, but the only thing that matters, for the sake of the test,
	// is that it doesn't have min/max redemption rate as attributes
	"github.com/Stride-Labs/stride/v9/x/claim/migrations/v2/types"
	oldclaimtypes "github.com/Stride-Labs/stride/v9/x/claim/migrations/v2/types"
)

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

	dummyUpgradeHeight := int64(5)

	s.SetupAirdropsBeforeUpgrade()
	s.ConfirmUpgradeSucceededs("v9", dummyUpgradeHeight)
	s.CheckAirdropsAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupAirdropsBeforeUpgrade() {
	// Create a list of airdrops of the old data type
	airdrops := []*oldclaimtypes.Airdrop{}
	for i, identifier := range utils.StringMapKeys(v9.AirdropChainIds) {
		airdrops = append(airdrops, &oldclaimtypes.Airdrop{
			AirdropIdentifier: identifier,
			ClaimDenom:        fmt.Sprintf("denom-%d", i),
		})
	}

	// Add in another airdrop that's not in the map
	airdrops = append(airdrops, &types.Airdrop{
		AirdropIdentifier: "different_airdrop",
	})

	// Store the airdrops using the old schema
	codec := app.MakeEncodingConfig().Marshaler
	claimStore := s.Ctx.KVStore(s.App.GetKey(claimtypes.StoreKey))

	paramsBz, err := codec.MarshalJSON(&oldclaimtypes.Params{Airdrops: airdrops})
	s.Require().NoError(err, "no error expected when marshalling claim params")
	claimStore.Set([]byte(claimtypes.ParamsKey), paramsBz)
}

func (s *UpgradeTestSuite) CheckAirdropsAfterUpgrade() {
	// Read in the airdrops using the new schema - which should include chainId and AirdropEnabled
	claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting claims params")
	s.Require().Len(claimParams.Airdrops, len(v9.AirdropChainIds)+1, "number of airdrops after migration")

	// Confirm the new fields were added and the old fields (e.g. ChainDenom) remain the same
	for i, identifier := range utils.StringMapKeys(v9.AirdropChainIds) {
		expectedChainId := v9.AirdropChainIds[identifier]
		expectedDenom := fmt.Sprintf("denom-%d", i)
		expectedAutopilotEnabled := identifier == v9.EvmosAirdropId

		actual := claimParams.Airdrops[i]
		s.Require().Equal(identifier, actual.AirdropIdentifier, "identifier after migration")
		s.Require().Equal(expectedChainId, actual.ChainId, "chain-id after migration")
		s.Require().Equal(expectedDenom, actual.ClaimDenom, "denom after migration")
		s.Require().Equal(expectedAutopilotEnabled, actual.AutopilotEnabled, "autopilot enabled after migration")
	}

	// Confirm the airdrop that was not in the map
	airdropWithoutChainId := claimParams.Airdrops[len(v9.AirdropChainIds)]
	s.Require().Equal("different_airdrop", airdropWithoutChainId.AirdropIdentifier, "airdrop id for outsider")
	s.Require().Equal("", airdropWithoutChainId.ChainId, "chain-id for outsider")
}

func (s *UpgradeTestSuite) TestAddFieldsToAirdropType() {
	s.SetupAirdropsBeforeUpgrade()

	err := v9.AddFieldsToAirdropType(s.Ctx, s.App.ClaimKeeper)
	s.Require().NoError(err, "no error expected when migrating airdrop schema")

	s.CheckAirdropsAfterUpgrade()
}
