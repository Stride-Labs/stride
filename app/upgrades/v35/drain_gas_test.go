package v35_test

import (
	"fmt"
	"sort"
	"time"

	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	epochstypes "github.com/Stride-Labs/stride/v35/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v35/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v35/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// mainnetBlockGasLimit is consensus_params.block.max_gas on stride-1 (checked 2026-10-05)
const mainnetBlockGasLimit = uint64(100_000_000)

// drainGasCeiling is the most a single wind-down drain tx or ack may cost: half the block
// limit, so the measured headroom is at least 2x
const drainGasCeiling = mainnetBlockGasLimit / 2

// drainGasMeasurement is the gas of the drain on one zone, against mainnet-sized state
type drainGasMeasurement struct {
	chainId          string
	fundedValidators int
	batches          uint64
	oneValidatorGas  uint64
	fullDrainGas     uint64
	ackGas           uint64
}

// TestDrainGasFromMainnetExport measures MsgUndelegateFromValidators against the mainnet export:
// every host zone and every epoch unbonding record are real, so the two costs that grow with
// state (the queued-record guard reads every epoch unbonding record, and each ICA batch and
// each ack rewrite the whole host zone) are paid at their mainnet size. The ICA channels are
// real ibctesting channels, so the packet send is metered too.
//
// The state is moved to where it will be when the drain runs: no record queued or retrying,
// no undelegate ICA in flight, no delegation change in progress.
func (s *MainnetExportTestSuite) TestDrainGasFromMainnetExport() {
	// Channels first: creating them commits blocks and replaces the suite context
	delegationChannels := map[string]string{}
	for _, chainId := range drainGasChainIds() {
		owner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_DELEGATION)
		channelId, _ := s.CreateICAChannel(owner)
		delegationChannels[chainId] = channelId
	}

	export := s.loadTrimmedExport()
	hostZones := s.populateStakeibcFromExport(export)
	records := s.populateRecordsFromExport(export)
	s.T().Logf("%d epoch unbonding records in the export", len(records.EpochUnbondingRecordList))

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           uint64(24 * time.Hour),
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.Add(12 * time.Hour).UnixNano()),
	})
	s.settleRecordsForDrain(records)

	measurements := []drainGasMeasurement{}
	for _, chainId := range drainGasChainIds() {
		hostZone := s.prepareZoneForDrain(hostZones[chainId])
		measurements = append(measurements, s.measureDrainGas(hostZone, delegationChannels[chainId]))
	}

	for _, measurement := range measurements {
		s.T().Logf("%-16s funded=%3d batches=%2d one-validator=%9d full-drain=%10d (%4.1f%% of block) ack=%9d",
			measurement.chainId, measurement.fundedValidators, measurement.batches, measurement.oneValidatorGas,
			measurement.fullDrainGas, 100*float64(measurement.fullDrainGas)/float64(mainnetBlockGasLimit), measurement.ackGas)

		s.Require().Less(measurement.fullDrainGas, drainGasCeiling, "%s full drain gas", measurement.chainId)
		s.Require().Less(measurement.ackGas, drainGasCeiling, "%s ack gas", measurement.chainId)
	}
}

// drainGasChainIds is the in-scope zones in a stable order. The first one names the ibctesting
// host chain, so it must be a chain id with a revision number
func drainGasChainIds() []string {
	chainIds := append([]string{}, v35.InScopeChainIds...)
	sort.SliceStable(chainIds, func(i, j int) bool { return chainIds[i] == "cosmoshub-4" && chainIds[j] != "cosmoshub-4" })
	return chainIds
}

// settleRecordsForDrain moves every queued or retrying record past submission, which is the
// state the drain's guard requires. The record count and sizes, which set the guard's gas,
// are unchanged
func (s *MainnetExportTestSuite) settleRecordsForDrain(records recordstypes.GenesisState) {
	for _, epochUnbondingRecord := range records.EpochUnbondingRecordList {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			queued := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE ||
				hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE
			if queued {
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
			}
			hostZoneUnbonding.UndelegationTxsInProgress = 0
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}
}

// prepareZoneForDrain points the mainnet zone at the test connection and clears the flags the
// drain refuses. Validators, delegations and the batch size stay as on mainnet
func (s *MainnetExportTestSuite) prepareZoneForDrain(hostZone stakeibctypes.HostZone) stakeibctypes.HostZone {
	hostZone.ConnectionId = ibctesting.FirstConnectionID
	for _, validator := range hostZone.Validators {
		validator.DelegationChangesInProgress = 0
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, hostZone.ChainId)
	return hostZone
}

