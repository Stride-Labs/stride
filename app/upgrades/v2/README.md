# Upgrade v2 Changelog

1. PENDING status for IBC/ICA function calls ([6660f60](https://github.com/Stride-Labs/stride/commit/6660f60094674b2e077f3775982ab4acc8a5ea96)). Added additional status field on internal accounting records to track when IBC/ICA calls are in flight and prevent re-submission. 
2. Add Validator through Governance ([c757364](https://github.com/Stride-Labs/stride/commit/c757364c4f532a8f7b9d17531f189c41cde90b14)). Added governance proposal type to enable adding validator's through governance. 
3. Validator Rebalancing ([725b991](https://github.com/Stride-Labs/stride/commit/725b9912073e4ff8c1fd5574ba4ebd68ec6aee88)). Added `rebalance-validators` transaction to redistribute delegations after validator weights are updated.