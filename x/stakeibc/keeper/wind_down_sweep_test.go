package keeper_test

import (
	"fmt"
	"strings"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	auctiontypes "github.com/Stride-Labs/stride/v34/x/auction/types"
	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
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
	s.setSweepHostZone()
	// Hub ATOM that travelled Hub -> Osmosis -> Stride: Stride's outermost hop is channel-5
	osmosisRoutedAtom := s.registerVoucher("uatom",
		transfertypes.NewHop(transfertypes.PortID, "channel-5"),
		transfertypes.NewHop(transfertypes.PortID, "channel-0"))
	wrongPortVoucher := s.registerVoucher("uatom", transfertypes.NewHop("wasm.contract", "channel-0"))
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
			name:     "osmosis-routed voucher unwinds one hop over its outer channel",
			denom:    osmosisRoutedAtom,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:     "staketia stToken goes to osmosis",
			denom:    "stutia",
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:        "stakedym stToken is rejected: the zone is deprecated",
			denom:       "stadym",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "stToken of a deprecated host zone is rejected",
			denom:       "stuosmo",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "made-up native denom is rejected",
			denom:       "ufake",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "native denom of a host zone (not its stToken) is rejected",
			denom:       "uatom",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "voucher whose outer hop is not the transfer port is rejected",
			denom:       wrongPortVoucher,
			expectedErr: types.ErrSweepDestinationUnavailable,
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

// setSweepHostZone registers the cosmoshub host zone so stuatom is a known stToken
func (s *KeeperTestSuite) setSweepHostZone() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{ChainId: "cosmoshub-4", HostDenom: "uatom"})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{ChainId: "osmosis-1", HostDenom: "uosmo", Deprecated: true})
}

// setBaseAccount stores a plain BaseAccount at the address, replacing any account already there
func (s *KeeperTestSuite) setBaseAccount(address sdk.AccAddress) {
	if existing := s.App.AccountKeeper.GetAccount(s.Ctx, address); existing != nil {
		s.App.AccountKeeper.RemoveAccount(s.Ctx, existing)
	}
	s.App.AccountKeeper.SetAccount(s.Ctx, s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, address))
}

// The copies in stakeibc/types must match the staketia and stakedym constants
func (s *KeeperTestSuite) TestSweepProtocolAddressesMatchModuleConstants() {
	s.Require().ElementsMatch([]string{
		staketiatypes.DepositAddress, staketiatypes.RedemptionAddress, staketiatypes.ClaimAddress,
		stakedymtypes.DepositAddress, stakedymtypes.RedemptionAddress, stakedymtypes.ClaimAddress,
		staketiatypes.SafeAddressOnStride, stakedymtypes.SafeAddressOnStride, staketiatypes.OperatorAddressOnStride,
	}, types.SweepProtocolAddresses())

	types.SweepOperatorAddress = s.TestAccs[0].String()
	s.T().Cleanup(func() { types.SweepOperatorAddress = "" })
	s.Require().Contains(types.SweepProtocolAddresses(), s.TestAccs[0].String())
}

