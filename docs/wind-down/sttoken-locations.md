# Where each stToken lives, and which locations the migration covers

Source: `sttoken_escrow.html` (Stride outgoing ICS-20 escrow balances, 22 Sep 2026 ~14:00 UTC). "Stride" is supply
minus everything escrowed. USD is redemption rate × CoinGecko price of the underlying at the snapshot. A row's chain is
the first hop out of Stride, not where the token ended up after further hops. Chains under 0.05% of a token's supply
are grouped as "Other".

**Status rules.** `in scope`: the pool covers holders there (Stride itself, Osmosis, the token's host chain, plus the
extra chains listed per token). `ignored`: holders there are not covered; the USD column is what that decision leaves
behind. Extra chains: stATOM on Injective, Secret, Penumbra, Kujira, Agoric; stTIA on Agoric, Neutron, Hub; stINJ on
Secret, Hub; stOSMO on Penumbra, Hub, Secret; stDYDX on Hub. Edit the Status column to change a decision.

Per-token flags present in the source data: ['evm', 'note', 'status'].

## Summary

| Token | Total USD | In scope USD | Ignored USD | Ignored locations |
|---|---:|---:|---:|---:|
| stATOM | $4,536,858 | $4,521,546 | $15,312 | 22 |
| stISLM | $416,792 | $416,571 | $221 | 2 |
| stTIA | $331,752 | $325,484 | $6,268 | 12 |
| stINJ | $164,102 | $162,942 | $1,159 | 5 |
| stOSMO | $144,709 | $143,760 | $949 | 18 |
| stBAND | $110,795 | $110,246 | $550 | 1 |
| stDYDX | $54,956 | $54,165 | $791 | 6 |
| stLUNA | $8,015 | $6,690 | $1,325 | 7 |
| stSAGA | $5,292 | $5,216 | $76 | 3 |
| stEVMOS | $5,216 | $5,183 | $33 | 11 |
| stDYM | $5,113 | $4,994 | $120 | 3 |
| stJUNO | $4,474 | $4,217 | $257 | 10 |
| stSTARS | $1,670 | $1,542 | $128 | 11 |
| stSOMM | $678 | $677 | $1 | 1 |
| stCMDX | $207 | $205 | $2 | 2 |
| stUMEE | $159 | $151 | $8 | 4 |
| **All** | **$5,790,789** | **$5,763,589** | **$27,199** | |


## stATOM (stuatom, host cosmoshub-4, status)

Supply 1,296,132.16 · RR 2.000174 · $4,536,858 total · 44.5% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 719,062.16 | 55.48% | $2,516,937 | in scope |
| Osmosis | osmosis-1 | channel-5 | 456,310.40 | 35.21% | $1,597,226 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 62,487.36 | 4.82% | $218,725 | in scope |
| Injective | injective-1 | channel-6 | 31,062.24 | 2.40% | $108,727 | in scope |
| Secret | secret-4 | channel-40 | 8,385.34 | 0.65% | $29,351 | in scope |
| Penumbra | penumbra-1 | channel-307 | 6,236.85 | 0.48% | $21,831 | in scope |
| Kujira | kaiyo-1 | channel-8 | 5,788.38 | 0.45% | $20,261 | in scope |
| Agoric | agoric-3 | channel-148 | 2,424.99 | 0.19% | $8,488 | in scope |
| Comdex | comdex-1 | channel-49 | 1,420.17 | 0.11% | $4,971 | ignored |
| Neutron | neutron-1 | channel-123 | 1,093.88 | 0.08% | $3,829 | ignored |
| Carbon | carbon-1 | channel-47 | 544.71 | 0.04% | $1,907 | ignored |
| Axelar | axelar-dojo-1 | channel-11, channel-69 | 374.77 | 0.03% | $1,312 | ignored |
| Terra | phoenix-1 | channel-52 | 234.48 | 0.02% | $821 | ignored |
| Canto | canto_7700-1 | channel-74 | 213.03 | 0.02% | $746 | ignored |
| Acrechain | acre_9052-1 | channel-57 | 163.85 | 0.01% | $574 | ignored |
| Stargaze | stargaze-1 | channel-19 | 90.36 | 0.01% | $316 | ignored |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 82.23 | 0.01% | $288 | ignored |
| Composable | centauri-1 | channel-134 | 60.27 | 0.00% | $211 | ignored |
| Umee | umee-1 | channel-29 | 49.97 | 0.00% | $175 | ignored |
| Saga | ssc-1 | channel-213 | 40.55 | 0.00% | $142 | ignored |
| Crescent | crescent-1 | channel-51 | 3.85 | 0.00% | $13 | ignored |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 0.92 | 0.00% | $3 | ignored |
| Juno | juno-1 | channel-24 | 0.64 | 0.00% | $2 | ignored |
| Sei | pacific-1 | channel-149 | 0.21 | 0.00% | $1 | ignored |
| Persistence | core-1 | channel-53 | 0.20 | 0.00% | $1 | ignored |
| Oraichain | Oraichain | channel-50 | 0.18 | 0.00% | $1 | ignored |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 0.10 | 0.00% | $0 | ignored |
| Evmos | evmos_9001-2 | channel-16 | 0.02 | 0.00% | $0 | ignored |
| Dymension | dymension_1100-1 | channel-197 | 0.02 | 0.00% | $0 | ignored |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored |

