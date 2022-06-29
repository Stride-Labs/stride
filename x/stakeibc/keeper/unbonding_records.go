package keeper

import (
	"fmt"
	"strconv"

	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	validators := hostZone.GetValidators()
	// we distribute the unbonding based on our target weights
	newUnbondingToValidator := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalAmtToUnbond)
	valAddrToUnbondAmt := make(map[string]int64)
	overflowAmt := uint64(0)
	for _, validator := range validators {
		valAddr := validator.GetAddress()
		valUnbondAmt := newUnbondingToValidator[valAddr]
		currentAmtStaked := validator.GetDelegationAmt()
		if valUnbondAmt > currentAmtStaked { // if we don't have enough assets to unbond
			overflowAmt += valUnbondAmt - currentAmtStaked
			valUnbondAmt = currentAmtStaked
		}
		valAddrToUnbondAmt[valAddr] = valUnbondAmt
	}
	if overflowAmt > 0 { // if we need to reallocate any weights
		for _, validator := range validators {
			valAddr := validator.GetAddress()
			valUnbondAmt := valAddrToUnbondAmt[valAddr]
			currentAmtStaked := validator.GetDelegationAmt()
			// store how many more tokens we could unbond, if needed
			amtToPotentiallyUnbond := currentAmtStaked - valUnbondAmt
			if amtToPotentiallyUnbond > 0 { // if we can afford to unbond more
				if amtToPotentiallyUnbond > overflowAmt { // we can fully cover the overflow
					valAddrToUnbondAmt[valAddr] += overflowAmt
					overflowAmt = 0
					break
				} else {
					valAddrToUnbondAmt[valAddr] += amtToPotentiallyUnbond
					overflowAmt -= amtToPotentiallyUnbond
				}
			}
		}
	}
	if overflowAmt > 0 { // what?? we still can't cover the overflow? something is very wrong
		k.Logger(ctx).Error(fmt.Sprintf("Could not unbond %d on Host Zone %s", totalAmtToUnbond, hostZone.ChainId))
		return false
	}
	for valAddr, valUnbondAmt := range valAddrToUnbondAmt {
		stakeAmt := sdk.NewInt64Coin(hostZone.HostDenom, int64(valUnbondAmt))

		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: delegationAccount.GetAddress(),
			ValidatorAddress: valAddr,
			Amount:           stakeAmt,
		})
	}
	// now we have to handle the overflow amount
	err := k.SubmitTxs(ctx, hostZone.GetConnectionId(), msgs, *delegationAccount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting unbonding tx: %s", err))
		return false
	}

	// mark all of these unbondings as done
	for _, unbonding := range *allHostZoneUnbondings {
		unbonding.UnbondingSent = true
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("newAmountUnbonding", strconv.FormatUint(totalAmtToUnbond, 10)),
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
	for i, epochUnbondingRecord := range k.recordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %d", i))
		if len(epochUnbondingRecord.HostZoneUnbondings) == 0 {
			k.recordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.GetId())
		}
	}
	return true
}

func (k Keeper) SweepAllUnbondedTokens(ctx sdk.Context) bool {
	// this function goes through each host zone, and sees if any tokens
	// have been unbonded and are ready to sweep. If so, it processes them
	for _, hostZone := range k.GetAllHostZone(ctx) {
		var queryBalanceCB icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
			k.Logger(ctx).Info(fmt.Sprintf("\tunbonding query callback on %s", hostZone.ChainId))
			queryRes := stakingtypes.QueryDelegatorUnbondingDelegationsResponse{}
			err := k.cdc.Unmarshal(args, &queryRes)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
				return err
			}
			for _, unbondingResponse := range queryRes.UnbondingResponses {
				// delegatorAddr := unbondingResponse.DelegatorAddress
				validatorAddr := unbondingResponse.ValidatorAddress
				unbondingEntries := unbondingResponse.Entries
				for _, unbondingEntry := range unbondingEntries {
					/*
						unbondingEntry has CreationHeight, CompletionTime, InitialBalance, Balance
					*/
					balance := unbondingEntry.Balance
					if !balance.IsZero() {
						// this entry is unbonded!
						k.Logger(ctx).Info(fmt.Sprintf("\t%s Tokens on %s Zone with validator %s are unbonded", balance.String(), hostZone.ChainId, validatorAddr))
					}
				}
			}

			/*
				TODO Handle logic here for how to unbond tokens
			*/

			return nil
		}
		k.Logger(ctx).Info(fmt.Sprintf("Checking if any unbondings occurred on host zone %s", hostZone.ChainId))
		delegationIca := hostZone.GetDelegationAccount()
		k.InterchainQueryKeeper.QueryUnbondingDelegation(ctx, hostZone, queryBalanceCB, delegationIca.Address)
	}
	return true
}
