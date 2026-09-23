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
