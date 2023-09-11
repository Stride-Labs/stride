package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Transfers tokens from the community pool deposit ICA account to the host zone holding address for that pool
func (k Keeper) IBCTransferCommunityPoolTokens(ctx sdk.Context, token sdk.Coin, communityPoolHostZone types.HostZone, shouldLiquidStake bool) error {

	// The memo may contain autopilot commands to atomically liquid stake tokens when transfer succeeds
	//  both transfer+liquid stake will succeed and stTokens will end in the stride side holding address, 
	//  or neither will and the original base tokens will return to the foreign deposit ICA address
	memoCommands := ""
	autopilotCmd := "{ \"autopilot\": { \"receiver\": \"%s\",  \"stakeibc\": { \"action\": \"LiquidStake\" } } }"
	if shouldLiquidStake {
		memoCommands = fmt.Sprintf(autopilotCmd, communityPoolHostZone.CommunityPoolHoldingAddress)
		stakedDenom := types.StAssetDenomFromHostZoneDenom(token.Denom)
		k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Transferring %v %s and then liquid staking to %s", token.Amount, token.Denom, stakedDenom))
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Transferring %v %s", token.Amount, token.Denom))
	}
	
	// NOTE: this assumes no clock drift between chains, which tendermint guarantees
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	msg := ibctypes.NewMsgTransfer(
		ibctypes.PortID,
		communityPoolHostZone.TransferChannelId,
		token,
		communityPoolHostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on foreign hub
		communityPoolHostZone.CommunityPoolHoldingAddress, // Stride address, unique to each community pool / hostzone
		clienttypes.Height{},
		timeoutTimestamp,
		memoCommands,
	)

	// Submit IBC transfer msg
	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Successfully submitted ibc transfer message %+v", msgTransferResponse))

	return nil
}
