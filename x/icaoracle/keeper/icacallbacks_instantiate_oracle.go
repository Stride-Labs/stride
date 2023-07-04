package keeper

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v11/utils"
	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v11/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Callback after an instantiating an oracle's CW contract
//
//	If successful: Stores the cosmwasm contract address on the oracle object
//	If timeout/failure: Does nothing
func InstantiateOracleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	instantiateCallback := types.InstantiateOracleCallback{}
	if err := proto.Unmarshal(args, &instantiateCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal instantiate oracle callback: %s", err.Error())
	}
	chainId := instantiateCallback.OracleChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_InstantiateOracle, "Starting instantiate oracle callback"))

	// Check for timeout/failure
	// No action is necessary on a timeout
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT ||
		ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_InstantiateOracle, ackResponse.Status, packet))
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
