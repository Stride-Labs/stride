package keeper

import (
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber int64) bool {
	hostZoneUnbondings := make(map[string]*recordstypes.HostZoneUnbonding)
	addEpochUndelegation := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			Amount:        uint64(0),
			Denom:         hostZone.HostDenom,
			HostZoneId:    hostZone.ChainId,
			Status: types.HostZoneUnbonding_BONDED,
		}
		hostZoneUnbondings[hostZone.ChainId] = &hostZoneUnbonding
		return nil
	}

	k.IterateHostZones(ctx, addEpochUndelegation)
	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		UnbondingEpochNumber:        epochNumber,
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.Logger(ctx).Info(fmt.Sprintf("epoch unbonding MOOSE %s", epochUnbondingRecord.String()))
	k.Logger(ctx).Info(fmt.Sprintf("hostZoneUnbondings MOOSE %v", hostZoneUnbondings))
	k.RecordsKeeper.AppendEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}

func (k Keeper) SendHostZoneUnbondings(ctx sdk.Context, hostZone types.HostZone) bool {
	// this function goes and processes all unbonded records for this hostZone
	// regardless of what epoch they belong to
	totalAmtToUnbond := uint64(0)
	var msgs []sdk.Msg
	var allHostZoneUnbondings *[]recordstypes.HostZoneUnbonding
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := epochUnbonding.HostZoneUnbondings[hostZone.ChainId]
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d", hostZone.ChainId, epochUnbonding.UnbondedEpochNumber))
			continue
		}
		if hostZoneRecord.Status == recordstypes.HostZoneUnbonding_BONDED{ // we only send the ICA call if this hostZone hasn't triggered yet
			totalAmtToUnbond += hostZoneRecord.Amount
			*allHostZoneUnbondings = append(*allHostZoneUnbondings, *hostZoneRecord)
		}
	}
	delegationAccount := hostZone.GetDelegationAccount()
	// TODO add proper validator selection on merge
	validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
	stakeAmt := sdk.NewInt64Coin(hostZone.HostDenom, int64(totalAmtToUnbond))

	msgs = append(msgs, &stakingtypes.MsgUndelegate{
		DelegatorAddress: delegationAccount.GetAddress(),
		ValidatorAddress: validator_address,
		Amount:           stakeAmt,
	})

	err := k.SubmitTxs(ctx, hostZone.GetConnectionId(), msgs, *delegationAccount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting unbonding tx: %s", err))
		return false
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("newAmountUnbonding", stakeAmt.String()),
		),
	})
	return true
}

func (k Keeper) InitiateAllHostZoneUnbondings(ctx sdk.Context, dayNumber uint64) bool {
	// this function goes through each host zone, and if it's the right time to
	// initiate an unbonding, it goes and tries to unbond all outstanding records
	for i, hostZone := range k.GetAllHostZone(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %d", i))
		// we only send the ICA call if this hostZone is supposed to be triggered
		if dayNumber%hostZone.UnbondingFrequency == 0 {
			k.Logger(ctx).Info(fmt.Sprintf("Sending unbondings for host zone %s", hostZone.ChainId))
			k.SendHostZoneUnbondings(ctx, hostZone)
		}
	}
	return true
}

func (k Keeper) CleanupEpochUnbondingRecords(ctx sdk.Context) bool {
	// this function goes through each EpochUnbondingRecord
	// if any of them don't have any hostZones, then it deletes the record
	for i, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %d", i))
		if len(epochUnbondingRecord.HostZoneUnbondings) == 0 {
			k.RecordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.GetId())
		}
	}
	return true
}

func (k Keeper) SweepAllUnbondedTokens(ctx sdk.Context) bool {
	// NOTE: at the beginning of the epoch we mark all PENDING_TRANSFER HostZoneUnbondingRecords as UNBONDED 
	// so that they're retried if the transfer fails
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_PENDING_TRANSFER {
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_UNBONDED
			}
		}
	}
	// this function goes through each host zone, and sees if any tokens
	// have been unbonded and are ready to sweep. If so, it processes them

	// TODO: replace with clock time check

	return true
}