In scope $4,521,546 · ignored $15,312 across 22 location(s)


## stISLM (staISLM, host haqq_11235-1, evm, status)

Supply 101,018,606.41 · RR 1.059801 · $416,792 total · 50.3% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 50,176,880.70 | 49.67% | $207,024 | in scope |
| HAQQ | haqq_11235-1 | channel-240 | 40,843,524.04 | 40.43% | $168,516 | in scope |
| Osmosis | osmosis-1 | channel-5 | 9,944,705.76 | 9.84% | $41,031 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 53,495.34 | 0.05% | $221 | ignored |
| Kujira | kaiyo-1 | channel-8 | 0.58 | 0.00% | $0 | ignored |

In scope $416,571 · ignored $221 across 2 location(s)


## stTIA (stutia, host celestia, status)

Supply 632,902.42 · RR 1.176364 · $331,752 total · 29.1% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 448,626.79 | 70.88% | $235,159 | in scope |
| Osmosis | osmosis-1 | channel-5 | 145,282.43 | 22.95% | $76,154 | in scope |
| Agoric | agoric-3 | channel-148 | 14,267.74 | 2.25% | $7,479 | in scope |
| Neutron | neutron-1 | channel-123 | 8,344.59 | 1.32% | $4,374 | in scope |
| Carbon | carbon-1 | channel-47 | 6,592.06 | 1.04% | $3,455 | ignored |
| Cosmos Hub | cosmoshub-4 | channel-0 | 3,809.09 | 0.60% | $1,997 | in scope |
| Dymension | dymension_1100-1 | channel-197 | 3,779.27 | 0.60% | $1,981 | ignored |
| Secret | secret-4 | channel-40 | 868.58 | 0.14% | $455 | ignored |
| Celestia | celestia | channel-162 | 614.16 | 0.10% | $322 | in scope |
| Penumbra | penumbra-1 | channel-307 | 409.55 | 0.06% | $215 | ignored |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 277.18 | 0.04% | $145 | ignored |
| Axelar | axelar-dojo-1 | channel-69 | 15.79 | 0.00% | $8 | ignored |
| Injective | injective-1 | channel-6 | 8.19 | 0.00% | $4 | ignored |
| Kujira | kaiyo-1 | channel-8 | 5.51 | 0.00% | $3 | ignored |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 1.29 | 0.00% | $1 | ignored |
| Astria | astria | channel-285 | 0.16 | 0.00% | $0 | ignored |
| Umee | umee-1 | channel-29 | 0.02 | 0.00% | $0 | ignored |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored |

In scope $325,484 · ignored $6,268 across 12 location(s)


## stINJ (stinj, host injective-1, evm, status)

Supply 13,685.82 · RR 1.543197 · $164,102 total · 64.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 4,821.81 | 35.23% | $57,817 | in scope |
| Injective | injective-1 | channel-6 | 4,447.27 | 32.50% | $53,326 | in scope |
| Secret | secret-4 | channel-40 | 3,763.10 | 27.50% | $45,122 | in scope |
| Osmosis | osmosis-1 | channel-5 | 283.91 | 2.07% | $3,404 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 273.04 | 2.00% | $3,274 | in scope |
| Carbon | carbon-1 | channel-47 | 46.61 | 0.34% | $559 | ignored |
| Terra | phoenix-1 | channel-52 | 25.76 | 0.19% | $309 | ignored |
| Kujira | kaiyo-1 | channel-8 | 24.23 | 0.18% | $291 | ignored |
| Evmos | evmos_9001-2 | channel-9 | 0.10 | 0.00% | $1 | ignored |
| Axelar | axelar-dojo-1 | channel-69 | 0.00 | 0.00% | $0 | ignored |

