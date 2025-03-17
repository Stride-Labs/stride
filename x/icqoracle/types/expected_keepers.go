package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v26/x/interchainquery/types"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
)

// IcqKeeper defines the expected interface needed to send ICQ requests.
type IcqKeeper interface {
	SubmitICQRequest(ctx sdk.Context, icqtypes types.Query, forceUnique bool) error
}

// IbcTransferKeeper defines the expected interface needed to convert an ibc token hash to its denom on the source chain.
type IbcTransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool)
}
