# Wind-down rehearsal log

Network: k8s `integration` (stride-test-1 4 vals, cosmoshub-test-1 8 vals, osmosis-test-1 3 vals).
Branches: `wind-down-rehearsal` (v35), `wind-down-rehearsal-v34` (old binary).
Spec: docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md. Every entry below is appended by the phase scripts.
2026-10-02T23:01:46Z ## Phase 0: pre-flight
2026-10-02T23:01:52Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-02T23:01:57Z CHECKPOINT FAIL: stride channel-0 client is cosmoshub
2026-10-02T23:03:22Z ## Phase 0: pre-flight
2026-10-02T23:03:27Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-02T23:03:34Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-02T23:03:39Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-02T23:03:44Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-02T23:03:50Z CHECKPOINT FAIL: hub ica host allows MsgTransfer
2026-10-02T23:04:08Z ## Phase 0: pre-flight
2026-10-02T23:04:13Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-02T23:04:19Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-02T23:04:24Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-02T23:04:29Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-02T23:04:34Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-02T23:04:36Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-02T23:04:36Z CHECKPOINT PASS: REST stride reachable
2026-10-02T23:04:36Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-02T23:04:37Z CHECKPOINT PASS: REST osmosis reachable
2026-10-02T23:04:47Z tx 63D3F172FA9D3BE92B106B9723B0FACDA64ABB5EE7000F935BDC044729279189 code=0 
2026-10-02T23:04:47Z CHECKPOINT PASS: vault receives
broadcast output (osmosisd):
```

```
2026-10-02T23:04:56Z ms_tx: no txhash in broadcast output for osmosisd
2026-10-02T23:04:56Z CHECKPOINT FAIL: vault spends (multisig)
2026-10-02T23:06:09Z ## Phase 0: pre-flight
2026-10-02T23:06:16Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-02T23:06:25Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-02T23:06:31Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-02T23:06:36Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-02T23:06:43Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-02T23:06:44Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-02T23:06:45Z CHECKPOINT PASS: REST stride reachable
2026-10-02T23:06:45Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-02T23:06:45Z CHECKPOINT PASS: REST osmosis reachable
2026-10-02T23:06:56Z tx D80335837A85734A11F4BD2EAF99B3F50A14E5E5939BCFC61788D634000D6D0F code=0 
2026-10-02T23:06:56Z CHECKPOINT PASS: vault receives
broadcast output (osmosisd):
```

ms_tx step failed: osmosisd tx sign
Error: accepts 1 arg(s), received 2
command terminated with exit code 1
```
2026-10-02T23:07:06Z ms_tx: no txhash in broadcast output for osmosisd
2026-10-02T23:07:06Z CHECKPOINT FAIL: vault spends (multisig)
2026-10-02T23:07:37Z ## Phase 0: pre-flight
2026-10-02T23:07:43Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-02T23:07:49Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-02T23:07:55Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-02T23:08:01Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-02T23:08:07Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-02T23:08:10Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-02T23:08:10Z CHECKPOINT PASS: REST stride reachable
2026-10-02T23:08:10Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-02T23:08:10Z CHECKPOINT PASS: REST osmosis reachable
2026-10-02T23:08:19Z tx AB139E00BC746C734F55E563CDC8616CFD6E9596C55C6F724C6A05742E4FF749 code=0 
2026-10-02T23:08:19Z CHECKPOINT PASS: vault receives
broadcast output (osmosisd):
```
{"height":"0","txhash":"83BEA1910572B6FB29A74C49805549EBF196B4F0518267CCDAE84364715727B8","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-02T23:08:36Z tx 83BEA1910572B6FB29A74C49805549EBF196B4F0518267CCDAE84364715727B8 code=0 
2026-10-02T23:08:36Z CHECKPOINT PASS: vault spends (multisig)
2026-10-02T23:08:50Z tx 992AC42C19F40244FAD2213D78E0CB4075E7DCF841C06A0A9C0FA9F693C54C0A code=0 
2026-10-02T23:08:50Z CHECKPOINT PASS: sweep operator spends
2026-10-02T23:08:55Z host zones not seeded yet: skipping the withdraw-address check
2026-10-02T23:09:12Z gaia v25.1.0, osmosis 28.0.0, strided 
2026-10-02T23:09:21Z ## Seed (v34)
2026-10-02T23:09:46Z CHECKPOINT PASS: admin-ms address
2026-10-02T23:09:49Z CHECKPOINT PASS: vault-ms address
2026-10-02T23:09:55Z CHECKPOINT PASS: hub-ms address
2026-10-02T23:09:55Z ### fund admin-ms
```
$ strided_old tx bank send faucet stride1mymazvsd79f9yhjq4n84dchyf6zvfmd8ed2nxc 1000000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:09:56 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 110745
{"height":"0","txhash":"87C1CED8EB389832C5CA49A909E674E44FC887FAD533AA2AF4A8276CDF036117","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:10:04Z ### fund sweep operator
```
$ strided_old tx bank send faucet stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy 100000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:10:05 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 75082
{"height":"0","txhash":"F2AED251C5184F75E6CFCC8A8E530E5ECB85DA40479F115EEF566AC20602798E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:10:14Z ### fund vault-ms
```
$ osmosisd tx bank send faucet osmo1mymazvsd79f9yhjq4n84dchyf6zvfmd8jaelyx 100000000uosmo --from faucet --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 138232
{"height":"0","txhash":"BC0726D1ADA6F4095E7DF1AA9DBBFA80B115CC14DAA274F7EF4F580C9F646F4F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:10:20Z ### fund hub-ms
```
$ gaiad tx bank send faucet cosmos1h0dup2qw23uhgn9nxyhyze4cxzrgu8rtrcnv7d 100000000uatom --from faucet --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 171159
{"height":"0","txhash":"22B4410A0E7DF39B94B033B37DB70682D2C364ACB053B2CE252C95A176195DEA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:10:35Z CHECKPOINT FAIL: 8 hub validators
