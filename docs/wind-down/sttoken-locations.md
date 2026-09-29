# Where each stToken lives, and which locations the migration covers

Source: `sttoken_escrow.html` (Stride outgoing ICS-20 escrow balances, 22 Sep 2026 ~14:00 UTC). "Stride" is supply
minus everything escrowed. USD is redemption rate × CoinGecko price of the underlying at the snapshot. A row's chain is
the first hop out of Stride, not where the token ended up after further hops. Chains under 0.05% of a token's supply
are grouped as "Other".

**Status values.**
- `in scope`: the pool covers holders there (Stride itself, Osmosis, the token's host chain, plus every live chain
  holding $1k or more of the token: stATOM on Injective, Secret, Penumbra, Agoric, Neutron, Carbon, Axelar; stTIA on
  Agoric, Neutron, Hub, Carbon, Dymension; stINJ on Secret, Hub; stOSMO on Penumbra, Hub, Secret; stDYDX on Hub;
  stLUNA on Carbon).
- `ignored · small`: a live chain we could serve but the balance is under $1k. Dymension counts as a live chain: it is a
  normal location for other stTokens. stSOMM is ignored entirely for size ($678 in total).
- `ignored · unrecoverable`: the chain has stopped producing blocks (checked 2026-09-23: Evmos, Stargaze and Umee are
  marked killed in the chain registry with no RPC answering; Kujira has no RPC answering; Comdex's last block is 123
  days old). Nothing on those chains can move again, and the stTokens whose host zone is one of them (stEVMOS,
  stSTARS, stUMEE, stCMDX) have no native side to migrate.
- `ignored · deprecated`: stDYM only. Dymension is alive, but stakedym is a deprecated zone we will not touch.

Edit the Status column to change a decision. The relayer map for the in-scope paths is `relayer-map.html` in this folder.

<!-- relayer-scope:start -->
## Relayer scope per chain

Generated 2026-09-29 by `scripts/wind-down/build_relayer_scope.py` from `relayer-map.html` (client ages) plus the youngest
packet a relayer actually delivered on each leg (`tx_search` on the Stride and Osmosis RPCs). Edit the constants at the
top of the script to change the rule, then rerun; `--offline` reuses the cached lookups in `relayer_scope_cache.json`.

**Reading the two decision columns.** *Stride leg* says whether anything still has to cross between Stride and the
chain after the upgrade: only host ICAs (the drain and the ICA transfers) and the sweep channel to Osmosis do, and we
already relay those; for every other chain the Stride leg is simply not needed, whatever its state. *Osmosis leg*
says how holders on the chain reach the pools: someone else already relays it (free), we will relay it, or the chain
is not served and why.

**The rule.** Relayers cost per chain and pool routes cost per token, so the minimum applies to a chain's total.
A leg is *free* when a packet crossed it within 7 days and its client is not expired (a fresh client header
alone proves someone updates the client, not that they relay our channel; a recent packet outranks the map's stale
label). A leg that is not free is run by *ops* when the
chain's in-scope value is at least $1,000, otherwise *none*: holders there move before the upgrade or
relay their own hop. The Stride leg only matters after the upgrade for hosts (our ICA relayers); nobody else gets one.
On a served chain every token worth at least $1,000 gets a pool route. "Last in / out" is the age of the
youngest packet received on the leg and the youngest acknowledgement delivered for the opposite direction.

Served: $2,454,250 across 16 chains, of which $29,542 needs a relayer from us (Penumbra). Not served: Injective ($162,057), Kujira ($20,850), Comdex ($5,145), Evmos ($3,395), Stargaze ($931), Canto ($748), Acrechain ($623), Namada ($611), Composable ($211), Umee ($211), Crescent ($13), Namada testnet ($4), Sei ($1), Persistence ($1), Oraichain ($1), Chihuahua ($1), Gravity Bridge ($0), Namada testnet ($0), Astria ($0), Indigo ($0), Sommelier ($0).

Every chain that holds any stToken is listed, largest first, whatever its status below; the USD column is the chain's
total across tokens. Legs are shown where the relayer map has a route for the chain.