// Protocol addresses (staketia and stakedym multisigs) as label -> address
func protocolAddressKinds() map[string]sdk.AccAddress {
	return map[string]sdk.AccAddress{
		"protocol_staketia_deposit":    sdk.MustAccAddressFromBech32(staketiatypes.DepositAddress),
		"protocol_staketia_redemption": sdk.MustAccAddressFromBech32(staketiatypes.RedemptionAddress),
		"protocol_staketia_claim":      sdk.MustAccAddressFromBech32(staketiatypes.ClaimAddress),
		"protocol_stakedym_deposit":    sdk.MustAccAddressFromBech32(stakedymtypes.DepositAddress),
		"protocol_stakedym_redemption": sdk.MustAccAddressFromBech32(stakedymtypes.RedemptionAddress),
		"protocol_stakedym_claim":      sdk.MustAccAddressFromBech32(stakedymtypes.ClaimAddress),
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

	// A blocked module address that a plain BaseAccount occupies (not a ModuleAccount)
	blockedBase := authtypes.NewModuleAddress(auctiontypes.ModuleName)
	s.Require().True(s.App.BankKeeper.BlockedAddr(blockedBase), "fixture: auction module address is blocked")
	s.setBaseAccount(blockedBase)
	kinds["blocked_base_account"] = blockedBase

	// Protocol multisigs are plain BaseAccounts that would pass every other rule
	for kind, address := range protocolAddressKinds() {
		s.setBaseAccount(address)
		kinds[kind] = address
	}

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

	// The operator signs the sweep and is never swept itself
	kinds["operator"] = s.TestAccs[0]
	s.setBaseAccount(kinds["operator"])
	types.SweepOperatorAddress = kinds["operator"].String()
	s.T().Cleanup(func() { types.SweepOperatorAddress = "" })
	protocol := keeper.SweepProtocolAddressSetForTest()

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
		{kind: "module", expectedSkip: true, expectedReason: "blocked module address"},
		{kind: "blocked_base_account", expectedSkip: true, expectedReason: "blocked module address"},
		{kind: "interchain_account", expectedSkip: true, expectedReason: "interchain account"},
		{kind: "operator", expectedSkip: true, expectedReason: "protocol address"},
	}
	for kind := range protocolAddressKinds() {
		testCases = append(testCases, struct {
			kind           string
			expectedSkip   bool
			expectedReason string
		}{kind: kind, expectedSkip: true, expectedReason: "protocol address"})
	}
	for _, tc := range testCases {
		s.Run(tc.kind, func() {
			reason, skip := keeper.SweepSkipReasonForTest(s.App.StakeibcKeeper, s.Ctx, kinds[tc.kind], escrows, protocol)
			s.Require().Equal(tc.expectedSkip, skip, fmt.Sprintf("skip for %s (reason %q)", tc.kind, reason))
			if tc.expectedSkip {
				s.Require().Equal(tc.expectedReason, reason)
			}
		})
	}
}

// openExtraTransferChannels opens n more transfer channels between Stride and the host chain
// on the connection CreateTransferChannel established, so tests can address channel-1 .. channel-n
func (s *KeeperTestSuite) openExtraTransferChannels(n int) {
	for i := 0; i < n; i++ {
		path := ibctesting.NewPath(s.StrideChain, s.HostChain).DisableUniqueChannelIDs()
		path.EndpointA.ClientID = s.TransferPath.EndpointA.ClientID
		path.EndpointA.ConnectionID = s.TransferPath.EndpointA.ConnectionID
		path.EndpointB.ClientID = s.TransferPath.EndpointB.ClientID
		path.EndpointB.ConnectionID = s.TransferPath.EndpointB.ConnectionID
		for _, endpoint := range []*ibctesting.Endpoint{path.EndpointA, path.EndpointB} {
			endpoint.ChannelConfig.PortID = ibctesting.TransferPort
			endpoint.ChannelConfig.Order = channeltypes.UNORDERED
			endpoint.ChannelConfig.Version = transfertypes.V1
		}

		s.Require().NoError(path.EndpointA.ChanOpenInit())
		apptesting.RunWithDifferentBechPrefix(sdk.Bech32MainPrefix, func() {
			s.Require().NoError(path.EndpointB.ChanOpenTry())
		})
		s.Require().NoError(path.EndpointA.ChanOpenAck())
		apptesting.RunWithDifferentBechPrefix(sdk.Bech32MainPrefix, func() {
			s.Require().NoError(path.EndpointB.ChanOpenConfirm())
		})
		s.Require().NoError(path.EndpointA.UpdateClient())
	}
	s.Ctx = s.StrideChain.GetContext()
}

type sweepTestCase struct {
	operator sdk.AccAddress
	holders  map[string]sdk.AccAddress
	atomIbc  string // single-hop uatom voucher over channel-0
	twoHop   string // Hub ATOM that came Hub -> Osmosis -> Stride, so its outer hop is channel-5
}

// SetupSweep opens channel-0 .. channel-5, funds a base account with an stToken, ustrd and
// a single-hop voucher, and registers the sweep operator
func (s *KeeperTestSuite) SetupSweep() sweepTestCase {
	s.CreateTransferChannel("GAIA")
	s.setSweepHostZone()
	s.openExtraTransferChannels(5)
	_, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, transfertypes.PortID, "channel-5")
	s.Require().True(found, "channel-5 should exist after opening five extra channels")

	operator := s.TestAccs[0]
	if s.App.AccountKeeper.GetAccount(s.Ctx, operator) == nil {
		s.SetNewAccount(operator) // assigns the next account number before SetAccount
	}
	types.SweepOperatorAddress = operator.String()
	s.T().Cleanup(func() { types.SweepOperatorAddress = "" })

	kinds := s.setupSweepAccountKinds()
	atomIbc := s.registerVoucher("uatom", transfertypes.NewHop(transfertypes.PortID, "channel-0"))
	twoHop := s.registerVoucher("uatom",
		transfertypes.NewHop(transfertypes.PortID, "channel-5"),
		transfertypes.NewHop(transfertypes.PortID, "channel-0"))

	return sweepTestCase{operator: operator, holders: kinds, atomIbc: atomIbc, twoHop: twoHop}
}

