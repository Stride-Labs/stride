package keeper

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v5/utils"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls oracle callback arguments
func (k Keeper) MarshalInstantiateOracleCallbackArgs(ctx sdk.Context, instantiateCallback types.InstantiateOracleCallback) ([]byte, error) {
	out, err := proto.Marshal(&instantiateCallback)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrMarshalFailure, "unable to marshal instantiate oracle callback: %s", err.Error())
	}
	return out, nil
}

// Unmarshalls oracle callback arguments into a OracleCallback struct
func (k Keeper) UnmarshalInstantiateOracleCallbackArgs(ctx sdk.Context, oracleCallbackBz []byte) (*types.InstantiateOracleCallback, error) {
	instantiateCallback := types.InstantiateOracleCallback{}
	if err := proto.Unmarshal(oracleCallbackBz, &instantiateCallback); err != nil {
		return nil, errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal instantiate oracle callback: %s", err.Error())
	}
	return &instantiateCallback, nil
}

// Callback after an instantiating an oracle's CW contract
//     If successful: Stores the cosmwasm contract address on the oracle object
//     If timeout/failure: Does nothing
func InstantiateOracleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	instantiateCallback, err := k.UnmarshalInstantiateOracleCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	chainId := instantiateCallback.OracleChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_InstantiateOracle, "Starting instantiate oracle callback"))

	// Check for timeout (ack nil)
	// No action is necessary on a timeout
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_InstantiateOracle,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// No action is necessary on a failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_InstantiateOracle,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_InstantiateOracle,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Get oracle from chainId
	oracle, found := k.GetOracle(ctx, chainId)
	if !found {
		return types.ErrOracleNotFound
	}

	// If the ICA was successful, store the contract address
	if len(ackResponse.MsgResponses) != 1 {
		return errorsmod.Wrapf(types.ErrInvalidICAResponse,
			"tx response from CW contract instantiation should have 1 message (%d found)", len(ackResponse.MsgResponses))
	}
	var instantiateContractResponse types.MsgInstantiateContractResponse
	if err := proto.Unmarshal(ackResponse.MsgResponses[0], &instantiateContractResponse); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal instantiate contract response: %s", err.Error())
	}
	if instantiateContractResponse.Address == "" {
		return errorsmod.Wrapf(types.ErrInvalidICAResponse, "response from CW contract instantiation ICA does not contain a contract address")
	}

	// Update contract address and mark the oracle as active
	oracle.ContractAddress = instantiateContractResponse.Address
	oracle.Active = true
	k.SetOracle(ctx, oracle)

	return nil
}
