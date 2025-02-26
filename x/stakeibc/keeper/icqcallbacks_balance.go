package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	icqkeeper "github.com/Stride-Labs/stride/v25/x/interchainquery/keeper"

	icqtypes "github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

func BalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	fmt.Printf("[DEBUG] %s | Starting balance callback, QueryId: %vs\n", query.ChainId, query.Id)

	balanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	fmt.Printf("[DEBUG] %s | Balance: %v\n", query.ChainId, balanceAmount)

	return nil
}