In scope $162,942 · ignored $1,159 across 5 location(s)


## stOSMO (stuosmo, host osmosis-1, status)

Supply 2,703,012.20 · RR 1.461651 · $144,709 total · 60.2% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 1,384,028.36 | 51.20% | $74,096 | in scope |
| Stride | stride-1 | – | 1,076,481.66 | 39.83% | $57,631 | in scope |
| Penumbra | penumbra-1 | channel-307 | 140,013.47 | 5.18% | $7,496 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 52,566.58 | 1.94% | $2,814 | in scope |
| Secret | secret-4 | channel-40 | 32,203.35 | 1.19% | $1,724 | in scope |
| Agoric | agoric-3 | channel-148 | 5,920.53 | 0.22% | $317 | ignored |
| Kujira | kaiyo-1 | channel-8 | 5,472.40 | 0.20% | $293 | ignored |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 3,329.36 | 0.12% | $178 | ignored |
| Carbon | carbon-1 | channel-47 | 1,162.03 | 0.04% | $62 | ignored |
| Acrechain | acre_9052-1 | channel-57 | 909.31 | 0.03% | $49 | ignored |
| Umee | umee-1 | channel-29 | 655.54 | 0.02% | $35 | ignored |
| Comdex | comdex-1 | channel-49 | 224.08 | 0.01% | $12 | ignored |
| Canto | canto_7700-1 | channel-74 | 39.42 | 0.00% | $2 | ignored |
| Juno | juno-1 | channel-24 | 2.31 | 0.00% | $0 | ignored |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 1.36 | 0.00% | $0 | ignored |
| Evmos | evmos_9001-2 | channel-9 | 1.04 | 0.00% | $0 | ignored |
| Injective | injective-1 | channel-6 | 0.70 | 0.00% | $0 | ignored |
| Neutron | neutron-1 | channel-123 | 0.47 | 0.00% | $0 | ignored |
| Stargaze | stargaze-1 | channel-19 | 0.11 | 0.00% | $0 | ignored |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 0.10 | 0.00% | $0 | ignored |
| Oraichain | Oraichain | channel-50 | 0.03 | 0.00% | $0 | ignored |
| Terra | phoenix-1 | channel-52 | 0.00 | 0.00% | $0 | ignored |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored |

In scope $143,760 · ignored $949 across 18 location(s)


## stBAND (stuband, host laozi-mainnet, status)

Supply 403,412.63 · RR 1.272885 · $110,795 total · 94.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 380,329.56 | 94.28% | $104,456 | in scope |
| Stride | stride-1 | – | 21,071.05 | 5.22% | $5,787 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 2,002.01 | 0.50% | $550 | ignored |
| Band | laozi-mainnet | channel-258 | 10.01 | 0.00% | $3 | in scope |

In scope $110,246 · ignored $550 across 1 location(s)


## stDYDX (stadydx, host dydx-mainnet-1, evm, status)

Supply 353,378.02 · RR 1.150952 · $54,956 total · 17.5% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 291,469.92 | 82.48% | $45,328 | in scope |
| Osmosis | osmosis-1 | channel-5 | 31,953.29 | 9.04% | $4,969 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 24,376.97 | 6.90% | $3,791 | in scope |
| Carbon | carbon-1 | channel-47 | 4,875.35 | 1.38% | $758 | ignored |
| dYdX | dydx-mainnet-1 | channel-160 | 488.85 | 0.14% | $76 | in scope |
| Neutron | neutron-1 | channel-123 | 200.01 | 0.06% | $31 | ignored |
| Kujira | kaiyo-1 | channel-8 | 13.32 | 0.00% | $2 | ignored |
| Persistence | core-1 | channel-53 | 0.20 | 0.00% | $0 | ignored |
| Umee | umee-1 | channel-29 | 0.10 | 0.00% | $0 | ignored |
| Indigo | indigo-1 | channel-256 | 0.00 | 0.00% | $0 | ignored |

In scope $54,165 · ignored $791 across 6 location(s)


## stLUNA (stuluna, host phoenix-1, status)

