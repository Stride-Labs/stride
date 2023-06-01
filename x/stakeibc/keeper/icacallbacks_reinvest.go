package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls reinvest callback arguments
func (k Keeper) MarshalReinvestCallbackArgs(ctx sdk.Context, reinvestCallback types.ReinvestCallback) ([]byte, error) {
	out, err := proto.Marshal(&reinvestCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalReinvestCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls reinvest callback arguments into a ReinvestCallback struct
func (k Keeper) UnmarshalReinvestCallbackArgs(ctx sdk.Context, reinvestCallback []byte) (*types.ReinvestCallback, error) {
	unmarshalledReinvestCallback := types.ReinvestCallback{}
	if err := proto.Unmarshal(reinvestCallback, &unmarshalledReinvestCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalReinvestCallbackArgs %s", err.Error()))
		return nil, err
	}
	return &unmarshalledReinvestCallback, nil
}

// ICA Callback after reinvestment
//   If successful:
//     * Creates a new DepositRecord with the reinvestment amount
//     * Issues an ICQ to query the rewards balance
//   If timeout/failure:
//     * Does nothing
func ReinvestCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	reinvestCallback, err := k.UnmarshalReinvestCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal reinvest callback args: %s", err.Error()))
	}
	chainId := reinvestCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Reinvest, "Starting reinvest callback"))

	// Grab the associated host zone
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone %s not found", chainId)
	}

	// Check for timeout (ack nil)
	// No action is necessary on a timeout
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Reinvest,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// No action is necessary on a failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Reinvest,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Reinvest,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Get the current stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find epoch")
		return errorsmod.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}

	// Create a new deposit record so that rewards are reinvested
	record := recordstypes.DepositRecord{
		Amount:             reinvestCallback.ReinvestAmount.Amount,
		Denom:              reinvestCallback.ReinvestAmount.Denom,
		HostZoneId:         reinvestCallback.HostZoneId,
		Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
		Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
		DepositEpochNumber: strideEpochTracker.EpochNumber,
	}
	k.RecordsKeeper.AppendDepositRecord(ctx, record)

	// Encode the fee account address for the query request
	// The query request consists of the fee account address and denom
	feeAccount := hostZone.FeeAccount
	if feeAccount == nil || feeAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no fee account found for %s", chainId)
	}
	_, feeAddressBz, err := bech32.DecodeAndConvert(feeAccount.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid fee account address, could not decode (%s)", err.Error())
	}
	queryData := append(banktypes.CreateAccountBalancesPrefix(feeAddressBz), []byte(hostZone.HostDenom)...)

	// The query should timeout before the next epoch
	timeout, err := k.GetICATimeoutNanos(ctx, epochtypes.STRIDE_EPOCH)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get ICATimeout from %s epoch", epochtypes.STRIDE_EPOCH)
	}

	// Submit an ICQ for the rewards balance in the fee account
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Reinvest, "Submitting ICQ for fee account balance"))
	return k.InterchainQueryKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		ICQCallbackID_FeeBalance,
		chainId,
		hostZone.ConnectionId,
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		queryData,
		timeout,
	)
}
