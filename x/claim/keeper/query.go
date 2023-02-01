package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/claim/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns legacy querier endpoint
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		default:
			err = fmt.Errorf("unknown %s query endpoint: %s: unknown request", types.ModuleName, path[0])
		}

		return res, err
	}
}
