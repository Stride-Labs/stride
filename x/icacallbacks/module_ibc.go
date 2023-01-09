package icacallbacks

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks/keeper"
)

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
	app    porttypes.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule) IBCModule {
	return IBCModule{
		keeper: k,
		app:    app,
	}
}

// func(ctx, order, connectionHops []string, portID string, channelID string, chanCap, counterparty, version string) (string, error)
// func(ctx , order , connectionHops []string, portID string, channelID string, channelCap , counterparty , version string) error)
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// Note: The channel capability is claimed by the underlying app.
	// call underlying app's OnChanOpenInit callback with the appVersion
	version, err := im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID,
		channelCap, counterparty, version)
	return version, err
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// call underlying app's OnChanOpenAck callback with the counterparty app version.
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// call underlying app's OnAcknowledgementPacket callback.
	return im.app.OnAcknowledgementPacket(ctx, modulePacket, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, modulePacket, relayer)
}

func (im IBCModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (version string, err error) {
	return proposedVersion, nil
}

// UnpackAcknowledgementResponse returns the msgs from an ICA transaction and can be reused across authentication modules
func UnpackAcknowledgementResponse(ctx sdk.Context, logger log.Logger, ack []byte, isICA bool) (*types.AcknowledgementResponse, error) {
	// TODO: Add logging after ack branches are determined

	fmt.Printf("ACK BYTES: %v\n", ack)

	// Unmarshal the raw ack response
	var acknowledgement channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(ack, &acknowledgement); err != nil {
		fmt.Println("CANNOT UNMARSHAL AS ACK")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %s", err.Error())
	}
	fmt.Printf("ACK UNMARSHALLED: %+v\n", acknowledgement.Response)

	// The ack can come back as either AcknowledgementResult or AcknowledgementError
	// If it comes back as AcknowledgementResult, the messages are encoded differently depending on the SDK version
	switch response := acknowledgement.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		fmt.Printf("ACK IN CASE ACK RESULT - Response: %+v, Result: %v\n", response, response.Result)
		if len(response.Result) == 0 {
			fmt.Println("RESULT IS EMPTY")
			return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "acknowledgement result cannot be empty")
		}

		// If this is an ack from a non-ICA transaction (e.g. an IBC transfer), there is no need to decode the data field
		if !isICA {
			return &types.AcknowledgementResponse{Status: types.SUCCESS}, nil
		}

		// Otherwise, if this ack is from an ICA, unmarshal the message data from within the ack
		txMsgData := &sdk.TxMsgData{}
		if err := proto.Unmarshal(acknowledgement.GetResult(), txMsgData); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 tx message data: %s", err.Error())
		}
		fmt.Printf("ACK TX MSG DATA: %+v\n", txMsgData)

		// Unpack all the message responses based on the sdk version (determined from the length of txMsgData.Data)
		switch len(txMsgData.Data) {
		case 0:
			// for SDK 0.46 and above
			msgResponses := make([][]byte, len(txMsgData.MsgResponses))
			for i, msgResponse := range txMsgData.MsgResponses {
				msgResponses[i] = msgResponse.GetValue()
			}
			return &types.AcknowledgementResponse{Status: types.SUCCESS, MsgResponses: msgResponses}, nil
		default:
			// for SDK 0.45 and below
			var msgResponses = make([][]byte, len(txMsgData.Data))
			for i, msgData := range txMsgData.Data {
				msgResponses[i] = msgData.Data
			}
			return &types.AcknowledgementResponse{Status: types.SUCCESS, MsgResponses: msgResponses}, nil
		}
	case *channeltypes.Acknowledgement_Error:
		fmt.Printf("ACK IN CASE ACK ERROR: %+v\n", response)
		logger.Error(fmt.Sprintf("acknowledgement error: %s", response.Error))
		return &types.AcknowledgementResponse{Status: types.FAILURE, Error: response.Error}, nil
	default:
		return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "unsupported acknowledgement response field type %T", response)
	}
}

// ###################################################################################
// 	Required functions to satisfy interface but not implemented for ICA auth modules
// ###################################################################################

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// icacontroller calls OnChanCloseConfirm but doesn't call the underlying app's OnChanCloseConfirm callback.
	return nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	panic("UNIMPLEMENTED")
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	panic("UNIMPLEMENTED")
}