func (s *KeeperTestSuite) sweep(tc sweepTestCase, denoms []string, addresses ...sdk.AccAddress) (*types.MsgSweepTokensOffStrideResponse, error) {
	holders := make([]string, 0, len(addresses))
	for _, address := range addresses {
		holders = append(holders, address.String())
	}
	msg := types.NewMsgSweepTokensOffStride(tc.operator.String(), denoms, holders)
	s.Require().NoError(msg.ValidateBasic())
	return s.GetMsgServer().SweepTokensOffStride(s.Ctx, msg)
}

func (s *KeeperTestSuite) escrowBalance(channelId, denom string) sdkmath.Int {
	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, channelId)
	return s.App.BankKeeper.GetBalance(s.Ctx, escrow, denom).Amount
}

// One base account holding an stToken and ustrd: both leave over channel-5 to the same bytes
// with the osmo prefix, the full balances are escrowed, and only the listed denoms move
func (s *KeeperTestSuite) TestSweepTokensOffStride_NativeDenomsToOsmosis() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 1_000_000))
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStrd, 250_000))
	s.FundAccount(holder, sdk.NewInt64Coin("stuosmo", 9)) // not on the list, must stay

	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumTransfers)
	s.Require().Equal(uint64(0), resp.NumSkipped)

	endSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")
	s.Require().Equal(startSequence+2, endSequence, "two packets on channel-5")

	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64())
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStrd).Amount.Int64())
	s.Require().Equal(int64(9), s.App.BankKeeper.GetBalance(s.Ctx, holder, "stuosmo").Amount.Int64(), "unlisted denom untouched")
	s.Require().Equal(int64(1_000_000), s.escrowBalance("channel-5", sweepTestStToken).Int64(), "stToken escrowed on channel-5")
	s.Require().Equal(int64(250_000), s.escrowBalance("channel-5", sweepTestStrd).Int64(), "ustrd escrowed on channel-5")

	expectedReceiver := sdk.MustBech32ifyAddressBytes("osmo", holder)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, expectedReceiver)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepChannel, "channel-5")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepDenom, sweepTestStToken)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepAmount, "1000000")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)

	// The derived receiver decodes to exactly the sender's bytes
	receiverBytes, err := sdk.GetFromBech32(expectedReceiver, "osmo")
	s.Require().NoError(err)
	s.Require().Equal([]byte(holder), receiverBytes)
}

// A single-hop voucher goes back over channel-0 with the cosmos prefix and is burned (Stride
// is not the source of that denom)
func (s *KeeperTestSuite) TestSweepTokensOffStride_SingleHopVoucherUnwinds() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(tc.atomIbc, 40_000))
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-0")

	resp, err := s.sweep(tc, []string{tc.atomIbc}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)

	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-0"))
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, tc.atomIbc).Amount.Int64())
	s.Require().Zero(s.App.BankKeeper.GetSupply(s.Ctx, tc.atomIbc).Amount.Int64(), "voucher burned on the way back")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, sdk.MustBech32ifyAddressBytes("cosmos", holder))
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepChannel, "channel-0")
}

// Hub ATOM that reached Stride through Osmosis (trace channel-5, then Osmosis's channel-0)
// unwinds one hop: it goes back over channel-5 to an osmo address and Osmosis is left holding
// the remaining path transfer/channel-0/uatom
func (s *KeeperTestSuite) TestSweepTokensOffStride_TwoHopVoucherUnwindsOneHop() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(tc.twoHop, 500))
	hash, err := transfertypes.ParseHexHash(strings.TrimPrefix(tc.twoHop, "ibc/"))
	s.Require().NoError(err)
	trace, found := s.App.TransferKeeper.GetDenom(s.Ctx, hash)
	s.Require().True(found)
	s.Require().Equal("transfer/channel-5/transfer/channel-0/uatom", trace.Path(), "fixture: Osmosis-routed Hub ATOM")
	s.Require().Equal("transfer/channel-0/uatom", transfertypes.NewDenom(trace.Base, trace.Trace[1:]...).Path(), "path left after one unwind")
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{tc.twoHop}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"))
	s.Require().Zero(s.App.BankKeeper.GetSupply(s.Ctx, tc.twoHop).Amount.Int64(), "two-hop voucher burned")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, sdk.MustBech32ifyAddressBytes("osmo", holder))
}