| Chain | Total USD | Stride channel(s) | Host | Stride leg: client · last in / out | Stride leg after the upgrade | Osmosis leg: client · last in / out | Osmosis leg for holders | Pool routes |
|---|---:|---|---|---|---|---|---|---|
| Osmosis (`osmosis-1`) | $1,905,914 | channel-5 |  | not mapped | sweep channel, we relay it | not mapped | destination, the pools live here | stATOM, stISLM, stTIA, stINJ, stOSMO, stBAND, stDYDX, stJUNO |
| Cosmos Hub (`cosmoshub-4`) | $231,762 | channel-0 | yes | channel-0: live 0.0d · 0.1d / 0.0d | ICA channel, we relay it | channel-0: live 0.0d · 0.0d / 0.0d | free, someone else relays it | stATOM, stTIA, stINJ, stOSMO, stDYDX |
| HAQQ (`haqq_11235-1`) | $168,516 | channel-240 | yes | channel-240: live 6.1d · 3.0d / 2.9d | ICA channel, we relay it | channel-1575: live 5.6d · 1.4d / 1.9d | free, someone else relays it | stISLM |
| Injective (`injective-1`) | $162,057 | channel-6 | yes | channel-6: live 2.1d · 0.2d / 0.0d | ICA channel, we relay it | channel-122: blocked 0.0d · 0.1d / 0.0d | not served: blocked (spec §12) | – |
| Secret (`secret-4`) | $76,719 | channel-40 |  | channel-40: live 13.7d · 3.2d / 4.7d | not needed after the upgrade | channel-88: live 0.0d · 0.1d / 0.1d | free, someone else relays it | stATOM, stINJ, stOSMO |
| Penumbra (`penumbra-1`) | $29,542 | channel-307 |  | channel-307: expired 6915.9d · never / never | not needed after the upgrade | channel-79703: expired 3194.7d · never / never | we relay it, after client recovery | stATOM, stOSMO |
| Kujira (`kaiyo-1`) | $20,850 | channel-8 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Agoric (`agoric-3`) | $16,284 | channel-148 |  | channel-148: expired 1137.3d · never / never | not needed after the upgrade | channel-320: live 0.4d · 0.1d / 0.5d | free, someone else relays it | stATOM, stTIA |
| Neutron (`neutron-1`) | $8,339 | channel-123 |  | channel-123: stale 36.3d · 7.4d / 7.4d | not needed after the upgrade | channel-874: stale 35.5d · 0.2d / 1.5d | free, someone else relays it | stATOM, stTIA |
| Carbon (`carbon-1`) | $8,031 | channel-47 |  | channel-47: stale 297.0d · 17.2d / never | not needed after the upgrade | channel-188: stale 40.8d · 4.2d / 8.5d | free, someone else relays it | stATOM, stTIA, stLUNA |
| Comdex (`comdex-1`) | $5,145 | channel-49 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Terra (`phoenix-1`) | $4,237 | channel-52 | yes | channel-52: live 5.3d · 0.2d / 0.2d | ICA channel, we relay it | channel-251: live 0.2d · 0.0d / 0.0d | free, someone else relays it | stLUNA |
| Evmos (`evmos_9001-2`) | $3,395 | channel-16 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Dymension (`dymension_1100-1`) | $2,884 | channel-197 |  | channel-197: live 2.5d · 5.3d / 0.5d | not needed after the upgrade | channel-19774: live 0.2d · 0.0d / 0.1d | free, someone else relays it | stTIA |
| Axelar (`axelar-dojo-1`) | $1,320 | channel-11, channel-69 |  | channel-69: expired 219.0d · never / never | not needed after the upgrade | channel-208: live 0.1d · 0.0d / 0.0d | free, someone else relays it | stATOM |
| Stargaze (`stargaze-1`) | $931 | channel-19 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Canto (`canto_7700-1`) | $748 | channel-74 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Acrechain (`acre_9052-1`) | $623 | channel-57 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Namada (`namada.5f5de2dd1b88cba30586420`) | $611 | channel-308 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Celestia (`celestia`) | $322 | channel-162 | yes | channel-162: live 5.4d · 0.6d / 0.3d | ICA channel, we relay it | channel-6994: live 0.0d · 0.0d / 0.0d | free, someone else relays it | – |
| Composable (`centauri-1`) | $211 | channel-134 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Umee (`umee-1`) | $211 | channel-29 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Saga (`ssc-1`) | $207 | channel-213 | yes | channel-213: live 5.2d · 0.2d / 0.2d | ICA channel, we relay it | channel-38946: live 0.1d · 0.1d / 0.1d | free, someone else relays it | – |
| Juno (`juno-1`) | $94 | channel-24 | yes | channel-24: live 5.9d · 0.2d / 0.2d | ICA channel, we relay it | channel-42: live 0.1d · 0.0d / 0.1d | free, someone else relays it | – |
| dYdX (`dydx-mainnet-1`) | $76 | channel-160 | yes | channel-160: live 5.2d · 0.2d / 0.1d | ICA channel, we relay it | channel-6787: live 0.0d · 0.2d / 0.2d | free, someone else relays it | – |
| Crescent (`crescent-1`) | $13 | channel-51 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Namada testnet (`housefire-alpaca.cc0d3e0c033be`) | $4 | channel-306 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Band (`laozi-mainnet`) | $3 | channel-258 | yes | channel-258: live 5.4d · never / never | ICA channel, we relay it | channel-148: live 9.3d · 0.3d / 0.2d | free, someone else relays it | – |
| Sei (`pacific-1`) | $1 | channel-149 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Persistence (`core-1`) | $1 | channel-53 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Oraichain (`Oraichain`) | $1 | channel-50 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Chihuahua (`chihuahua-1`) | $1 | channel-99 |  | not mapped | not needed after the upgrade | not mapped | not served: chain dead | – |
| Gravity Bridge (`gravity-bridge-3`) | $0 | channel-121 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Namada testnet (`campfire-square.ff09671d333707`) | $0 | channel-297 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Astria (`astria`) | $0 | channel-285 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Indigo (`indigo-1`) | $0 | channel-256 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |
| Sommelier (`sommelier-3`) | $0 | channel-150 |  | not mapped | not needed after the upgrade | not mapped | not served: below minimum | – |

Per-token value on each chain, with the status from the tables below:

