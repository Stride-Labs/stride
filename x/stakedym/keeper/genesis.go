package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v20/x/stakedym/types"
)

// Initializes the genesis state in the store
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Validate that all required fields are specified
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// Create fee module account (calling GetModuleAccount will set it for the first time)
	k.accountKeeper.GetModuleAccount(ctx, types.FeeAddress)

	// Set the main host zone config
	k.SetHostZone(ctx, genState.HostZone)

	// Set all the records to their respective stores
	for _, delegationRecord := range genState.DelegationRecords {
		k.SetDelegationRecord(ctx, delegationRecord)
	}
	for _, unbondingRecord := range genState.UnbondingRecords {
		k.SetUnbondingRecord(ctx, unbondingRecord)
	}
	for _, redemptionRecord := range genState.RedemptionRecords {
		k.SetRedemptionRecord(ctx, redemptionRecord)
	}
	for _, slashRecord := range genState.SlashRecords {
		k.SetSlashRecord(ctx, slashRecord)
	}
	for _, transfer := range genState.TransferInProgressRecordIds {
		k.SetTransferInProgressRecordId(ctx, transfer.ChannelId, transfer.Sequence, transfer.RecordId)
	}
}

// Exports the current state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		panic(err)
	}

	genesis.HostZone = hostZone
	genesis.DelegationRecords = append(k.GetAllActiveDelegationRecords(ctx), k.GetAllArchivedDelegationRecords(ctx)...)
	genesis.UnbondingRecords = append(k.GetAllActiveUnbondingRecords(ctx), k.GetAllArchivedUnbondingRecords(ctx)...)
	genesis.RedemptionRecords = k.GetAllRedemptionRecords(ctx)
	genesis.SlashRecords = k.GetAllSlashRecords(ctx)
	genesis.TransferInProgressRecordIds = k.GetAllTransferInProgressId(ctx)

	return genesis
}
