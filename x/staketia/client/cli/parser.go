package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v26/x/staketia/types"
)

//////////////////////////////////////////////
// LOGIC FOR PARSING OVERWRITE RECORD JSONS //
//////////////////////////////////////////////

// Parse the overwrite delegationrecord json file into a proto message
func parseOverwriteDelegationRecordFile(cdc codec.JSONCodec,
	parseOverWriteRecordFile string,
	delegationRecord proto.Message,
) (err error) {
	// Defer with a recover to set the error,
	// if an expected sdk.Int field is not included in the JSON
	// will panic and the CLI can SEGFAULT
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.ErrInvalidRequest
		}
	}()

	contents, err := os.ReadFile(parseOverWriteRecordFile)
	if err != nil {
		return err
	}

	if err = cdc.UnmarshalJSON(contents, delegationRecord); err != nil {
		return err
	}

	return err
}

// Parse the overwrite unbondingrecord json file into a proto message
func parseOverwriteUnbondingRecordFile(cdc codec.JSONCodec,
	parseOverWriteRecordFile string,
	unbondingRecord proto.Message,
) (err error) {
	// Defer with a recover to set the error,
	// if an expected sdk.Int field is not included in the JSON
	// will panic and the CLI can SEGFAULT
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.ErrInvalidRequest
		}
	}()

	contents, err := os.ReadFile(parseOverWriteRecordFile)
	if err != nil {
		return err
	}

	if err = cdc.UnmarshalJSON(contents, unbondingRecord); err != nil {
		return err
	}

	return err
}

// Parse the overwrite redemptionrecord json file into a proto message
func parseOverwriteRedemptionRecordFile(cdc codec.JSONCodec,
	parseOverWriteRecordFile string,
	redemptionRecord proto.Message,
) (err error) {
	// Defer with a recover to set the error,
	// if an expected sdk.Int field is not included in the JSON
	// will panic and the CLI can SEGFAULT
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.ErrInvalidRequest
		}
	}()

	contents, err := os.ReadFile(parseOverWriteRecordFile)
	if err != nil {
		return err
	}

	if err = cdc.UnmarshalJSON(contents, redemptionRecord); err != nil {
		return err
	}

	return err
}

//////////////////////////////////////////////
// LOGIC FOR BROADCASTING OVERWRITE RECORDS //
//////////////////////////////////////////////

// helper to parse delegation record and broadcast OverwriteDelegationRecord
func parseAndBroadcastOverwriteDelegation(clientCtx client.Context, cmd *cobra.Command, recordContents string) error {
	var delegationRecord types.DelegationRecord
	// parse the input json
	if err := parseOverwriteDelegationRecordFile(clientCtx.Codec, recordContents, &delegationRecord); err != nil {
		return err
	}
	msg := types.NewMsgOverwriteDelegationRecord(
		clientCtx.GetFromAddress().String(),
		delegationRecord,
	)
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// helper to parse unbonding record and broadcast OverwriteUnbondingRecord
func parseAndBroadcastOverwriteUnbonding(clientCtx client.Context, cmd *cobra.Command, recordContents string) error {
	var unbondingRecord types.UnbondingRecord
	// parse the input json
	if err := parseOverwriteUnbondingRecordFile(clientCtx.Codec, recordContents, &unbondingRecord); err != nil {
		return err
	}
	msg := types.NewMsgOverwriteUnbondingRecord(
		clientCtx.GetFromAddress().String(),
		unbondingRecord,
	)
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// helper to parse redemption record and broadcast OverwriteRedemptionRecord
func parseAndBroadcastOverwriteRedemption(clientCtx client.Context, cmd *cobra.Command, recordContents string) error {
	var redemptionRecord types.RedemptionRecord
	// parse the input json
	if err := parseOverwriteRedemptionRecordFile(clientCtx.Codec, recordContents, &redemptionRecord); err != nil {
		return err
	}
	msg := types.NewMsgOverwriteRedemptionRecord(
		clientCtx.GetFromAddress().String(),
		redemptionRecord,
	)
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}
