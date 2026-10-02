package v35_test

import (
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// seedFlaggedZone stores a host zone with two validators carrying non-zero in-progress counters
func (s *UpgradeTestSuite) seedFlaggedZone(chainId, connectionId string, deprecated bool) {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, withInBoundsRates(stakeibctypes.HostZone{
		ChainId:      chainId,
		ConnectionId: connectionId,
		Deprecated:   deprecated,
		Validators: []*stakeibctypes.Validator{
			{Address: chainId + "valoper1", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 12},
			{Address: chainId + "valoper2", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 1},
			{Address: chainId + "valoper3", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 0},
		},
	}))
}

func (s *UpgradeTestSuite) flags(chainId string) []int64 {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
	s.Require().True(found)
	flags := []int64{}
	for _, validator := range hostZone.Validators {
		flags = append(flags, validator.DelegationChangesInProgress)
	}
	return flags
}

func (s *UpgradeTestSuite) mockDelegationChannel(chainId, connectionId, channelId string) (portId string) {
	owner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_DELEGATION)
	s.MockICAChannel(connectionId, channelId, owner, chainId+"ica")
	portId, _ = icatypes.NewControllerPortID(owner)
	return portId
}

func (s *UpgradeTestSuite) TestResetStaleDelegationChangesInProgress() {
	// cosmoshub-4: open channel, nothing in flight -> reset
	s.seedFlaggedZone("cosmoshub-4", "connection-0", false)
	s.mockDelegationChannel("cosmoshub-4", "connection-0", "channel-863")

	// juno-1: open channel with an unacked packet -> skipped. The fixture is a bare packet
	// commitment, which is exactly what ibc-go's SendPacket leaves behind for every packet
	// until its ack or timeout is processed (channel keeper: SendPacket -> SetPacketCommitment,
	// AcknowledgePacket/TimeoutPacket -> deletePacketCommitment); PR 2's day-epoch hook test
	// (TestBeforeEpochStart_DayEpoch_KeptFlowsRunNoNewRecord) drives a real SubmitTxsDayEpoch
	// and asserts the sequence advance that goes with such a commitment
	s.seedFlaggedZone("juno-1", "connection-1", false)
	junoPort := s.mockDelegationChannel("juno-1", "connection-1", "channel-10")
	s.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(s.Ctx, junoPort, "channel-10", 7, []byte{1})

	// haqq_11235-1: no active channel at all -> skipped
	s.seedFlaggedZone("haqq_11235-1", "connection-2", false)

	// evmos_9001-2: deprecated, open channel -> skipped
	s.seedFlaggedZone("evmos_9001-2", "connection-3", true)
	s.mockDelegationChannel("evmos_9001-2", "connection-3", "channel-20")

	// osmosis-1: the delegation channel is open on a different connection than the host zone's
	// ConnectionId -> not the zone's channel, skipped
	s.seedFlaggedZone("osmosis-1", "connection-4", false)
	s.mockDelegationChannel("osmosis-1", "connection-99", "channel-30")

	v35.ResetStaleDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper)

	s.Require().Equal([]int64{12, 1, 0}, s.flags("osmosis-1"), "channel on another connection: untouched")
	s.Require().Equal([]int64{0, 0, 0}, s.flags("cosmoshub-4"), "open channel, no packets: reset")
	s.Require().Equal([]int64{12, 1, 0}, s.flags("juno-1"), "unacked packet: untouched")
	s.Require().Equal([]int64{12, 1, 0}, s.flags("haqq_11235-1"), "no channel: untouched")
	s.Require().Equal([]int64{12, 1, 0}, s.flags("evmos_9001-2"), "deprecated: untouched")
}

func (s *UpgradeTestSuite) TestResetStaleDelegationChangesInProgress_NoZones() {
	s.Require().NotPanics(func() { v35.ResetStaleDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper) })
}