Supply 75,786.52 · RR 1.979438 · $8,015 total · 58.4% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 31,511.96 | 41.58% | $3,333 | in scope |
| Terra | phoenix-1 | channel-13, channel-52 | 29,379.24 | 38.77% | $3,107 | in scope |
| Carbon | carbon-1 | channel-47 | 11,436.45 | 15.09% | $1,210 | ignored |
| Osmosis | osmosis-1 | channel-5 | 2,366.49 | 3.12% | $250 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 1,006.96 | 1.33% | $106 | ignored |
| Secret | secret-4 | channel-40 | 76.62 | 0.10% | $8 | ignored |
| Comdex | comdex-1 | channel-49 | 8.27 | 0.01% | $1 | ignored |
| Kujira | kaiyo-1 | channel-8 | 0.42 | 0.00% | $0 | ignored |
| Axelar | axelar-dojo-1 | channel-69 | 0.10 | 0.00% | $0 | ignored |
| Evmos | evmos_9001-2 | channel-9 | 0.01 | 0.00% | $0 | ignored |

In scope $6,690 · ignored $1,325 across 7 location(s)


## stSAGA (stusaga, host ssc-1, status)

Supply 109,843.16 · RR 1.279077 · $5,292 total · 13.6% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 94,869.65 | 86.37% | $4,571 | in scope |
| Osmosis | osmosis-1 | channel-5 | 12,047.75 | 10.97% | $580 | in scope |
| Carbon | carbon-1 | channel-47 | 1,404.83 | 1.28% | $68 | ignored |
| Saga | ssc-1 | channel-213 | 1,352.80 | 1.23% | $65 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 167.50 | 0.15% | $8 | ignored |
| Neutron | neutron-1 | channel-123 | 0.64 | 0.00% | $0 | ignored |

In scope $5,216 · ignored $76 across 3 location(s)


## stEVMOS (staevmos, host evmos_9001-2, status)

Supply 7,955,610.62 · RR 1.589778 · $5,216 total · 76.0% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Evmos | evmos_9001-2 | channel-9 | 5,176,921.61 | 65.07% | $3,394 | in scope |
| Stride | stride-1 | – | 1,905,584.08 | 23.95% | $1,249 | in scope |
| Osmosis | osmosis-1 | channel-5 | 822,152.94 | 10.33% | $539 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 47,423.48 | 0.60% | $31 | ignored |
| Carbon | carbon-1 | channel-47 | 3,074.38 | 0.04% | $2 | ignored |
| Crescent | crescent-1 | channel-51 | 273.64 | 0.00% | $0 | ignored |
| Kujira | kaiyo-1 | channel-8 | 75.11 | 0.00% | $0 | ignored |
| Canto | canto_7700-1 | channel-74 | 56.55 | 0.00% | $0 | ignored |
| Comdex | comdex-1 | channel-49 | 25.73 | 0.00% | $0 | ignored |
| Injective | injective-1 | channel-6 | 9.99 | 0.00% | $0 | ignored |
| Juno | juno-1 | channel-24 | 8.02 | 0.00% | $0 | ignored |
| Acrechain | acre_9052-1 | channel-57 | 4.78 | 0.00% | $0 | ignored |
| Composable | centauri-1 | channel-134 | 0.30 | 0.00% | $0 | ignored |
| Secret | secret-4 | channel-40 | 0.01 | 0.00% | $0 | ignored |

In scope $5,183 · ignored $33 across 11 location(s)


## stDYM (stadym, host dymension_1100-1, note, evm, status)

Supply 257,169.44 · RR 1.096130 · $5,113 total · 29.7% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 180,780.26 | 70.30% | $3,594 | in scope |
| Dymension | dymension_1100-1 | channel-197 | 45,433.60 | 17.67% | $903 | in scope |
| Osmosis | osmosis-1 | channel-5 | 24,940.62 | 9.70% | $496 | in scope |
| Neutron | neutron-1 | channel-123 | 5,300.13 | 2.06% | $105 | ignored |
| Cosmos Hub | cosmoshub-4 | channel-0 | 714.70 | 0.28% | $14 | ignored |
| Carbon | carbon-1 | channel-47 | 0.13 | 0.00% | $0 | ignored |

In scope $4,994 · ignored $120 across 3 location(s)


## stJUNO (stujuno, host juno-1, status)

