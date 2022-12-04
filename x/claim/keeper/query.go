package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/claim/types"

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
			err = fmt.Errorf("Unknown %s query endpoint: %s", types.ModuleName, path[0])
		}

		return res, err
	}
}