// measureDrainGas runs a one-validator drain and a full drain, each on its own branch of state,
// then the ack of the full drain's first batch
func (s *MainnetExportTestSuite) measureDrainGas(hostZone stakeibctypes.HostZone, channelId string) drainGasMeasurement {
	funded := []stakeibctypes.ValidatorUndelegation{}
	for _, validator := range hostZone.Validators {
		if validator.Delegation.IsPositive() {
			funded = append(funded, stakeibctypes.ValidatorUndelegation{Address: validator.Address, Offset: sdkmath.ZeroInt()})
		}
	}
	s.Require().NotEmpty(funded, "%s has funded validators", hostZone.ChainId)

	oneValidatorGas, _ := s.drainGas(hostZone.ChainId, funded[:1])
	fullDrainGas, fullDrainCtx := s.drainGas(hostZone.ChainId, nil)

	batchSize := int(hostZone.MaxMessagesPerIcaTx)
	if batchSize == 0 {
		batchSize = int(stakeibckeeper.DefaultMaxMessagesPerIcaTx)
	}
	firstBatch := funded
	if len(firstBatch) > batchSize {
		firstBatch = funded[:batchSize]
	}

	return drainGasMeasurement{
		chainId:          hostZone.ChainId,
		fundedValidators: len(funded),
		batches:          uint64((len(funded) + batchSize - 1) / batchSize),
		oneValidatorGas:  oneValidatorGas,
		fullDrainGas:     fullDrainGas,
		ackGas:           s.ackGas(fullDrainCtx, hostZone, channelId, firstBatch),
	}
}

// drainGas submits the drain through the msg server on a branch of state with a fresh gas
// meter, and returns the gas with the branch so the ack can run on top of it
func (s *MainnetExportTestSuite) drainGas(chainId string, validators []stakeibctypes.ValidatorUndelegation) (uint64, sdk.Context) {
	branch, _ := s.Ctx.CacheContext()
	branch = branch.WithGasMeter(storetypes.NewInfiniteGasMeter())

	msgServer := stakeibckeeper.NewMsgServerImpl(s.App.StakeibcKeeper)
	_, err := msgServer.UndelegateFromValidators(branch, &stakeibctypes.MsgUndelegateFromValidators{
		Creator:    "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh",
		ChainId:    chainId,
		Validators: validators,
	})
	s.Require().NoError(err, "%s drain of %d listed validator(s)", chainId, len(validators))

	return branch.GasMeter().GasConsumed(), branch
}

// ackGas runs the undelegate callback for one successful batch of the full drain. The split
// amounts are each validator's recorded delegation, as the drain submits them
func (s *MainnetExportTestSuite) ackGas(
	drainCtx sdk.Context,
	hostZone stakeibctypes.HostZone,
	channelId string,
	batch []stakeibctypes.ValidatorUndelegation,
) uint64 {
	// Built from the zone as it was before the drain flagged its validators
	_, splits, err := s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(hostZone, batch)
	s.Require().NoError(err, "%s batch splits", hostZone.ChainId)
	callbackArgs, err := proto.Marshal(&stakeibctypes.UndelegateCallback{HostZoneId: hostZone.ChainId, SplitUndelegations: splits})
	s.Require().NoError(err)

	completionTime := s.Coordinator.CurrentTime.Add(21 * 24 * time.Hour)
	msgResponse, err := proto.Marshal(&stakingtypes.MsgUndelegateResponse{CompletionTime: completionTime})
	s.Require().NoError(err)
	msgResponses := make([][]byte, len(batch))
	for index := range msgResponses {
		msgResponses[index] = msgResponse
	}

	owner := stakeibctypes.FormatHostZoneICAOwner(hostZone.ChainId, stakeibctypes.ICAAccountType_DELEGATION)
	packet := channeltypes.Packet{SourcePort: fmt.Sprintf("icacontroller-%s", owner), SourceChannel: channelId, Sequence: 1}
	ackResponse := icacallbackstypes.AcknowledgementResponse{
		Status:       icacallbackstypes.AckResponseStatus_SUCCESS,
		MsgResponses: msgResponses,
	}

	ackCtx := drainCtx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	s.Require().NoError(s.App.StakeibcKeeper.UndelegateCallback(ackCtx, packet, &ackResponse, callbackArgs), "%s ack", hostZone.ChainId)
	return ackCtx.GasMeter().GasConsumed()
}
