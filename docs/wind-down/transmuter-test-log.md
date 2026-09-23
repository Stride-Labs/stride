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
