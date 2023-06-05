package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackskeeper "github.com/Stride-Labs/stride/v9/x/icacallbacks/keeper"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/Stride-Labs/stride/v9/x/records/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc                codec.BinaryCodec
		storeKey           storetypes.StoreKey
		memKey             storetypes.StoreKey
		paramstore         paramtypes.Subspace
		AccountKeeper      types.AccountKeeper
		TransferKeeper     ibctransferkeeper.Keeper
		IBCKeeper          ibckeeper.Keeper
		ICACallbacksKeeper icacallbackskeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	AccountKeeper types.AccountKeeper,
	TransferKeeper ibctransferkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	ICACallbacksKeeper icacallbackskeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:                Cdc,
		storeKey:           storeKey,
		memKey:             memKey,
		paramstore:         ps,
		AccountKeeper:      AccountKeeper,
		TransferKeeper:     TransferKeeper,
		IBCKeeper:          ibcKeeper,
		ICACallbacksKeeper: ICACallbacksKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Transfer(ctx sdk.Context, msg *ibctypes.MsgTransfer, depositRecord types.DepositRecord) error {
	goCtx := sdk.WrapSDKContext(ctx)
	msgTransferResponse, err := k.TransferKeeper.Transfer(goCtx, msg)
	if err != nil {
		return err
	}
	sequence := msgTransferResponse.Sequence
	// add callback data
	transferCallback := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Marshalling TransferCallback args: %+v", transferCallback))
	marshalledCallbackArgs, err := k.MarshalTransferCallbackArgs(ctx, transferCallback)
	if err != nil {
		return err
	}
	// Store the callback data
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, sequence),
		PortId:       msg.SourcePort,
		ChannelId:    msg.SourceChannel,
		Sequence:     sequence,
		CallbackId:   TRANSFER,
		CallbackArgs: marshalledCallbackArgs,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Storing callback data: %+v", callback))
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	// update the record state to TRANSFER_IN_PROGRESS
	depositRecord.Status = types.DepositRecord_TRANSFER_IN_PROGRESS
	k.SetDepositRecord(ctx, depositRecord)

	return nil
}
