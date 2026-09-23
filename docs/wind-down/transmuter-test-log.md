# Transmuter mainnet test log (stATOM)

Plan: `docs/superpowers/plans/2026-09-23-transmuter-mainnet-test.md`. Reference: `docs/wind-down/transmuter.md`.
Every tx below was signed from the `transmuter-test-*` keys in the local `test` keyring; amounts are real mainnet funds.

## Address book

| Role | Chain | Key | Address |
|---|---|---|---|
| `$KEY` admin / moderator / LP | osmosis-1 | transmuter-test-1 | osmo1v0694qqq6ztzxvzl807dgq7h3e857hdxvpmdlc |
| `$KEY2` stranger / admin candidate | osmosis-1 | transmuter-test-2 | osmo1yju8w83cc3z2jtqfkawus29yspw4dygecwmyfp |
| `$SKEY` seeder | stride-1 | transmuter-test-1 | stride1v0694qqq6ztzxvzl807dgq7h3e857hdx83gpax |
| Hub forwarder | cosmoshub-4 | transmuter-test-1 | cosmos1v0694qqq6ztzxvzl807dgq7h3e857hdxy6gaf2 |
| Injective forwarder | injective-1 | Keplr (same mnemonic) | to be recorded when used |
| Secret forwarder | secret-4 | Keplr | to be recorded when used |

## Task 1: snapshot (2026-09-23 20:00 UTC, osmosis-1 height 71,182,675)

| Input | Value |
|---|---|
| stATOM redemption rate (`stakeibc/host_zone/cosmoshub-4`) | `2.000174393066540432` |
| stATOM/ATOM taker fee | `0.0002` |
| Pool creation fee | `20000000 factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC` |
| Market, 10 stATOM → ATOM (SQS, pool 1283) | `19797431` uatom, 1.02% under RR |
| `$KEY` balances | 8.838973 ATOM, 25 USDC (noble, `ibc/498A…`), 5.037953 stATOM, 7.903662 OSMO |
| `$KEY2` balances | 1.007161 stATOM, 1 OSMO |
| `$SKEY` balances | 4 stATOM, 5 STRD |
| Hub balances | 0.5 ATOM |

Deviations from the plan's prerequisites: `$KEY` holds 8.84 ATOM instead of 10 (fund with what is
there); `$KEY2` has no ATOM yet (it will receive ATOM from its first swap); the fee token is
noble USDC, not allUSDC, so Task 2 gains a step: join the allUSDC pool with 25 USDC to mint allUSDC.

Overflow arithmetic for the real pool, factors `1e18` (stATOM) and `2000174393066540432` (ATOM):
gcd 16, lcm `1.25e35` (fits Uint128), normalized total for 1.3M stATOM plus its ATOM `3.25e29` (fits),
alloyed minted for all stATOM `2.6e12` (fits). Expected swap math: 1,000,000 ustatom exact-in →
2,000,174 uatom; 2,000,000 uatom exact-out ← 999,913 ustatom.

Decisions: subdenom `stATOMtest`; at the end the pool is frozen (`set_active_status false`) and
not emptied.

## Tx log

| # | Task | What | Tx hash | Height | Result |
|---|---|---|---|---|---|
| 1 | 2 | Mint 25 allUSDC: `join_pool` on the allUSDC pool (osmo147h…) with 25 noble USDC | `1FE232464F6B30B7D28DB60652BDE0B4C0E7C90EF268F36AD764D0328AB96906` | 71182733 | 25,000,000 allUSDC minted 1:1, no fee. First attempt at 0.0025uosmo was refused at CheckTx (`insufficient fees … required: 7499uosmo`): Osmosis base fee is 0.03 uosmo; use `--gas-prices 0.04uosmo` |
| 2 | 2 | `MsgCreateCosmWasmPool` code 996, `instantiate.json` (stATOM `1e18`, ATOM and alloyed `2000174393066540432`) | `74F5E96583F15BE84906C36754AD5F220190E8C2D5F4CC3928A9D312539D7439` | 71182754 | **pool 3590**, contract `osmo13wcvdtkcu459zjuqdh9f7jshlg64sps0xsduxdsn9r6yez0t270qeg6z6e`, alloyed `factory/osmo13wcvdtkcu459zjuqdh9f7jshlg64sps0xsduxdsn9r6yez0t270qeg6z6e/alloyed/stATOMtest`, cw2 version 3.2.0, gas 1,437,728. `--dry-run` cannot resolve a key name in simulation mode; skip it |