- Osmosis: stATOM $1,597,226 (in scope), stBAND $104,456 (in scope), stTIA $76,154 (in scope), stOSMO $74,096 (in scope), stISLM $41,031 (in scope), stDYDX $4,969 (in scope), stINJ $3,404 (in scope), stJUNO $1,663 (in scope), stSTARS $680 (ignored · unrecoverable), stSAGA $580 (in scope), stEVMOS $539 (ignored · unrecoverable), stDYM $496 (ignored · deprecated), stSOMM $330 (ignored · small), stLUNA $250 (in scope), stUMEE $38 (ignored · unrecoverable), stCMDX $2 (ignored · unrecoverable)
- Cosmos Hub: stATOM $218,725 (in scope), stDYDX $3,791 (in scope), stINJ $3,274 (in scope), stOSMO $2,814 (in scope), stTIA $1,997 (in scope), stBAND $550 (ignored · small), stISLM $221 (ignored · small), stJUNO $195 (ignored · small), stLUNA $106 (ignored · small), stEVMOS $31 (ignored · unrecoverable), stSTARS $25 (ignored · unrecoverable), stDYM $14 (ignored · deprecated), stSAGA $8 (ignored · small), stUMEE $8 (ignored · unrecoverable), stCMDX $2 (ignored · unrecoverable), stSOMM $1 (ignored · small)
- HAQQ: stISLM $168,516 (in scope)
- Injective: stATOM $108,727 (in scope), stINJ $53,326 (in scope), stTIA $4 (ignored · small), stOSMO $0 (ignored · small), stEVMOS $0 (ignored · unrecoverable), stUMEE $0 (ignored · unrecoverable)
- Secret: stINJ $45,122 (in scope), stATOM $29,351 (in scope), stOSMO $1,724 (in scope), stTIA $455 (ignored · small), stJUNO $59 (ignored · small), stLUNA $8 (ignored · small), stEVMOS $0 (ignored · unrecoverable)
- Penumbra: stATOM $21,831 (in scope), stOSMO $7,496 (in scope), stTIA $215 (ignored · small)
- Kujira: stATOM $20,261 (ignored · unrecoverable), stOSMO $293 (ignored · unrecoverable), stINJ $291 (ignored · unrecoverable), stTIA $3 (ignored · unrecoverable), stDYDX $2 (ignored · unrecoverable), stISLM $0 (ignored · unrecoverable), stLUNA $0 (ignored · unrecoverable), stEVMOS $0 (ignored · unrecoverable), stJUNO $0 (ignored · unrecoverable), stSTARS $0 (ignored · unrecoverable), stCMDX $0 (ignored · unrecoverable), stUMEE $0 (ignored · unrecoverable)
- Agoric: stATOM $8,488 (in scope), stTIA $7,479 (in scope), stOSMO $317 (ignored · small)
- Neutron: stTIA $4,374 (in scope), stATOM $3,829 (in scope), stDYM $105 (ignored · deprecated), stDYDX $31 (ignored · small), stOSMO $0 (ignored · small), stSAGA $0 (ignored · small), stJUNO $0 (ignored · small)
- Carbon: stTIA $3,455 (in scope), stATOM $1,907 (in scope), stLUNA $1,210 (in scope), stDYDX $758 (ignored · small), stINJ $559 (ignored · small), stSAGA $68 (ignored · small), stOSMO $62 (ignored · small), stSTARS $10 (ignored · unrecoverable), stEVMOS $2 (ignored · unrecoverable), stDYM $0 (ignored · deprecated), stJUNO $0 (ignored · small)
- Comdex: stATOM $4,971 (ignored · unrecoverable), stCMDX $160 (ignored · unrecoverable), stOSMO $12 (ignored · unrecoverable), stLUNA $1 (ignored · unrecoverable), stJUNO $1 (ignored · unrecoverable), stEVMOS $0 (ignored · unrecoverable)
- Terra: stLUNA $3,107 (in scope), stATOM $821 (ignored · small), stINJ $309 (ignored · small), stOSMO $0 (ignored · small), stSTARS $0 (ignored · unrecoverable)
- Evmos: stEVMOS $3,394 (ignored · unrecoverable), stINJ $1 (ignored · unrecoverable), stATOM $0 (ignored · unrecoverable), stOSMO $0 (ignored · unrecoverable), stLUNA $0 (ignored · unrecoverable), stJUNO $0 (ignored · unrecoverable), stSTARS $0 (ignored · unrecoverable)
- Dymension: stTIA $1,981 (in scope), stDYM $903 (ignored · deprecated), stATOM $0 (ignored · small)
- Axelar: stATOM $1,312 (in scope), stTIA $8 (ignored · small), stINJ $0 (ignored · small), stLUNA $0 (ignored · small)
- Stargaze: stSTARS $615 (ignored · unrecoverable), stATOM $316 (ignored · unrecoverable), stOSMO $0 (ignored · unrecoverable)
- Canto: stATOM $746 (ignored · small), stOSMO $2 (ignored · small), stEVMOS $0 (ignored · unrecoverable), stJUNO $0 (ignored · small), stSTARS $0 (ignored · unrecoverable)
- Acrechain: stATOM $574 (ignored · small), stOSMO $49 (ignored · small), stEVMOS $0 (ignored · unrecoverable), stJUNO $0 (ignored · small), stSTARS $0 (ignored · unrecoverable)
- Namada: stATOM $288 (ignored · small), stOSMO $178 (ignored · small), stTIA $145 (ignored · small)
- Celestia: stTIA $322 (in scope)
- Composable: stATOM $211 (ignored · small), stEVMOS $0 (ignored · unrecoverable)
- Umee: stATOM $175 (ignored · unrecoverable), stOSMO $35 (ignored · unrecoverable), stJUNO $1 (ignored · unrecoverable), stTIA $0 (ignored · unrecoverable), stDYDX $0 (ignored · unrecoverable), stSTARS $0 (ignored · unrecoverable), stUMEE $0 (ignored · unrecoverable)
- Saga: stATOM $142 (ignored · small), stSAGA $65 (in scope)
- Juno: stSTARS $91 (ignored · unrecoverable), stATOM $2 (ignored · small), stJUNO $1 (in scope), stOSMO $0 (ignored · small), stEVMOS $0 (ignored · unrecoverable)
- dYdX: stDYDX $76 (in scope)
- Crescent: stATOM $13 (ignored · small), stEVMOS $0 (ignored · unrecoverable), stUMEE $0 (ignored · unrecoverable)
- Namada testnet: stATOM $3 (ignored · small), stTIA $1 (ignored · small), stOSMO $0 (ignored · small)
- Band: stBAND $3 (in scope)
- Sei: stATOM $1 (ignored · small)
- Persistence: stATOM $1 (ignored · small), stDYDX $0 (ignored · small)
- Oraichain: stATOM $1 (ignored · small), stOSMO $0 (ignored · small)
- Chihuahua: stSTARS $1 (ignored · unrecoverable)
- Gravity Bridge: stATOM $0 (ignored · small), stOSMO $0 (ignored · small), stSTARS $0 (ignored · unrecoverable)
- Namada testnet: stATOM $0 (ignored · small), stTIA $0 (ignored · small), stOSMO $0 (ignored · small)
- Astria: stTIA $0 (ignored · small)
- Indigo: stDYDX $0 (ignored · small)
- Sommelier: stSOMM $0 (ignored · small)

