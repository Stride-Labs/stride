package keeper

import (
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber int64) bool {
	var hostZoneUnbondings = map[string]*recordstypes.EpochUnbondingRecordHostZoneUnbonding{}
	addEpochUndelegation := func(index int64, hostZone types.HostZone) (stop bool) {
		hostZoneUnbonding := recordstypes.EpochUnbondingRecordHostZoneUnbonding{
			Amount:        uint64(0),
			Denom:         hostZone.HostDenom,
			HostZoneId:    hostZone.ChainId,
			UnbondingSent: false,
		}
		hostZoneUnbondings[hostZone.ChainId] = &hostZoneUnbonding
		return false
	}
	k.IterateHostZones(ctx, addEpochUndelegation)
	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.recordsKeeper.AppendEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}

func (k Keeper) SendHostZoneUnbondings(ctx sdk.Context, hostZone types.HostZone) bool {
	// this function goes and processes all unbonded records for this hostZone
	// regardless of what epoch they belong to
	totalAmtToUnbond := uint64(0)
	var msgs []sdk.Msg
	var allHostZoneUnbondings *[]recordstypes.EpochUnbondingRecordHostZoneUnbonding
	for _, epochUnbonding := range k.recordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := epochUnbonding.HostZoneUnbondings[hostZone.ChainId]
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d", hostZone.ChainId, epochUnbonding.EpochNumber))
			continue
		}
		if !hostZoneRecord.UnbondingSent { // we only send the ICA call if this hostZone hasn't triggered yet
			totalAmtToUnbond += hostZoneRecord.Amount
			*allHostZoneUnbondings = append(*allHostZoneUnbondings, *hostZoneRecord)
		}
	}
	delegationAccount := hostZone.GetDelegationAccount()
	// TODO add proper validator selection on merge
	validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
	stakeAmt := sdk.NewInt64Coin(hostZone.HostDenom, int64(totalAmtToUnbond))

	msgs = append(msgs, &stakingTypes.MsgUndelegate{
		DelegatorAddress: delegationAccount.GetAddress(),
		ValidatorAddress: validator_address,
		Amount:           stakeAmt,
	})

	err := k.SubmitTxs(ctx, hostZone.GetConnectionId(), msgs, *delegationAccount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting unbonding tx: %s", err))
		return false
	}

	// mark all of these unbondings as done
	for _, unbonding := range *allHostZoneUnbondings {
		unbonding.UnbondingSent = true
	}
	return true
}

func (k Keeper) ProcessAllEpochUnbondings(ctx sdk.Context, dayNumber uint64) bool {
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

func (k Keeper) VerifyAllUnbondings(ctx sdk.Context, dayNumber uint64) bool {
	// this function goes through each host zone, and sees if any
	// tokens have been unbonded and are ready to sweep. If so, it
	// processes them
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