### Task 2 state checks (all as predicted)

asset configs stATOM `1e18`, ATOM `2000174393066540432`, alloyed `2000174393066540432`; admin and moderator `$KEY`;
active; liquidity `0 stATOM, 0 ATOM`; no limiters; swap fee `0`; spot price stATOM/ATOM `2.000174393066540432`,
ATOM/stATOM `0.49995640553465114`; `calc_out_amt_given_in` 1 stATOM → `Insufficient pool asset: required: 2000174ATOM,
available: 0ATOM`; `calc_in_amt_given_out` 2 ATOM → same shape, required 2000000; swap fee 0.001 → `Invalid swap fee:
expected: 0, actual: 0.001`; wasm contract admin and creator `osmo1rxjakgd8yhks2j7hc7pt6a22z3zd64grexpyf7` (module).
| 3 | 3 | Funding join: `join_pool` with 8,000,000 uatom from `$KEY` (8 ATOM, not 10: that is what the key held) | `045BBF7923CA2D9D180B5BC3D9E6BCE78601740BA035C8FFC42A82B17A59D2D9` | 71182801 | 8,000,000 alloyed minted to `$KEY`; liquidity 0 stATOM / 8,000,000 ATOM; total shares 8,000,000 |

### Task 3 quotes after funding (all as predicted)

1,000,000 ustatom exact-in → `2000174` uatom; 2,000,000 uatom exact-out ← `999913` ustatom; 1 ustatom → `2` uatom;
1 uatom → `0` ustatom. SQS at 20:07 UTC (about 2 minutes after funding) still lists only pools 1136, 1283, 803 as
stATOM→ATOM candidates and quotes 1 stATOM → 1,979,748 uatom via 1283; its candidate-route cache expires every
20 minutes, re-check later. `GET /pools/3590` on SQS: Not Found at that time.
| 4 | 4.1 | Router exact-in from `$KEY2`: 1,000,000 ustatom → ATOM via pool 3590, min out 1,990,000 | `27BF9766A753177FCD3139C4F44E79434D220FEE215FBCB75EAD36CE0667C862` | 71182845 | `token_swapped` tokens_in `999800` ustatom (taker fee 200 = 0.02%), tokens_out `1999774` uatom, exactly floor(999800 × RR). Pool 999,800 stATOM / 6,000,226 ATOM. `$KEY2` is left with 7,161 ustatom, so the next steps run from `$KEY` |
| 5 | 4.2 | Router exact-out from `$KEY`: 2,000,000 uatom out, max in 1,010,000 ustatom | `E216956893CAC39839C974C7BE1AAFD04BACB5BF658C217606FB450185E2B3DE` | 71182874 | tokens_in `999913` ustatom (= ceil(2e6 / RR)), tokens_out `2000000` uatom; taker fee `201` ustatom charged on top (rounded up); 10,087 of the max refunded. Pool 1,999,713 stATOM / 4,000,226 ATOM |
| 6 | 4.3 | Direct path, join: `join_pool` with 1,000,000 ustatom from `$KEY` | `931D5AD79AEB90CDDB4B45E8ACA45C5D06A265C70EA1E0C7930FF9B917C6FCC0` | 71182899 | `2000174` alloyed minted (= floor(1e6 × RR)); `$KEY` shares 10,000,174 |
| 7 | 4.3 | Direct path, exit: `exit_pool` 2,000,174 uatom from `$KEY` | `DF40B6D73D475B9FB1DC9D2023C711C6FC95BB1644D15CF9F0188D698C65327E` | 71182921 | burned exactly `2000174` alloyed from `$KEY`, sent `2000174` uatom; shares back to 8,000,000. Net 1 stATOM → 2.000174 ATOM with no fee at all |
| – | 4.4 | Dust: 1 uatom exact-in with min out 0 | – | – | rejected by the CLI's ValidateBasic: `min out amount or max in amount should be positive, was (0)` |
| 8 | 4.4 | Dust: 1 uatom exact-in with min out 1, from `$KEY2` | `04D2650AE18392F46E31B6502295D7A2A0F8DF560E15E4EBC1263136E839BC76` | 71182969 | code 5: `Amount of coin to be operated on must be greater than zero` (the 0.02% taker fee rounds 1 uatom to 0 before the contract sees it). Gas paid, nothing moved |
| 9 | 4.5 | Reverse direction from `$KEY2`: 1,000,000 uatom → stATOM | `F47424054EAD44FB5EA734B6EDAE1A1C16B457B00468C75F23B0E62134EC9A0F` | 71182950 | tokens_in `999800` uatom (200 taker fee), tokens_out `499856` ustatom = floor(999800 / RR). Two-way at the same rate, as expected |
| 10 | – | `$KEY` → `$KEY2` 1,500,000 ustatom (top-up for later tasks) | `21274457DD14A2644E2D7A0A5A49A718C27337B184D5C66FB948079A760E9B3B` | 71182986 | ok |
| 11 | 4.6 | Oversize from `$KEY`: 2,000,000 ustatom exact-in | `60F255F1B45642A21D293BA6A3346F9B1258967397577D6466AB3B5C3981A9F5` | 71182996 | code 5 `insufficient funds` from the bank: `$KEY` only held 1.537 stATOM. Wrong failure, retried from `$KEY2` |
| 12 | 4.6 | Oversize from `$KEY2`: 2,000,000 ustatom exact-in against 2,999,852 uatom in the pool | `8C42AB7FE3723977A7D4D4AF99F2DC7FC32EA984901CE16F79A86CC297004377` | 71183019 | code 5: `Insufficient pool asset: required: 3999548ibc/2739…, available: 2999852ibc/2739…`. Nothing moved. This is the user-facing error for an under-funded pool |

