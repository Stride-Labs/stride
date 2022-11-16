<!--
order: 1
-->

# Concepts

In Stride, users are required to claim their airdrop by participating in core network activities. An Airdrop recipient is given 20% of the airdrop amount which is not in vesting, and then they have to perform the following activities to get the rest:

* 20% vesting over 3 months by staking
* 60% vesting over 3 months by liquid staking

At initial, module stores all airdrop users with amounts from genesis inside KVStore.

Airdrop users are eligible to claim their vesting or free amount only once in the initial period of 3 months and after the initial period, users can claim tokens monthly not in vesting format.
