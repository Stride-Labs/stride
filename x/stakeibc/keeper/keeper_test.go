package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/suite"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v18/app/apptesting"
	icqtypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

var (
	Atom         = "uatom"
	StAtom       = "stuatom"
	IbcAtom      = "ibc/uatom"
	GaiaPrefix   = "cosmos"
	HostChainId  = "GAIA"
	Bech32Prefix = "cosmos"

	Osmo        = "uosmo"
	StOsmo      = "stuosmo"
	IbcOsmo     = "ibc/uosmo"
	OsmoPrefix  = "osmo"
	OsmoChainId = "OSMO"

	HostDenom   = "udenom"
	RewardDenom = "ureward"

	ValAddress        = "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"
	StrideICAAddress  = "stride1gcx4yeplccq9nk6awzmm0gq8jf7yet80qj70tkwy0mz7pg87nepsen0l38"
	HostICAAddress    = "cosmos1gcx4yeplccq9nk6awzmm0gq8jf7yet80qj70tkwy0mz7pg87nepswn2dj8"
	LSMTokenBaseDenom = ValAddress + "/32"

	DepositAddress                    = "deposit"
	CommunityPoolStakeHoldingAddress  = "staking-holding"
	CommunityPoolRedeemHoldingAddress = "redeem-holding"

	Authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
)

type ICQCallbackArgs struct {
	Query        icqtypes.Query
	CallbackArgs []byte
}

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.StakeibcKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Helper function to get a host zone and confirm it was found
func (s *KeeperTestSuite) MustGetHostZone(chainId string) types.HostZone {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
	s.Require().True(found, "host zone should have been found")
	return hostZone
}

// Helper function to create an epoch tracker that dictates the timeout
func (s *KeeperTestSuite) CreateEpochForICATimeout(epochType string, timeoutDuration time.Duration) {
	epochEndTime := uint64(s.Ctx.BlockTime().Add(timeoutDuration).UnixNano())
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochType,
		NextEpochStartTime: epochEndTime,
		Duration:           uint64(timeoutDuration),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
}

// Validates the query object stored after an ICQ submission, using some default testing
// values (e.g. HostChainId, stakeibc module name, etc.), and returning the query
// NOTE: This assumes there was only one submission and grabs the first query from the store
func (s *KeeperTestSuite) ValidateQuerySubmission(
	queryType string,
	queryData []byte,
	callbackId string,
	timeoutDuration time.Duration,
	timeoutPolicy icqtypes.TimeoutPolicy,
) icqtypes.Query {
	// Check that there's only one query
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Validate the chainId and connectionId
	s.Require().Equal(HostChainId, query.ChainId, "query chain ID")
	s.Require().Equal(ibctesting.FirstConnectionID, query.ConnectionId, "query connection ID")
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query module")

	// Validate the query type and request data
	s.Require().Equal(queryType, query.QueryType, "query type")
	s.Require().Equal(string(queryData), string(query.RequestData), "query request data")
	s.Require().Equal(callbackId, query.CallbackId, "query callback ID")

	// Validate the query timeout
	expectedTimeoutTimestamp := s.Ctx.BlockTime().Add(timeoutDuration).UnixNano()
	s.Require().Equal(timeoutDuration, query.TimeoutDuration, "query timeout duration")
	s.Require().Equal(expectedTimeoutTimestamp, int64(query.TimeoutTimestamp), "query timeout timestamp")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	return query
}

func (s *KeeperTestSuite) TestIsRedemptionRateWithinSafetyBounds() {
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.DefaultMinRedemptionRateThreshold = 75
	params.DefaultMaxRedemptionRateThreshold = 150
	hostZones := make(map[string]types.HostZone)
	hostZones["gaia-1"] = types.HostZone{
		ChainId:           "gaia-1",
		MinRedemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
		MaxRedemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
	}
	hostZones["osmosis-1"] = types.HostZone{
		ChainId:           "osmosis-1",
		MinRedemptionRate: sdk.NewDecWithPrec(3, 1),  // 0.3
		MaxRedemptionRate: sdk.NewDecWithPrec(20, 1), // 2
	}
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	for _, tc := range []struct {
		chainId        string
		redemptionRate sdk.Dec
		expSafe        bool
	}{
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
	} {
		hostZone, ok := hostZones[tc.chainId]
		if !ok {
			hostZone = types.HostZone{
				ChainId: tc.chainId,
			}
		}
		hostZone.RedemptionRate = tc.redemptionRate
		rrSafe, err := s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, hostZone)
		if tc.expSafe {
			s.Require().NoError(err)
			s.Require().True(rrSafe)
		} else {
			s.Require().Error(err)
			s.Require().False(rrSafe)
		}
	}
}