<!-- relayer-scope:end -->

## Summary

| Token | Total USD | In scope | Ignored · small | Ignored · unrecoverable | Ignored · deprecated |
|---|---:|---:|---:|---:|---:|
| stATOM | $4,536,858 | $4,508,333 | $2,802 | $25,723 | $0 |
| stISLM | $416,792 | $416,571 | $221 | $0 | $0 |
| stTIA | $331,752 | $330,921 | $829 | $3 | $0 |
| stINJ | $164,102 | $162,942 | $868 | $292 | $0 |
| stOSMO | $144,709 | $143,760 | $608 | $340 | $0 |
| stBAND | $110,795 | $110,246 | $550 | $0 | $0 |
| stDYDX | $54,956 | $54,165 | $789 | $2 | $0 |
| stLUNA | $8,015 | $7,900 | $115 | $1 | $0 |
| stSAGA | $5,292 | $5,216 | $76 | $0 | $0 |
| stEVMOS | $5,216 | $0 | $0 | $5,216 | $0 |
| stDYM | $5,113 | $0 | $0 | $0 | $5,113 |
| stJUNO | $4,474 | $4,217 | $255 | $2 | $0 |
| stSTARS | $1,670 | $0 | $0 | $1,670 | $0 |
| stSOMM | $678 | $0 | $678 | $0 | $0 |
| stCMDX | $207 | $0 | $0 | $207 | $0 |
| stUMEE | $159 | $0 | $0 | $159 | $0 |
| **All** | **$5,790,789** | **$5,744,270** | **$7,790** | **$33,615** | **$5,113** |


## stATOM (stuatom, host cosmoshub-4)