// Every sweepable account type is swept; every non-sweepable one is skipped with its reason
// while the rest of the batch goes through and num_skipped counts it
func (s *KeeperTestSuite) TestSweepTokensOffStride_SkipRulesInOneBatch() {
	tc := s.SetupSweep()
	swept := []string{"base", "continuous_vesting", "delayed_vesting", "periodic_vesting", "stride_periodic_vesting"}
	skipped := map[string]string{
		"thirty_two_bytes":     "address is not 20 bytes",
		"unknown":              "account not found",
		"escrow":               "transfer escrow address",
		"module":               "blocked module address",
		"blocked_base_account": "blocked module address",
		"interchain_account":   "interchain account",
	}
	for kind := range protocolAddressKinds() {
		skipped[kind] = "protocol address"
	}
	addresses := []sdk.AccAddress{}
	for _, kind := range swept {
		s.FundAccount(tc.holders[kind], sdk.NewInt64Coin(sweepTestStToken, 1_000))
		addresses = append(addresses, tc.holders[kind])
	}
	// The module account and the escrow hold a balance too: skipping must leave it in place
	s.FundModuleAccount(distrtypes.ModuleName, sdk.NewInt64Coin(sweepTestStToken, 777))
	s.FundAccount(tc.holders["escrow"], sdk.NewInt64Coin(sweepTestStToken, 555))
	s.FundAccount(tc.holders["interchain_account"], sdk.NewInt64Coin(sweepTestStToken, 333))
	for kind := range skipped {
		addresses = append(addresses, tc.holders[kind])
	}
	// Protocol multisigs hold real balances that module code spends from: none may move
	for kind := range protocolAddressKinds() {
		s.FundAccount(tc.holders[kind], sdk.NewInt64Coin(sweepTestStToken, 111))
	}

	resp, err := s.sweep(tc, []string{sweepTestStToken}, addresses...)
	s.Require().NoError(err)
	s.Require().Equal(uint64(len(swept)), resp.NumTransfers)
	s.Require().Equal(uint64(len(skipped)), resp.NumSkipped)

	for _, kind := range swept {
		s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, tc.holders[kind], sweepTestStToken).Amount.Int64(), kind)
	}
	s.Require().Equal(int64(777), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["module"], sweepTestStToken).Amount.Int64())
	s.Require().Equal(int64(555), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["escrow"], sweepTestStToken).Amount.Int64())
	s.Require().Equal(int64(333), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["interchain_account"], sweepTestStToken).Amount.Int64())
	for kind := range protocolAddressKinds() {
		s.Require().Equal(int64(111), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders[kind], sweepTestStToken).Amount.Int64(), kind)
	}

	for kind, reason := range skipped {
		s.CheckEventValueEmitted(types.EventTypeSweepSkipped, types.AttributeKeySweepAddress, tc.holders[kind].String())
		s.CheckEventValueEmitted(types.EventTypeSweepSkipped, types.AttributeKeySweepReason, reason)
	}
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepSkipped), len(skipped))
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), len(swept))
}

