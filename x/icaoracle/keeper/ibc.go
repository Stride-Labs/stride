package keeper

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func (k Keeper) OnChanOpenInit(ctx sdk.Context, portID, channelID string, channelCap *capabilitytypes.Capability) error {
	// TODO: Update IBC-go to v6/v7 and then there's no longer a need to claim the channel capability here
	// Until then, we need to make sure we only claim for oracle ports
	if strings.Contains(portID, types.ICAAccountType_Oracle) {
		k.Logger(ctx).Info(fmt.Sprintf("%s claimed the channel capability for %s %s", types.ModuleName, channelID, portID))
		return k.scopedKeeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID))
	}
	return nil
}

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Get the connectionId from the port and channel
	connectionId, _, err := k.IBCKeeper.ChannelKeeper.GetChannelConnection(ctx, portID, channelID)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get connection from channel %s and port %s", channelID, portID)
	}

	// If this callback is for an oracle channel, store the ICA address and channel on the oracle struct
	// If the callback is not for an oracle ICA, it should do nothing and then pass the ack down to stakeibc
	oracle, found := k.GetOracleFromConnectionId(ctx, connectionId)
	if found {
		// Confirm the portId is for an oracle ICA
		expectedOraclePort, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(oracle.ChainId, types.ICAAccountType_Oracle))
		if err != nil {
			return err
		}
		if portID == expectedOraclePort {
			// Get the associated ICA address from the port and connection
			icaAddress, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
			if !found {
				return errorsmod.Wrapf(err, "unable to get ica address from connection %s", connectionId)
			}
			k.Logger(ctx).Info(fmt.Sprintf("Oracle ICA registered to channel %s and address %s", channelID, icaAddress))

			// Update the ICA address and channel in the oracle
			oracle.IcaAddress = icaAddress
			oracle.ChannelId = channelID

			k.SetOracle(ctx, oracle)
		}
	}
	return nil
}