Supply 1,296,132.16 · RR 2.000174 · $4,536,858 total · 44.5% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 719,062.16 | 55.48% | $2,516,937 | in scope |
| Osmosis | osmosis-1 | channel-5 | 456,310.40 | 35.21% | $1,597,226 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 62,487.36 | 4.82% | $218,725 | in scope |
| Injective | injective-1 | channel-6 | 31,062.24 | 2.40% | $108,727 | in scope |
| Secret | secret-4 | channel-40 | 8,385.34 | 0.65% | $29,351 | in scope |
| Penumbra | penumbra-1 | channel-307 | 6,236.85 | 0.48% | $21,831 | in scope |
| Kujira | kaiyo-1 | channel-8 | 5,788.38 | 0.45% | $20,261 | ignored · unrecoverable |
| Agoric | agoric-3 | channel-148 | 2,424.99 | 0.19% | $8,488 | in scope |
| Comdex | comdex-1 | channel-49 | 1,420.17 | 0.11% | $4,971 | ignored · unrecoverable |
| Neutron | neutron-1 | channel-123 | 1,093.88 | 0.08% | $3,829 | in scope |
| Carbon | carbon-1 | channel-47 | 544.71 | 0.04% | $1,907 | in scope |
| Axelar | axelar-dojo-1 | channel-11, channel-69 | 374.77 | 0.03% | $1,312 | in scope |
| Terra | phoenix-1 | channel-52 | 234.48 | 0.02% | $821 | ignored · small |
| Canto | canto_7700-1 | channel-74 | 213.03 | 0.02% | $746 | ignored · small |
| Acrechain | acre_9052-1 | channel-57 | 163.85 | 0.01% | $574 | ignored · small |
| Stargaze | stargaze-1 | channel-19 | 90.36 | 0.01% | $316 | ignored · unrecoverable |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 82.23 | 0.01% | $288 | ignored · small |
| Composable | centauri-1 | channel-134 | 60.27 | 0.00% | $211 | ignored · small |
| Umee | umee-1 | channel-29 | 49.97 | 0.00% | $175 | ignored · unrecoverable |
| Saga | ssc-1 | channel-213 | 40.55 | 0.00% | $142 | ignored · small |
| Crescent | crescent-1 | channel-51 | 3.85 | 0.00% | $13 | ignored · small |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 0.92 | 0.00% | $3 | ignored · small |
| Juno | juno-1 | channel-24 | 0.64 | 0.00% | $2 | ignored · small |
| Sei | pacific-1 | channel-149 | 0.21 | 0.00% | $1 | ignored · small |
| Persistence | core-1 | channel-53 | 0.20 | 0.00% | $1 | ignored · small |
| Oraichain | Oraichain | channel-50 | 0.18 | 0.00% | $1 | ignored · small |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 0.10 | 0.00% | $0 | ignored · small |
| Evmos | evmos_9001-2 | channel-16 | 0.02 | 0.00% | $0 | ignored · unrecoverable |
| Dymension | dymension_1100-1 | channel-197 | 0.02 | 0.00% | $0 | ignored · small |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored · small |

In scope $4,508,333 · small $2,802 · unrecoverable $25,723 · deprecated $0


## stISLM (staISLM, host haqq_11235-1)

Supply 101,018,606.41 · RR 1.059801 · $416,792 total · 50.3% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 50,176,880.70 | 49.67% | $207,024 | in scope |
| HAQQ | haqq_11235-1 | channel-240 | 40,843,524.04 | 40.43% | $168,516 | in scope |
| Osmosis | osmosis-1 | channel-5 | 9,944,705.76 | 9.84% | $41,031 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 53,495.34 | 0.05% | $221 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 0.58 | 0.00% | $0 | ignored · unrecoverable |

In scope $416,571 · small $221 · unrecoverable $0 · deprecated $0


## stTIA (stutia, host celestia)

Supply 632,902.42 · RR 1.176364 · $331,752 total · 29.1% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 448,626.79 | 70.88% | $235,159 | in scope |
| Osmosis | osmosis-1 | channel-5 | 145,282.43 | 22.95% | $76,154 | in scope |
| Agoric | agoric-3 | channel-148 | 14,267.74 | 2.25% | $7,479 | in scope |
| Neutron | neutron-1 | channel-123 | 8,344.59 | 1.32% | $4,374 | in scope |
| Carbon | carbon-1 | channel-47 | 6,592.06 | 1.04% | $3,455 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 3,809.09 | 0.60% | $1,997 | in scope |
| Dymension | dymension_1100-1 | channel-197 | 3,779.27 | 0.60% | $1,981 | in scope |
| Secret | secret-4 | channel-40 | 868.58 | 0.14% | $455 | ignored · small |
| Celestia | celestia | channel-162 | 614.16 | 0.10% | $322 | in scope |
| Penumbra | penumbra-1 | channel-307 | 409.55 | 0.06% | $215 | ignored · small |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 277.18 | 0.04% | $145 | ignored · small |
| Axelar | axelar-dojo-1 | channel-69 | 15.79 | 0.00% | $8 | ignored · small |
| Injective | injective-1 | channel-6 | 8.19 | 0.00% | $4 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 5.51 | 0.00% | $3 | ignored · unrecoverable |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 1.29 | 0.00% | $1 | ignored · small |
| Astria | astria | channel-285 | 0.16 | 0.00% | $0 | ignored · small |
| Umee | umee-1 | channel-29 | 0.02 | 0.00% | $0 | ignored · unrecoverable |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored · small |

In scope $330,921 · small $829 · unrecoverable $3 · deprecated $0


## stINJ (stinj, host injective-1)

Supply 13,685.82 · RR 1.543197 · $164,102 total · 64.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 4,821.81 | 35.23% | $57,817 | in scope |
| Injective | injective-1 | channel-6 | 4,447.27 | 32.50% | $53,326 | in scope |
| Secret | secret-4 | channel-40 | 3,763.10 | 27.50% | $45,122 | in scope |
| Osmosis | osmosis-1 | channel-5 | 283.91 | 2.07% | $3,404 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 273.04 | 2.00% | $3,274 | in scope |
| Carbon | carbon-1 | channel-47 | 46.61 | 0.34% | $559 | ignored · small |
| Terra | phoenix-1 | channel-52 | 25.76 | 0.19% | $309 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 24.23 | 0.18% | $291 | ignored · unrecoverable |
| Evmos | evmos_9001-2 | channel-9 | 0.10 | 0.00% | $1 | ignored · unrecoverable |
| Axelar | axelar-dojo-1 | channel-69 | 0.00 | 0.00% | $0 | ignored · small |