// A vesting account is swept for its spendable balance only: the STRD its schedule still locks
// stays behind (the escrow send would refuse it and fail the batch), a fully locked account is
// skipped silently like a zero balance, and its non-vesting denoms move in full
func (s *KeeperTestSuite) TestSweepTokensOffStride_VestingSweepsOnlySpendable() {
	tc := s.SetupSweep()
	now := s.Ctx.BlockTime().Unix()
	holders := apptesting.CreateRandomAccounts(2)
	original := sdk.NewCoins(sdk.NewInt64Coin(sweepTestStrd, 1_000))

	// 40% of a linear schedule has elapsed: 400 vested, 600 still locked
	partlyVested, err := vestingtypes.NewContinuousVestingAccount(
		s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, holders[0]).(*authtypes.BaseAccount), original, now-400, now+600)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, partlyVested)
	s.FundAccount(holders[0], sdk.NewInt64Coin(sweepTestStrd, 1_000))
	s.FundAccount(holders[0], sdk.NewInt64Coin(sweepTestStToken, 50))
	s.Require().Equal(int64(600), s.App.BankKeeper.LockedCoins(s.Ctx, holders[0]).AmountOf(sweepTestStrd).Int64(), "fixture: 600 locked")

	// Nothing vested yet: every ustrd is locked, nothing is spendable
	fullyLocked, err := vestingtypes.NewContinuousVestingAccount(
		s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, holders[1]).(*authtypes.BaseAccount), original, now, now+1_000)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, fullyLocked)
	s.FundAccount(holders[1], sdk.NewInt64Coin(sweepTestStrd, 1_000))

	resp, err := s.sweep(tc, []string{sweepTestStrd, sweepTestStToken}, holders[0], holders[1])
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumTransfers, "400 ustrd and 50 stuatom from the first holder; nothing from the second")
	s.Require().Equal(uint64(0), resp.NumSkipped, "a fully locked balance is silent, not a skipped address")

	s.Require().Equal(int64(600), s.App.BankKeeper.GetBalance(s.Ctx, holders[0], sweepTestStrd).Amount.Int64(), "locked ustrd stays")
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holders[0], sweepTestStToken).Amount.Int64(), "stToken moves in full")
	s.Require().Equal(int64(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holders[1], sweepTestStrd).Amount.Int64(), "fully locked holder untouched")
	s.Require().Equal(int64(400), s.escrowBalance("channel-5", sweepTestStrd).Int64(), "only the spendable 400 escrowed")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepAmount, "400")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)
}

// A vesting account whose locked coins are partly delegated has less locked than its schedule
// says (the delegated part has already left the balance), and only the spendable amount moves
func (s *KeeperTestSuite) TestSweepTokensOffStride_VestingWithDelegatedLockedSweepsSpendable() {
	tc := s.SetupSweep()
	now := s.Ctx.BlockTime().Unix()
	holder := apptesting.CreateRandomAccounts(1)[0]
	original := sdk.NewCoins(sdk.NewInt64Coin(sweepTestStrd, 1_000))

	// 600 still locked by the schedule, 300 of the vesting coins are delegated: 300 stay locked
	account, err := vestingtypes.NewContinuousVestingAccount(
		s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, holder).(*authtypes.BaseAccount), original, now-400, now+600)
	s.Require().NoError(err)
	account.DelegatedVesting = sdk.NewCoins(sdk.NewInt64Coin(sweepTestStrd, 300))
	s.App.AccountKeeper.SetAccount(s.Ctx, account)
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStrd, 700)) // 1000 vesting minus 300 delegated away
	s.Require().Equal(int64(300), s.App.BankKeeper.LockedCoins(s.Ctx, holder).AmountOf(sweepTestStrd).Int64(), "fixture: 300 locked")

	resp, err := s.sweep(tc, []string{sweepTestStrd}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)
	s.Require().Equal(int64(300), s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStrd).Amount.Int64(), "locked ustrd stays")
	s.Require().Equal(int64(400), s.escrowBalance("channel-5", sweepTestStrd).Int64(), "only the spendable 400 escrowed")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepAmount, "400")
}

// A zero balance is skipped silently: no transfer, no skip event, not counted
func (s *KeeperTestSuite) TestSweepTokensOffStride_ZeroBalanceSilent() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), resp.NumTransfers)
	s.Require().Equal(uint64(0), resp.NumSkipped)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"))
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
}

// A holder with two of three listed denoms yields exactly two transfers
func (s *KeeperTestSuite) TestSweepTokensOffStride_TwoOfThreeDenoms() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(holder, sdk.NewInt64Coin(tc.atomIbc, 20))

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd, tc.atomIbc}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumTransfers)
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), 2)
}

// A denom with no destination rejects the whole tx before any address is touched
func (s *KeeperTestSuite) TestSweepTokensOffStride_UnwhitelistedVoucherRejectsBatch() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	luna := s.registerVoucher("uluna", transfertypes.NewHop(transfertypes.PortID, "channel-1")) // channel-1 exists, not whitelisted
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(holder, sdk.NewInt64Coin(luna, 10))

	_, err := s.sweep(tc, []string{sweepTestStToken, luna}, holder)
	s.Require().ErrorIs(err, types.ErrSweepDestinationUnavailable)
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "nothing moved")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
}

