# Live-test validator per zone

Generated 2026-09-29 by `scripts/wind-down/pick_live_test_validators.py`; rerun on the day, after the day-0 refresh.
Prices: the snapshot in sttoken-locations.md (total USD / (supply × rate) per token).

The first `MsgUndelegateFromValidators` on each zone drains exactly one validator in full, as the live test of the tx
and its callback, before the empty-list drain of the rest (spec §7, §9 step 3). The pick is the validator with the
smallest recorded delegation of at least one whole token that has no unbonding entry in flight from the delegation ICA (the SDK allows 7
concurrent entries per delegator-validator pair; a validator drained in full never needs a second one). "Next" is the
second-smallest delegation, to show how much the pick matters.

| Zone | Validator | Recorded delegation | USD | Entries in flight | Funded validators | Next smallest (USD) |
|---|---|---:|---:|---:|---:|---:|
| celestia | mhventures (`celestiavaloper1q2kaajedxm0r5xc0twdqz6atap96502d67yjyj`) | 15,861,063 utia | $7.07 | 0 | 92 | $8.48 |
| cosmoshub-4 | icycro (`cosmosvaloper1ukpah0340rx7k3x2njnavwyjv6pfpvn632df9q`) | 14,207,897 uatom | $24.86 | 0 | 62 | $26.51 |
| dydx-mainnet-1 | luganodes (`dydxvaloper1fs0t34g628xdqc8alfefnadq2x3qawt8g88mav`) | 14,328,832,501,947,858,372 adydx | $1.94 | 0 | 25 | $2.25 |
| haqq_11235-1 | digiser2 (`haqqvaloper1nekpsmetpxx2crsuzznuy4epv9eqvj03rtmxae`) | 3,877,996,874,608,234,702,464 aISLM | $15.10 | 0 | 41 | $15.10 |
| injective-1 | autostake (`injvaloper1acgud5qpn3frwzjrayqcdsdr9vkl3p6hrz34ts`) | 9,813,332,455,629,722,717 inj | $76.25 | 0 | 38 | $76.79 |
| juno-1 | cosmosspaces (`junovaloper1836fhsg6yqpu98vezfc7caakchqe8pvske7t8q`) | 9,007,060,654 ujuno | $84.44 | 0 | 21 | $95.05 |
| laozi-mainnet | meria (`bandvaloper1plau7keptn9qdt7nmhphltakv5t054f8lgwjdn`) | 4,158,959,467 uband | $897.36 | 0 | 32 | $897.36 |
| osmosis-1 | haannode (`osmovaloper1hqqzynrdqxzky82mw92ugwsrry0ntrse84g5nr`) | 7,791,194,655 uosmo | $285.37 | 0 | 23 | $485.13 |
| phoenix-1 | coinhall (`terravaloper1ge3vqn6cjkk2xkfwpg5ussjwxvahs2f6at87yp`) | 411,068,185 uluna | $21.96 | 0 | 40 | $27.49 |
| sommelier-3 | ztakeorg (`sommvaloper13ul4wf2gwuwfrqrx70h2j9evje05vtglpc44sc`) | 2,450,134,223 usomm | $0.00 | 0 | 19 | $0.00 |
| ssc-1 | solva (`sagavaloper13pcp0cstupahzz3n36x0dlhpa9fr8m9vlcy78y`) | 117,083,715 usaga | $4.41 | 4 | 16 | $26.46 |

Total value put at risk by the eleven live tests: $1,418.75.