Pool after Task 4: 2,499,857 stATOM / 2,999,852 ATOM; total shares 8,000,000, all held by `$KEY`. No third-party swap
was seen through 20:20 UTC.
| 13 | 7.1 | `add_new_assets` Penumbra two-hop denom (supply 0) from `$KEY` | `CC64DB1C1BA5EF932F5ECB918121FA6F2CCC04D597D4FD34BDFB4B2766BC5ADF` | 71183085 | code 5: `Denom has no supply, it might be an invalid denom: ibc/B667…` |
| 14 | 7.1 | same with `--amount 1uosmo` from `$KEY` | `84873EAC6E3A0D4CA8CEC82FE2D0EE84F3EE59D2D78D6E6137429707CFD3FFEA` | 71183096 | code 5: `Funds must be empty` |
| 15 | 7.1 | same from `$KEY2` | `C57911AFB749B7AB737538086A5AA8178E723B1A42A8D39B33D8C8FBABDBBAA2` | 71183106 | code 5: `Unauthorized` |
| 16 | 5 | Stride → Hub: 1,000,000 stuatom over channel-0 to the Hub test address | `4D66536AA998D662BE0273F33F2D0CEA4532EBBB4A70260ECFFCD45842BE44F5` (stride-1) | 40503885 | packet 227560 → channel-391; landed as `1000000 ibc/B05539…` within ~3 minutes, relayed by public relayers, commitment cleared |
| 17 | 6 | Hub → Osmosis: 1,000,000 `ibc/B05539…` over channel-141 to `$KEY` | `F7403661C69DEAAA9B4E6C4E0BF17C7B2E871F7EA9A4F032283E547AB5D5FB09` (cosmoshub-4) | 33091749 | packet 4998328 → channel-0; arrival pending |
| 18 | 7.2 | `add_new_assets` Hub (`ibc/7451…`) and Agoric (`ibc/C86C…`) two-hop denoms, factor `1e18`, from `$KEY` | `DDBBDBC69EE85F248EA9EA0F6C581D814466717AF2F481489FDC6244D0F3C14D` | 71183148 | ok; five pool assets plus alloyed; spot price Hub-stATOM/ATOM `2.000174393066540432`, Hub-stATOM/canonical stATOM `1`. Injective and Secret denoms wait for their seed transfers (zero supply today) |
| 19 | 6 | Hub-hop stATOM arrived on `$KEY` as `1000000 ibc/7451074F…` (trace `transfer/channel-0/transfer/channel-391/stuatom`), about 4 minutes after the Hub send, public relayers | – | – | as predicted in the reference doc |
| 20 | 7.3 | Router: 400,000 `ibc/7451…` → ATOM from `$KEY` | `56C06A51FDD4DBCB1AE4357B396DC4E1E850E14015C1CDE4275A30EBF220C0AD` | 71183203 | tokens_in `399600` (taker fee 400 = **0.1% default**, not the 0.02% stATOM/ATOM override), tokens_out `799269` uatom = floor(399600 × RR) |
| 21 | 7.3 | `join_pool` 300,000 `ibc/7451…` from `$KEY` | `1705617ABFADE1E27A75BB43EC119F19423719DBEE5C5AACDAF5DEFDC0A6DA8D` | 71183211 | `600052` alloyed minted = floor(300000 × RR); no fee |
| 22 | 7.3 | Router: 300,000 `ibc/7451…` → canonical stATOM from `$KEY` | `4C59BF14BA9BC6B0DEE90432CCED002911D147AC85E9CA048D4B46F4BBEE070A` | 71183216 | tokens_in `299700` (0.1% fee), tokens_out `299700` canonical stATOM: the pool de-hops a foreign route 1:1 |