In scope $162,942 · small $868 · unrecoverable $292 · deprecated $0


## stOSMO (stuosmo, host osmosis-1)

Supply 2,703,012.20 · RR 1.461651 · $144,709 total · 60.2% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 1,384,028.36 | 51.20% | $74,096 | in scope |
| Stride | stride-1 | – | 1,076,481.66 | 39.83% | $57,631 | in scope |
| Penumbra | penumbra-1 | channel-307 | 140,013.47 | 5.18% | $7,496 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 52,566.58 | 1.94% | $2,814 | in scope |
| Secret | secret-4 | channel-40 | 32,203.35 | 1.19% | $1,724 | in scope |
| Agoric | agoric-3 | channel-148 | 5,920.53 | 0.22% | $317 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 5,472.40 | 0.20% | $293 | ignored · unrecoverable |
| Namada | namada.5f5de2dd1b88cba30586420 | channel-308 | 3,329.36 | 0.12% | $178 | ignored · small |
| Carbon | carbon-1 | channel-47 | 1,162.03 | 0.04% | $62 | ignored · small |
| Acrechain | acre_9052-1 | channel-57 | 909.31 | 0.03% | $49 | ignored · small |
| Umee | umee-1 | channel-29 | 655.54 | 0.02% | $35 | ignored · unrecoverable |
| Comdex | comdex-1 | channel-49 | 224.08 | 0.01% | $12 | ignored · unrecoverable |
| Canto | canto_7700-1 | channel-74 | 39.42 | 0.00% | $2 | ignored · small |
| Juno | juno-1 | channel-24 | 2.31 | 0.00% | $0 | ignored · small |
| Namada testnet | housefire-alpaca.cc0d3e0c033be | channel-306 | 1.36 | 0.00% | $0 | ignored · small |
| Evmos | evmos_9001-2 | channel-9 | 1.04 | 0.00% | $0 | ignored · unrecoverable |
| Injective | injective-1 | channel-6 | 0.70 | 0.00% | $0 | ignored · small |
| Neutron | neutron-1 | channel-123 | 0.47 | 0.00% | $0 | ignored · small |
| Stargaze | stargaze-1 | channel-19 | 0.11 | 0.00% | $0 | ignored · unrecoverable |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 0.10 | 0.00% | $0 | ignored · small |
| Oraichain | Oraichain | channel-50 | 0.03 | 0.00% | $0 | ignored · small |
| Terra | phoenix-1 | channel-52 | 0.00 | 0.00% | $0 | ignored · small |
| Namada testnet | campfire-square.ff09671d333707 | channel-297 | 0.00 | 0.00% | $0 | ignored · small |

In scope $143,760 · small $608 · unrecoverable $340 · deprecated $0


## stBAND (stuband, host laozi-mainnet)

Supply 403,412.63 · RR 1.272885 · $110,795 total · 94.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 380,329.56 | 94.28% | $104,456 | in scope |
| Stride | stride-1 | – | 21,071.05 | 5.22% | $5,787 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 2,002.01 | 0.50% | $550 | ignored · small |
| Band | laozi-mainnet | channel-258 | 10.01 | 0.00% | $3 | in scope |

In scope $110,246 · small $550 · unrecoverable $0 · deprecated $0


## stDYDX (stadydx, host dydx-mainnet-1)

Supply 353,378.02 · RR 1.150952 · $54,956 total · 17.5% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 291,469.92 | 82.48% | $45,328 | in scope |
| Osmosis | osmosis-1 | channel-5 | 31,953.29 | 9.04% | $4,969 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 24,376.97 | 6.90% | $3,791 | in scope |
| Carbon | carbon-1 | channel-47 | 4,875.35 | 1.38% | $758 | ignored · small |
| dYdX | dydx-mainnet-1 | channel-160 | 488.85 | 0.14% | $76 | in scope |
| Neutron | neutron-1 | channel-123 | 200.01 | 0.06% | $31 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 13.32 | 0.00% | $2 | ignored · unrecoverable |
| Persistence | core-1 | channel-53 | 0.20 | 0.00% | $0 | ignored · small |
| Umee | umee-1 | channel-29 | 0.10 | 0.00% | $0 | ignored · unrecoverable |
| Indigo | indigo-1 | channel-256 | 0.00 | 0.00% | $0 | ignored · small |

In scope $54,165 · small $789 · unrecoverable $2 · deprecated $0


## stLUNA (stuluna, host phoenix-1)

Supply 75,786.52 · RR 1.979438 · $8,015 total · 58.4% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 31,511.96 | 41.58% | $3,333 | in scope |
| Terra | phoenix-1 | channel-13, channel-52 | 29,379.24 | 38.77% | $3,107 | in scope |
| Carbon | carbon-1 | channel-47 | 11,436.45 | 15.09% | $1,210 | in scope |
| Osmosis | osmosis-1 | channel-5 | 2,366.49 | 3.12% | $250 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 1,006.96 | 1.33% | $106 | ignored · small |
| Secret | secret-4 | channel-40 | 76.62 | 0.10% | $8 | ignored · small |
| Comdex | comdex-1 | channel-49 | 8.27 | 0.01% | $1 | ignored · unrecoverable |
| Kujira | kaiyo-1 | channel-8 | 0.42 | 0.00% | $0 | ignored · unrecoverable |
| Axelar | axelar-dojo-1 | channel-69 | 0.10 | 0.00% | $0 | ignored · small |
| Evmos | evmos_9001-2 | channel-9 | 0.01 | 0.00% | $0 | ignored · unrecoverable |

