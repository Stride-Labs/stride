package types

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ibctransfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v32/x/interchainquery/types"
)

// IcqKeeper defines the expected interface needed to send ICQ requests.
type IcqKeeper interface {
	SubmitICQRequest(ctx sdk.Context, icqtypes types.Query, forceUnique bool) error
}

// IbcTransferKeeper defines the expected interface needed to convert an ibc token hash to its denom on the source chain.
type IbcTransferKeeper interface {
	GetDenom(ctx sdk.Context, denomHash tmbytes.HexBytes) (ibctransfertypes.Denom, bool)
}
