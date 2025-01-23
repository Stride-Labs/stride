package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/interchainquery/types"
)

// IcqKeeper defines the expected interface needed to send ICQ requests.
type IcqKeeper interface {
	SubmitICQRequest(ctx sdk.Context, icqtypes types.Query, forceUnique bool) error
}