// A transfer error (a whitelisted channel that does not exist on chain) rejects the whole tx
// even when earlier holders in the batch were fine, and the failed tx moves nothing.
//
// In production baseapp runs every message in a cache-wrapped context and only writes it on
// success (runMsgs / runTx), so a keeper error discards every state change the message made.
// The keeper suite calls the keeper directly with no baseapp in front, so this test reproduces
// that boundary by hand: it runs the sweep on s.Ctx.CacheContext(), never calls write, first
// proves on the cache that the first holder's transfer really happened (otherwise any
// implementation that errors before touching a holder would pass), then asserts on s.Ctx that
// none of it survived.
//
// channel-24 does not exist on the test chain. ibc-go v11's Transfer looks for a v1 channel of
// that id, finds none, and falls through to the V2 (client-id) path, which fails on the unknown
// id; the "channel-24" in the error text comes from this keeper's own wrap ("unable to sweep ...
// over channel-24"), not from ibc-go, which is fine because the wrap is what ops read.
func (s *KeeperTestSuite) TestSweepTokensOffStride_TransferErrorRejectsBatch() {
	tc := s.SetupSweep()
	junoVoucher := s.registerVoucher("ujuno", transfertypes.NewHop(transfertypes.PortID, "channel-24")) // whitelisted, no such channel
	first := tc.holders["base"]
	second := tc.holders["continuous_vesting"]
	s.FundAccount(first, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(second, sdk.NewInt64Coin(junoVoucher, 10))

	sequenceBefore := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")
	escrowBefore := s.escrowBalance("channel-5", sweepTestStToken)

	msg := types.NewMsgSweepTokensOffStride(tc.operator.String(), []string{sweepTestStToken, junoVoucher},
		[]string{first.String(), second.String()})
	s.Require().NoError(msg.ValidateBasic())

	// The first holder's transfer succeeds inside the cache, the second holder's fails
	cacheCtx, _ := s.Ctx.CacheContext()
	_, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(cacheCtx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "channel-24")

	// Inside the cache the first transfer did happen: balance moved to escrow, one sequence
	// consumed, one commitment written. This is what makes the rollback assertions below meaningful
	s.Require().Zero(s.App.BankKeeper.GetBalance(cacheCtx, first, sweepTestStToken).Amount.Int64(), "first holder drained inside the cache")
	escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-5")
	s.Require().Equal(escrowBefore.AddRaw(10), s.App.BankKeeper.GetBalance(cacheCtx, escrowAddress, sweepTestStToken).Amount, "escrowed inside the cache")
	cacheSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(cacheCtx, transfertypes.PortID, "channel-5")
	s.Require().True(found)
	s.Require().Equal(sequenceBefore+1, cacheSequence, "one packet sequence consumed inside the cache")
	s.Require().Len(s.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(cacheCtx, transfertypes.PortID, "channel-5"), 1, "one commitment inside the cache")

	// Nothing written to the cache reaches s.Ctx: balances, the channel sequence, the packet
	// commitment and the events are all as they were before the call
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, first, sweepTestStToken).Amount.Int64(), "first holder untouched")
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, second, junoVoucher).Amount.Int64(), "second holder untouched")
	s.Require().Equal(escrowBefore, s.escrowBalance("channel-5", sweepTestStToken), "escrow untouched")
	s.Require().Equal(sequenceBefore, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"), "no packet sequence consumed")
	s.Require().Empty(s.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(s.Ctx, transfertypes.PortID, "channel-5"), "no packet commitment")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)
}

// The positive counterpart: through the msg server on s.Ctx, a good batch leaves its state
// changes in place (the cache boundary only discards on error)
func (s *KeeperTestSuite) TestSweepTokensOffStride_SuccessfulBatchPersists() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	sequenceBefore := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)

	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "holder drained")
	s.Require().Equal(int64(10), s.escrowBalance("channel-5", sweepTestStToken).Int64(), "escrowed")
	s.Require().Equal(sequenceBefore+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"), "one packet sent")
	s.Require().Len(s.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(s.Ctx, transfertypes.PortID, "channel-5"), 1, "one commitment")
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), 1)
}