Supply 246,844.95 · RR 1.933365 · $4,474 total · 42.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 140,869.86 | 57.07% | $2,553 | in scope |
| Osmosis | osmosis-1 | channel-5 | 91,743.04 | 37.17% | $1,663 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 10,780.01 | 4.37% | $195 | ignored |
| Secret | secret-4 | channel-40 | 3,253.61 | 1.32% | $59 | ignored |
| Umee | umee-1 | channel-29 | 77.89 | 0.03% | $1 | ignored |
| Juno | juno-1 | channel-24 | 55.30 | 0.02% | $1 | in scope |
| Comdex | comdex-1 | channel-39, channel-49 | 37.57 | 0.02% | $1 | ignored |
| Evmos | evmos_9001-2 | channel-9 | 11.91 | 0.00% | $0 | ignored |
| Canto | canto_7700-1 | channel-74 | 10.00 | 0.00% | $0 | ignored |
| Carbon | carbon-1 | channel-47 | 4.39 | 0.00% | $0 | ignored |
| Kujira | kaiyo-1 | channel-8 | 0.53 | 0.00% | $0 | ignored |
| Acrechain | acre_9052-1 | channel-57 | 0.49 | 0.00% | $0 | ignored |
| Neutron | neutron-1 | channel-123 | 0.35 | 0.00% | $0 | ignored |

In scope $4,217 · ignored $257 across 10 location(s)


## stSTARS (stustars, host stargaze-1, status)

Supply 14,563,642.58 · RR 1.917700 · $1,670 total · 85.2% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 5,930,860.51 | 40.72% | $680 | in scope |
| Stargaze | stargaze-1 | channel-19 | 5,365,052.27 | 36.84% | $615 | in scope |
| Stride | stride-1 | – | 2,151,137.02 | 14.77% | $247 | in scope |
| Juno | juno-1 | channel-24 | 795,680.41 | 5.46% | $91 | ignored |
| Cosmos Hub | cosmoshub-4 | channel-0 | 219,822.43 | 1.51% | $25 | ignored |
| Carbon | carbon-1 | channel-47 | 89,046.95 | 0.61% | $10 | ignored |
| Chihuahua | chihuahua-1 | channel-99 | 11,364.18 | 0.08% | $1 | ignored |
| Canto | canto_7700-1 | channel-74 | 550.08 | 0.00% | $0 | ignored |
| Acrechain | acre_9052-1 | channel-57 | 84.86 | 0.00% | $0 | ignored |
| Terra | phoenix-1 | channel-52 | 20.54 | 0.00% | $0 | ignored |
| Kujira | kaiyo-1 | channel-8 | 10.71 | 0.00% | $0 | ignored |
| Evmos | evmos_9001-2 | channel-9 | 8.50 | 0.00% | $0 | ignored |
| Umee | umee-1 | channel-29 | 3.12 | 0.00% | $0 | ignored |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 1.00 | 0.00% | $0 | ignored |

In scope $1,542 · ignored $128 across 11 location(s)


## stSOMM (stusomm, host sommelier-3, status)

Supply 1,334,084.22 · RR 1.080333 · $678 total · 48.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 683,170.78 | 51.21% | $347 | in scope |
| Osmosis | osmosis-1 | channel-5 | 648,687.88 | 48.62% | $330 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 2,225.42 | 0.17% | $1 | ignored |
| Sommelier | sommelier-3 | channel-150 | 0.14 | 0.00% | $0 | in scope |

In scope $677 · ignored $1 across 1 location(s)


## stCMDX (stucmdx, host comdex-1, status)

Supply 1,599,360.16 · RR 1.445795 · $207 total · 78.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Comdex | comdex-1 | channel-49 | 1,237,229.86 | 77.36% | $160 | in scope |
| Stride | stride-1 | – | 337,462.75 | 21.10% | $44 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 12,050.69 | 0.75% | $2 | ignored |
| Osmosis | osmosis-1 | channel-5 | 12,002.17 | 0.75% | $2 | in scope |
| Kujira | kaiyo-1 | channel-8 | 614.69 | 0.04% | $0 | ignored |

In scope $205 · ignored $2 across 2 location(s)


## stUMEE (stuumee, host umee-1, status)

Supply 22,431,695.36 · RR 1.479574 · $159 total · 28.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 15,953,084.27 | 71.12% | $113 | in scope |
| Osmosis | osmosis-1 | channel-5 | 5,354,748.58 | 23.87% | $38 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 1,123,289.31 | 5.01% | $8 | ignored |
| Crescent | crescent-1 | channel-51 | 285.32 | 0.00% | $0 | ignored |
| Kujira | kaiyo-1 | channel-8 | 174.75 | 0.00% | $0 | ignored |
| Umee | umee-1 | channel-29 | 62.23 | 0.00% | $0 | in scope |
| Injective | injective-1 | channel-6 | 50.91 | 0.00% | $0 | ignored |

In scope $151 · ignored $8 across 4 location(s)

