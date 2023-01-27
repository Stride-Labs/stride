package keeper

// import (
// 	"fmt"

// 	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
// 	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
// 	"github.com/golang/protobuf/proto" //nolint:staticcheck
// )

// // Marshalls reinvest callback arguments
// func (k Keeper) MarshalReinvestCallbackArgs(ctx sdk.Context, reinvestCallback types.ReinvestCallback) ([]byte, error) {
// 	out, err := proto.Marshal(&reinvestCallback)
// 	if err != nil {
// 		k.Logger(ctx).Error(fmt.Sprintf("MarshalReinvestCallbackArgs %v", err.Error()))
// 		return nil, err
// 	}
// 	return out, nil
// }

// // Unmarshalls castvoteonhost callback arguments into a CastVoteOnHostCallback struct
// func (k Keeper) UnmarshalCastVoteOnHostCallbackArgs(ctx sdk.Context, castVoteOnHostCallback []byte) (*types.CastVoteOnHostCallback, error) {
// 	unmarshalledCastVoteOnHostCallback := types.CastVoteOnHostCallback{}
// 	if err := proto.Unmarshal(castVoteOnHostCallback, &unmarshalledCastVoteOnHostCallback); err != nil {
// 		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalReinvestCallbackArgs %s", err.Error()))
// 		return nil, err
// 	}
// 	return &unmarshalledCastVoteOnHostCallback, nil
// }

// func CastVoteOnHostCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
// 	// Fetch callback args
// 	castVoteCallback, err := k.UnmarshalCastVoteOnHostCallbackArgs(ctx, args)
// 	// Check for timeout (ack nil)

// 	// Check for a failed transaction (ack error)

// 	// on success delete proposal
// 	k.DeleteProposal(ctx, castVoteCallback.hostZone, castVoteCallback.proposal)
// }