In scope $7,900 · small $115 · unrecoverable $1 · deprecated $0


## stSAGA (stusaga, host ssc-1)

Supply 109,843.16 · RR 1.279077 · $5,292 total · 13.6% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 94,869.65 | 86.37% | $4,571 | in scope |
| Osmosis | osmosis-1 | channel-5 | 12,047.75 | 10.97% | $580 | in scope |
| Carbon | carbon-1 | channel-47 | 1,404.83 | 1.28% | $68 | ignored · small |
| Saga | ssc-1 | channel-213 | 1,352.80 | 1.23% | $65 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 167.50 | 0.15% | $8 | ignored · small |
| Neutron | neutron-1 | channel-123 | 0.64 | 0.00% | $0 | ignored · small |

In scope $5,216 · small $76 · unrecoverable $0 · deprecated $0


## stEVMOS (staevmos, host evmos_9001-2 · host zone stopped, ignored entirely)

Supply 7,955,610.62 · RR 1.589778 · $5,216 total · 76.0% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Evmos | evmos_9001-2 | channel-9 | 5,176,921.61 | 65.07% | $3,394 | ignored · unrecoverable |
| Stride | stride-1 | – | 1,905,584.08 | 23.95% | $1,249 | ignored · unrecoverable |
| Osmosis | osmosis-1 | channel-5 | 822,152.94 | 10.33% | $539 | ignored · unrecoverable |
| Cosmos Hub | cosmoshub-4 | channel-0 | 47,423.48 | 0.60% | $31 | ignored · unrecoverable |
| Carbon | carbon-1 | channel-47 | 3,074.38 | 0.04% | $2 | ignored · unrecoverable |
| Crescent | crescent-1 | channel-51 | 273.64 | 0.00% | $0 | ignored · unrecoverable |
| Kujira | kaiyo-1 | channel-8 | 75.11 | 0.00% | $0 | ignored · unrecoverable |
| Canto | canto_7700-1 | channel-74 | 56.55 | 0.00% | $0 | ignored · unrecoverable |
| Comdex | comdex-1 | channel-49 | 25.73 | 0.00% | $0 | ignored · unrecoverable |
| Injective | injective-1 | channel-6 | 9.99 | 0.00% | $0 | ignored · unrecoverable |
| Juno | juno-1 | channel-24 | 8.02 | 0.00% | $0 | ignored · unrecoverable |
| Acrechain | acre_9052-1 | channel-57 | 4.78 | 0.00% | $0 | ignored · unrecoverable |
| Composable | centauri-1 | channel-134 | 0.30 | 0.00% | $0 | ignored · unrecoverable |
| Secret | secret-4 | channel-40 | 0.01 | 0.00% | $0 | ignored · unrecoverable |

In scope $0 · small $0 · unrecoverable $5,216 · deprecated $0


## stDYM (stadym, host dymension_1100-1 · deprecated zone, ignored entirely)

Supply 257,169.44 · RR 1.096130 · $5,113 total · 29.7% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 180,780.26 | 70.30% | $3,594 | ignored · deprecated |
| Dymension | dymension_1100-1 | channel-197 | 45,433.60 | 17.67% | $903 | ignored · deprecated |
| Osmosis | osmosis-1 | channel-5 | 24,940.62 | 9.70% | $496 | ignored · deprecated |
| Neutron | neutron-1 | channel-123 | 5,300.13 | 2.06% | $105 | ignored · deprecated |
| Cosmos Hub | cosmoshub-4 | channel-0 | 714.70 | 0.28% | $14 | ignored · deprecated |
| Carbon | carbon-1 | channel-47 | 0.13 | 0.00% | $0 | ignored · deprecated |

In scope $0 · small $0 · unrecoverable $0 · deprecated $5,113


## stJUNO (stujuno, host juno-1)

Supply 246,844.95 · RR 1.933365 · $4,474 total · 42.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 140,869.86 | 57.07% | $2,553 | in scope |
| Osmosis | osmosis-1 | channel-5 | 91,743.04 | 37.17% | $1,663 | in scope |
| Cosmos Hub | cosmoshub-4 | channel-0 | 10,780.01 | 4.37% | $195 | ignored · small |
| Secret | secret-4 | channel-40 | 3,253.61 | 1.32% | $59 | ignored · small |
| Umee | umee-1 | channel-29 | 77.89 | 0.03% | $1 | ignored · unrecoverable |
| Juno | juno-1 | channel-24 | 55.30 | 0.02% | $1 | in scope |
| Comdex | comdex-1 | channel-39, channel-49 | 37.57 | 0.02% | $1 | ignored · unrecoverable |
| Evmos | evmos_9001-2 | channel-9 | 11.91 | 0.00% | $0 | ignored · unrecoverable |
| Canto | canto_7700-1 | channel-74 | 10.00 | 0.00% | $0 | ignored · small |
| Carbon | carbon-1 | channel-47 | 4.39 | 0.00% | $0 | ignored · small |
| Kujira | kaiyo-1 | channel-8 | 0.53 | 0.00% | $0 | ignored · unrecoverable |
| Acrechain | acre_9052-1 | channel-57 | 0.49 | 0.00% | $0 | ignored · small |
| Neutron | neutron-1 | channel-123 | 0.35 | 0.00% | $0 | ignored · small |

