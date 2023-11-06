# The StakeIBC Module

The StakeIBC Module contains Stride's main app logic. It relies on some assumptions. We document those here and game what impact violating those assumptions would have.

Violatins of `x/stakeibc` assumptions can be categorized by their potential worst-case consequence:
- **SEVERE** - user funds are at risk (e.g. minting bugs, accounting mismatches).
- **MEDIUM** - app downtime, unexpected delays (e.g. delayed delivery of user funds) impacting users.
- **MILD** - unexpected behavior but no measurable impact on users.



### Assumptions in `x/stakeibc`

- **[SEVERE]** If there is an error in the `beginBlocker` or `endBlocker`, all other state changes made in the same `beginBlocker` or `endBlocker` are reverted (e.g. `CacheContext`). Alternatively, Stride's `beginBlocker` or `endBlocker` logic is modular, so that it can fail at any intermediate step without causing any accounting issues. 

- **[SEVERE]** CallbackIDS are unique when issuing Interchain Account logic.

- **[MEDIUM]** Automated beginBlocker and endBlocker logic runs at the beginning and end of each block, with no error.

- **[MEDIUM]** There are no gas constraints in the beginBlocker and endBlocker.

- **[MEDIUM]** Passive accounting runs on the host executes upon `msgDelegate`, `msgUndelegate` and `msgRedelegate`.
