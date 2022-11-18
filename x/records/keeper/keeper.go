package keeper

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	icacallbackstypes "github.com/Stride-Labs/stride/v3/x/icacallbacks/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	icacallbackskeeper "github.com/Stride-Labs/stride/v3/x/icacallbacks/keeper"

	"github.com/Stride-Labs/stride/v3/x/records/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc                codec.BinaryCodec
		storeKey           sdk.StoreKey
		memKey             sdk.StoreKey
		paramstore         paramtypes.Subspace
		scopedKeeper       capabilitykeeper.ScopedKeeper
		AccountKeeper      types.AccountKeeper
		TransferKeeper     ibctransferkeeper.Keeper
		IBCKeeper          ibckeeper.Keeper
		ICACallbacksKeeper icacallbackskeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	scopedKeeper capabilitykeeper.ScopedKeeper,
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
		scopedKeeper:       scopedKeeper,
		AccountKeeper:      AccountKeeper,
		TransferKeeper:     TransferKeeper,
		IBCKeeper:          ibcKeeper,
		ICACallbacksKeeper: ICACallbacksKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) Transfer(ctx sdk.Context, msg *ibctypes.MsgTransfer, depositRecord types.DepositRecord) error {
	goCtx := sdk.WrapSDKContext(ctx)

	// because TransferKeeper.Transfer doesn't return a sequence number, we need to fetch it manually
	// the sequence number isn't actually incremented here, that happens in `SendPacket`, which is triggered
	// by calling `Transfer`
	// see: https://github.com/cosmos/ibc-go/blob/48a6ae512b4ea42c29fdf6c6f5363f50645591a2/modules/core/04-channel/keeper/packet.go#L125
	sequence, found := k.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, msg.SourcePort, msg.SourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", msg.SourcePort, msg.SourceChannel,
		)
	}

	// trigger transfer
	_, err := k.TransferKeeper.Transfer(goCtx, msg)
	if err != nil {
		return err
	}

	// add callback data
	transferCallback := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling TransferCallback args: %v", transferCallback))
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
	k.Logger(ctx).Info(fmt.Sprintf("Storing callback data: %v", callback))
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	// update the record state to TRANSFER_IN_PROGRESS
	depositRecord.Status = types.DepositRecord_TRANSFER_IN_PROGRESS
	k.SetDepositRecord(ctx, depositRecord)

	return nil
}
