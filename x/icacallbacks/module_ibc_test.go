package icacallbacks_test

import (
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

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
		expectedErr         string
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
			expectedErr:    exampleAckError.Error(),
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
			expectedErr:    exampleAckError.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// marshal ack and call Unpack function
			ackBz, err := ibctransfertypes.ModuleCdc.MarshalJSON(&tc.ack)
			require.NoError(t, err, "no error expected when marshalling ack")

			ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(sdk.Context{}, log.NewNopLogger(), ackBz, tc.isICA)
			require.NoError(t, err, "no error expected when unpacking ack")

			// Confirm the response and error status
			require.Equal(t, tc.expectedStatus, ackResponse.Status, "Acknowledgement response status")
			require.Equal(t, tc.expectedErr, ackResponse.Error, "AcknowledgementError")

			// Confirm expected messages
			require.Len(t, ackResponse.MsgResponses, tc.expectedNumMessages)
		})
	}
}