// The ICS-20 timeout refund lands the escrowed balance back on the holder
func (s *KeeperTestSuite) TestSweepTokensOffStride_TimeoutRefundsHolder() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 1_000))

	_, err := s.sweep(tc, []string{sweepTestStToken}, holder)
	s.Require().NoError(err)
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64())

	// Replay the timeout through the transfer keeper with the packet the sweep built
	data := transfertypes.NewInternalTransferRepresentation(
		transfertypes.Token{Denom: transfertypes.NewDenom(sweepTestStToken), Amount: "1000"},
		holder.String(),
		sdk.MustBech32ifyAddressBytes("osmo", holder),
		"",
	)
	err = s.App.TransferKeeper.OnTimeoutPacket(s.Ctx, transfertypes.PortID, "channel-5", data)
	s.Require().NoError(err)
	s.Require().Equal(int64(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "refunded")
	s.Require().Zero(s.escrowBalance("channel-5", sweepTestStToken).Int64())
}

// TestSweepRealMainnetValues pins values read from Stride mainnet REST on 2026-09-30, so the
// denom hashes, destination lookups, address derivation and escrow address are checked against
// reality rather than against constants the code under test also produced
func (s *KeeperTestSuite) TestSweepRealMainnetValues() {
	s.Run("denom hashes resolve to the expected channel and prefix", func() {
		testCases := []struct {
			name           string
			base           string
			hops           []transfertypes.Hop
			expectedDenom  string
			expectedTarget keeper.SweepDestinationForTest
		}{
			{
				name:           "atom over channel-0",
				base:           "uatom",
				hops:           []transfertypes.Hop{transfertypes.NewHop(transfertypes.PortID, "channel-0")},
				expectedDenom:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
				expectedTarget: keeper.SweepDestinationForTest{ChannelId: "channel-0", Bech32Prefix: "cosmos"},
			},
			{
				name:           "tia over channel-162",
				base:           "utia",
				hops:           []transfertypes.Hop{transfertypes.NewHop(transfertypes.PortID, "channel-162")},
				expectedDenom:  "ibc/BF3B4F53F3694B66E13C23107C84B6485BD2B96296BB7EC680EA77BBA75B4801",
				expectedTarget: keeper.SweepDestinationForTest{ChannelId: "channel-162", Bech32Prefix: "celestia"},
			},
			{
				name:           "osmo over channel-5",
				base:           "uosmo",
				hops:           []transfertypes.Hop{transfertypes.NewHop(transfertypes.PortID, "channel-5")},
				expectedDenom:  "ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34",
				expectedTarget: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
			},
			{
				name: "atom routed through osmosis (channel-5 outermost, then channel-0)",
				base: "uatom",
				hops: []transfertypes.Hop{
					transfertypes.NewHop(transfertypes.PortID, "channel-5"),
					transfertypes.NewHop(transfertypes.PortID, "channel-0"),
				},
				expectedDenom:  "ibc/070039AB58034252D2E86210276E05FE25B2FB41D15603185FD57960AFBAEDB6",
				expectedTarget: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
			},
		}
		for _, tc := range testCases {
			denom := transfertypes.NewDenom(tc.base, tc.hops...)
			s.App.TransferKeeper.SetDenom(s.Ctx, denom)
			s.Require().Equal(tc.expectedDenom, denom.IBCDenom(), tc.name)

			destination, err := keeper.ResolveSweepDestinationForTest(s.App.StakeibcKeeper, s.Ctx, tc.expectedDenom)
			s.Require().NoError(err, tc.name)
			s.Require().Equal(tc.expectedTarget, destination, tc.name)
		}

		s.Require().Equal(staketiatypes.CelestiaNativeTokenIBCDenom,
			"ibc/BF3B4F53F3694B66E13C23107C84B6485BD2B96296BB7EC680EA77BBA75B4801")
	})

	s.Run("stride address converts to the real osmo address", func() {
		holder := sdk.MustAccAddressFromBech32("stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c")

		// Same derivation the keeper uses for the receiver
		osmoAddress := sdk.MustBech32ifyAddressBytes("osmo", holder)
		s.Require().Equal("osmo1am99pcvynqqhyrwqfvfmnvxjk96rn46lj4pkkx", osmoAddress)

		cosmosAddress := sdk.MustBech32ifyAddressBytes("cosmos", holder)
		roundTripped, err := sdk.GetFromBech32(cosmosAddress, "cosmos")
		s.Require().NoError(err)
		s.Require().Len(roundTripped, 20)
		s.Require().Equal([]byte(holder), roundTripped)
	})

	s.Run("channel-5 escrow address", func() {
		escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-5")
		s.Require().Equal("stride16h2ynrzwhxgjnd0hswkvdvq9nav9kklq08fhf4",
			sdk.MustBech32ifyAddressBytes("stride", escrow))
	})
}
