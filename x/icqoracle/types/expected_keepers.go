package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

// IcqKeeper defines the expected interface needed to send ICQ requests.
type IcqKeeper interface {
	SubmitICQRequest(ctx sdk.Context, icqtypes types.Query, forceUnique bool) error
	ForceQueryTimeout(ctx sdk.Context, queryId string) error
	GetQueryId(ctx sdk.Context, query types.Query, forceUnique bool) string
}

// IbcTransferKeeper defines the expected interface needed to convert an ibc token hash to its denom on the source chain.
type IbcTransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool)
}
