package keeper

import (
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber int64) bool {
	hostZoneUnbondings := make(map[string]*recordstypes.HostZoneUnbonding)
	addEpochUndelegation := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			Amount:     uint64(0),
			Denom:      hostZone.HostDenom,
			HostZoneId: hostZone.ChainId,
			Status:     recordstypes.HostZoneUnbonding_BONDED,
		}
		hostZoneUnbondings[hostZone.ChainId] = &hostZoneUnbonding
		return nil
	}

	k.IterateHostZones(ctx, addEpochUndelegation)
	//TODO(TEST-112) replace this with a check downstream for nil hostZoneUnbonding => replacing with empty struct
	hostZoneUnbondings[""] = &recordstypes.HostZoneUnbonding{}
	latestEpochUnbondingRecordCount := k.RecordsKeeper.GetEpochUnbondingRecordCount(ctx)
	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		Id:                   latestEpochUnbondingRecordCount + 1,
		UnbondingEpochNumber: uint64(epochNumber),
		HostZoneUnbondings:   hostZoneUnbondings,
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
	var allHostZoneUnbondings []*recordstypes.HostZoneUnbonding
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := epochUnbonding.HostZoneUnbondings[hostZone.ChainId]
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d", hostZone.ChainId, epochUnbonding.GetUnbondingEpochNumber()))
			continue
		}
		if hostZoneRecord.Status == recordstypes.HostZoneUnbonding_BONDED { // we only send the ICA call if this hostZone hasn't triggered yet
			totalAmtToUnbond += hostZoneRecord.Amount
			allHostZoneUnbondings = append(allHostZoneUnbondings, hostZoneRecord)
		}
	}
	delegationAccount := hostZone.GetDelegationAccount()
	// TODO add proper validator selection on merge
	// validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
	validator_address := "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne" // local validator
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
	for _, hostZone := range k.GetAllHostZone(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %s", hostZone.GetChainId()))
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
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for epoch unbonding record from epoch %d", epochUnbondingRecord.GetId()))
		if len(epochUnbondingRecord.HostZoneUnbondings) == 0 {
			k.RecordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.GetId())
		}
	}
	return true
}

func (k Keeper) SweepAllUnbondedTokens(ctx sdk.Context) {
	// this function goes through each host zone, and sees if any tokens
	// have been unbonded and are ready to sweep. If so, it processes them

	// get latest epoch unbonding record
	unbondingRecord, found := k.RecordsKeeper.GetLatestEpochUnbondingRecord(ctx)
	if !found {
		k.Logger(ctx).Error("No epoch unbonding record found")
		sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "No latest unbonding record found")
	}

	sweepUnbondedTokens := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {

		// total amount of tokens to be swept
		totalAmtToUnbond := uint64(0)

		// iterate through all host zone unbondings and process them if they're ready to be swept
		// TODO() index into the HostZoneUnbonding map with chainID rather than iterating and checking chainID equality
		for _, unbonding := range unbondingRecord.HostZoneUnbondings {
			if unbonding.HostZoneId == zoneInfo.ChainId {
				k.Logger(ctx).Info(fmt.Sprintf("\tProcessing batch SweepAllUnbondedTokens for host zone %s", zoneInfo.ChainId))
				zone, found := k.GetHostZone(ctx, unbonding.HostZoneId)
				if !found {
					k.Logger(ctx).Error(fmt.Sprintf("\t\tHost zone not found for hostZoneId %s", unbonding.HostZoneId))
					return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Host zone not found")
				}

				// get latest blockTime from light client
				blockTime, found := k.GetLightClientTimeSafely(ctx, zone.ConnectionId)
				if !found {
					k.Logger(ctx).Error(fmt.Sprintf("\t\tCould not find blockTime for host zone %s", zone.ChainId))
					sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "\t\tCould not find blockTime for host zone")
				}

				// if the unbonding period has elapsed, then we can send the ICA call to sweep this hostZone's unbondings to the rewards account (in a batch)
				if unbonding.UnbondingTime < blockTime {
					// we have a match, so we can process this unbonding
					k.Logger(ctx).Info(fmt.Sprintf("\t\tAdding %d to amt to batch transfer from delegation acct to rewards acct for host zone %s", unbonding.Amount, zone.ChainId))
					totalAmtToUnbond += unbonding.Amount
				}
			}
		}

		// if we have any amount to sweep, then we can send the ICA call to sweep them
		if totalAmtToUnbond > 0 {
			zoneInfo, found := k.GetHostZone(ctx, "GAIA")
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("Could not find GAIA host zone"))
			}
			k.Logger(ctx).Info(fmt.Sprintf("\tSending batch SweepAllUnbondedTokens for %d amt to host zone %s", totalAmtToUnbond, zoneInfo.ChainId))
			// Issue ICA bank send from delegation account to rewards account
			if (&zoneInfo).WithdrawalAccount != nil && (&zoneInfo).RedemptionAccount != nil { // only process host zones once withdrawal accounts are registered

				// get the delegation account and rewards account
				delegationAccount := zoneInfo.GetDelegationAccount()
				redemptionAccount := zoneInfo.GetRedemptionAccount()

				//HARDCODED
				sweepCoin := sdk.NewCoin(zoneInfo.HostDenom, sdk.NewInt(int64(99)))
				var msgs []sdk.Msg
				// construct the msg
				msgs = append(msgs, &banktypes.MsgSend{FromAddress: delegationAccount.GetAddress(),
					ToAddress: redemptionAccount.GetAddress(), Amount: sdk.NewCoins(sweepCoin)})

				ctx.Logger().Info(fmt.Sprintf("Bank sending unbonded tokens batch, from delegation to redemption account. Msg: %v", msgs))

				// Send the transaction through SubmitTx
				err := k.SubmitTxs(ctx, zoneInfo.ConnectionId, msgs, *delegationAccount)
				if err != nil {
					// ctx.Logger().Info(fmt.Sprintf("MICE Failed to SubmitTxs for %s, %s, %v", zoneInfo.ConnectionId, zoneInfo.ChainId, msgs))
					// return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %v", zoneInfo.ConnectionId, zoneInfo.ChainId, msgs)
				}
				ctx.Logger().Info(fmt.Sprintf("MICE Successfully completed SubmitTxs for %s, %s, %v", zoneInfo.ConnectionId, zoneInfo.ChainId, msgs))
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("\tNo unbonded tokens this day to sweep for host zone %s", zoneInfo.ChainId))
		}

		return nil
	}
	// Iterate the zones and sweep their unbonded tokens
	k.IterateHostZones(ctx, sweepUnbondedTokens)
}