In scope $4,217 · small $255 · unrecoverable $2 · deprecated $0


## stSTARS (stustars, host stargaze-1 · host zone stopped, ignored entirely)

Supply 14,563,642.58 · RR 1.917700 · $1,670 total · 85.2% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Osmosis | osmosis-1 | channel-5 | 5,930,860.51 | 40.72% | $680 | ignored · unrecoverable |
| Stargaze | stargaze-1 | channel-19 | 5,365,052.27 | 36.84% | $615 | ignored · unrecoverable |
| Stride | stride-1 | – | 2,151,137.02 | 14.77% | $247 | ignored · unrecoverable |
| Juno | juno-1 | channel-24 | 795,680.41 | 5.46% | $91 | ignored · unrecoverable |
| Cosmos Hub | cosmoshub-4 | channel-0 | 219,822.43 | 1.51% | $25 | ignored · unrecoverable |
| Carbon | carbon-1 | channel-47 | 89,046.95 | 0.61% | $10 | ignored · unrecoverable |
| Chihuahua | chihuahua-1 | channel-99 | 11,364.18 | 0.08% | $1 | ignored · unrecoverable |
| Canto | canto_7700-1 | channel-74 | 550.08 | 0.00% | $0 | ignored · unrecoverable |
| Acrechain | acre_9052-1 | channel-57 | 84.86 | 0.00% | $0 | ignored · unrecoverable |
| Terra | phoenix-1 | channel-52 | 20.54 | 0.00% | $0 | ignored · unrecoverable |
| Kujira | kaiyo-1 | channel-8 | 10.71 | 0.00% | $0 | ignored · unrecoverable |
| Evmos | evmos_9001-2 | channel-9 | 8.50 | 0.00% | $0 | ignored · unrecoverable |
| Umee | umee-1 | channel-29 | 3.12 | 0.00% | $0 | ignored · unrecoverable |
| Gravity Bridge | gravity-bridge-3 | channel-121 | 1.00 | 0.00% | $0 | ignored · unrecoverable |

In scope $0 · small $0 · unrecoverable $1,670 · deprecated $0


## stSOMM (stusomm, host sommelier-3 · too small, ignored entirely)

Supply 1,334,084.22 · RR 1.080333 · $678 total · 48.8% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 683,170.78 | 51.21% | $347 | ignored · small |
| Osmosis | osmosis-1 | channel-5 | 648,687.88 | 48.62% | $330 | ignored · small |
| Cosmos Hub | cosmoshub-4 | channel-0 | 2,225.42 | 0.17% | $1 | ignored · small |
| Sommelier | sommelier-3 | channel-150 | 0.14 | 0.00% | $0 | ignored · small |

In scope $0 · small $678 · unrecoverable $0 · deprecated $0


## stCMDX (stucmdx, host comdex-1 · host zone stopped, ignored entirely)

Supply 1,599,360.16 · RR 1.445795 · $207 total · 78.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Comdex | comdex-1 | channel-49 | 1,237,229.86 | 77.36% | $160 | ignored · unrecoverable |
| Stride | stride-1 | – | 337,462.75 | 21.10% | $44 | ignored · unrecoverable |
| Cosmos Hub | cosmoshub-4 | channel-0 | 12,050.69 | 0.75% | $2 | ignored · unrecoverable |
| Osmosis | osmosis-1 | channel-5 | 12,002.17 | 0.75% | $2 | ignored · unrecoverable |
| Kujira | kaiyo-1 | channel-8 | 614.69 | 0.04% | $0 | ignored · unrecoverable |

In scope $0 · small $0 · unrecoverable $207 · deprecated $0


## stUMEE (stuumee, host umee-1 · host zone stopped, ignored entirely)

Supply 22,431,695.36 · RR 1.479574 · $159 total · 28.9% escrowed off Stride

| Chain | Chain id | Stride channel(s) | Amount | % of supply | USD | Status |
|---|---|---|---:|---:|---:|---|
| Stride | stride-1 | – | 15,953,084.27 | 71.12% | $113 | ignored · unrecoverable |
| Osmosis | osmosis-1 | channel-5 | 5,354,748.58 | 23.87% | $38 | ignored · unrecoverable |
| Cosmos Hub | cosmoshub-4 | channel-0 | 1,123,289.31 | 5.01% | $8 | ignored · unrecoverable |
| Crescent | crescent-1 | channel-51 | 285.32 | 0.00% | $0 | ignored · unrecoverable |
| Kujira | kaiyo-1 | channel-8 | 174.75 | 0.00% | $0 | ignored · unrecoverable |
| Umee | umee-1 | channel-29 | 62.23 | 0.00% | $0 | ignored · unrecoverable |
| Injective | injective-1 | channel-6 | 50.91 | 0.00% | $0 | ignored · unrecoverable |

In scope $0 · small $0 · unrecoverable $159 · deprecated $0

