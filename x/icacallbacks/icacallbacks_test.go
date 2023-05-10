package icacallbacks_test

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
)

func TestParseTxMsgDataCurrent(t *testing.T) {
	expectedMessages := [][]byte{{1}, {2, 2}, {3, 3}}

	msgData := &sdk.TxMsgData{
		MsgResponses: make([]*codectypes.Any, len(expectedMessages)),
	}
	for i, msgBytes := range expectedMessages {
		typeUrl := "type" + strconv.Itoa(i)
		msgData.MsgResponses[i] = &codectypes.Any{
			TypeUrl: typeUrl,
			Value:   msgBytes,
		}
	}

	msgDataBz, err := proto.Marshal(msgData)
	require.NoError(t, err, "marshaling of current messages should not error")

	parsedMsgResponses, err := icacallbacks.ParseTxMsgData(msgDataBz)
	require.NoError(t, err, "parsing tx message data for current messages should not error")

	require.ElementsMatch(t, expectedMessages, parsedMsgResponses, "parsed current messages")
}

func TestParseTxMsgDataLegacy(t *testing.T) {
	expectedMessages := [][]byte{{1}, {2, 2}, {3, 3}}

	msgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, len(expectedMessages)), //nolint:staticcheck
	}
	for i, msgBytes := range expectedMessages {
		typeUrl := "type" + strconv.Itoa(i)
		msgData.Data[i] = &sdk.MsgData{ //nolint:staticcheck
			MsgType: typeUrl,
			Data:    msgBytes,
		}
	}

	msgDataBz, err := proto.Marshal(msgData)
	require.NoError(t, err, "marshaling of legacy messages should not error")

	parsedMsgResponses, err := icacallbacks.ParseTxMsgData(msgDataBz)
	require.NoError(t, err, "parsing tx message data for legacy messages should not error")

	require.ElementsMatch(t, expectedMessages, parsedMsgResponses, "parsed legacy messages")
}

func TestUnwrapAcknowledgement(t *testing.T) {
	msgDelegate := "/cosmos.staking.v1beta1.MsgDelegate"
	msgUndelegate := "/cosmos.staking.v1beta1.MsgUndelegate"
	exampleAckError := errors.New("ABCI code: 1: error handling packet: see events for details")

	testCases := []struct {
		name                string
		isICA               bool
		ack                 channeltypes.Acknowledgement
		expectedStatus      icacallbacktypes.AckResponseStatus
		expectedNumMessages int
		packetError         string
		functionError       string
	}{
		{
			name:           "ibc_transfer_success",
			isICA:          false,
			ack:            channeltypes.NewResultAcknowledgement([]byte{1}),
			expectedStatus: icacallbacktypes.AckResponseStatus_SUCCESS,
		},
		{
			name:           "ibc_transfer_failure",
			isICA:          false,
			ack:            channeltypes.NewErrorAcknowledgement(exampleAckError),
			expectedStatus: icacallbacktypes.AckResponseStatus_FAILURE,
			packetError:    exampleAckError.Error(),
		},
		{
			name:  "ica_delegate_success",
			isICA: true,
			ack: apptesting.ICAPacketAcknowledgement(
				t,
				msgDelegate,
				[]proto.Message{nil, nil},
			),
			expectedStatus:      icacallbacktypes.AckResponseStatus_SUCCESS,
			expectedNumMessages: 2,
		},
		{
			name:  "ica_undelegate_success",
			isICA: true,
			ack: apptesting.ICAPacketAcknowledgement(
				t,
				msgUndelegate,
				[]proto.Message{
					&stakingtypes.MsgUndelegateResponse{CompletionTime: time.Now()},
					&stakingtypes.MsgUndelegateResponse{CompletionTime: time.Now().Add(time.Duration(10))},
				},
			),
			expectedStatus:      icacallbacktypes.AckResponseStatus_SUCCESS,
			expectedNumMessages: 2,
		},
		{
			name:  "ica_delegate_success_legacy",
			isICA: true,
			ack: apptesting.ICAPacketAcknowledgementLegacy(
				t,
				msgDelegate,
				[]proto.Message{nil, nil},
			),
			expectedStatus:      icacallbacktypes.AckResponseStatus_SUCCESS,
			expectedNumMessages: 2,
		},
		{
			name:  "ica_undelegate_success_legacy",
			isICA: true,
			ack: apptesting.ICAPacketAcknowledgementLegacy(
				t,
				msgUndelegate,
				[]proto.Message{
					&stakingtypes.MsgUndelegateResponse{CompletionTime: time.Now()},
					&stakingtypes.MsgUndelegateResponse{CompletionTime: time.Now().Add(time.Duration(10))},
				},
			),
			expectedStatus:      icacallbacktypes.AckResponseStatus_SUCCESS,
			expectedNumMessages: 2,
		},
		{
			name:           "ica_failure",
			isICA:          true,
			ack:            channeltypes.NewErrorAcknowledgement(exampleAckError),
			expectedStatus: icacallbacktypes.AckResponseStatus_FAILURE,
			packetError:    exampleAckError.Error(),
		},
		{
			name:          "ack_unmarshal_failure",
			isICA:         false,
			ack:           channeltypes.Acknowledgement{},
			functionError: "cannot unmarshal ICS-20 transfer packet acknowledgement",
		},
		{
			name:          "ack_empty_result",
			isICA:         false,
			ack:           apptesting.ICAPacketAcknowledgementLegacy(t, "", []proto.Message{}),
			functionError: "acknowledgement result cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// If the ack is not empty, marshal it
			var err error
			var ackBz []byte
			if !bytes.Equal(tc.ack.Acknowledgement(), []byte("{}")) {
				ackBz, err = ibctransfertypes.ModuleCdc.MarshalJSON(&tc.ack)
				require.NoError(t, err, "no error expected when marshalling ack")
			}

			// Call unpack ack response and check error
			ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(sdk.Context{}, log.NewNopLogger(), ackBz, tc.isICA)
			if tc.functionError != "" {
				require.ErrorContains(t, err, tc.functionError, "unpacking acknowledgement reponse should have resulted in a function error")
				return
			}
			require.NoError(t, err, "no error expected when unpacking ack")

			// Confirm the response and error status
			require.Equal(t, tc.expectedStatus, ackResponse.Status, "Acknowledgement response status")
			require.Equal(t, tc.packetError, ackResponse.Error, "AcknowledgementError")

			// Confirm expected messages
			require.Len(t, ackResponse.MsgResponses, tc.expectedNumMessages)
		})
	}
}
