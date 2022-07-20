## Vesting Module

Options to choose from:

Evmos x/vesting (from Agoric)
- https://docs.evmos.org/modules/vesting/
- https://github.com/evmos/evmos/blob/main/x/vesting/spec/README.md
- https://docs.cosmos.network/main/modules/auth/05_vesting.html
- https://github.com/Agoric/agoric-sdk/issues/4085
- https://github.com/agoric-labs/cosmos-sdk/tree/Agoric/x/auth/vesting/cmd/vestcalc


Terra vesting upgrades to cosmos-sdk/x/vesting
- https://github.com/terra-money/core/blob/beeff192329372e2bb993f897b8c866fd8be917d/app/upgrade.go#L24


## Distribution Custom Modules

Basic distribution module
- withdraw delegator reward could be altered to blacklist addresses from withdrawing rewards, but that would not prevent them from _accruing_ rewards, which we cannot allow https://github.com/cosmos/cosmos-sdk/blob/main/x/distribution/keeper/msg_server.go
- core logic for distribution: https://github.com/cosmos/cosmos-sdk/blob/main/x/distribution/keeper/allocation.go, could probably include logic a la community tax that siphens off rewards that would otherwise have gone to the addresses in question, so that they receive 0 rewards. then would need to increase the rewards paid to all other stakers by this amount
    - better solution to this is to blacklist a single validator to receive no rewards, then renormalize all the others so that they are upscaled to receive more rewards!
