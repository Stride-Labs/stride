package keeper_test

import (
	"fmt"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	sweepTestStToken = "stuatom"
	sweepTestStrd    = "ustrd"
)

// registerVoucher stores a denom trace (outermost hop first) and returns its ibc/ denom
func (s *KeeperTestSuite) registerVoucher(base string, hops ...transfertypes.Hop) string {
	denom := transfertypes.NewDenom(base, hops...)
	s.App.TransferKeeper.SetDenom(s.Ctx, denom)
	return denom.IBCDenom()
}

func (s *KeeperTestSuite) TestResolveSweepDestination() {
	singleHopAtom := s.registerVoucher("uatom", transfertypes.NewHop(transfertypes.PortID, "channel-0"))
	twoHopStAtomViaOsmosis := s.registerVoucher(sweepTestStToken,
		transfertypes.NewHop(transfertypes.PortID, "channel-5"),
		transfertypes.NewHop(transfertypes.PortID, "channel-326"))
	unwhitelistedLuna := s.registerVoucher("uluna", transfertypes.NewHop(transfertypes.PortID, "channel-52"))

	testCases := []struct {
		name        string
		denom       string
		expected    keeper.SweepDestinationForTest
		expectedErr error
	}{
		{
			name:     "stToken goes to osmosis over channel-5",
			denom:    sweepTestStToken,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:     "ustrd goes to osmosis over channel-5",
			denom:    sweepTestStrd,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:     "single-hop atom voucher unwinds over channel-0 to cosmos",
			denom:    singleHopAtom,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-0", Bech32Prefix: "cosmos"},
		},
		{
			name:     "two-hop voucher unwinds one hop over its outer channel",
			denom:    twoHopStAtomViaOsmosis,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:        "voucher whose outer channel is not whitelisted is rejected",
			denom:       unwhitelistedLuna,
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "ibc denom with no trace in the store is rejected",
			denom:       "ibc/0000000000000000000000000000000000000000000000000000000000000000",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "ibc denom with a malformed hash is rejected",
			denom:       "ibc/NOTAHASH",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			destination, err := keeper.ResolveSweepDestinationForTest(s.App.StakeibcKeeper, s.Ctx, tc.denom)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expected, destination)
		})
	}
}

// Sets up one account of each kind the skip rules distinguish and returns them keyed by label
func (s *KeeperTestSuite) setupSweepAccountKinds() map[string]sdk.AccAddress {
	accounts := apptesting.CreateRandomAccounts(8)
	kinds := map[string]sdk.AccAddress{}
	// NewAccountWithAddress assigns the next account number. A bare NewBaseAccountWithAddress
	// leaves it at 0, and SetAccount then panics on the auth store's unique account-number index
	// because genesis already owns number 0 (apptesting.SetNewAccount does the same dance)
	newBase := func(address sdk.AccAddress) *authtypes.BaseAccount {
		return s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, address).(*authtypes.BaseAccount)
	}
	vesting := sdk.NewCoins(sdk.NewInt64Coin(sweepTestStrd, 1))
	now := s.Ctx.BlockTime().Unix()

	s.App.AccountKeeper.SetAccount(s.Ctx, newBase(accounts[0]))
	kinds["base"] = accounts[0]

	continuous, err := vestingtypes.NewContinuousVestingAccount(newBase(accounts[1]), vesting, now, now+1000)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, continuous)
	kinds["continuous_vesting"] = accounts[1]

	delayed, err := vestingtypes.NewDelayedVestingAccount(newBase(accounts[2]), vesting, now+1000)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, delayed)
	kinds["delayed_vesting"] = accounts[2]

	periodic, err := vestingtypes.NewPeriodicVestingAccount(newBase(accounts[3]), vesting, now,
		vestingtypes.Periods{{Length: 1000, Amount: vesting}})
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, periodic)
	kinds["periodic_vesting"] = accounts[3]

	stridePeriodic := claimvestingtypes.NewStridePeriodicVestingAccount(newBase(accounts[4]), vesting,
		claimvestingtypes.Periods{{StartTime: now, Length: 1000, Amount: vesting}})
	s.App.AccountKeeper.SetAccount(s.Ctx, stridePeriodic)
	kinds["stride_periodic_vesting"] = accounts[4]

	ica := icatypes.NewInterchainAccount(newBase(accounts[5]), "cosmos1owner")
	s.App.AccountKeeper.SetAccount(s.Ctx, ica)
	kinds["interchain_account"] = accounts[5]

	kinds["unknown"] = accounts[6]

	kinds["module"] = authtypes.NewModuleAddress(distrtypes.ModuleName)
	kinds["thirty_two_bytes"] = sdk.AccAddress(make([]byte, 32))
	kinds["escrow"] = transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-0")
	return kinds
}

func (s *KeeperTestSuite) TestSweepSkipReason() {
	kinds := s.setupSweepAccountKinds()

	// channel-0 must exist for its escrow address to be in the set
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, "channel-0", channeltypes.Channel{
		State:          channeltypes.OPEN,
		Ordering:       channeltypes.UNORDERED,
		Counterparty:   channeltypes.NewCounterparty(transfertypes.PortID, "channel-0"),
		ConnectionHops: []string{ibctesting.FirstConnectionID},
		Version:        transfertypes.V1,
	})
	escrows := keeper.TransferEscrowAddressesForTest(s.App.StakeibcKeeper, s.Ctx)
	s.Require().True(escrows[kinds["escrow"].String()], "channel-0 escrow should be in the set")

	testCases := []struct {
		kind           string
		expectedSkip   bool
		expectedReason string
	}{
		{kind: "base", expectedSkip: false},
		{kind: "continuous_vesting", expectedSkip: false},
		{kind: "delayed_vesting", expectedSkip: false},
		{kind: "periodic_vesting", expectedSkip: false},
		{kind: "stride_periodic_vesting", expectedSkip: false},
		{kind: "thirty_two_bytes", expectedSkip: true, expectedReason: "address is not 20 bytes"},
		{kind: "unknown", expectedSkip: true, expectedReason: "account not found"},
		{kind: "escrow", expectedSkip: true, expectedReason: "transfer escrow address"},
		{kind: "module", expectedSkip: true, expectedReason: "account type *types.ModuleAccount is not sweepable"},
		{kind: "interchain_account", expectedSkip: true, expectedReason: "account type *types.InterchainAccount is not sweepable"},
	}
	for _, tc := range testCases {
		s.Run(tc.kind, func() {
			reason, skip := keeper.SweepSkipReasonForTest(s.App.StakeibcKeeper, s.Ctx, kinds[tc.kind], escrows)
			s.Require().Equal(tc.expectedSkip, skip, fmt.Sprintf("skip for %s (reason %q)", tc.kind, reason))
			if tc.expectedSkip {
				s.Require().Equal(tc.expectedReason, reason)
			}
		})
	}
}
