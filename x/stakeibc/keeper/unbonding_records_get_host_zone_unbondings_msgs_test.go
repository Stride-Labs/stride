package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v16/x/records/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

type ValidatorUnbonding struct {
	Validator    string
	UnbondAmount sdkmath.Int
}

type UnbondingTestCase struct {
	hostZone                   types.HostZone
	totalUnbondAmount          sdkmath.Int
	delegationChannelID        string
	delegationPortID           string
	channelStartSequence       uint64
	expectedUnbondingRecordIds []uint64
}

func (s *KeeperTestSuite) SetupTestUnbondFromHostZone(
	totalWeight int64,
	totalStake sdkmath.Int,
	unbondAmount sdkmath.Int,
	validators []*types.Validator,
) UnbondingTestCase {
	delegationAccountOwner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Sanity checks:
	//  - total stake matches
	//  - total weights sum to 100
	actualTotalStake := sdkmath.ZeroInt()
	actualTotalWeights := uint64(0)
	for _, validator := range validators {
		actualTotalStake = actualTotalStake.Add(validator.Delegation)
		actualTotalWeights += validator.Weight
	}
	s.Require().Equal(totalStake.Int64(), actualTotalStake.Int64(), "test setup failed - total stake does not match")
	s.Require().Equal(totalWeight, int64(actualTotalWeights), "test setup failed - total weight does not match")

	// Store the validators on the host zone
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators:           validators,
		TotalDelegations:     totalStake,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store the total unbond amount across two epoch unbonding records
	halfUnbondAmount := unbondAmount.Quo(sdkmath.NewInt(2))
	for i := uint64(1); i <= 2; i++ {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: i,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					NativeTokenAmount: halfUnbondAmount,
				},
			},
		})
	}

	// Mock the epoch tracker to timeout 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get tx seq number before the ICA was submitted to check whether an ICA was submitted
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found before ica")

	return UnbondingTestCase{
		hostZone:                   hostZone,
		totalUnbondAmount:          unbondAmount,
		delegationChannelID:        delegationChannelID,
		delegationPortID:           delegationPortID,
		channelStartSequence:       startSequence,
		expectedUnbondingRecordIds: []uint64{1, 2},
	}
}