### Router visibility (Task 3.3 / 7.4), checked repeatedly 20:07–20:50 UTC

SQS (`sqs.osmosis.zone`, the router behind app.osmosis.zone) ingested the pool: `GET /pools?filter[id]=3590` returns it
with `liquidity_cap: 12` (USD) and `liquidity_cap_error: zero cap for denom (ibc/7451…); zero cap for denom (ibc/C86C…)`,
i.e. it prices ATOM and canonical stATOM but has no price for the two-hop denoms. It never offered pool 3590 as a
stATOM→ATOM route (`/router/routes` stayed at 1136, 1283, 803; quotes stayed on 1283 at 1,979,748 per stATOM, worse
than our 1,999,774). Reason, from the SQS config: `DynamicMinLiquidityCapFiltersDesc` is keyed on the tokens' total
liquidity across the chain, not the trade size; for a pair whose tokens have ≥ $1M of liquidity only pools with a cap of
≥ $40,000 are route candidates. A $12 pool can never qualify; a real pool holding the zone's backing will. A quote for
the two-hop denom fails outright: `denom is not a valid chain denom (ibc/7451…)`, so the app cannot quote or swap
foreign-route stATOM; those holders need a contract-execute path (join then exit, or a hosted page) or the pool's
own `calc_out_amt_given_in`. Not a contract problem, a frontend one; recorded for §7.
| 23–29 | 8.1 | From `$KEY2`: `set_active_status`, `mark_corrupted_assets`, `rescale_normalization_factor`, `register_limiter`, `assign_moderator`, `transfer_admin`, `claim_admin` | `FD7F1892…`, `AF44321E…`, `BEACA1F0…`, `BCAB4CD9…`, `E3CAD757…`, `A4965886…`, `5CCFB498…` | 71183301–71183344 | all code 5 `Unauthorized`; nothing changed |
| 30 | 8.2 | `set_active_status false` from `$KEY` (moderator) | `D546A277C1293482AC27CDA3EF1F3CB048216B84E2A4C295DECAF417ECBC423F` | 71183369 | `is_active: false` |
| 31 | 8.2 | router swap 100,000 ustatom from `$KEY2` while frozen | `A72CFA9CC0D3B4DDE19E8534BE7EF21A65F602D214E0AA69F67699FDDCAD3184` | 71183378 | code 5 `The pool is currently inactive` |
| 32 | 8.2 | `exit_pool` 1,000 uatom from `$KEY` (admin) while frozen | `AD34F97FC91A3A368F3E3637C20B2B50B15D5A563EF74291DEC469255ABF4373` | 71183384 | code 5 `The pool is currently inactive` (the freeze binds the admin too) |
| 33 | 8.2 | `join_pool` 100,000 ustatom from `$KEY2` while frozen | `A374BE30E7F799C9965A5704F8B91A5AAD5400F9D7294EC9787E9ADEB72C72D4` | 71183394 | code 5 `The pool is currently inactive` |
| 34 | 8.2 | `set_active_status false` again | `D801F439F5F036965E2A0630F9600D986DC924BA08B127CC1061F9C8D9833761` | 71183400 | code 5 `Attempt to set pool to active status to false when it is already false` |
| 35 | 8.2 | `set_active_status true` | `3AC184AD1B7604A17F2722833B841B5F1253A228D4856826FF68B8003608BAD2` | 71183408 | `is_active: true` |
| 36 | 8.3 | `mark_corrupted_assets [ATOM]` from `$KEY` (moderator) | h 71183435 | | `get_corrupted_denoms` = [ATOM] |
| 37 | 8.3 | router 100,000 ustatom → ATOM from `$KEY2` with ATOM corrupted | h 71183443 | | ok: in 99,980, out 199,977 (ATOM leaving is allowed) |
| 38 | 8.3 | router 100,000 uatom → stATOM from `$KEY2` | h 71183448 | | code 5 `Corrupted asset: ibc/2739… must not increase` |
| 39 | 8.3 | `join_pool` 100,000 uatom from `$KEY` (the vault's top-up) | h 71183457 | | same error: marking the native corrupted also blocks funding, so it can only be flipped after the pool is fully funded |
| 40 | 8.3 | `exit_pool` 100,000 ustatom only from `$KEY` | h 71183463 | | same error: ATOM's weight would rise |
| 41 | 8.3 | `unmark_corrupted_assets [ATOM]` | h 71183472 | | `get_corrupted_denoms` = [] |
| 42 | 8.4 | `rescale_normalization_factor 1/7` | h 71183497 | | code 5 `Rescaling parameter is not divisible: rescale 1000000000000000000 by 1/7` (the ATOM factor happens to be divisible by 7; the stATOM one is not) |
| 43 | 8.4 | `rescale_normalization_factor 1/1e12` | h 71183507 | | code 5 `… rescale 2000174393066540432 by 1/1000000000000` |
| 44 | 8.4 | `rescale_normalization_factor 2/1` | h 71183513 | | ok: every factor doubled (stATOM `2e18`, ATOM and alloyed `4000348786133080864`, both two-hop `2e18`); spot price unchanged `2.000174393066540432` |
| 45 | 8.4 | `rescale_normalization_factor 1/2` | h 71183520 | | back to the original factors, price unchanged |
| 46 | 8.5 | router 300,000 canonical stATOM → Hub-stATOM from `$KEY` (to hold some again) | h 71183564 | | 299,700 → 299,700; weights afterwards stATOM 0.6047, ATOM 0.2326, Hub 0.1627, Agoric 0; total value 8,600,055 uatom-equivalent |
| 47 | 8.5 | `register_limiter` Hub denom, label `route-cap`, static `0.173` (current weight + 0.01) | h 71183573 | | listed |
| 48 | 8.5 | router 10,000 Hub-stATOM → ATOM (weight 0.1627 → 0.165, under the cap) | h 71183580 | | ok |
| 49 | 8.5 | router 200,000 Hub-stATOM → ATOM (would reach ~0.21) | h 71183589 | | code 5 `Upper limit exceeded for ibc/7451…, upper limit is 0.173, …`; nothing moved |
| 50 | 8.5 | router 50,000 ATOM → Hub-stATOM (Hub weight falls) | h 71183596 | | ok: a capped denom can always leave |
| 51 | 8.5 | `deregister_limiter` route-cap | h 71183602 | | code 5 `Denom: ibc/7451… cannot have an empty limiter after it has been registered` (the last limiter is permanent) |
| 52 | 8.5 | `set_static_limiter_upper_limit` route-cap → `1` | h 71183609 | | ok |
| 53 | 8.5 | router 200,000 Hub-stATOM → ATOM again | h 71183616 | | ok: 199,800 → 399,634; weights stATOM 0.6047, ATOM 0.1896, Hub 0.2057; total value 8,600,058 (constant under swaps, +3 from rounding) |
| 54–62 | 8.6 | Admin hand-over: `$KEY` `transfer_admin`→`$KEY2` (h 71183655; candidate visible, admin unchanged); `$KEY2` `add_new_assets` before claiming → `Unauthorized` (h 71183657); `$KEY2` `claim_admin` (h 71183663, admin = `$KEY2`); `$KEY2` `assign_moderator` `$KEY2` (h 71183673); `$KEY2` `transfer_admin`→`$KEY` (h 71183678); `$KEY` `claim_admin` (h 71183686); `$KEY` `assign_moderator` `$KEY` (h 71183693); cancel path: transfer then `cancel_admin_transfer` by the admin (h 71183703/709, candidate null); reject path: transfer then `reject_admin_transfer` by the candidate (h 71183715/721, candidate null, admin unchanged) | | | all as expected; this is the procedure for moving both roles to the vault multisig |
| 63 | 8.7 | `$KEY` bank-sends 1,500,000 alloyed to `$KEY2` | h 71183754 | | `get_shares($KEY2)` = 1,500,000: the LP receipt is a plain bearer token |
| 64 | 8.7 | `$KEY2` `exit_pool` 1,000,000 uatom | h 71183760 | | ok, burned 1,000,000 alloyed from `$KEY2` |
| 65 | 8.7 | `$KEY2` `exit_pool` 600,000 uatom with 500,000 alloyed | h 71183767 | | code 5 `Insufficient shares: required: 600000, available: 500000` |
| 66 | 8.7 | `$KEY2` router 500,000 alloyed → stATOM | h 71183772 | | ok: 499,500 alloyed burned (500 taken as the 0.1% taker fee, so the fee collector now holds 500 alloyed), 249,728 ustatom out; `$KEY2` shares 0; total shares 7,100,552 |
| 67 | 5 | Stride → Secret 1,000,000 stuatom over channel-40 (sent by the operator from Keplr, packet 39378, default 10-minute height timeout) | – | 40504388 | **never relayed**: Stride→Secret has no public relayer; packets from 20–21 Sep (39374–39377) were still pending, and the Secret→Stride direction had 9 unreceived packets from other users. 39378 timed out at Secret height 27,290,944 |
| 68 | 5 | Stride → Secret resend, 6-hour timeout, from `$SKEY` | `8AA2F865EB8FD8C4E0EB20D6C9119A24662D0A865A9C15D4E214F551A22B762C` (stride-1) | 40505362 | packet 39379 |
| 69 | 5 | Manual relay: `hermes tx packet-recv --dst-chain secret-4 --src-chain stride-1 --src-port transfer --src-channel channel-40 --packet-sequences 39379` | – | – | landed on the third attempt; the only live Secret RPC (lavenderfive) answers 429 under hermes's query burst, so single-packet commands with 40 s pauses were needed. Secret key imported into hermes with hd-path `m/44'/529'/0'/0/0`. hermes 1.13.2 flags Stride's SDK 0.54 as unsupported in `health-check` but relays fine with `compat_mode = '0.38'` |
| 70 | 6 | Secret → Osmosis 500,000 `ibc/A0E80E…` over channel-1 to `$KEY` (secretcli, 1 h timeout) | `24B23B62AD87971DF52F7157B7FAF73DAED25FB0B86054B4D3560C3AB3E5346E` (secret-4) | 27291377 | packet 431047 → Osmosis channel-88 |
| – | 6 | Secret → Osmosis over channel-44 | – | – | **no such transfer channel on Secret**: Osmosis channel-476's counterparty is port `wasm.secret1tqmms5awftpuhalcv5h5mg76fa0tkdz4jv9ex4` channel-44, Secret's private-token (SNIP-20) IBC bridge contract, not `transfer`. Plain stATOM on Secret only has channel-1. The reference doc's second Secret denom is therefore for tokens that leave through that bridge contract, if any; dropped from the add list |
| 71 | 6 | Injective → Osmosis 1,000,000 `ibc/A8F392…` over channel-8 to `$KEY`, signed by `injectived` v1.17.2 running under Docker (`--platform linux/amd64`, host CA bundle mounted; there is no macOS build) | `BA21FFA17BB2930300ED7BC078C742D4D52C1D543EB3CC3C05AA9795EDA86AEA` (injective-1) | 184283660 | packet 649337 → Osmosis channel-122. Keplr's send flow would not let the operator pick the channel |
| 72 | 6 | Secret-hop stATOM arrived on `$KEY` as `500000 ibc/8AEB813E…` (trace `transfer/channel-88/transfer/channel-37/stuatom`), public relayers, ~5 minutes | – | – | denom hash as predicted |
| 73 | 3 | `join_pool` 3,000,000 uatom top-up from `$KEY` | h 71187348 | | 3,000,000 alloyed minted |
| 74 | 7.2 | `add_new_assets` Secret two-hop denom `ibc/8AEB…`, factor `1e18` | h 71187354 | | ok; spot price vs ATOM `2.000174393066540432` |
| 75 | 7.3 | router 250,000 `ibc/8AEB…` → ATOM | h 71187364 | | in 249,750 (0.1% default fee), out `499543` = floor(249750 × RR) |
| 76 | 7.3 | `join_pool` 250,000 `ibc/8AEB…` | h 71187368 | | `500043` alloyed minted = floor(250000 × RR) |
| 77 | 6 | Injective packet 649337 received on Osmosis (tx `0E6C27145AF4FBC7D15A3123D141CD0E84B089E7D4EBC00DF05499AB0D3014A8`, h 71187282) | | | **error ack** `{"error":"ABCI code: 2: error handling packet"}`, event `ibccallbackerror-ibc-acknowledgement-error: rate limit exceeded`; Injective refunded the 1 stATOM |
| 78 | 6 | Control: 1,000 ustatom Injective → Osmosis over channel-8 | `3A0EA1A8444EC4E6F169753D3973FE2235C3B1EE9A798BF17D2E55496B98B090` (injective-1) → Osmosis seq 649340, h 71187650 | | same error ack; INJ, USDT and ERC-20 packets on channel-122 in the same minutes succeeded |

### Finding: Injective-hop stATOM cannot enter Osmosis (rate-limiter prefix bug)

Osmosis wraps every contract error from its IBC rate limiter (`x/ibc-rate-limit`, contract
`osmo17r7qdw2zk6jyw62cvwm6flmhtj9q7zd26r8zc6sqyf0pnaq46cfss8hgxg`, cw2 `rate-limiter 0.1.1`) as
`rate limit exceeded`; no quota exists for channel-122 or for the denom (full state scan: 804 flow entries, none
for channel-122). The contract's `Packet::receiver_chain_is_source` tests
`denom.starts_with("transfer/{source_channel}")` **without a trailing slash**. Injective's channel to Osmosis is
`channel-8` and its channel to Stride is `channel-89`, so `transfer/channel-89/stuatom` is mistaken for a token
returning home; the strip of `transfer/channel-8/` fails, the denom becomes empty, the supply query errors and the
packet is rejected regardless of amount. ibc-go itself compares with the trailing slash. Secret (channel-1 vs
channel-37) and the Hub (channel-141 vs channel-391) don't collide, which is why those hops worked. Any denom whose
first hop on Injective is channel-80…89 hits this; other pairs can collide the same way (a source channel that is
a decimal prefix of the trace's first channel).

Consequence: 31k stATOM on Injective (~$108k) cannot be sent to Osmosis as a two-hop denom until Osmosis migrates
the contract (governance). Injective holders can still redeem through Stride in window 1 (both clients active), or
after the halt route Injective → Hub → Osmosis as a three-hop denom (Hub's channel to Injective does not collide
with channel-141), which the pool would then need added. Report to Osmosis before window 2.
| 79 | 9 | `set_active_status false` | `DAC709653638A08E5224301F3F927687D5CACF9417D3D5AECCC7E6F7226964D5` | 71187802 | pool 3590 frozen, not emptied (operator's choice). Liquidity left: 2,350,109 stATOM, 3,131,398 ATOM, 884,418 Hub-stATOM, 499,750 Secret-stATOM, 0 Agoric-stATOM; total shares 10,600,595, of which `$KEY` holds 10,600,095 and the taker-fee collector 500 |

## Task 9: reconciliation (21:55 UTC)

Everything in ATOM-equivalent at RR 2.000174393066540432. Pool value at the freeze is 10,600,603 against
10,600,595 shares: **+8 uatom** for the pool, all rounding in its favour.

| | uatom-equivalent |
|---|---:|
| Start: `$KEY` 8,838,973 ATOM + `$KEY`/`$KEY2` 6,045,114 stATOM | 20,930,255 |
| Inflows: Hub-hop 1,000,000 + Secret-hop 500,000 stATOM | 3,000,261 |
| End: `$KEY` 3,507,574 ATOM, 1,537,539 stATOM, 114,672 Hub-stATOM, 10,600,095 alloyed; `$KEY2` 2,199,751 ATOM, 2,156,745 stATOM | 23,925,996 |
| Difference | 4,520 |
| Explained: taker fees (1,881 stToken-side, 251 ATOM-side, 500 alloyed) ≈ 4,513 + pool rounding 8 | 4,521 |

Unexplained: 1 uatom (rounding of the fee estimate). Gas: 995,252 uosmo across both keys (~$0.04); pool creation 20
allUSDC; 5 allUSDC left on `$KEY`. Off Osmosis: 1 stATOM back on Injective (refunded), 0.5 stATOM still on Secret,
1 stATOM in Stride's channel-40 escrow from the timed-out packet 39378 (refunds when someone relays the timeout;
hermes could not get through the Secret RPC's rate limit for the full `clear packets` run), 0.4975 ATOM on the Hub.

## Findings that change the spec

1. **Normalization factors were inverted in the spec** (caught reading the code before the test; corrected). On chain
   the corrected orientation gives spot price exactly `RR`.
2. **Osmosis's rate limiter rejects Injective-hop stATOM** (prefix bug, channel-8 vs channel-89). ~$108k of stATOM
   on Injective cannot enter Osmosis until Osmosis migrates the contract. Report to Osmosis; Injective holders
   redeem via Stride in window 1, or route via the Hub after the halt.
3. **The Osmosis app cannot see or quote foreign-route stATOM**: SQS calls the two-hop denom "not a valid chain
   denom". Holders need a contract path (CLI, or a page we host). The chain handles it fine.
4. **The app router ignores small pools** (≈ $40k floor for this pair); irrelevant for the real pool, but it means
   no test pool will ever get organic traffic.
5. **Stride → Secret has no relayer** (packets from 20 Sep still pending; the Secret → Stride direction too), and
   the only public Secret RPC rate-limits hermes hard. Window 1 needs us to relay Secret in both directions with
   single-packet commands and pauses, or a private Secret node.
6. **Osmosis channel-476 to Secret is a wasm-port channel** (Secret's SNIP-20 bridge), not a transfer channel;
   bank-held stATOM on Secret has one route, channel-1.
7. **Two-hop pairs pay the 0.1% default taker fee**, canonical stATOM/ATOM 0.02%.
8. **The corrupted-asset lever blocks the vault's own funding and stToken-only exits**: usable only after funding.
9. **A denom's last limiter is permanent**; widen to 1 to disable. Caps behave exactly as designed and total pool
   value is invariant under swaps.
10. **`--dry-run` cannot resolve key names**, Osmosis base fee is 0.03 uosmo, hermes 1.13.2 needs
    `compat_mode = '0.38'` for Stride, `injectived` has no macOS build (Docker works).

Everything else matched the plan's predictions exactly: rounding, fees, freeze, hand-over, bearer alloyed,
`add_new_assets` rules, `Insufficient pool asset` on under-funding.
