package keeper

// NewQuerier returns legacy querier endpoint
// func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
// 	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
// 		var (
// 			res []byte
// 			err error
// 		)

// 		switch path[0] {
// 		default:
// 			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
// 		}

// 		return res, err
// 	}
// }