// Helper function to check that an undelegation ICA was submitted and that the callback data
// holds the expected unbondings for each validator
func (s *KeeperTestSuite) CheckUnbondingMessages(tc UnbondingTestCase, expectedUnbondings []ValidatorUnbonding) {
	// Trigger unbonding
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected when calling unbond from host")

	// Check that sequence number incremented from a sent ICA
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "sequence number should have incremented")

	// Check that callback data was stored
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1, "there should only be one callback data stored")

	// Check host zone and epoch unbonding record id's
	var actualCallback types.UndelegateCallback
	err = proto.Unmarshal(callbackData[0].CallbackArgs, &actualCallback)
	s.Require().NoError(err, "no error expected when unmarshalling callback args")

	s.Require().Equal(HostChainId, actualCallback.HostZoneId, "chain-id on callback")
	s.Require().Equal(tc.expectedUnbondingRecordIds, actualCallback.EpochUnbondingRecordIds, "unbonding record id's on callback")

	// Print all of the true unbondings
	for _, actualSplit := range actualCallback.SplitDelegations {
		s.T().Logf("actualSplit: %s, %s", actualSplit.Validator, actualSplit.Amount)
	}

	// Check splits from callback data align with expected unbondings
	// s.Require().Len(actualCallback.SplitDelegations, len(expectedUnbondings), "number of unbonding messages")
	// for i, expected := range expectedUnbondings {
	// 	actualSplit := actualCallback.SplitDelegations[i]
	// 	s.Require().Equal(expected.Validator, actualSplit.Validator, "callback message validator - index %d", i)
	// 	s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.Amount.Int64(), "callback message amount - index %d", i)
	// }

	// // Check the delegation change in progress was incremented from each that had an unbonding
	// actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	// s.Require().True(found, "host zone should have been found")

	// for _, actualValidator := range actualHostZone.Validators {
	// 	validatorUnbonded := false
	// 	for _, unbondedVal := range expectedUnbondings {
	// 		if actualValidator.Address == unbondedVal.Validator {
	// 			validatorUnbonded = true
	// 		}
	// 	}

	// 	expectedDelegationChangesInProgress := 0
	// 	if validatorUnbonded {
	// 		expectedDelegationChangesInProgress = 1
	// 	}
	// 	s.Require().Equal(expectedDelegationChangesInProgress, int(actualValidator.DelegationChangesInProgress),
	// 		"validator %s delegation changes in progress", actualValidator.Address)
	// }

	// // Check that the unbond event was emitted with the proper unbond amount
	// s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, tc.totalUnbondAmount.String())
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondOnlyZeroWeightVals() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:        50
	// Stake After Unbond: 1200
	totalUnbondAmount := sdkmath.NewInt(50)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 100-120 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1200 = 420, Capacity: 420-420 = 0
		// No capacity -> unbondings
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1200 = 360, Capacity: 300-360 = -60 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 200 - 120 = 80
		// >>> Ratio: 110/200 = 0.55 -> #3 Priority <<<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1200 = 180, Capacity: 160-180 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 50 - 40 = 10 unbond remaining
		// valE has #2 priority - unbond up to remaining
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondTotalLessThanTotalLSM() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:       150
	// Stake After Unbond: 1100
	totalUnbondAmount := sdkmath.NewInt(150)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 100-110 = -10 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1100 = 385, Capacity: 420-385 = 35
		// >>> Ratio: 385/420 = 0.91 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1100 = 330, Capacity: 300-330 = -30 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 200 - 110 = 90
		// >>> Ratio: 110/200 = 0.55 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1100 = 165, Capacity: 160-165 = -5 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 150 - 40 = 110 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 150 - 40 - 30 = 80 unbond remaining
		// valF has #3 priority - unbond up to remaining
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(80)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_HubUnbondingsTest() {
	validators := []*types.Validator{
		{Address: "notional", Weight: 5619, Delegation: sdkmath.NewInt(289702451015)},
		{Address: "binaryholdings", Weight: 5623, Delegation: sdkmath.NewInt(256289463439)},
		{Address: "cryptocrew", Weight: 5621, Delegation: sdkmath.NewInt(235481269896)},
		{Address: "lavenderfive", Weight: 4371, Delegation: sdkmath.NewInt(219938271925)},
		{Address: "cephalopod", Weight: 4369, Delegation: sdkmath.NewInt(201224336140)},
		{Address: "jabbey", Weight: 4375, Delegation: sdkmath.NewInt(200179203682)},
		{Address: "keplr", Weight: 4373, Delegation: sdkmath.NewInt(199287672546)},
		{Address: "citadelone", Weight: 2869, Delegation: sdkmath.NewInt(196608485155)},
		{Address: "polkachu", Weight: 3121, Delegation: sdkmath.NewInt(163925247132)},
		{Address: "icycro", Weight: 3123, Delegation: sdkmath.NewInt(160785497206)},
		{Address: "silknodes", Weight: 1873, Delegation: sdkmath.NewInt(155405500030)},
		{Address: "stakely", Weight: 2871, Delegation: sdkmath.NewInt(150979731610)},
		{Address: "stakelab", Weight: 2619, Delegation: sdkmath.NewInt(150806796149)},
		{Address: "cosmostation", Weight: 3119, Delegation: sdkmath.NewInt(149219956121)},
		{Address: "a41", Weight: 3125, Delegation: sdkmath.NewInt(142413440797)},
		{Address: "stakin", Weight: 2621, Delegation: sdkmath.NewInt(140614308769)},
		{Address: "cosmicvalidator", Weight: 2875, Delegation: sdkmath.NewInt(131020365529)},
		{Address: "bronbro", Weight: 2873, Delegation: sdkmath.NewInt(130987220925)},
		{Address: "nodestake", Weight: 2623, Delegation: sdkmath.NewInt(121922567724)},
		{Address: "stakewolle", Weight: 2123, Delegation: sdkmath.NewInt(112197355120)},
		{Address: "cosmosspaces", Weight: 2625, Delegation: sdkmath.NewInt(111416747252)},
		{Address: "enigma", Weight: 2375, Delegation: sdkmath.NewInt(108296049161)},
		{Address: "simplystaking", Weight: 2371, Delegation: sdkmath.NewInt(108051925780)},
		{Address: "stakewithus", Weight: 1871, Delegation: sdkmath.NewInt(95029022733)},
		{Address: "tienthuattoan", Weight: 1875, Delegation: sdkmath.NewInt(94961525727)},
		{Address: "forbole", Weight: 1869, Delegation: sdkmath.NewInt(94413775492)},
		{Address: "allnodes", Weight: 1218, Delegation: sdkmath.NewInt(54249917896)},
		{Address: "sg1", Weight: 593, Delegation: sdkmath.NewInt(26191549170)},
		{Address: "posthuman", Weight: 2369, Delegation: sdkmath.NewInt(23195510401)},
		{Address: "multichain", Weight: 516, Delegation: sdkmath.NewInt(22748654714)},
		{Address: "shapeshiftdao", Weight: 2119, Delegation: sdkmath.NewInt(19020339797)},
		{Address: "stakesystems", Weight: 411, Delegation: sdkmath.NewInt(18003000000)},
		{Address: "stakecito", Weight: 407, Delegation: sdkmath.NewInt(17847274790)},
		{Address: "prism", Weight: 357, Delegation: sdkmath.NewInt(15590639920)},
		{Address: "imperator", Weight: 347, Delegation: sdkmath.NewInt(15145607350)},
		{Address: "ubikcapital", Weight: 324, Delegation: sdkmath.NewInt(14118447728)},
		{Address: "chillvalidation", Weight: 2373, Delegation: sdkmath.NewInt(14092952691)},
		{Address: "s16researchventures", Weight: 2121, Delegation: sdkmath.NewInt(12568673845)},
		{Address: "dsrv", Weight: 2125, Delegation: sdkmath.NewInt(12556710011)},
		{Address: "goldenratio", Weight: 278, Delegation: sdkmath.NewInt(12039749700)},
		{Address: "prodelegators", Weight: 276, Delegation: sdkmath.NewInt(11967441720)},
		{Address: "provalidator", Weight: 273, Delegation: sdkmath.NewInt(11835174281)},
		{Address: "cyphercore", Weight: 257, Delegation: sdkmath.NewInt(11127049044)},
		{Address: "bharvest", Weight: 255, Delegation: sdkmath.NewInt(11003603387)},
		{Address: "swissstaking", Weight: 119, Delegation: sdkmath.NewInt(4926041540)},
		{Address: "sikka", Weight: 107, Delegation: sdkmath.NewInt(4396037806)},
		{Address: "figment", Weight: 101, Delegation: sdkmath.NewInt(4086850000)},
		{Address: "dforce", Weight: 98, Delegation: sdkmath.NewInt(3961300000)},
		{Address: "stakefish", Weight: 73, Delegation: sdkmath.NewInt(2832319160)},
		{Address: "frens", Weight: 70, Delegation: sdkmath.NewInt(2737211090)},
		{Address: "gatadao", Weight: 49, Delegation: sdkmath.NewInt(1770364440)},
		{Address: "westaking", Weight: 44, Delegation: sdkmath.NewInt(1542120270)},
		{Address: "meria", Weight: 44, Delegation: sdkmath.NewInt(1529599324)},
		{Address: "0base", Weight: 43, Delegation: sdkmath.NewInt(1513481350)},
		{Address: "whispernode", Weight: 39, Delegation: sdkmath.NewInt(1328361300)},
		{Address: "everstake", Weight: 33, Delegation: sdkmath.NewInt(1070988967)},
		{Address: "blockpower", Weight: 32, Delegation: sdkmath.NewInt(1011568993)},
		{Address: "01node", Weight: 25, Delegation: sdkmath.NewInt(678540000)},
		{Address: "ledger", Weight: 24, Delegation: sdkmath.NewInt(630085660)},
		{Address: "pupmos", Weight: 23, Delegation: sdkmath.NewInt(620632350)},
		{Address: "klubstaking", Weight: 23, Delegation: sdkmath.NewInt(587500000)},
		{Address: "oni", Weight: 22, Delegation: sdkmath.NewInt(562776160)},
		{Address: "game", Weight: 22, Delegation: sdkmath.NewInt(542445260)},
		{Address: "ezstaking", Weight: 21, Delegation: sdkmath.NewInt(534036530)},
		{Address: "highstakes", Weight: 21, Delegation: sdkmath.NewInt(502499996)},
		{Address: "chorusone", Weight: 19, Delegation: sdkmath.NewInt(424507080)},
		{Address: "iqlusion", Weight: 19, Delegation: sdkmath.NewInt(420055380)},
		{Address: "smartstake", Weight: 19, Delegation: sdkmath.NewInt(410444889)},
		{Address: "nodeguardians", Weight: 17, Delegation: sdkmath.NewInt(334470000)},
		{Address: "p2p", Weight: 13, Delegation: sdkmath.NewInt(170178580)},
		{Address: "stir", Weight: 13, Delegation: sdkmath.NewInt(144773556)},
		{Address: "blockhunters", Weight: 13, Delegation: sdkmath.NewInt(137068940)},
		{Address: "crosnest", Weight: 12, Delegation: sdkmath.NewInt(125264450)},
		{Address: "coinbasecustody", Weight: 11, Delegation: sdkmath.NewInt(86678850)},
		{Address: "witval", Weight: 11, Delegation: sdkmath.NewInt(66061360)},
		{Address: "zkv", Weight: 11, Delegation: sdkmath.NewInt(64803420)},
		{Address: "danku", Weight: 11, Delegation: sdkmath.NewInt(59100000)},
		{Address: "dhkdao", Weight: 11, Delegation: sdkmath.NewInt(52453179)},
		{Address: "binancenode", Weight: 11, Delegation: sdkmath.NewInt(52100031)},
		{Address: "crowdcontrol", Weight: 11, Delegation: sdkmath.NewInt(50809130)},
		{Address: "auditone", Weight: 10, Delegation: sdkmath.NewInt(34990000)},
		{Address: "blocksunited", Weight: 10, Delegation: sdkmath.NewInt(31224699)},
		{Address: "citizencosmos", Weight: 5625, Delegation: sdkmath.NewInt(31050240), SlashQueryInProgress: true},
		{Address: "stargaze", Weight: 10, Delegation: sdkmath.NewInt(28999998)},
		{Address: "in3s", Weight: 10, Delegation: sdkmath.NewInt(26000000)},
		{Address: "interstellarlounge", Weight: 10, Delegation: sdkmath.NewInt(23068000)},
		{Address: "dokiacapital", Weight: 10, Delegation: sdkmath.NewInt(22391760)},
		{Address: "blockscape", Weight: 10, Delegation: sdkmath.NewInt(21000000)},
		{Address: "madeinblock", Weight: 10, Delegation: sdkmath.NewInt(19323690)},
		{Address: "freshstaking", Weight: 10, Delegation: sdkmath.NewInt(19022000)},
		{Address: "rockawayxinfra", Weight: 10, Delegation: sdkmath.NewInt(14000000)},
		{Address: "genesislab", Weight: 10, Delegation: sdkmath.NewInt(9000000)},
		{Address: "terranodes", Weight: 10, Delegation: sdkmath.NewInt(5000000)},
		{Address: "coinbasecloud", Weight: 10, Delegation: sdkmath.NewInt(4999000)},
		{Address: "irisnetbianjie", Weight: 10, Delegation: sdkmath.NewInt(4262570)},
		{Address: "polychain", Weight: 10, Delegation: sdkmath.NewInt(1000000)},
		{Address: "bitvalidator", Weight: 10, Delegation: sdkmath.NewInt(1000000)},
		{Address: "dacm", Weight: 10, Delegation: sdkmath.NewInt(1000000)},
		{Address: "wetez", Weight: 10, Delegation: sdkmath.NewInt(0)},
		{Address: "blockdaemon", Weight: 10, Delegation: sdkmath.NewInt(0)},
	}

	totalUnbondAmount := sdkmath.NewInt(540313933163)
	totalStake := sdkmath.NewInt(0)
	for _, val := range validators {
		totalStake = totalStake.Add(val.Delegation)
	}
	totalWeight := int64(0)
	for _, val := range validators {
		totalWeight += int64(val.Weight)
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 150 - 40 = 110 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 150 - 40 - 30 = 80 unbond remaining
		// valF has #3 priority - unbond up to remaining
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(80)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_OldHubUnbondingsTest() {
	validators := []*types.Validator{
		{Address: "everstake", Weight: 0, Delegation: sdkmath.NewInt(1070988967)},
		{Address: "cosmostation", Weight: 3119, Delegation: sdkmath.NewInt(149109333065)},
		{Address: "citadelone", Weight: 2869, Delegation: sdkmath.NewInt(196542216268)},
		{Address: "cephalopod", Weight: 4369, Delegation: sdkmath.NewInt(201177237990)},
		{Address: "provalidator", Weight: 0, Delegation: sdkmath.NewInt(11835174281)},
		{Address: "forbole", Weight: 1869, Delegation: sdkmath.NewInt(94393627532)},
		{Address: "bharvest", Weight: 0, Delegation: sdkmath.NewInt(11003603387)},
		{Address: "imperator", Weight: 0, Delegation: sdkmath.NewInt(15145607350)},
		{Address: "simplystaking", Weight: 2371, Delegation: sdkmath.NewInt(108026366222)},
		{Address: "lavenderfive", Weight: 4371, Delegation: sdkmath.NewInt(219891152215)},
		{Address: "bronbro", Weight: 2873, Delegation: sdkmath.NewInt(130956249768)},
		{Address: "notional", Weight: 5619, Delegation: sdkmath.NewInt(289641877770)},
		{Address: "cryptocrew", Weight: 5621, Delegation: sdkmath.NewInt(235420675091)},
		{Address: "icycro", Weight: 3123, Delegation: sdkmath.NewInt(160251831030)},
		{Address: "danku", Weight: 0, Delegation: sdkmath.NewInt(59100000)},
		{Address: "stakin", Weight: 2621, Delegation: sdkmath.NewInt(140586054192)},
		{Address: "polkachu", Weight: 3121, Delegation: sdkmath.NewInt(163891602516)},
		{Address: "cyphercore", Weight: 0, Delegation: sdkmath.NewInt(11127049044)},
		{Address: "stakely", Weight: 2871, Delegation: sdkmath.NewInt(150948782013)},
		{Address: "blockscape", Weight: 0, Delegation: sdkmath.NewInt(21000000)},
		{Address: "wetez", Weight: 0, Delegation: sdkmath.NewInt(0)},
		{Address: "smartstake", Weight: 0, Delegation: sdkmath.NewInt(251401420)},
		{Address: "blockhunters", Weight: 0, Delegation: sdkmath.NewInt(137068940)},
		{Address: "frens", Weight: 0, Delegation: sdkmath.NewInt(2737211090)},
		{Address: "oni", Weight: 0, Delegation: sdkmath.NewInt(562776160)},
		{Address: "madeinblock", Weight: 0, Delegation: sdkmath.NewInt(19323690)},
		{Address: "genesislab", Weight: 0, Delegation: sdkmath.NewInt(9000000)},
		{Address: "auditone", Weight: 0, Delegation: sdkmath.NewInt(34990000)},
		{Address: "binaryholdings", Weight: 5623, Delegation: sdkmath.NewInt(256228847039)},
		{Address: "blockdaemon", Weight: 0, Delegation: sdkmath.NewInt(0)},
		{Address: "goldenratio", Weight: 0, Delegation: sdkmath.NewInt(11981619700)},
		{Address: "dhkdao", Weight: 0, Delegation: sdkmath.NewInt(52453179)},
		{Address: "ubikcapital", Weight: 0, Delegation: sdkmath.NewInt(14118447728)},
		{Address: "stakelab", Weight: 2619, Delegation: sdkmath.NewInt(150778563132)},
		{Address: "s16researchventures", Weight: 2121, Delegation: sdkmath.NewInt(12545809306)},
		{Address: "freshstaking", Weight: 0, Delegation: sdkmath.NewInt(15000000)},
		{Address: "zkv", Weight: 0, Delegation: sdkmath.NewInt(35803420)},
		{Address: "a41", Weight: 3125, Delegation: sdkmath.NewInt(142379753061)},
		{Address: "sg1", Weight: 0, Delegation: sdkmath.NewInt(26000549170)},
		{Address: "crosnest", Weight: 0, Delegation: sdkmath.NewInt(125264450)},
		{Address: "p2p", Weight: 0, Delegation: sdkmath.NewInt(170178580)},
		{Address: "whispernode", Weight: 0, Delegation: sdkmath.NewInt(1328361300)},
		{Address: "polychain", Weight: 0, Delegation: sdkmath.NewInt(1000000)},
		{Address: "shapeshiftdao", Weight: 2119, Delegation: sdkmath.NewInt(18997496818)},
		{Address: "witval", Weight: 0, Delegation: sdkmath.NewInt(66061360)},
		{Address: "tienthuattoan", Weight: 1875, Delegation: sdkmath.NewInt(94902313086)},
		{Address: "jabbey", Weight: 4375, Delegation: sdkmath.NewInt(200132040851)},
		{Address: "irisnetbianjie", Weight: 0, Delegation: sdkmath.NewInt(4262570)},
		{Address: "stakewolle", Weight: 2123, Delegation: sdkmath.NewInt(112174469020)},
		{Address: "in3s", Weight: 0, Delegation: sdkmath.NewInt(26000000)},
		{Address: "dokiacapital", Weight: 0, Delegation: sdkmath.NewInt(22391760)},
		{Address: "allnodes", Weight: 0, Delegation: sdkmath.NewInt(54249917896)},
		{Address: "stargaze", Weight: 0, Delegation: sdkmath.NewInt(28999998)},
		{Address: "stakecito", Weight: 0, Delegation: sdkmath.NewInt(17547274790)},
		{Address: "binancenode", Weight: 0, Delegation: sdkmath.NewInt(52100031)},
		{Address: "westaking", Weight: 0, Delegation: sdkmath.NewInt(1542120270)},
		{Address: "silknodes", Weight: 1873, Delegation: sdkmath.NewInt(155385308949)},
		{Address: "stir", Weight: 0, Delegation: sdkmath.NewInt(144773556)},
		{Address: "pupmos", Weight: 0, Delegation: sdkmath.NewInt(620632350)},
		{Address: "ezstaking", Weight: 0, Delegation: sdkmath.NewInt(534036530)},
		{Address: "terranodes", Weight: 0, Delegation: sdkmath.NewInt(5000000)},
		{Address: "stakefish", Weight: 0, Delegation: sdkmath.NewInt(2805319160)},
		{Address: "chorusone", Weight: 0, Delegation: sdkmath.NewInt(424507080)},
		{Address: "iqlusion", Weight: 0, Delegation: sdkmath.NewInt(420055380)},
		{Address: "interstellarlounge", Weight: 0, Delegation: sdkmath.NewInt(23068000)},
		{Address: "bitvalidator", Weight: 0, Delegation: sdkmath.NewInt(1000000)},
		{Address: "sikka", Weight: 0, Delegation: sdkmath.NewInt(4384859898)},
		{Address: "stakewithus", Weight: 1871, Delegation: sdkmath.NewInt(95008853212)},
		{Address: "crowdcontrol", Weight: 0, Delegation: sdkmath.NewInt(50809130)},
		{Address: "nodeguardians", Weight: 0, Delegation: sdkmath.NewInt(334470000)},
		{Address: "gatadao", Weight: 0, Delegation: sdkmath.NewInt(1770364440)},
		{Address: "stakesystems", Weight: 0, Delegation: sdkmath.NewInt(18003000000)},
		{Address: "swissstaking", Weight: 0, Delegation: sdkmath.NewInt(4906041540)},
		{Address: "coinbasecloud", Weight: 0, Delegation: sdkmath.NewInt(4999000)},
		{Address: "keplr", Weight: 4373, Delegation: sdkmath.NewInt(199240531276)},
		{Address: "figment", Weight: 0, Delegation: sdkmath.NewInt(4086850000)},
		{Address: "klubstaking", Weight: 0, Delegation: sdkmath.NewInt(587500000)},
		{Address: "coinbasecustody", Weight: 0, Delegation: sdkmath.NewInt(86678850)},
		{Address: "01node", Weight: 0, Delegation: sdkmath.NewInt(678540000)},
		{Address: "posthuman", Weight: 2369, Delegation: sdkmath.NewInt(23149972403)},
		{Address: "rockawayxinfra", Weight: 0, Delegation: sdkmath.NewInt(14000000)},
		{Address: "dacm", Weight: 0, Delegation: sdkmath.NewInt(1000000)},
		{Address: "citizencosmos", Weight: 5625, Delegation: sdkmath.NewInt(31050240)},
		{Address: "0base", Weight: 0, Delegation: sdkmath.NewInt(1513481350)},
		{Address: "ledger", Weight: 0, Delegation: sdkmath.NewInt(630085660)},
		{Address: "highstakes", Weight: 0, Delegation: sdkmath.NewInt(502499996)},
		{Address: "nodestake", Weight: 2623, Delegation: sdkmath.NewInt(121894291587)},
		{Address: "prism", Weight: 0, Delegation: sdkmath.NewInt(15590639920)},
		{Address: "blocksunited", Weight: 0, Delegation: sdkmath.NewInt(31224699)},
		{Address: "prodelegators", Weight: 0, Delegation: sdkmath.NewInt(11875441720)},
		{Address: "blockpower", Weight: 0, Delegation: sdkmath.NewInt(1011568993)},
		{Address: "meria", Weight: 0, Delegation: sdkmath.NewInt(1524599326)},
		{Address: "multichain", Weight: 0, Delegation: sdkmath.NewInt(22748654714)},
		{Address: "dforce", Weight: 0, Delegation: sdkmath.NewInt(3961300000)},
		{Address: "game", Weight: 0, Delegation: sdkmath.NewInt(542445260)},
		{Address: "cosmicvalidator", Weight: 2875, Delegation: sdkmath.NewInt(130989372813)},
		{Address: "cosmosspaces", Weight: 2625, Delegation: sdkmath.NewInt(111388449554)},
		{Address: "chillvalidation", Weight: 2373, Delegation: sdkmath.NewInt(14067371572)},
		{Address: "enigma", Weight: 2375, Delegation: sdkmath.NewInt(108270446482)},
		{Address: "dsrv", Weight: 2125, Delegation: sdkmath.NewInt(12533802351)},
	}

	totalUnbondAmount := sdkmath.NewInt(428227320496)
	totalStake := sdkmath.NewInt(0)
	for _, val := range validators {
		totalStake = totalStake.Add(val.Delegation)
	}
	totalWeight := int64(0)
	for _, val := range validators {
		totalWeight += int64(val.Weight)
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 150 - 40 = 110 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 150 - 40 - 30 = 80 unbond remaining
		// valF has #3 priority - unbond up to remaining
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(80)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondTotalGreaterThanTotalLSM() {
	// Native Stake: 1000
	// LSM Stake:     250
	// Total Stake:  1250
	//
	// Unbond Amount:      350
	// Stake After Unbond: 900
	totalUnbondAmount := sdkmath.NewInt(350)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 100-90 = 10
		// >>> Ratio: 90/100 = 0.9 -> Priority #7 <<<
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 900 = 315, Capacity: 420-315 = 105
		// >>> Ratio: 315/420 = 0.75 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 900 = 270, Capacity: 300-270 = 30
		// >>> Ratio: 270/300 = 0.9 -> Priority #6 <<<
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 200 - 90 = 110
		// >>> Ratio: 90/200 = 0.45 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 900 = 135, Capacity: 160-135 = 25
		// >>> Ratio: 135/160 = 0.85 -> Priority #5 <<<
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 350 - 40 = 310 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 310 - 30 = 280 unbond remaining
		// valF has #3 priority - unbond up to capacity at 110
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(110)},
		// 280 - 110 = 170 unbond remaining
		// valB has #4 priority - unbond up to capacity at 105
		{Validator: "valB", UnbondAmount: sdkmath.NewInt(105)},
		// 170 - 105 = 65 unbond remaining
		// valG has #5 priority - unbond up to capacity at 25
		{Validator: "valG", UnbondAmount: sdkmath.NewInt(25)},
		// 65 - 25 = 40 unbond remaining
		// valD has #6 priority - unbond up to capacity at 30
		{Validator: "valD", UnbondAmount: sdkmath.NewInt(30)},
		// 40 - 30 = 10 unbond remaining
		// valA has #7 priority - unbond up to remaining
		{Validator: "valA", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_NoDelegationAccount() {
	// Call unbond on a host zone without a delegation account - it should error
	invalidHostZone := types.HostZone{
		ChainId:              HostChainId,
		DelegationIcaAddress: "",
	}
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "no delegation account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ZeroUnbondAmount() {
	totalWeight := int64(0)
	totalStake := sdkmath.ZeroInt()
	totalUnbondAmount := sdkmath.ZeroInt()
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, []*types.Validator{})

	// Call unbond - it should NOT error since the unbond amount was 0 - but it should short circuit
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().Nil(err, "unbond should not have thrown an error - it should have simply ignored the host zone")

	// Confirm no ICAs were sent
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence, endSequence, "sequence number should stay the same since no messages were sent")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ZeroValidatorWeights() {
	// Setup the test with all zero-weight validators
	totalWeight := int64(0)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(10)
	validators := []*types.Validator{
		{Address: "valA", Weight: 0, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 0, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(25)},
	}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_InsufficientDelegations() {
	// Setup the test where the total unbond amount is greater than the current delegations
	totalWeight := int64(100)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(200)
	validators := []*types.Validator{
		{Address: "valA", Weight: 25, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 50, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 25, Delegation: sdkmath.NewInt(25)},
	}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ICAFailed() {
	// Validator setup here is arbitrary as long as the totals match
	totalWeight := int64(100)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(10)
	validators := []*types.Validator{{Address: "valA", Weight: 100, Delegation: sdkmath.NewInt(100)}}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Remove the connection ID from the host zone so that the ICA fails
	invalidHostZone := tc.hostZone
	invalidHostZone.ConnectionId = ""

	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "unable to submit unbonding ICA for GAIA")
}

func (s *KeeperTestSuite) TestGetBalanceRatio() {
	testCases := []struct {
		unbondCapacity keeper.ValidatorUnbondCapacity
		expectedRatio  sdk.Dec
		errorExpected  bool
	}{
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(0),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.ZeroDec(),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(25),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("0.25"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(75),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("0.75"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(150),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("1.5"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(100),
				CurrentDelegation:  sdkmath.NewInt(0),
			},
			errorExpected: true,
		},
	}
	for _, tc := range testCases {
		balanceRatio, err := tc.unbondCapacity.GetBalanceRatio()
		if tc.errorExpected {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRatio.String(), balanceRatio.String())
		}
	}
}

func (s *KeeperTestSuite) TestGetTotalUnbondAmountAndRecordsIds() {
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: uint64(1),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(2),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			EpochNumber: uint64(2),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(3),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(4),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			EpochNumber: uint64(3),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Status
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(5),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
				{
					// Different Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(6),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
			},
		},
		{
			EpochNumber: uint64(4),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Host and Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(7),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(8),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	expectedUnbondAmount := int64(1 + 3 + 8)
	expectedRecordIds := []uint64{1, 2, 4}

	actualUnbondAmount, actualRecordIds := s.App.StakeibcKeeper.GetTotalUnbondAmountAndRecordsIds(s.Ctx, HostChainId)
	s.Require().Equal(expectedUnbondAmount, actualUnbondAmount.Int64(), "unbonded amount")
	s.Require().Equal(expectedRecordIds, actualRecordIds, "epoch unbonding record IDs")
}

func (s *KeeperTestSuite) TestGetValidatorUnbondCapacity() {
	// Start with the expected returned list of validator capacities
	expectedUnbondCapacity := []keeper.ValidatorUnbondCapacity{
		{
			ValidatorAddress:   "valA",
			CurrentDelegation:  sdkmath.NewInt(50),
			BalancedDelegation: sdkmath.NewInt(0),
			Capacity:           sdkmath.NewInt(50),
		},
		{
			ValidatorAddress:   "valB",
			CurrentDelegation:  sdkmath.NewInt(200),
			BalancedDelegation: sdkmath.NewInt(5),
			Capacity:           sdkmath.NewInt(195),
		},
		{
			ValidatorAddress:   "valC",
			CurrentDelegation:  sdkmath.NewInt(1089),
			BalancedDelegation: sdkmath.NewInt(1000),
			Capacity:           sdkmath.NewInt(89),
		},
	}

	// Build list of input validators and map of balanced delegations from expected list
	validators := []*types.Validator{}
	balancedDelegations := map[string]sdkmath.Int{}
	for _, validatorCapacity := range expectedUnbondCapacity {
		validators = append(validators, &types.Validator{
			Address:    validatorCapacity.ValidatorAddress,
			Delegation: validatorCapacity.CurrentDelegation,
		})
		balancedDelegations[validatorCapacity.ValidatorAddress] = validatorCapacity.BalancedDelegation
	}

	// Add validators with no capacity - none of these should be in the returned list
	deficits := []int64{0, 10, 50}
	valAddresses := []string{"valD", "valE", "valF"}
	for i, deficit := range deficits {
		address := valAddresses[i]

		// the delegation amount is arbitrary here
		// all that mattesr is that it's less than the balance delegation
		currentDelegation := sdkmath.NewInt(50)
		balancedDelegation := currentDelegation.Add(sdkmath.NewInt(deficit))

		validators = append(validators, &types.Validator{
			Address:    address,
			Delegation: currentDelegation,
		})
		balancedDelegations[address] = balancedDelegation
	}

	// Check capacity matches expectations
	actualUnbondCapacity := s.App.StakeibcKeeper.GetValidatorUnbondCapacity(s.Ctx, validators, balancedDelegations)
	s.Require().Len(actualUnbondCapacity, len(expectedUnbondCapacity), "number of expected unbondings")

	for i, expected := range expectedUnbondCapacity {
		address := expected.ValidatorAddress
		actual := actualUnbondCapacity[i]
		s.Require().Equal(expected.ValidatorAddress, actual.ValidatorAddress, "address for %s", address)
		s.Require().Equal(expected.CurrentDelegation.Int64(), actual.CurrentDelegation.Int64(), "current for %s", address)
		s.Require().Equal(expected.BalancedDelegation.Int64(), actual.BalancedDelegation.Int64(), "balanced for %s", address)
		s.Require().Equal(expected.Capacity.Int64(), actual.Capacity.Int64(), "capacity for %s", address)
	}
}

func (s *KeeperTestSuite) TestSortUnbondingCapacityByPriority() {
	// First we define what the ideal list will look like after sorting
	expectedSortedCapacities := []keeper.ValidatorUnbondCapacity{
		// Zero-weight validator's
		{
			// (1) Ratio: 0, Capacity: 100
			ValidatorAddress:   "valE",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(100), // ratio = 0/100
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (2) Ratio: 0, Capacity: 25
			ValidatorAddress:   "valC",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(25), // ratio = 0/25
			Capacity:           sdkmath.NewInt(25),
		},
		{
			// (3) Ratio: 0, Capacity: 25
			// Same ratio and capacity as above but name is tie breaker
			ValidatorAddress:   "valD",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(25), // ratio = 0/25
			Capacity:           sdkmath.NewInt(25),
		},
		// Non-zero-weight validator's
		{
			// (4) Ratio: 0.1
			ValidatorAddress:   "valB",
			BalancedDelegation: sdkmath.NewInt(1),
			CurrentDelegation:  sdkmath.NewInt(10), // ratio = 1/10
			Capacity:           sdkmath.NewInt(9),
		},
		{
			// (5) Ratio: 0.25
			ValidatorAddress:   "valH",
			BalancedDelegation: sdkmath.NewInt(250),
			CurrentDelegation:  sdkmath.NewInt(1000), // ratio = 250/1000
			Capacity:           sdkmath.NewInt(750),
		},
		{
			// (6) Ratio: 0.5, Capacity: 100
			ValidatorAddress:   "valF",
			BalancedDelegation: sdkmath.NewInt(100),
			CurrentDelegation:  sdkmath.NewInt(200), // ratio = 100/200
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (7) Ratio: 0.5, Capacity: 100
			// Same ratio and capacity as above - name is tie breaker
			ValidatorAddress:   "valI",
			BalancedDelegation: sdkmath.NewInt(100),
			CurrentDelegation:  sdkmath.NewInt(200), // ratio = 100/200
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (8) Ratio: 0.5, Capacity: 50
			// Same ratio as above but capacity is lower
			ValidatorAddress:   "valG",
			BalancedDelegation: sdkmath.NewInt(50),
			CurrentDelegation:  sdkmath.NewInt(100), // ratio = 50/100
			Capacity:           sdkmath.NewInt(50),
		},
		{
			// (9) Ratio: 0.6
			ValidatorAddress:   "valA",
			BalancedDelegation: sdkmath.NewInt(6),
			CurrentDelegation:  sdkmath.NewInt(10), // ratio = 6/10
			Capacity:           sdkmath.NewInt(4),
		},
	}

	// Define the shuffled ordering of the array above by just specifying
	// the validator addresses an a randomized order
	shuffledOrder := []string{
		"valA",
		"valD",
		"valG",
		"valF",
		"valE",
		"valB",
		"valH",
		"valI",
		"valC",
	}

	// Use ordering above in combination with the data structures from the
	// expected list to shuffle the expected list into a list that will be the
	// input to this function
	inputCapacities := []keeper.ValidatorUnbondCapacity{}
	for _, shuffledValAddress := range shuffledOrder {
		for _, capacity := range expectedSortedCapacities {
			if capacity.ValidatorAddress == shuffledValAddress {
				inputCapacities = append(inputCapacities, capacity)
			}
		}
	}

	// Sort the list
	actualSortedCapacities, err := keeper.SortUnbondingCapacityByPriority(inputCapacities)
	s.Require().NoError(err)
	s.Require().Len(actualSortedCapacities, len(expectedSortedCapacities), "number of capacities")

	// To make the error easier to understand, we first compare just the list of validator addresses
	actualValidators := []string{}
	for _, actual := range actualSortedCapacities {
		actualValidators = append(actualValidators, actual.ValidatorAddress)
	}
	expectedValidators := []string{}
	for _, expected := range expectedSortedCapacities {
		expectedValidators = append(expectedValidators, expected.ValidatorAddress)
	}
	s.Require().Equal(expectedValidators, actualValidators, "validator order")

	// Then we'll do a sanity check on each field
	// If the above passes and this fails, that likely means the test was setup improperly
	for i, expected := range expectedSortedCapacities {
		actual := actualSortedCapacities[i]
		address := expected.ValidatorAddress
		s.Require().Equal(expected.ValidatorAddress, actual.ValidatorAddress, "validator %d address", i+1)
		s.Require().Equal(expected.BalancedDelegation, actual.BalancedDelegation, "validator %s balanced", address)
		s.Require().Equal(expected.CurrentDelegation, actual.CurrentDelegation, "validator %s current", address)
		s.Require().Equal(expected.Capacity, actual.Capacity, "validator %s capacity", address)
	}
}

func (s *KeeperTestSuite) TestGetUnbondingICAMessages() {
	delegationAddress := "cosmos_DELEGATION"

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		DelegationIcaAddress: delegationAddress,
	}

	validatorCapacities := []keeper.ValidatorUnbondCapacity{
		{ValidatorAddress: "val1", Capacity: sdkmath.NewInt(100)},
		{ValidatorAddress: "val2", Capacity: sdkmath.NewInt(200)},
		{ValidatorAddress: "val3", Capacity: sdkmath.NewInt(300)},
		{ValidatorAddress: "val4", Capacity: sdkmath.NewInt(400)},
	}

	testCases := []struct {
		name               string
		totalUnbondAmount  sdkmath.Int
		expectedUnbondings []ValidatorUnbonding
		expectedError      string
	}{
		{
			name:              "unbond val1 partially",
			totalUnbondAmount: sdkmath.NewInt(50),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(50)},
			},
		},
		{
			name:              "unbond val1 fully",
			totalUnbondAmount: sdkmath.NewInt(100),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
			},
		},
		{
			name:              "unbond val1 fully and val2 partially",
			totalUnbondAmount: sdkmath.NewInt(200),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(100)},
			},
		},
		{
			name:              "unbond val1 val2 fully",
			totalUnbondAmount: sdkmath.NewInt(300),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
			},
		},
		{
			name:              "unbond val1 val2 fully and val3 partially",
			totalUnbondAmount: sdkmath.NewInt(450),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(150)},
			},
		},
		{
			name:              "unbond val1 val2 and val3 fully",
			totalUnbondAmount: sdkmath.NewInt(600),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(300)},
			},
		},
		{
			name:              "full unbonding",
			totalUnbondAmount: sdkmath.NewInt(1000),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(300)},
				{Validator: "val4", UnbondAmount: sdkmath.NewInt(400)},
			},
		},
		{
			name:              "insufficient delegation",
			totalUnbondAmount: sdkmath.NewInt(1001),
			expectedError:     "unable to unbond full amount",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Get the unbonding ICA messages for the test case
			actualMessages, actualSplits, actualError := s.App.StakeibcKeeper.GetUnbondingICAMessages(
				hostZone,
				tc.totalUnbondAmount,
				validatorCapacities,
			)

			// If this is an error test case, check the error message
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError, "error expected")
				return
			}

			// For the success case, check the error number of unbondings
			s.Require().NoError(actualError, "no error expected when unbonding %v", tc.totalUnbondAmount)
			s.Require().Len(actualMessages, len(tc.expectedUnbondings), "number of undelegate messages")
			s.Require().Len(actualSplits, len(tc.expectedUnbondings), "number of validator splits")

			// Check each unbonding
			for i, expected := range tc.expectedUnbondings {
				valAddress := expected.Validator
				actualMsg := actualMessages[i].(*stakingtypes.MsgUndelegate)
				actualSplit := actualSplits[i]

				// Check the ICA message
				s.Require().Equal(valAddress, actualMsg.ValidatorAddress, "ica message validator")
				s.Require().Equal(delegationAddress, actualMsg.DelegatorAddress, "ica message delegator for %s", valAddress)
				s.Require().Equal(Atom, actualMsg.Amount.Denom, "ica message denom for %s", valAddress)
				s.Require().Equal(expected.UnbondAmount.Int64(), actualMsg.Amount.Amount.Int64(),
					"ica message amount for %s", valAddress)

				// Check the callback
				s.Require().Equal(expected.Validator, actualSplit.Validator, "callback validator for %s", valAddress)
				s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.Amount.Int64(), "callback amount %s", valAddress)
			}
		})
	}
}
