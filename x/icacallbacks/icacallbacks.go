package icacallbacks

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
)

// Parses ICA tx responses and returns a list of each serialized response
// The format of the raw ack differs depending on which version of ibc-go is used
// For v4 and prior, the message responses are stored under the `Data` attribute of TxMsgData
// For v5 and later, the message responses are stored under the `MsgResponse` attribute of TxMsgdata
func ParseTxMsgData(acknowledgementResult []byte) ([][]byte, error) {
	txMsgData := &sdk.TxMsgData{}
	if err := proto.Unmarshal(acknowledgementResult, txMsgData); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 tx message data: %s", err.Error())
	}

	// Unpack all the message responses based on the sdk version (determined from the length of txMsgData.Data)
	switch len(txMsgData.Data) {
	case 0:
		// for SDK 0.46 and above
		msgResponses := make([][]byte, len(txMsgData.MsgResponses))
		for i, msgResponse := range txMsgData.MsgResponses {
			msgResponses[i] = msgResponse.GetValue()
		}
		return msgResponses, nil
	default:
		// for SDK 0.45 and below
		var msgResponses = make([][]byte, len(txMsgData.Data))
		for i, msgData := range txMsgData.Data {
			msgResponses[i] = msgData.Data
		}
		return msgResponses, nil
	}
}

// UnpackAcknowledgementResponse unmarshals IBC Acknowledgements, determines the status of the
// acknowledgement (success or failure), and, if applicable, assembles the message responses
//
// ICA transactions have associated messages responses. IBC transfer do not.
//
// With ICA transactions, the schema of the response differs depending on the version of ibc-go used,
// however, this function unifies the format into a common response (a slice of byte arrays)
func UnpackAcknowledgementResponse(ctx sdk.Context, logger log.Logger, ack []byte, isICA bool) (*types.AcknowledgementResponse, error) {
	// Unmarshal the raw ack response
	var acknowledgement channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(ack, &acknowledgement); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %s", err.Error())
	}

	// The ack can come back as either AcknowledgementResult or AcknowledgementError
	// If it comes back as AcknowledgementResult, the messages are encoded differently depending on the SDK version
	switch response := acknowledgement.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		if len(response.Result) == 0 {
			return nil, errorsmod.Wrapf(channeltypes.ErrInvalidAcknowledgement, "acknowledgement result cannot be empty")
		}

		// If this is an ack from a non-ICA transaction (e.g. an IBC transfer), there is no need to decode the data field
		if !isICA {
			logger.Info(fmt.Sprintf("IBC transfer acknowledgement success: %+v", response))
			return &types.AcknowledgementResponse{Status: types.AckResponseStatus_SUCCESS}, nil
		}

		// Otherwise, if this ack is from an ICA, unmarshal the message data from within the ack
		msgResponses, err := ParseTxMsgData(acknowledgement.GetResult())
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot parse TxMsgData from ICA acknowledgement packet: %s", err.Error())
		}
		return &types.AcknowledgementResponse{Status: types.AckResponseStatus_SUCCESS, MsgResponses: msgResponses}, nil

	case *channeltypes.Acknowledgement_Error:
		logger.Error(fmt.Sprintf("acknowledgement error: %s", response.Error))
		return &types.AcknowledgementResponse{Status: types.AckResponseStatus_FAILURE, Error: response.Error}, nil
	default:
		return nil, errorsmod.Wrapf(channeltypes.ErrInvalidAcknowledgement, "unsupported acknowledgement response field type %T", response)
	}
}
