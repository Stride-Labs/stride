<!--
order: 1
-->

# Concepts

In Stride, users are required to claim their airdrop by participating in core network activities. An Airdrop recipient needs to perform the following activities to get the full airdrop amounts:

* 50% is claimed by staking
* 50% is claimed by liquid staking

At initial, module stores all airdrop users with amounts from genesis inside KVStore.

Furthermore, to incentivize users to claim in a timely manner, the amount of claimable airdrop reduces over time. Users can claim the full airdrop amount for two months (`DurationUntilDecay`).
After two months, the claimable amount linearly decays until 6 months after launch. (At which point none of it is claimable) This is controlled by the parameter `DurationOfDecay` in the code, which is set to 4 months. (6 months - 2 months).

After 6 months from launch, all unclaimed airdrop tokens are sent to the community pool.
