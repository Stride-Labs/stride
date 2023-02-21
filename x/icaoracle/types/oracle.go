package types

import (
	errorsmod "cosmossdk.io/errors"
)

func (o Oracle) ValidateICASetup() error {
	if o.ConnectionId == "" {
		return errorsmod.Wrapf(ErrOracleICANotRegistered, "connectionId is empty")
	}
	if o.ChannelId == "" {
		return errorsmod.Wrapf(ErrOracleICANotRegistered, "channelId is empty")
	}
	if o.PortId == "" {
		return errorsmod.Wrapf(ErrOracleICANotRegistered, "portId is empty")
	}
	if o.IcaAddress == "" {
		return errorsmod.Wrapf(ErrOracleICANotRegistered, "ICAAddress is empty")
	}
	return nil
}

func (o Oracle) ValidateContractInstantiated() error {
	if o.ContractAddress == "" {
		return errorsmod.Wrapf(ErrOracleNotInstantiated, "contract address is empty")
	}
	return nil
}
