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
2026-10-02T23:39:13Z ## Seed (v34)
2026-10-02T23:39:32Z CHECKPOINT PASS: admin-ms address
2026-10-02T23:39:34Z CHECKPOINT PASS: vault-ms address
2026-10-02T23:39:42Z CHECKPOINT PASS: hub-ms address
2026-10-02T23:39:42Z ### fund admin-ms
```
$ strided_old tx bank send faucet stride1mymazvsd79f9yhjq4n84dchyf6zvfmd8ed2nxc 1000000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:39:43 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 74904
{"height":"0","txhash":"CBFF482E8582B67C770A07CD8E141A93E91F406E85AFBFACEE5ACDC526CBD0B4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:39:52Z ### fund sweep operator
```
$ strided_old tx bank send faucet stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy 100000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:39:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 75087
{"height":"0","txhash":"01B07052DE7843D361CC5B83216A49865B41A85007A6CB02E9F68D65289DB0F5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:01Z ### fund vault-ms
```
$ osmosisd tx bank send faucet osmo1mymazvsd79f9yhjq4n84dchyf6zvfmd8jaelyx 100000000uosmo --from faucet --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 118902
{"height":"0","txhash":"381B11AB53275787030C6ADD21058A4E240302769ED6A99B5BFE073D6C51AF0B","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:07Z ### fund hub-ms
```
$ gaiad tx bank send faucet cosmos1h0dup2qw23uhgn9nxyhyze4cxzrgu8rtrcnv7d 100000000uatom --from faucet --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 130440
{"height":"0","txhash":"EE4E406F4D3050EF122B9DDA456B4814FE36FE4BE24114B0A77EC53223D86545","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:31Z CHECKPOINT PASS: 8 hub validators
2026-10-02T23:40:31Z CHECKPOINT PASS: 3 osmosis validators
2026-10-02T23:40:31Z ### register hub zone
```
$ strided_old tx stakeibc register-host-zone connection-0 uatom cosmos ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 channel-0 1 false --max-messages-per-ica-tx 3 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:40:32 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 663705
{"height":"0","txhash":"4CB0767EF4A481F63FAC5BD4C2675F4178839ECF9278E9B96CEF88884E99BDE4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:39Z ### register osmo zone
```
$ strided_old tx stakeibc register-host-zone connection-1 uosmo osmo ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B channel-1 1 false --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:40:40 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 647544
{"height":"0","txhash":"251C6019D042D6DF8A3BB6DC7E2E0B74FD45C63F28991E1B074836340102FCCE","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:50Z ### add hub validators
```
$ strided_old tx stakeibc add-validators cosmoshub-test-1 /tmp/hub_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:40:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 1098741
{"height":"0","txhash":"A516142CC6DAD1069F3EEC4187A72815914FE429F8B407C6E34EC082785F0541","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:40:59Z ### add osmo validators
```
$ strided_old tx stakeibc add-validators osmosis-test-1 /tmp/osmo_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:40:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:40:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:41:00 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 402217
{"height":"0","txhash":"5A393AD106E75789EBA9A036548BE1463971129A9766DDB391372E1EBA0141DC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:41:15Z ready: cosmoshub-test-1 delegation_ica_address
2026-10-02T23:41:20Z ready: cosmoshub-test-1 fee_ica_address
2026-10-02T23:41:26Z ready: cosmoshub-test-1 withdrawal_ica_address
2026-10-02T23:41:34Z ready: cosmoshub-test-1 redemption_ica_address
2026-10-02T23:41:38Z ready: osmosis-test-1 delegation_ica_address
2026-10-02T23:41:43Z ready: osmosis-test-1 fee_ica_address
2026-10-02T23:41:50Z ready: osmosis-test-1 withdrawal_ica_address
2026-10-02T23:41:55Z ready: osmosis-test-1 redemption_ica_address
2026-10-02T23:41:55Z ### atom to stride
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 2000000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 223386
{"height":"0","txhash":"7C4FBE8E1E2352AA8538C8CB258B85D26FF2F9F49BD5B143C1D603494E8FA4D5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:42:04Z ### osmo to stride
```
$ osmosisd tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 184975
{"height":"0","txhash":"F62BD73CF850D0CC934580455ACB62DB52B48E6D3D00B8CEF612B737AC8E1DB6","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:42:18Z ready: atom on stride
2026-10-02T23:42:24Z ready: osmo on stride
2026-10-02T23:42:24Z ### liquid stake 1000 ATOM
```
$ strided_old tx stakeibc liquid-stake 1000000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:42:25 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 198025
{"height":"0","txhash":"F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:42:33Z ### liquid stake 300 OSMO
```
$ strided_old tx stakeibc liquid-stake 300000000 uosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:42:34 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:42:34 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:42:35 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 160689
{"height":"0","txhash":"256362526079C6ACB1A46454F6B23095C78A1AA691A1E35E703A280C0422DA87","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:44:17Z ready: hub delegated
2026-10-02T23:54:24Z TIMEOUT waiting for: osmo delegated
2026-10-02T23:55:30Z FINDING (harness/race): osmosis delegation ICA seq 1 sent before OpenConfirm; relayer refused recv, 36s stride-epoch timeout closed the channel. Restoring with restore-interchain-account.
2026-10-02T23:55:44Z tx ACAD50F60EA1829C8CD649C58E2502C9F1DF3787B738916692F95C2D29C46EFE code=0 
2026-10-02T23:56:27Z ## Seed (v34)
2026-10-02T23:56:27Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-02T23:56:43Z resume: HIST_TX=F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA,        8 hub validators,        3 osmosis validators
2026-10-02T23:56:49Z ready: osmo delegated
2026-10-02T23:56:49Z ### holder base
```
$ strided_old tx bank send user1 stride1ef2axra0mrwwqacf2l33ye62qzavgtwrypqmgs 50000000stuatom,10000000ustrd,20000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:56:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:56:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:56:51 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 133923
{"height":"0","txhash":"4D2474FB7224E1C2D869A5192A3187D4D6A38F7FBF07F0C79E7370747D0DEF81","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:00Z ### vesting acct
```
$ strided_old tx vesting create-vesting-account stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 1000000ustrd 1822521420 --delayed --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:01 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 94638
{"height":"0","txhash":"F23D160ECE9F5A7423CDB1385EF5FD7C5785DFA7B0A25A85ED19F59FBA63E018","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:09Z ### holder vesting
```
$ strided_old tx bank send user1 stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 30000000stuatom,5000000ustrd --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:10 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 97917
{"height":"0","txhash":"299B3361D406F70577910166DAD03D99ACA7C6A214CC3C9AF190BC22BEA89262","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:19Z ### distribution holds stATOM
```
$ strided_old tx distribution fund-community-pool 5000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:21 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 81744
{"height":"0","txhash":"ADFF0BC4656BD4371876E05B5F1524EF6155162EE93BE8F1451D239E8D0F5C86","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:29Z ### statom to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 100000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:31 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138193
{"height":"0","txhash":"77912577D6FD78C587C192E6081A3E0D7D0E6684B039DDCD682A34C51EA231CA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:40Z ### statom to hub
```
$ strided_old tx ibc-transfer transfer transfer channel-0 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l 60000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:41 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138294
{"height":"0","txhash":"2C98A68600EFC775DD9ACE7EDFB09A2758D0112CA3E9F98B3DB984F4A508B2F1","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:57:51Z ### stosmo to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 50000000stuosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/02 23:57:52 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 121983
{"height":"0","txhash":"1BF849B56AE1154F141079CBD40CC690F54DD3E9E0F53EF89795672CEDE7A993","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:58:06Z ready: statom on hub
2026-10-02T23:58:06Z ### statom hub->osmosis (two-hop)
```
$ gaiad tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 30000000ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 221128
{"height":"0","txhash":"534971A73E1D2B9FB1F00484BE0005A4296C834FB2E6265B29A7CF218EED2BDF","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:58:23Z ### fund hub fee ICA
```
$ gaiad tx bank send user1 cosmos1gnmd482qhaplvplxhw08w7x0thyraae8x2c5emewhs5pl86xr23q6yupdn 3000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131751
{"height":"0","txhash":"339A67C4BDFCFCB41F0D66F1D26E2FE0D362A73881E31F16C61704945A178C0E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:58:37Z ### fund hub withdrawal ICA
```
$ gaiad tx bank send user1 cosmos1tcpc9ew35pc944kw2etdfwgele9wrt8vcqxvqwrey9yu23avn30qs5fem8 4000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131755
{"height":"0","txhash":"1256A2B489B15F6BF33BCFFD4B0F0EAD18B559204907378D84ADA6D2DF073C4C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:58:52Z ### fund osmo fee ICA
```
$ osmosisd tx bank send user1 osmo1r7yph0aqz4efgc85kywvav2gjj0ctr8skvthc5esgfv7hd23rtpsjt43vp 3000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120492
{"height":"0","txhash":"6DD01931D725AF8884F89916C66CF0764F674BB1BE122162D4CC8AD9CAB942AC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-02T23:59:05Z ### fund osmo withdrawal ICA
```
$ osmosisd tx bank send user1 osmo1d5apuzu4p7c63a8mpaaw20yzhdgdctlq4g3z3rq2wseq78renpys6fh2r2 4000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120492
{"height":"0","txhash":"6E61E823C29105A3E8D6317FC3774E24036E40D0A70E8C555C1731FFDFF2E6E6","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
broadcast output (gaiad):
```
{"height":"0","txhash":"A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-02T23:59:48Z tx A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:00:31Z tx 0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:01:07Z tx CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:01:44Z tx 02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437 code=0 
2026-10-03T00:01:44Z ### multisign + broadcast transfer grant
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx sign /tmp/unsigned.json --from d1 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s1.json
  gaiad tx sign /tmp/unsigned.json --from d2 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s2.json
  gaiad tx multisign /tmp/unsigned.json hub-ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/signed.json
  gaiad tx broadcast /tmp/signed.json --chain-id cosmoshub-test-1 -o json
{"height":"0","txhash":"75666F8E7605777CE72F5BBCBA517D3BE18C0C45DC288833ABF099D6A3C28CCC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:02:24Z CHECKPOINT FAIL: transfer grant present
2026-10-03T00:03:24Z ## Seed (v34)
2026-10-03T00:03:49Z CHECKPOINT PASS: admin-ms address
2026-10-03T00:03:53Z CHECKPOINT PASS: vault-ms address
2026-10-03T00:04:00Z CHECKPOINT PASS: hub-ms address
2026-10-03T00:04:00Z ### fund admin-ms
```
$ strided_old tx bank send faucet stride1mymazvsd79f9yhjq4n84dchyf6zvfmd8ed2nxc 1000000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026-10-03T00:04:11Z ## Seed (v34)
2026-10-03T00:04:11Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-03T00:04:27Z resume: HIST_TX=F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA,        8 hub validators,        3 osmosis validators
2026-10-03T00:04:34Z ready: osmo delegated
2026-10-03T00:04:34Z SEED_RESUME=2: skipping the holders/transfers half of Part B
2026-10-03T00:04:39Z staketia grants already present: skipping the multisig delegate and grants (rerun)
2026-10-03T00:04:45Z CHECKPOINT PASS: transfer grant present
2026-10-03T00:04:45Z ### rate limit proposal
```
$ strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:04:46 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 205068
{"height":"0","txhash":"296C45AAC1E0888A471F74886EF88E64017D6F0AB64B86DC0A6B2E974EBF075F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:07:28Z ## Seed (v34)
2026-10-03T00:07:28Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-03T00:07:45Z resume: HIST_TX=F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA,        8 hub validators,        3 osmosis validators
2026-10-03T00:07:52Z ready: osmo delegated
2026-10-03T00:07:52Z SEED_RESUME=2: skipping the holders/transfers half of Part B
2026-10-03T00:07:58Z staketia grants already present: skipping the multisig delegate and grants (rerun)
2026-10-03T00:08:04Z CHECKPOINT PASS: transfer grant present
2026-10-03T00:08:05Z ### rate limit proposal
```
$ strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:08:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:08:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:08:06 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 204429
{"height":"0","txhash":"81CFA2FF428BB9E4D5FB66DB93FF3ABC33B220BC14856A8C5F4BF2CCAC3C5261","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:08:41Z tx F7E7FB8A391A53DB5F0177A36518CB535BC275E74459070B83B3611005A8A6AC code=11 out of gas in location: WritePerByte; gasWanted: 54894, gasUsed: 56271: out of gas
2026-10-03T00:08:41Z vote from val4 failed: {"height":"0","txhash":"F7E7FB8A391A53DB5F0177A36518CB535BC275E74459070B83B3611005A8A6AC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"
2026-10-03T00:08:42Z tx 81D89BEE6F39B871D53CDE627E88F8345BBCA625EC518E513DCBC0CB2A738395 code=11 out of gas in location: WritePerByte; gasWanted: 54894, gasUsed: 56271: out of gas
2026-10-03T00:08:42Z vote from val2 failed: {"height":"0","txhash":"81D89BEE6F39B871D53CDE627E88F8345BBCA625EC518E513DCBC0CB2A738395","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"
2026-10-03T00:08:46Z tx 968865A639084186ABCC5FB3BA8489853355C65CD0C64B209ED1CBD5894A9DBF code=11 out of gas in location: WritePerByte; gasWanted: 54894, gasUsed: 56271: out of gas
2026-10-03T00:08:46Z vote from val3 failed: {"height":"0","txhash":"968865A639084186ABCC5FB3BA8489853355C65CD0C64B209ED1CBD5894A9DBF","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"
2026-10-03T00:08:47Z tx EFBBC92E8DFA23531A6FB586E8857AFD0D7D0A542A5CD4505CFE868A5FFF6713 code=11 out of gas in location: WritePerByte; gasWanted: 54741, gasUsed: 56169: out of gas
2026-10-03T00:08:47Z vote from val1 failed: {"height":"0","txhash":"EFBBC92E8DFA23531A6FB586E8857AFD0D7D0A542A5CD4505CFE868A5FFF6713","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"
2026-10-03T00:10:52Z TIMEOUT waiting for: rate limit live
2026-10-03T00:11:46Z ## Seed (v34)
2026-10-03T00:11:46Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-03T00:12:17Z ## Seed (v34)
2026-10-03T00:12:17Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-03T00:12:40Z resume: HIST_TX=F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA,        8 hub validators,        3 osmosis validators
2026-10-03T00:12:47Z ready: osmo delegated
2026-10-03T00:12:47Z SEED_RESUME=2: skipping the holders/transfers half of Part B
2026-10-03T00:12:55Z staketia grants already present: skipping the multisig delegate and grants (rerun)
2026-10-03T00:13:03Z CHECKPOINT PASS: transfer grant present
2026-10-03T00:13:04Z ### rate limit proposal
```
$ strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:13:04 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 204429
{"height":"0","txhash":"2E35EACBE655B61DFAEE1CE2026451D5C376F56A78F602D14B47DA4AFFBF0EA4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:14:18Z tx 7FDAE316EC474788FD538519B2CAE3595EB352A7BE042C4ABDD93B000928B2BA code=0 
2026-10-03T00:14:18Z tx AE807711CC9C3F640861A65A8F279BFB204B5F2AE9B076217B2EC421A8A24009 code=0 
2026-10-03T00:14:19Z tx 2B5CF346DE40CAA2CA52619A2C645835B81541774C22D9842F3266A264348224 code=0 
2026-10-03T00:14:19Z tx A045DF6238CA42C71C504550A7C43B156C145A570373E318762511F52A60B7DA code=0 
2026-10-03T00:14:27Z ready: rate limit live
2026-10-03T00:14:36Z day epoch now=26: D0=1790986481 D1=1790986841 D2=1790987021 D3=1790987201 D4=1790987381 (staketia prepare epoch) upgrade target U=1790987531
2026-10-03T00:14:36Z ### hub RA redeem
```
$ strided_old tx stakeibc redeem-stake 30000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:14:37 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:14:37 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:14:38 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125562
{"height":"0","txhash":"1A2BFB06F4AF32AE011EA192192C9D81231A1461405B53F93B560B436BF6C9E5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:14:48Z ### osmo val3 weight 0
```
$ strided_old tx stakeibc change-validator-weight osmosis-test-1 osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4 0 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:14:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 117798
{"height":"0","txhash":"522F483B7F7365F61693E71527F379FBF299E4E8F7475B242E0B20DC6FCF8824","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:15:02Z ### stop signing osmosis-validator-2
```
$ pause_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T00:16:01Z ready: osmo val3 jailed
2026-10-03T00:16:01Z ### resume osmosis-validator-2
```
$ resume_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T00:16:01Z ### osmo RE redeem (retry)
```
$ strided_old tx stakeibc redeem-stake 120000000 osmosis-test-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:16:02 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 127254
{"height":"0","txhash":"3593E79AAC547218164F54D91416D994F63BFBCC3132463221205F648D86EB2C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:23:51Z ### hub RB redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:23:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:23:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:23:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125485
{"height":"0","txhash":"1825C8150B36829D48878D16F9EE968A17F70B7C4C3AA93CF445C3FF7DDC7338","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:26:51Z ### hub RC redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:26:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:26:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:26:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125485
{"height":"0","txhash":"6CCD9A5DCF83E35216F7A0AFBD153AAC182CA3ED6A7E6F1CA99B36F8D11BA0F1","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:27:03Z ### staketia R1
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:27:04 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 152445
{"height":"0","txhash":"C38160DD2528B4D9E14513B7D653143C92A3AFFC0BE101715FEDE2570592BD97","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:29:54Z ready: staketia unbonding record in UNBONDING_QUEUE
2026-10-03T00:30:10Z staketia record 28 native_amount=20294372
2026-10-03T00:30:10Z ### hub-ms undelegate via authz exec
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx staking unbond cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p 20294372uatom --from hub-ms --generate-only --keyring-backend test --chain-id cosmoshub-test-1 > /tmp/unbond.json
  gaiad tx authz exec /tmp/unbond.json --from st-operator --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 390205
{"height":"0","txhash":"F445821B04637A82D5F67F5026F633D3E495574920EAAC066ED1D6441201DB21","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:30:22Z CHECKPOINT PASS: hub undelegate tx hash captured
2026-10-03T00:30:28Z ### staketia confirm-undelegation
```
$ strided_old tx staketia confirm-undelegation 28 F445821B04637A82D5F67F5026F633D3E495574920EAAC066ED1D6441201DB21 --from st-operator --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:30:29 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:30:29 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:30:30 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 227749
{"height":"0","txhash":"05BEDD7051FB138609B83D92DF87578BF778272CA6A0F0E740625F4A9A48EFEB","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:30:38Z ### hub RD redeem (queue)
```
$ strided_old tx stakeibc redeem-stake 10000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:30:40 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125485
{"height":"0","txhash":"15B82C37A52057E72D86BA7DFCCD0075F10E6C38E8E83ADA18CCD384E78C6B44","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:30:48Z ### staketia R2
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:30:49 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:30:49 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:30:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 137302
{"height":"0","txhash":"144758047728F829CE9EBE18B017049136BD9DA5F0D161C304D7E32BCC8D1A2D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:30:58Z ### staketia R3 spillover
```
$ strided_old tx staketia redeem-stake 40000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:30:59 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 308310
{"height":"0","txhash":"19F327BA93A215FF2E2E1F45BC7EF57810160DC68F66097D7C6AE886A5FE9B85","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:31:13Z ready: osmo RE in UNBONDING_RETRY_QUEUE
2026-10-03T00:31:17Z CHECKPOINT PASS: hub RA CLAIMABLE
2026-10-03T00:31:25Z CHECKPOINT PASS: hub RB+RC EXIT_TRANSFER_QUEUE
2026-10-03T00:31:30Z CHECKPOINT PASS: hub RB+RC both EXIT_TRANSFER_QUEUE
2026-10-03T00:31:36Z CHECKPOINT PASS: hub RD UNBONDING_QUEUE
2026-10-03T00:31:48Z pre-upgrade rates: hub=1.015345809038518546 osmo=1.012052853692850543
2026-10-03T00:31:48Z CHECKPOINT PASS: state.env has HIST_TX
2026-10-03T00:31:48Z ### records at upgrade
```
$ strided_old q records list-epoch-unbonding-record -o json
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:31:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"17","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"18","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"19","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"20","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"21","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"22","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"23","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"24","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"25","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"26","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"27","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"30347305","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"30347305","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790986907385336141","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.27.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"120000000","native_token_amount":"121444411","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121444411","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.27.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"28","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"29","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"30","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20278671","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987447598221085","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.30.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"31","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20294372","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987627696177706","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.31.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"32","host_zone_unbondings":[{"st_token_amount":"40724984","native_token_amount":"41340684","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":["cosmoshub-test-1.32.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T00:31:55Z ### staketia records at upgrade
```
$ strided_old q staketia unbonding-records -o json
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:31:56 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"unbonding_records":[{"id":"28","status":"UNBONDING_IN_PROGRESS","st_token_amount":"20000000","native_amount":"20294372","unbonding_completion_time_seconds":"1790987675","undelegation_tx_hash":"F445821B04637A82D5F67F5026F633D3E495574920EAAC066ED1D6441201DB21","unbonded_token_sweep_tx_hash":""},{"id":"32","status":"ACCUMULATING_REDEMPTIONS","st_token_amount":"29275016","native_amount":"29717610","unbonding_completion_time_seconds":"0","undelegation_tx_hash":"","unbonded_token_sweep_tx_hash":""}]}
```
2026-10-03T00:31:59Z seed done at 1790987519; upgrade target U=1790987531 (now - U = -12s)
2026-10-03T00:32:10Z seed overran D4 by ~70s; shifting the upgrade target one day epoch: U=1790987531 -> 1790987711 (D5=1790987561). A fresh RD is redeemed after D5 so a queued record exists at the upgrade.
2026-10-03T00:33:04Z tx 872C0EE04716718BD43C43B4FC38B9C1C125703D8DE3597B73A34E1945DEEB3E code=0 
2026-10-03T00:33:04Z hub RD2 redeem (queue) after D5: 872C0EE04716718BD43C43B4FC38B9C1C125703D8DE3597B73A34E1945DEEB3E
2026-10-03T00:33:21Z ## Phase 1: upgrade to v35
2026-10-03T00:33:26Z upgrade height 4018 (now 3913, target time 1790987711)

Submitting proposal for v35 at height 4018...

code: 0
txhash: 7C280A297E2C726AF249A704C75DC6F36B73052E218EDEC9C28D4306C5D5C0EE

Proposal:

proposal:
  deposit_end_time: "2026-10-03T00:34:09.516865551Z"
  final_tally_result:
    abstain_count: "0"
    no_count: "0"
    no_with_veto_count: "0"
    yes_count: "0"
  id: "4"
  messages:
  - type: /cosmos.upgrade.v1beta1.MsgSoftwareUpgrade
    value:
      authority: stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
      plan:
        height: "4018"
        name: v35
        time: "0001-01-01T00:00:00Z"
  proposer: stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
  status: PROPOSAL_STATUS_VOTING_PERIOD
  submit_time: "2026-10-03T00:33:39.516865551Z"
  summary: Upgrade v35
  title: Upgrade v35
  total_deposit:
  - amount: "2000000000"
    denom: ustrd
  voting_end_time: "2026-10-03T00:34:09.516865551Z"
  voting_start_time: "2026-10-03T00:33:39.516865551Z"

Voting on proposal #4...

code: 0
txhash: 292CCA1E4B78C0D31EEB5225DBA59D5C7A84F8B115E8673BB50161B21DDB2BC6
code: 0
txhash: 8C5AA488FE67957EB3641C566612D9DBE46B121179305FAB608241E45CA4908C
code: 0
txhash: 7A508431E9A724005F0E63CEA9F26A1C6656DD658FF3DAB37B71795BDCF259E0
code: 0
txhash: 233A9156D65F7AC3E67F6D55100CF61950AD8C3C6DE14CFDE1929124DA3984EA

Vote confirmation:

tally:
  abstain_count: "0"
  no_count: "0"
  no_with_veto_count: "0"
  yes_count: "4000000000"

Proposal Status:

Proposal passed!
2026-10-03T00:34:51Z ### deposit in flight
```
$ strided_old tx stakeibc liquid-stake 10000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 00:34:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 00:34:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 00:34:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 162168
{"height":"0","txhash":"B01658820746BD0683F90F93C65E0B8FB75B232389268288DA3E2808C76C7F64","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T00:35:04Z CHECKPOINT PASS: hub deposit in TRANSFER_QUEUE
2026-10-03T00:36:25Z ready: v35 running
2026-10-03T00:36:25Z ### handler log lines
```
$ handler_log_lines
[90m12:34AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m4 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m12:35AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 4018: [0m [36mmodule=[0mx/upgrade
[90m12:35AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mstate
[90m12:35AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 506 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc003bd8ee0?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc002c76e08, 0xfb2, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270?, 0xc00c1016c0?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc002c76e08, {{0x7dccca0, 0xc004477778}, {0xc0027c4270, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc002c76e08, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 474\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m12:36AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m12:36AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 4018[0m [36mmodule=[0mx/upgrade
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m4018 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1gnmd482qhaplvplxhw08w7x0thyraae8x2c5emewhs5pl86xr23q6yupdn -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1r7yph0aqz4efgc85kywvav2gjj0ctr8skvthc5esgfv7hd23rtpsjt43vp -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos10vtu56ne6rvtcrwe8j2shuph933qzcgk6hhy6y65lzdcmkx5gyksqmdzkl removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos1f64067vtvvr3lrv9f4kdw43qdxcsw529h3fmw8pnm6f7u4kpgerqdrvxyq removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1u96zlqsfa4ehw2ua46z092nlqu498rfvmcfk9ql9p4x6kk7gxvzqzkcrrp removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1rhvcrvqsxkem7nuuw3zlq8duqldyk32p97s7fmfvsg83aczjtu7supceqw removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 2 unacked packet(s) on channel-2, skipping flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-14[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T00:36:27Z handler error lines: none
2026-10-03T00:36:27Z CHECKPOINT PASS: no handler error
broadcast output (strided_new):
```
{"height":"0","txhash":"8EEBBD69B3777E8EFD45DF302A4A3AF1E06FB1CCBBB52009F12CEBDDC22F9DA9","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:40:45Z tx 8EEBBD69B3777E8EFD45DF302A4A3AF1E06FB1CCBBB52009F12CEBDDC22F9DA9 not found on strided_new
2026-10-03T00:40:45Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T00:40:51Z CHECKPOINT FAIL: liquid-stake cannot route
broadcast output (strided_new):
```
{"height":"0","txhash":"8EEBBD69B3777E8EFD45DF302A4A3AF1E06FB1CCBBB52009F12CEBDDC22F9DA9","codespace":"sdk","code":19,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:47:58Z tx 8EEBBD69B3777E8EFD45DF302A4A3AF1E06FB1CCBBB52009F12CEBDDC22F9DA9 not found on strided_new
broadcast output (strided_new):
```
{"height":"0","txhash":"754CF10D5FF0B01761D4B5CB97E4279405B8AF3AFD7AA0A03D44E8EE59552E54","codespace":"sdk","code":32,"data":"","raw_log":"account sequence mismatch, expected 1, got 0: incorrect account sequence","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T00:53:08Z tx 754CF10D5FF0B01761D4B5CB97E4279405B8AF3AFD7AA0A03D44E8EE59552E54 not found on strided_new
2026-10-03T00:59:06Z FINDING (harness): after the v35 halt at 4018, stride-validator-1..3 (daemon restarted in-process by cosmovisor) came back on v35 with zero peers and never redialed val0 (whose container restarted); consensus stalled ~20 min until the three processes were killed and restarted. Multisig txs sat in the mempool meanwhile (looked like a multisig signing problem; it was not).
2026-10-03T00:59:06Z PHASE 1 manual: single-key admin drain was refused by the guard: 'validator ... has 1 delegation change(s) in progress' (RD2 undelegate ICAs in flight) -> CHECKPOINT PASS (by hand): drain refused on the hub
2026-10-03T00:59:47Z ## Phase 1: upgrade to v35
2026-10-03T00:59:53Z ready: v35 running
2026-10-03T00:59:53Z ### handler log lines
```
$ handler_log_lines
[90m12:34AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m4 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m12:35AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 4018: [0m [36mmodule=[0mx/upgrade
[90m12:35AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mstate
[90m12:35AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 506 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc003bd8ee0?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc002c76e08, 0xfb2, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270?, 0xc00c1016c0?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc002c76e08, {{0x7dccca0, 0xc004477778}, {0xc0027c4270, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc002c76e08, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 474\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m12:36AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m12:36AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 4018[0m [36mmodule=[0mx/upgrade
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m4018 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1gnmd482qhaplvplxhw08w7x0thyraae8x2c5emewhs5pl86xr23q6yupdn -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1r7yph0aqz4efgc85kywvav2gjj0ctr8skvthc5esgfv7hd23rtpsjt43vp -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos10vtu56ne6rvtcrwe8j2shuph933qzcgk6hhy6y65lzdcmkx5gyksqmdzkl removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos1f64067vtvvr3lrv9f4kdw43qdxcsw529h3fmw8pnm6f7u4kpgerqdrvxyq removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1u96zlqsfa4ehw2ua46z092nlqu498rfvmcfk9ql9p4x6kk7gxvzqzkcrrp removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1rhvcrvqsxkem7nuuw3zlq8duqldyk32p97s7fmfvsg83aczjtu7supceqw removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 2 unacked packet(s) on channel-2, skipping flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-14[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T00:59:56Z handler error lines: none
2026-10-03T00:59:56Z CHECKPOINT PASS: no handler error
broadcast output (strided_new):
```
{"height":"0","txhash":"DCC3E635B00753B71EB81C06B62AB50CB361064E4D24323B3D7AFEA24FFD2DE1","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:00:45Z tx DCC3E635B00753B71EB81C06B62AB50CB361064E4D24323B3D7AFEA24FFD2DE1 code=1563 failed to execute message; message index: 0: validator cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75 has 1 delegation change(s) in progress: invalid delegation changes in progress
2026-10-03T01:00:45Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T01:00:49Z CHECKPOINT PASS: liquid-stake cannot route
2026-10-03T01:00:54Z CHECKPOINT PASS: autopilot stakeibc off
2026-10-03T01:01:01Z CHECKPOINT PASS: rate limits removed
2026-10-03T01:01:07Z CHECKPOINT PASS: wasm upload gov-only
2026-10-03T01:01:14Z CHECKPOINT PASS: ica host allow-list trimmed
2026-10-03T01:01:18Z CHECKPOINT PASS: historical tx decodes
2026-10-03T01:01:25Z CHECKPOINT FAIL: hub rate frozen
2026-10-03T01:02:19Z RATE_HUB/RATE_OSMO corrected to the values at the upgrade height 4018 (seed-end snapshot was one v34 stride epoch stale): hub 1.016554547041740302 osmo 1.012057704535602316; both unchanged 25 min after the upgrade -> CHECKPOINT PASS: hub rate frozen / osmo rate frozen
2026-10-03T01:02:19Z ## Phase 1: upgrade to v35
2026-10-03T01:02:25Z ready: v35 running
2026-10-03T01:02:25Z ### handler log lines
```
$ handler_log_lines
[90m12:34AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m4 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m12:35AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 4018: [0m [36mmodule=[0mx/upgrade
[90m12:35AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mstate
[90m12:35AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 4018: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 506 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc003bd8ee0?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc002c76e08, 0xfb2)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc002c76e08, 0xfb2, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc002c76e08, 0xc001788a90, {0xc0027c4270?, 0xc00c1016c0?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc002c76e08, {{0x7dccca0, 0xc004477778}, {0xc0027c4270, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc002c76e08, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 474\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m12:36AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m12:36AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 4018[0m [36mmodule=[0mx/upgrade
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m4018 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m12:36AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1gnmd482qhaplvplxhw08w7x0thyraae8x2c5emewhs5pl86xr23q6yupdn -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1jhy2rwce86a9t8yuauqxz47mdkes67qtjlkhdsgteg9r2l4y89ps0dzj5j -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1egea5jh75lnazm8f8qshq2pxvt8yy8r62mq03u38uv6jyhw447uqdlsueh -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1r7yph0aqz4efgc85kywvav2gjj0ctr8skvthc5esgfv7hd23rtpsjt43vp -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos10vtu56ne6rvtcrwe8j2shuph933qzcgk6hhy6y65lzdcmkx5gyksqmdzkl removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos1f64067vtvvr3lrv9f4kdw43qdxcsw529h3fmw8pnm6f7u4kpgerqdrvxyq removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1u96zlqsfa4ehw2ua46z092nlqu498rfvmcfk9ql9p4x6kk7gxvzqzkcrrp removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1rhvcrvqsxkem7nuuw3zlq8duqldyk32p97s7fmfvsg83aczjtu7supceqw removed[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 2 unacked packet(s) on channel-2, skipping flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-14[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m12:36AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T01:02:28Z handler error lines: none
2026-10-03T01:02:28Z CHECKPOINT PASS: no handler error
2026-10-03T01:02:35Z hub drain-refusal probe skipped: D5 too close
2026-10-03T01:02:35Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T01:02:41Z CHECKPOINT PASS: liquid-stake cannot route
2026-10-03T01:02:46Z CHECKPOINT PASS: autopilot stakeibc off
2026-10-03T01:02:53Z CHECKPOINT PASS: rate limits removed
2026-10-03T01:03:01Z CHECKPOINT PASS: wasm upload gov-only
2026-10-03T01:03:08Z CHECKPOINT PASS: ica host allow-list trimmed
2026-10-03T01:03:16Z CHECKPOINT PASS: historical tx decodes
2026-10-03T01:03:23Z CHECKPOINT PASS: hub rate frozen
2026-10-03T01:03:29Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T01:03:29Z ### records after upgrade
```
$ strided_new q records list-epoch-unbonding-record -o json
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:03:30 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"17","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"18","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"19","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"20","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"21","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"22","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"23","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"24","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"25","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"26","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"27","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"30347305","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"30347305","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790986907385336141","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.27.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"120000000","native_token_amount":"121444411","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121444411","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.27.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"28","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"29","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"30","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20278671","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"20278671","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987447598221085","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.30.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"31","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20294372","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"20294372","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987627696177706","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.31.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"32","host_zone_unbondings":[{"st_token_amount":"40724984","native_token_amount":"41357921","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987807423293749","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.32.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"33","host_zone_unbondings":[{"st_token_amount":"10000000","native_token_amount":"10163658","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790987988072313161","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.33.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"34","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T01:03:50Z ### autopilot liquid-stake memo from Hub
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000uatom --memo {"autopilot":{"receiver":"stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn","stakeibc":{"action":"LiquidStake"}}} --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
Error: rpc error: code = Unknown desc = rpc error: code = Unknown desc = failed to execute message; message index: 0: cannot send packet using client (07-tendermint-0) with status Expired: client state is not active [cosmos/ibc-go/v10@v10.3.0/modules/core/04-channel/keeper/packet.go:60] with gas used: '104106': unknown request
Usage:
  gaiad tx ibc-transfer transfer [src-port] [src-channel] [receiver] [coin] [flags]

Examples:
gaiad tx ibc-transfer transfer [src-port] [src-channel] [receiver] [coin]

Flags:
      --absolute-timeouts               Timeout flags are used as absolute timeouts.
  -a, --account-number uint             The account number of the signing account (offline mode only)
      --aux                             Generate aux signer data instead of sending a tx
  -b, --broadcast-mode string           Transaction broadcasting mode (sync|async) (default "sync")
      --chain-id string                 The network chain ID
      --dry-run                         ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)
      --fee-granter string              Fee granter grants fees for the transaction
      --fee-payer string                Fee payer pays fees for the transaction instead of deducting from the signer
      --fees string                     Fees to pay along with transaction; eg: 10uatom
      --from string                     Name or address of private key with which to sign
      --gas string                      gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically. Note: "auto" option doesn't always report accurate results. Set a valid coin value to adjust the result. Can be used instead of "fees". (default 200000)
      --gas-adjustment float            adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
      --gas-prices string               Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)
      --generate-only                   Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase only accessed when providing a key name)
  -h, --help                            help for transfer
      --keyring-backend string          Select keyring's backend (os|file|kwallet|pass|test|memory) (default "os")
      --keyring-dir string              The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                          Use a connected Ledger device
      --memo string                     Memo to be sent along with the packet.
      --node string                     <host>:<port> to CometBFT rpc interface for this chain (default "tcp://localhost:26657")
      --note string                     Note to add a description to the transaction (previously --memo)
      --offline                         Offline mode (does not allow any online functionality)
  -o, --output string                   Output format (text|json) (default "json")
      --packet-timeout-height string    Packet timeout block height in the format {revision}-{height}. (default "0-0")
      --packet-timeout-timestamp uint   Packet timeout timestamp in nanoseconds from now. Default is 10 minutes. On IBC v1 protocol, either timeout timestamp or timeout height must be set. On IBC v2 protocol timeout timestamp must be set. (default 600000000000)
  -s, --sequence uint                   The sequence number of the signing account (offline mode only)
      --sign-mode string                Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature
      --timeout-duration duration       TimeoutDuration is the duration the transaction will be considered valid in the mempool. The transaction's unordered nonce will be set to the time of transaction creation + the duration value passed. If the transaction is still in the mempool, and the block time has passed the time of submission + TimeoutTimestamp, the transaction will be rejected.
      --timeout-height uint             DEPRECATED: Please use --timeout-duration instead. Set a block timeout height to prevent the tx from being committed past a certain height
      --tip string                      Tip is the amount that is going to be transferred to the fee payer on the target chain. This flag is only valid when used with --aux, and is ignored if the target chain didn't enable the TipDecorator
      --unordered                       Enable unordered transaction delivery; must be used in conjunction with --timeout-duration
  -y, --yes                             Skip tx broadcasting prompt confirmation

Global Flags:
      --home string         directory for config and data (default "/home/validator/.gaia")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>') (default "info")
      --log_no_color        Disable colored logs
      --trace               print out full stack trace on errors

command terminated with exit code 1
```
2026-10-03T01:05:07Z ## RUN 1 ENDED after phase 1: the post-upgrade consensus stall expired all four Stride<->host light clients (trusting period 204s). Restarting the network for RUN 2 with the harness fixes (peer redial cap, phase-1 self-heal).
2026-10-03T01:20:18Z ## Phase 0: pre-flight
2026-10-03T01:20:23Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-03T01:20:29Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-03T01:20:35Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-03T01:20:39Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-03T01:20:44Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-03T01:20:47Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-03T01:20:48Z CHECKPOINT PASS: REST stride reachable
2026-10-03T01:20:48Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-03T01:20:48Z CHECKPOINT PASS: REST osmosis reachable
2026-10-03T01:20:52Z CHECKPOINT FAIL: vault receives
2026-10-03T01:21:22Z ## Phase 0: pre-flight
2026-10-03T01:21:28Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-03T01:21:35Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-03T01:21:40Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-03T01:21:47Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-03T01:21:51Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-03T01:21:54Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-03T01:21:55Z CHECKPOINT PASS: REST stride reachable
2026-10-03T01:21:55Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-03T01:21:55Z CHECKPOINT PASS: REST osmosis reachable
2026-10-03T01:22:00Z CHECKPOINT FAIL: vault receives
2026-10-03T01:24:18Z tx  code= 
2026-10-03T01:24:34Z tx 171FC7CFBFED35E6D5C730201858E278A38CB30C98AED4A22A8F06BC1FCEA2A4 code=0 
2026-10-03T01:24:38Z ## Phase 0: pre-flight
2026-10-03T01:24:42Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-03T01:24:46Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-03T01:24:50Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-03T01:24:56Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-03T01:25:01Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-03T01:25:04Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-03T01:25:04Z CHECKPOINT PASS: REST stride reachable
2026-10-03T01:25:05Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-03T01:25:05Z CHECKPOINT PASS: REST osmosis reachable
2026-10-03T01:25:13Z tx C1B1036888CC3490C7EEEDA1004472F9B3D4B0DBF3BDAA53C91FD35A06C4553D code=0 
2026-10-03T01:25:13Z CHECKPOINT PASS: vault receives
broadcast output (osmosisd):
```
{"height":"0","txhash":"D6C46DEA89EC90D631D799D4EAB14B6CDEBD4542DCCC8F7B2813EFCD94AF4680","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:25:29Z tx D6C46DEA89EC90D631D799D4EAB14B6CDEBD4542DCCC8F7B2813EFCD94AF4680 code=0 
2026-10-03T01:25:29Z CHECKPOINT PASS: vault spends (multisig)
2026-10-03T01:25:43Z tx 992AC42C19F40244FAD2213D78E0CB4075E7DCF841C06A0A9C0FA9F693C54C0A code=0 
2026-10-03T01:25:43Z CHECKPOINT PASS: sweep operator spends
2026-10-03T01:25:48Z host zones not seeded yet: skipping the withdraw-address check
2026-10-03T01:26:02Z gaia v25.1.0, osmosis 28.0.0, strided 
2026-10-03T01:26:11Z ## Seed (v34)
2026-10-03T01:26:29Z CHECKPOINT PASS: admin-ms address
2026-10-03T01:26:31Z CHECKPOINT PASS: vault-ms address
2026-10-03T01:26:38Z CHECKPOINT PASS: hub-ms address
2026-10-03T01:26:38Z ### fund admin-ms
```
$ strided_old tx bank send faucet stride1mymazvsd79f9yhjq4n84dchyf6zvfmd8ed2nxc 1000000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:26:39 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 110745
{"height":"0","txhash":"87C1CED8EB389832C5CA49A909E674E44FC887FAD533AA2AF4A8276CDF036117","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:26:46Z ### fund sweep operator
```
$ strided_old tx bank send faucet stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy 100000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:26:47 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 75082
{"height":"0","txhash":"F2AED251C5184F75E6CFCC8A8E530E5ECB85DA40479F115EEF566AC20602798E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:26:55Z ### fund vault-ms
```
$ osmosisd tx bank send faucet osmo1mymazvsd79f9yhjq4n84dchyf6zvfmd8jaelyx 100000000uosmo --from faucet --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 138232
{"height":"0","txhash":"BC0726D1ADA6F4095E7DF1AA9DBBFA80B115CC14DAA274F7EF4F580C9F646F4F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:27:01Z ### fund hub-ms
```
$ gaiad tx bank send faucet cosmos1h0dup2qw23uhgn9nxyhyze4cxzrgu8rtrcnv7d 100000000uatom --from faucet --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 171159
{"height":"0","txhash":"22B4410A0E7DF39B94B033B37DB70682D2C364ACB053B2CE252C95A176195DEA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:27:20Z CHECKPOINT PASS: 8 hub validators
2026-10-03T01:27:20Z CHECKPOINT PASS: 3 osmosis validators
2026-10-03T01:27:20Z ### register hub zone
```
$ strided_old tx stakeibc register-host-zone connection-0 uatom cosmos ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 channel-0 1 false --max-messages-per-ica-tx 3 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:27:21 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 663705
{"height":"0","txhash":"4CB0767EF4A481F63FAC5BD4C2675F4178839ECF9278E9B96CEF88884E99BDE4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:27:30Z ### register osmo zone
```
$ strided_old tx stakeibc register-host-zone connection-1 uosmo osmo ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B channel-1 1 false --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:27:31 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 651189
{"height":"0","txhash":"7B412777F74D7C070398E4BB1B36DCCCD1BE10A40DDE4B8C4A1FFF27E659CDFE","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:27:41Z ### add hub validators
```
$ strided_old tx stakeibc add-validators cosmoshub-test-1 /tmp/hub_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:27:42 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 1098741
{"height":"0","txhash":"A516142CC6DAD1069F3EEC4187A72815914FE429F8B407C6E34EC082785F0541","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:27:49Z ### add osmo validators
```
$ strided_old tx stakeibc add-validators osmosis-test-1 /tmp/osmo_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:27:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 402217
{"height":"0","txhash":"5A393AD106E75789EBA9A036548BE1463971129A9766DDB391372E1EBA0141DC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:28:04Z ready: cosmoshub-test-1 delegation_ica_address
2026-10-03T01:28:09Z ready: cosmoshub-test-1 fee_ica_address
2026-10-03T01:28:15Z ready: cosmoshub-test-1 withdrawal_ica_address
2026-10-03T01:28:20Z ready: cosmoshub-test-1 redemption_ica_address
2026-10-03T01:28:26Z ready: osmosis-test-1 delegation_ica_address
2026-10-03T01:28:31Z ready: osmosis-test-1 fee_ica_address
2026-10-03T01:28:37Z ready: osmosis-test-1 withdrawal_ica_address
2026-10-03T01:28:43Z ready: osmosis-test-1 redemption_ica_address
2026-10-03T01:28:43Z ### atom to stride
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 2000000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 223377
{"height":"0","txhash":"2D6096178F5F640ACB623186A24993030BA05FC71E784B32BCF756D884464356","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:28:51Z ### osmo to stride
```
$ osmosisd tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 185025
{"height":"0","txhash":"CA482F3D93672693B09F181B420880F7D2EA2721F1AB029295034E236E170B9B","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:29:02Z ready: atom on stride
2026-10-03T01:29:07Z ready: osmo on stride
2026-10-03T01:29:07Z ### liquid stake 1000 ATOM
```
$ strided_old tx stakeibc liquid-stake 1000000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:29:08 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 198025
{"height":"0","txhash":"F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:29:15Z ### liquid stake 300 OSMO
```
$ strided_old tx stakeibc liquid-stake 300000000 uosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:29:16 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 160689
{"height":"0","txhash":"256362526079C6ACB1A46454F6B23095C78A1AA691A1E35E703A280C0422DA87","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:30:20Z ready: hub delegated
2026-10-03T01:30:26Z ready: osmo delegated
2026-10-03T01:30:26Z ### holder base
```
$ strided_old tx bank send user1 stride1ef2axra0mrwwqacf2l33ye62qzavgtwrypqmgs 50000000stuatom,10000000ustrd,20000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:30:27 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 133923
{"height":"0","txhash":"4D2474FB7224E1C2D869A5192A3187D4D6A38F7FBF07F0C79E7370747D0DEF81","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:30:36Z ### vesting acct
```
$ strided_old tx vesting create-vesting-account stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 1000000ustrd 1822527036 --delayed --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:30:36 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:30:36 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:30:37 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 94638
{"height":"0","txhash":"FCC5BD51A282CEE33C2D2D50B90828F1EAA37F695ED08AFB2CDDE3C3DE1B1608","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:30:45Z ### holder vesting
```
$ strided_old tx bank send user1 stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 30000000stuatom,5000000ustrd --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:30:46 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 97917
{"height":"0","txhash":"299B3361D406F70577910166DAD03D99ACA7C6A214CC3C9AF190BC22BEA89262","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:30:53Z ### distribution holds stATOM
```
$ strided_old tx distribution fund-community-pool 5000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:30:54 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:30:55 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 81694
{"height":"0","txhash":"66E35BB64578B53B10A3DF2FE9E79E628A6A63B139A6C9D8437EE5FE0C594A02","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:31:02Z ### statom to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 100000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:31:03 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138193
{"height":"0","txhash":"8E05346900ADC03FBD0C5A0727EF729FDC553BDB720C29A37B499D43D471583F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:31:11Z ### statom to hub
```
$ strided_old tx ibc-transfer transfer transfer channel-0 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l 60000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:31:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:31:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:31:13 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138240
{"height":"0","txhash":"9CFFF0B018A4128CF172F385BBFB57C91868A86ED66206660BCFF74722BCC41C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:31:21Z ### stosmo to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 50000000stuosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:31:22 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 121965
{"height":"0","txhash":"17DC229DB0974A0F4E02BE515798BD51B1EA0DA890944AAB09825B9DF4D6C226","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:31:35Z ready: statom on hub
2026-10-03T01:31:35Z ### statom hub->osmosis (two-hop)
```
$ gaiad tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 30000000ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 221119
{"height":"0","txhash":"F4C7D2F8A9056AB9D164C97D7FC35C720CCE9EB8FD405C13A57105C9A4D1B62B","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:31:50Z ### fund hub fee ICA
```
$ gaiad tx bank send user1 cosmos1r7satedqtx7qgxah809t2c4cgz0j3w7cswccleg7p8y3fdg5zpgsh5myw7 3000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131728
{"height":"0","txhash":"4F367E96C26FD1879225C4F2EA8E0C4AFD8F91E5B1723F7E25573B0C0A84F084","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:32:08Z ### fund hub withdrawal ICA
```
$ gaiad tx bank send user1 cosmos147k5g8sl63ju3hg7afzhpraqq8awkk80awptt3tc6aaww9nh6hps0g0huw 4000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131755
{"height":"0","txhash":"E02C203A4200A7057771DB8ED2F95910A224231AEED93EEB642BC010A360F5B2","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:32:22Z ### fund osmo fee ICA
```
$ osmosisd tx bank send user1 osmo1t5n089p0k0rfv3p2yr9g5vtzjr855mgqc6zch7tfgace64phr0jspff9yq 3000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120501
{"height":"0","txhash":"0117D01A030B3A9B52B6AAE36A2F26E16DA8B542CDAF5DC9A1B7030A5049A301","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:32:35Z ### fund osmo withdrawal ICA
```
$ osmosisd tx bank send user1 osmo120qfrhv0vd60mldt22qyppm4tuunf4jhttfjdyppx8zade54456qf8e80q 4000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120492
{"height":"0","txhash":"A8D8B6632BAF3ACA6B37181C497ACEA770296A3589072E60BA71A3B62498CF7A","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
broadcast output (gaiad):
```
{"height":"0","txhash":"A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:33:24Z tx A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:33:59Z tx 0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:34:32Z tx CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T01:35:05Z tx 02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437 code=0 
2026-10-03T01:35:06Z ### multisign + broadcast transfer grant
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx sign /tmp/unsigned.json --from d1 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s1.json
  gaiad tx sign /tmp/unsigned.json --from d2 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s2.json
  gaiad tx multisign /tmp/unsigned.json hub-ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/signed.json
  gaiad tx broadcast /tmp/signed.json --chain-id cosmoshub-test-1 -o json
{"height":"0","txhash":"75666F8E7605777CE72F5BBCBA517D3BE18C0C45DC288833ABF099D6A3C28CCC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:35:39Z CHECKPOINT PASS: transfer grant present
2026-10-03T01:35:40Z ### rate limit proposal
```
$ strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:35:41 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 205068
{"height":"0","txhash":"296C45AAC1E0888A471F74886EF88E64017D6F0AB64B86DC0A6B2E974EBF075F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:36:16Z tx C48801B75D5848CB36A00083E16815E6EA074EF995F405F2531C86B69547C49E code=0 
2026-10-03T01:36:21Z tx 030A8072954A2BCEF91B9E36673E0D665105C91ACF77F737300E06926FC906C6 code=0 
2026-10-03T01:36:23Z tx C48452FA0CC5EE2A3A704EC10C61EBDAACA774377B3C56148E806BBEA220897F code=0 
2026-10-03T01:36:24Z tx CA0009CA05B0C65469AE083B09C5E73358866D011096617CE091AA5621D4D4A1 code=0 
2026-10-03T01:36:31Z ready: rate limit live
2026-10-03T01:36:37Z day epoch now=11: D0=1790991548 D1=1790991728 D2=1790991908 D3=1790992088 D4=1790992268 (staketia prepare epoch) upgrade target U=1790992418
2026-10-03T01:36:37Z ### hub RA redeem
```
$ strided_old tx stakeibc redeem-stake 30000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:36:37 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:36:37 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:36:38 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125490
{"height":"0","txhash":"C9117889AD40E87E45CFFA14BFA6910F56DEEAF775C9A1B5B65CEA524F993284","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:36:45Z ### osmo val3 weight 0
```
$ strided_old tx stakeibc change-validator-weight osmosis-test-1 osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4 0 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:36:45 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:36:45 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:36:46 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 117798
{"height":"0","txhash":"A99AAAC32E7A51F97D243811D15CA50C0B9F437DE22FF865D20D042CE789E01A","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:36:53Z ### stop signing osmosis-validator-2
```
$ pause_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T01:37:44Z ready: osmo val3 jailed
2026-10-03T01:37:44Z ### resume osmosis-validator-2
```
$ resume_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T01:37:45Z ### osmo RE redeem (retry)
```
$ strided_old tx stakeibc redeem-stake 120000000 osmosis-test-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:37:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:37:46 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:37:47 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 127272
{"height":"0","txhash":"55A2F5B9509732589B830B68AC62FC12038C5EFF10E589652EA2E855E41ED303","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:45:18Z ### hub RB redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:45:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125485
{"height":"0","txhash":"1825C8150B36829D48878D16F9EE968A17F70B7C4C3AA93CF445C3FF7DDC7338","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:48:18Z ### hub RC redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:48:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:48:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:48:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125485
{"height":"0","txhash":"6CCD9A5DCF83E35216F7A0AFBD153AAC182CA3ED6A7E6F1CA99B36F8D11BA0F1","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:48:30Z ### staketia R1
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:48:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:48:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:48:31 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 152445
{"height":"0","txhash":"C38160DD2528B4D9E14513B7D653143C92A3AFFC0BE101715FEDE2570592BD97","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:51:21Z ready: staketia unbonding record in UNBONDING_QUEUE
2026-10-03T01:51:34Z staketia record 12 native_amount=20169539
2026-10-03T01:51:34Z ### hub-ms undelegate via authz exec
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx staking unbond cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p 20169539uatom --from hub-ms --generate-only --keyring-backend test --chain-id cosmoshub-test-1 > /tmp/unbond.json
  gaiad tx authz exec /tmp/unbond.json --from st-operator --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 390025
{"height":"0","txhash":"55C8108DF6645AC414FEDFBD60C517C78001BF6F445C6F522398C86BB4F7429E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:51:44Z CHECKPOINT PASS: hub undelegate tx hash captured
2026-10-03T01:51:50Z ### staketia confirm-undelegation
```
$ strided_old tx staketia confirm-undelegation 12 55C8108DF6645AC414FEDFBD60C517C78001BF6F445C6F522398C86BB4F7429E --from st-operator --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:51:51 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 226957
{"height":"0","txhash":"C241419DCEA4A50619580FC07236DD51D8851CF558C88BDD1DAD3C82D7721962","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:51:58Z ### hub RD redeem (queue)
```
$ strided_old tx stakeibc redeem-stake 10000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:51:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:51:59 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:52:00 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125530
{"height":"0","txhash":"C6517E7B702CCC98D31DEC98A9F9D5FA85EF6896A55952B8034CCE25FB751149","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:52:09Z ### staketia R2
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:52:10 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 137302
{"height":"0","txhash":"144758047728F829CE9EBE18B017049136BD9DA5F0D161C304D7E32BCC8D1A2D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:52:20Z ### staketia R3 spillover
```
$ strided_old tx staketia redeem-stake 40000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:52:21 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:52:21 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:52:22 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 308310
{"height":"0","txhash":"19F327BA93A215FF2E2E1F45BC7EF57810160DC68F66097D7C6AE886A5FE9B85","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:52:36Z ready: osmo RE in UNBONDING_RETRY_QUEUE
2026-10-03T01:52:42Z CHECKPOINT PASS: hub RA CLAIMABLE
2026-10-03T01:52:48Z CHECKPOINT PASS: hub RB+RC EXIT_TRANSFER_QUEUE
2026-10-03T01:52:53Z CHECKPOINT FAIL: hub RB+RC both EXIT_TRANSFER_QUEUE
2026-10-03T01:52:59Z CHECKPOINT PASS: hub RD UNBONDING_QUEUE
2026-10-03T01:53:13Z pre-upgrade rates: hub=1.009096150729838263 osmo=1.012011324650629984
2026-10-03T01:53:13Z CHECKPOINT PASS: state.env has HIST_TX
2026-10-03T01:53:13Z ### records at upgrade
```
$ strided_old q records list-epoch-unbonding-record -o json
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:53:14 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"8","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"9","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"10","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"11","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"30159427","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"30159427","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790991794583417186","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.11.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"120000000","native_token_amount":"121439456","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121439456","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.11.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"12","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"13","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"14","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20153593","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"20153593","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992333352600755","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.14.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20169539","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992514850198583","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.15.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"40420027","native_token_amount":"40778792","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":["cosmoshub-test-1.16.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T01:53:20Z ### staketia records at upgrade
```
$ strided_old q staketia unbonding-records -o json
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:53:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"unbonding_records":[{"id":"12","status":"UNBONDING_IN_PROGRESS","st_token_amount":"20000000","native_amount":"20169539","unbonding_completion_time_seconds":"1790992554","undelegation_tx_hash":"55C8108DF6645AC414FEDFBD60C517C78001BF6F445C6F522398C86BB4F7429E","unbonded_token_sweep_tx_hash":""},{"id":"16","status":"ACCUMULATING_REDEMPTIONS","st_token_amount":"29579973","native_amount":"29842521","unbonding_completion_time_seconds":"0","undelegation_tx_hash":"","unbonded_token_sweep_tx_hash":""}]}
```
2026-10-03T01:53:24Z seed done at 1790992404; upgrade target U=1790992418 (now - U = -14s)
2026-10-03T01:53:25Z ## Phase 1: upgrade to v35
2026-10-03T01:53:33Z ABORT: only 5s to U (need >= 60s for the proposal to pass); pushing the height out would cross D5. Re-seed.
2026-10-03T01:53:55Z seed ended 13s before U; shifting the upgrade target one day epoch: U=1790992418 -> 1790992598 (D5=1790992448); a fresh RD is redeemed after D5
2026-10-03T01:54:28Z tx CBAEB54F9438ECC0748C0AA6E7F0712421C21B7C7C0CCCD9EE3E37EBC3513B1C code=0 
2026-10-03T01:54:28Z hub RD2 redeem (queue) after D5: CBAEB54F9438ECC0748C0AA6E7F0712421C21B7C7C0CCCD9EE3E37EBC3513B1C
2026-10-03T01:54:28Z ## Phase 1: upgrade to v35
2026-10-03T01:54:36Z upgrade height 2081 (now 1959, target time 1790992598)

Submitting proposal for v35 at height 2081...

code: 0
txhash: 17A477008FB4C257B9664701B75228D229F459F48539C6A8391D0105A350071A

Proposal:

proposal:
  deposit_end_time: "2026-10-03T01:55:19.297307119Z"
  final_tally_result:
    abstain_count: "0"
    no_count: "0"
    no_with_veto_count: "0"
    yes_count: "0"
  id: "2"
  messages:
  - type: /cosmos.upgrade.v1beta1.MsgSoftwareUpgrade
    value:
      authority: stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
      plan:
        height: "2081"
        name: v35
        time: "0001-01-01T00:00:00Z"
  proposer: stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
  status: PROPOSAL_STATUS_VOTING_PERIOD
  submit_time: "2026-10-03T01:54:49.297307119Z"
  summary: Upgrade v35
  title: Upgrade v35
  total_deposit:
  - amount: "2000000000"
    denom: ustrd
  voting_end_time: "2026-10-03T01:55:19.297307119Z"
  voting_start_time: "2026-10-03T01:54:49.297307119Z"

Voting on proposal #2...

code: 0
txhash: 19C7A8442FC29CF8AE1FAAE0350D5C747920DE4D6287AF925E8446198A780918
code: 0
txhash: 38C621318C31BACE5FA2350341595FAD03BE35E2A19F35F9EBBDD67D5DD81CC1
code: 0
txhash: 736F281539A971E902B49D86361B1E3622B963E4F5827A50B308A76728382840
code: 0
txhash: 618A61BCAEA228FB3A48343994B97D76CE74696821BFCD4C9BBFAF2BA975274D

Vote confirmation:

tally:
  abstain_count: "0"
  no_count: "0"
  no_with_veto_count: "0"
  yes_count: "4000000000"

Proposal Status:

Proposal passed!
2026-10-03T01:56:18Z ### deposit in flight
```
$ strided_old tx stakeibc liquid-stake 10000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 01:56:18 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 01:56:18 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 01:56:19 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 162474
{"height":"0","txhash":"B460740DEF9D5475F7E5C499A7D1871F2CA0E1C5478FE0E7DE8731686D8C53BD","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T01:56:31Z CHECKPOINT PASS: hub deposit in TRANSFER_QUEUE
2026-10-03T01:57:55Z ready: v35 running
2026-10-03T01:59:35Z FINDING (harness): chain stalled at 2081 after the upgrade; restarting stride-validator-1..3 processes
2026-10-03T01:59:55Z ready: chain advancing after restart
2026-10-03T01:59:55Z ### handler log lines
```
$ handler_log_lines
[90m1:55AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m2 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m1:57AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 2081: [0m [36mmodule=[0mx/upgrade
[90m1:57AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 2081: " [36mmodule=[0mstate
[90m1:57AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 2081: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 497 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc00e8f9c10?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc0035a5188, 0x821)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc0035a5188, 0x821)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc0035a5188, 0x821, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc0035a5188, 0xc00d1ec820, {0xc00637eae0, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc0035a5188, 0xc00d1ec820, {0xc00637eae0?, 0xc00a9433c0?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc0035a5188, {{0x7dccca0, 0xc002093698}, {0xc00637eae0, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc0035a5188, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 484\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m1:57AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m1:57AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 2081[0m [36mmodule=[0mx/upgrade
[90m1:57AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m2081 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m1:57AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1pk4d832aes52mzhyrw35ekf3euqz2lwylkpshsywytc6pmu9yfpsxvdu9e -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1pk4d832aes52mzhyrw35ekf3euqz2lwylkpshsywytc6pmu9yfpsxvdu9e -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1r7satedqtx7qgxah809t2c4cgz0j3w7cswccleg7p8y3fdg5zpgsh5myw7 -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1namcl5x5wwapn2sma975el9w6jmsn6s3ml8cwf2hnjxxey7kn7tq9p99yt -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1namcl5x5wwapn2sma975el9w6jmsn6s3ml8cwf2hnjxxey7kn7tq9p99yt -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1t5n089p0k0rfv3p2yr9g5vtzjr855mgqc6zch7tfgace64phr0jspff9yq -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos19sfkjww7xqtlsly3lsxkvjepfy3xvkxzfcvlzn9x2vf83tptsrfswgd46v removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos1nqzn0039jjfsrg0ad3ylmnktj3vv0vcc6yu0qtuwzdzar6l3w5tsgx4uj5 removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1w6qtw2hhnexszeme7sw0vdj2m6zum0saynlrcws8w0vtanfa0p5q6qsk7l removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 0 stale in-progress flag(s) reset on channel-2[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-8[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T02:00:03Z handler error lines: none
2026-10-03T02:00:03Z CHECKPOINT PASS: no handler error
2026-10-03T02:00:10Z hub drain-refusal probe skipped: D5 too close
2026-10-03T02:00:10Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T02:00:17Z CHECKPOINT PASS: liquid-stake cannot route
2026-10-03T02:00:25Z CHECKPOINT PASS: autopilot stakeibc off
2026-10-03T02:00:34Z CHECKPOINT PASS: rate limits removed
2026-10-03T02:00:41Z CHECKPOINT PASS: wasm upload gov-only
2026-10-03T02:00:48Z CHECKPOINT PASS: ica host allow-list trimmed
2026-10-03T02:00:54Z CHECKPOINT PASS: historical tx decodes
2026-10-03T02:01:00Z CHECKPOINT FAIL: hub rate frozen
2026-10-03T02:01:35Z RUN 2: RATE_HUB/RATE_OSMO set to the values at upgrade height 2081 (hub 1.010309872498677576, osmo 1.012017369290889547); unchanged since -> rate frozen PASS. The seed-end snapshot is one v34 stride epoch stale by construction (script finding: capture the rate at the upgrade height, not at seed end).
2026-10-03T02:01:42Z ## Phase 1: upgrade to v35
2026-10-03T02:01:48Z ready: v35 running
2026-10-03T02:03:30Z ### handler log lines
```
$ handler_log_lines
[90m1:55AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m2 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m1:57AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 2081: [0m [36mmodule=[0mx/upgrade
[90m1:57AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 2081: " [36mmodule=[0mstate
[90m1:57AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 2081: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 497 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc00e8f9c10?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc0035a5188, 0x821)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc0035a5188, 0x821)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc0035a5188, 0x821, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc0035a5188, 0xc00d1ec820, {0xc00637eae0, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc0035a5188, 0xc00d1ec820, {0xc00637eae0?, 0xc00a9433c0?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc0035a5188, {{0x7dccca0, 0xc002093698}, {0xc00637eae0, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc0035a5188, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 484\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m1:57AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m1:57AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 2081[0m [36mmodule=[0mx/upgrade
[90m1:57AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m2081 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m1:57AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1pk4d832aes52mzhyrw35ekf3euqz2lwylkpshsywytc6pmu9yfpsxvdu9e -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1pk4d832aes52mzhyrw35ekf3euqz2lwylkpshsywytc6pmu9yfpsxvdu9e -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1r7satedqtx7qgxah809t2c4cgz0j3w7cswccleg7p8y3fdg5zpgsh5myw7 -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1namcl5x5wwapn2sma975el9w6jmsn6s3ml8cwf2hnjxxey7kn7tq9p99yt -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1namcl5x5wwapn2sma975el9w6jmsn6s3ml8cwf2hnjxxey7kn7tq9p99yt -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1t5n089p0k0rfv3p2yr9g5vtzjr855mgqc6zch7tfgace64phr0jspff9yq -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos19sfkjww7xqtlsly3lsxkvjepfy3xvkxzfcvlzn9x2vf83tptsrfswgd46v removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos1nqzn0039jjfsrg0ad3ylmnktj3vv0vcc6yu0qtuwzdzar6l3w5tsgx4uj5 removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1w6qtw2hhnexszeme7sw0vdj2m6zum0saynlrcws8w0vtanfa0p5q6qsk7l removed[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 0 stale in-progress flag(s) reset on channel-2[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-8[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m1:57AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T02:03:34Z handler error lines: none
2026-10-03T02:03:34Z CHECKPOINT PASS: no handler error
broadcast output (strided_new):
```
{"height":"0","txhash":"2292DA6B81A0EAFB3CC790BD35EA6A25BB546178B0B403F31EB8C6CAB3A6AE95","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:04:22Z tx 2292DA6B81A0EAFB3CC790BD35EA6A25BB546178B0B403F31EB8C6CAB3A6AE95 code=0 
2026-10-03T02:04:22Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T02:04:29Z CHECKPOINT PASS: liquid-stake cannot route
2026-10-03T02:04:34Z CHECKPOINT PASS: autopilot stakeibc off
2026-10-03T02:04:40Z CHECKPOINT PASS: rate limits removed
2026-10-03T02:04:44Z CHECKPOINT PASS: wasm upload gov-only
2026-10-03T02:04:50Z CHECKPOINT PASS: ica host allow-list trimmed
2026-10-03T02:04:57Z CHECKPOINT PASS: historical tx decodes
2026-10-03T02:05:02Z CHECKPOINT PASS: hub rate frozen
2026-10-03T02:05:08Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T02:05:08Z ### records after upgrade
```
$ strided_new q records list-epoch-unbonding-record -o json
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 02:05:09 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"8","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"9","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"10","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"11","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"30159427","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"30159427","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790991794583417186","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.11.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"120000000","native_token_amount":"121439456","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121439456","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.11.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"12","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"13","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"14","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20153593","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"20153593","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992333352600755","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.14.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"20169539","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"20169539","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992514850198583","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.15.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"40420027","native_token_amount":"40795510","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"40795510","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992694111633436","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.16.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"17","host_zone_unbondings":[{"st_token_amount":"10000000","native_token_amount":"10101176","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"10101176","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1790992874320177648","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.17.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"18","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T02:05:24Z ### autopilot liquid-stake memo from Hub
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000uatom --memo {"autopilot":{"receiver":"stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn","stakeibc":{"action":"LiquidStake"}}} --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 184306
{"height":"0","txhash":"EA9ADE1103147C332604F37D757CB7138180A00F71C824650F4BBE64F047812C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T02:06:37Z CHECKPOINT PASS: autopilot route refused (no stATOM minted, ATOM refunded)
2026-10-03T02:06:37Z ### register ICA on Stride from Hub
```
$ gaiad tx interchain-accounts controller register connection-0 --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
command terminated with exit code 137
```
2026-10-03T02:08:50Z TIMEOUT waiting for: hub-controlled ICA open on Stride
2026-10-03T02:08:50Z CHECKPOINT FAIL: ICA host route (channel did not open; not testable)
2026-10-03T02:08:50Z phase 1 done
2026-10-03T02:09:00Z ## Phase 2: day 0
broadcast output (strided_new):
```
{"height":"0","txhash":"D5817F74FA671AD002BD75966950AF151943A94C46FF65A7E59CF4078B669AAB","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:09:48Z tx D5817F74FA671AD002BD75966950AF151943A94C46FF65A7E59CF4078B669AAB code=1568 failed to execute message; message index: 0: epoch 11 record for osmosis-test-1 is UNBONDING_RETRY_QUEUE with 121439456; wait for the day epoch to submit it before draining: host zone has an unbonding record queued or retrying
2026-10-03T02:09:48Z CHECKPOINT PASS: drain refused while retry record (osmo)
2026-10-03T02:09:55Z CHECKPOINT PASS: non-admin refresh rejected
2026-10-03T02:10:00Z osmo val3 recorded delegation before refresh: 101370870
broadcast output (strided_new):
```
{"height":"0","txhash":"2F784D69A8F0D36A2ADE3ADA2C125B906C6F071C692A6EC6A666E1EF8371270D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:10:33Z tx 2F784D69A8F0D36A2ADE3ADA2C125B906C6F071C692A6EC6A666E1EF8371270D code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"361D5725327E01EBDEF4B1ACDE8FE41C12BA8F4FA46AA8FBC511A26A1E99B280","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:11:05Z tx 361D5725327E01EBDEF4B1ACDE8FE41C12BA8F4FA46AA8FBC511A26A1E99B280 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"1A03328F2E19F2BE43D3E56DA2EF736B9F1D9AFB5A6046C48119983A7C15025E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:11:33Z tx 1A03328F2E19F2BE43D3E56DA2EF736B9F1D9AFB5A6046C48119983A7C15025E code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"FD906889AF00141F815B3FF17B7E057A7FC8E8668C74F491A15E9EDA0CC41935","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:12:04Z tx FD906889AF00141F815B3FF17B7E057A7FC8E8668C74F491A15E9EDA0CC41935 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"0DF62EB1AAE35D8FEA9C9A14000893CC2885BC6C6452DF77F9C34A059D0714E9","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:12:38Z tx 0DF62EB1AAE35D8FEA9C9A14000893CC2885BC6C6452DF77F9C34A059D0714E9 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"04861ECBDB63D99DAA41BEBAF5DC75BB8980F251A861F04C91B020AFA4F56C55","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:13:11Z tx 04861ECBDB63D99DAA41BEBAF5DC75BB8980F251A861F04C91B020AFA4F56C55 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"7F6570A2FE01F1277627C390F60777A1B11F129A6854F16C6DC022331D398522","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:13:44Z tx 7F6570A2FE01F1277627C390F60777A1B11F129A6854F16C6DC022331D398522 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"0BD13888D6BA800064CE929CF92B9BF4D9C159E2623752325C878EBF5846D66E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:14:15Z tx 0BD13888D6BA800064CE929CF92B9BF4D9C159E2623752325C878EBF5846D66E code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"0DE70360578F7FDADFB69CF0E6D7B3F1E5106E814B29308606D280661CAD38A5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:14:47Z tx 0DE70360578F7FDADFB69CF0E6D7B3F1E5106E814B29308606D280661CAD38A5 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"50665679611C6D4534CD282A4FBDD5F856BA85B0BE80EAB5309C6F18E2B7BBB6","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:15:24Z tx 50665679611C6D4534CD282A4FBDD5F856BA85B0BE80EAB5309C6F18E2B7BBB6 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"A08FE97C73D40C3BCCDB30C1CC360A68DAC0842A629EFCF8A335F1A035C3417C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:15:59Z tx A08FE97C73D40C3BCCDB30C1CC360A68DAC0842A629EFCF8A335F1A035C3417C code=0 
2026-10-03T02:16:05Z ready: osmo val3 recorded delegation reduced by the slash
2026-10-03T02:16:12Z ready: no slash query in flight
2026-10-03T02:16:12Z ### osmo validators after refresh
```
$ strided_new q stakeibc show-validators osmosis-test-1 -o json
2026/10/03 02:16:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 02:16:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 02:16:13 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"validators":[{"name":"val1","address":"osmovaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6n3re8","weight":"10","delegation":"101374956","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val2","address":"osmovaloper17kht2x2ped6qytr2kklevtvmxpw7wq9r2mr7dy","weight":"10","delegation":"101374946","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val3","address":"osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4","weight":"0","delegation":"100357502","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"0.990003367348911271","delegation_changes_in_progress":"0","slash_query_in_progress":false}]}
```
2026-10-03T02:16:22Z CHECKPOINT PASS: osmo rate still frozen
2026-10-03T02:16:22Z ### drift
```
$ python3 /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/measure_delegation_drift.py --chain-id cosmoshub-test-1 --chain-id osmosis-test-1
=== cosmoshub-test-1 ===
  using REST endpoint: https://cosmoshub-api.internal.stridenet.co
  host delegations: 0
=== osmosis-test-1 ===
  using REST endpoint: https://osmosis-api.internal.stridenet.co
  host delegations: 3

Done. Wrote /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/drift.json and /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/report.md
```
2026-10-03T02:16:24Z CHECKPOINT FAIL: zero over-recorded
2026-10-03T02:23:54Z ## Run 2 recovery: hub delegation channel wedged (acks 309-311 of the accidental phase-1 drain fail in the undelegate callback: TotalDelegations < sum(validators) by 20169539, a genesis setup error - staketia's 50M remaining balance was never added to the hub zone's TotalDelegations, and the seed's R1 confirm-undelegation then subtracted 20.17M). Recovery = the spec's lost-ack path: close-delegation-channel -> restore -> calibrate-delegation.
2026-10-03T02:23:54Z FINDING (script): phase1 hub_drain_refused probe counted zero-amount UNBONDING_QUEUE entries as 'RD still queued', so after D6 had submitted RD2 it sent a REAL --all drain (the guard rightly accepted it: nothing queued).
2026-10-03T02:23:54Z FINDING (runbook): a full drain needs TotalDelegations >= sum(validator delegations) on the zone; otherwise the last batch's ack fails in the callback and the ordered delegation channel wedges. Mainnet check 2026-10-03: every in-scope zone has sum == total except celestia (total > sum by the 156.87B utia staketia portion, safe; pending staketia confirmations 14.3B fit inside it). Add the check to the pre-drain checklist.
broadcast output (strided_new):
```
{"height":"0","txhash":"FA797396B7D16D2CA5131B36E4430538D5C70C8E7E4A72FB47A2FAB409879891","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:24:51Z tx FA797396B7D16D2CA5131B36E4430538D5C70C8E7E4A72FB47A2FAB409879891 code=0 
2026-10-03T02:24:59Z ready: hub delegation channel-2 closed
broadcast output (strided_new):
```
{"height":"0","txhash":"FF10BC47AB47CFBA60A483ADAF852D2A130CE35B62CE6A2741E4350A57F6C874","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:26:03Z tx FF10BC47AB47CFBA60A483ADAF852D2A130CE35B62CE6A2741E4350A57F6C874 code=0 
2026-10-03T02:26:11Z ready: hub delegation ICA reopened
2026-10-03T02:26:24Z CHECKPOINT PASS: restore reset every in-progress flag
2026-10-03T02:26:24Z ### host-side delegations of the hub delegation ICA after restore (lost-ack check)
```
$ bash -c gaiad q staking delegations $(strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json 2>/dev/null | jq -r .host_zone.delegation_ica_address) -o json 2>/dev/null | jq -c '[.delegation_responses[]?]|length'
0
```
broadcast output (strided_new):
```
{"height":"0","txhash":"B363747CF01D728D99D7EAB48D6EC70FE1E7B85B22CCA4E811E48D4119AEB614","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:27:59Z tx B363747CF01D728D99D7EAB48D6EC70FE1E7B85B22CCA4E811E48D4119AEB614 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"85083D68C492D9E91BEB24EAB3DC5321070DB027EBD21BB37E84AB0ABCF150CC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:28:56Z tx 85083D68C492D9E91BEB24EAB3DC5321070DB027EBD21BB37E84AB0ABCF150CC code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"E6E799E86B94ED48E3C2EF81BDBC3E2AEC128345E016B8E32EC32C560CE5B1C7","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:29:56Z tx E6E799E86B94ED48E3C2EF81BDBC3E2AEC128345E016B8E32EC32C560CE5B1C7 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"39485ADAF406839C734C28B30616F4E7A6BC5AB2FBB2D7E9D5207F4E7C3693D7","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:31:11Z tx 39485ADAF406839C734C28B30616F4E7A6BC5AB2FBB2D7E9D5207F4E7C3693D7 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"68317EAFADC0A88B70485A8EB8E2B7D762CB90C957C019D2591A0352E79FB8E7","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:32:17Z tx 68317EAFADC0A88B70485A8EB8E2B7D762CB90C957C019D2591A0352E79FB8E7 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"28D6DA95964276F1EF65EEA05B4A948774BA08B3D97136B471451D3C73FE4CF5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:33:15Z tx 28D6DA95964276F1EF65EEA05B4A948774BA08B3D97136B471451D3C73FE4CF5 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"7AF846F8D15A8A42B677F9E23A63674B471730342C6025F02D78B0DF728AB8FD","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:34:15Z tx 7AF846F8D15A8A42B677F9E23A63674B471730342C6025F02D78B0DF728AB8FD code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"5F1BA30167B50A1C4570ACEDF7BCF883A4C96D7E023E107189B50CACE6CD2F0A","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:35:08Z tx 5F1BA30167B50A1C4570ACEDF7BCF883A4C96D7E023E107189B50CACE6CD2F0A code=0 
2026-10-03T02:36:08Z ### hub validators after calibration
```
$ bash -c strided_new q stakeibc show-validators cosmoshub-test-1 -o json 2>/dev/null | jq -c '[.validators[] | {d: .delegation, p: .delegation_changes_in_progress}]'; strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json 2>/dev/null | jq -c '.host_zone | {total_delegations}'
[{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"},{"d":"0","p":"0"}]
{"total_delegations":"-20169539"}
```
2026-10-03T02:36:48Z ## Phase 2: day 0
2026-10-03T02:37:03Z ## Phase 2 (resume: drift + staketia day 0)
2026-10-03T02:37:03Z ### drift
```
$ python3 /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/measure_delegation_drift.py --chain-id cosmoshub-test-1 --chain-id osmosis-test-1
=== cosmoshub-test-1 ===
  using REST endpoint: https://cosmoshub-api.internal.stridenet.co
  host delegations: 0
=== osmosis-test-1 ===
  using REST endpoint: https://osmosis-api.internal.stridenet.co
  host delegations: 3

Done. Wrote /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/drift.json and /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/report.md
```
2026-10-03T02:37:05Z CHECKPOINT PASS: zero over-recorded
2026-10-03T02:37:35Z tx 352C62351001228032DC00DCD15B9D2D783699B34DB24563C1B61E2C872CF3DA code=0 
2026-10-03T02:37:42Z ### confirm-undelegation 16
```
$ strided_new tx staketia confirm-undelegation 16 352C62351001228032DC00DCD15B9D2D783699B34DB24563C1B61E2C872CF3DA --from st-operator --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 02:37:43 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
Error: rpc error: code = Unknown desc = rpc error: code = Unknown desc = failed to execute message; message index: 0: host zone's delegated balance would be negative after undelegation: negative value not allowed [Stride-Labs/stride/v35/x/staketia/keeper/unbonding.go:154] with gas used: '49012': unknown request
Usage:
  strided tx staketia confirm-undelegation [record-id] [tx-hash] [flags]

Flags:
  -a, --account-number uint         The account number of the signing account (offline mode only)
      --aux                         Generate aux signer data instead of sending a tx
  -b, --broadcast-mode string       Transaction broadcasting mode (sync|async) (default "sync")
      --chain-id string             The network chain ID
      --dry-run                     ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)
      --fee-granter string          Fee granter grants fees for the transaction
      --fee-payer string            Fee payer pays fees for the transaction instead of deducting from the signer
      --fees string                 Fees to pay along with transaction; eg: 10uatom
      --from string                 Name or address of private key with which to sign
      --gas string                  gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically. Note: "auto" option doesn't always report accurate results. Set a valid coin value to adjust the result. Can be used instead of "fees". (default 200000)
      --gas-adjustment float        adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored (default 1)
      --gas-prices string           Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)
      --generate-only               Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase only accessed when providing a key name)
  -h, --help                        help for confirm-undelegation
      --keyring-backend string      Select keyring's backend (os|file|kwallet|pass|test|memory) (default "test")
      --keyring-dir string          The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                      Use a connected Ledger device
      --node string                 <host>:<port> to CometBFT rpc interface for this chain (default "tcp://localhost:26657")
      --note string                 Note to add a description to the transaction (previously --memo)
      --offline                     Offline mode (does not allow any online functionality)
  -o, --output string               Output format (text|json) (default "json")
  -s, --sequence uint               The sequence number of the signing account (offline mode only)
      --sign-mode string            Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature
      --timeout-duration duration   TimeoutDuration is the duration the transaction will be considered valid in the mempool. The transaction's unordered nonce will be set to the time of transaction creation + the duration value passed. If the transaction is still in the mempool, and the block time has passed the time of submission + TimeoutDuration, the transaction will be rejected.
      --timeout-height uint         DEPRECATED: Please use --timeout-duration instead. Set a block timeout height to prevent the tx from being committed past a certain height
      --tip string                  Tip is the amount that is going to be transferred to the fee payer on the target chain. This flag is only valid when used with --aux, and is ignored if the target chain didn't enable the TipDecorator
      --unordered                   Enable unordered transaction delivery; must be used in conjunction with --timeout-duration
  -y, --yes                         Skip tx broadcasting prompt confirmation

Global Flags:
      --home string                directory for config and data (default "/home/validator/.stride")
      --log_format string          The logging format (json|plain) (default "plain")
      --log_level string           The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>') (default "info")
      --log_no_color               Disable colored logs
      --trace                      print out full stack trace on errors
      --verbose_log_level string   The logging level (trace|debug|info|warn|error|fatal|panic|disabled|none) to use when performing operations which require extra verbosity (such as upgrades). When enabled, verbose mode disables any custom log filters. Set this to none to make verbose mode equivalent to normal logging. (default "debug")

command terminated with exit code 1
```
2026-10-03T02:38:23Z Harness repair: staketia adjust-delegated-balance +80000000 (st-safe) so the hub zone's TotalDelegations (-20169539 after calibration) has headroom for the pending staketia confirmations. Compensates the genesis error; staketia's RemainingDelegatedBalance ends overstated by the same amount (irrelevant to the wind-down).
2026-10-03T02:38:41Z tx A922A67E29714A73317A5019E70AE79D2DE46F07D2BDA7FA709ED2D81874A430 code=0 
2026-10-03T02:39:02Z tx 554CB1D8BC08ACA3F7559051E7D37A0DA92B2BCB62BCA97226F3B4CDE9A99E16 code=0 
2026-10-03T02:39:36Z tx 30A69EC12DDBE4BA7E45861DEC6A6435407ADF773C6DF69321F6259962191712 code=0 
2026-10-03T02:39:36Z confirm-undelegation 16 ok
2026-10-03T02:39:52Z ## Phase 3: last redemption cycle
2026-10-03T02:54:52Z TIMEOUT waiting for: every hub/osmo unbonding CLAIMABLE
2026-10-03T02:56:26Z FINDING (code/runbook): v35 deletes RebalanceAllHostZones and MsgRebalanceValidators. A queued record whose amount needs a FULL drain of a zero-weight validator with stored rate < 1 comes up short by applySharesRoundingSafety's trim when the zone's excess is exhausted ('unable to unbond full amount') and retries forever; the drain refuses the zone while it is queued -> the zone deadlocks. Rehearsal: osmosis RE (121.44M) vs capacity 121.42M. Mainnet 2026-10-03: no RETRY records; queued records are small relative to their zones, but add a pre-upgrade capacity check per zone. Escape hatch that still exists in v35: change-validator-weight (used below).
broadcast output (strided_new):
```
{"height":"0","txhash":"F048C62D83B8FB08957A8F06D20F204B9A000C85FAE111567004B9CDA491BD8C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T02:57:21Z tx F048C62D83B8FB08957A8F06D20F204B9A000C85FAE111567004B9CDA491BD8C code=0 
2026-10-03T02:57:25Z ## Phase 3: last redemption cycle
2026-10-03T03:05:34Z ready: every hub/osmo unbonding CLAIMABLE
2026-10-03T03:05:40Z ### claim cosmoshub-test-1 11
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 11 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:05:41 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156214
{"height":"0","txhash":"A2E30475A64A54F9654303BD9EBD0AABEE477CE5EE51699E7D82DABFDA29E44C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:05:52Z ### claim cosmoshub-test-1 14
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 14 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:05:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 155764
{"height":"0","txhash":"B7DCAF49B0C980C11B711850949A8A15E097FACBC093ED11F84107E10C5FFCBA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:06:04Z ### claim cosmoshub-test-1 15
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 15 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:06:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:06:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:06:05 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 155764
{"height":"0","txhash":"9F15826AB7284E58743DBA3278F669136FFFA295BB63524B8D19569B71E93DE6","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:06:18Z ### claim cosmoshub-test-1 16
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 16 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:06:18 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 155755
{"height":"0","txhash":"470DCD3B7393CF83E88B52F989F34A057A51573825278E4254E524954F2A733A","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:06:31Z ### claim cosmoshub-test-1 17
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 17 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:06:31 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:06:31 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:06:32 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 155764
{"height":"0","txhash":"521DACC70B688590BD2C0A5644920ACF4B9BC9CF7ECB51F440053093CA180EA2","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:06:44Z ### claim osmosis-test-1 11
```
$ strided_new tx stakeibc claim-undelegated-tokens osmosis-test-1 11 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:06:44 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:06:44 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:06:45 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 150799
{"height":"0","txhash":"7633EEC9EE51EA153E5AFEC7B14440E47A8ADF5392F13FC8CD79E292098074DD","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:08:09Z ready: zero user redemption records
2026-10-03T03:08:09Z ### deposit records
```
$ strided_new q records list-deposit-record -o json
2026/10/03 03:08:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:08:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:08:11 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"deposit_record":[{"id":"131","amount":"10002819","denom":"uatom","host_zone_id":"cosmoshub-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"68","source":"STRIDE","delegation_txs_in_progress":"0"},{"id":"132","amount":"8","denom":"uosmo","host_zone_id":"osmosis-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"68","source":"STRIDE","delegation_txs_in_progress":"0"},{"id":"137","amount":"166540","denom":"uatom","host_zone_id":"cosmoshub-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"69","source":"WITHDRAWAL_ICA","delegation_txs_in_progress":"0"}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T03:08:22Z CHECKPOINT PASS: stranded deposit never staked
2026-10-03T03:11:54Z CHECKPOINT PASS: hub rate frozen
2026-10-03T03:12:01Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T03:12:08Z CHECKPOINT PASS: no new epoch unbonding records
2026-10-03T03:12:09Z CHECKPOINT PASS: no reinvest/claim-rewards ICA
2026-10-03T03:12:42Z ## Phase 4: admin drain
2026-10-03T03:12:48Z ready: osmo flags clear
2026-10-03T03:13:00Z STATOM_SUPPLY=840515683
2026-10-03T03:13:00Z ## Phase 4 (osmosis zone; the hub zone was already drained by --all in phase 1, which exercised the 3-batch split and, via recovery, the lost-ack path)
broadcast output (strided_new):
```
{"height":"0","txhash":"984A7F2BD92871481F37FA3338FFA391C34BEB3526244DCCD970138307844D22","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:14:12Z tx 984A7F2BD92871481F37FA3338FFA391C34BEB3526244DCCD970138307844D22 code=0 
2026-10-03T03:15:20Z ready: osmo delegation channel closed
broadcast output (strided_new):
```
{"height":"0","txhash":"63CA25B03B02D5B27B2692EBF520E044B7596C38913D88FD24A849C93DDF96B2","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:16:05Z tx 63CA25B03B02D5B27B2692EBF520E044B7596C38913D88FD24A849C93DDF96B2 code=0 
2026-10-03T03:16:12Z ready: osmo delegation channel reopened
2026-10-03T03:16:20Z CHECKPOINT PASS: flags reset by restore
2026-10-03T03:16:27Z ### osmo host-side delegations of the delegation ICA after restore
```
$ osmosisd q staking delegations osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr -o json
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr",
        "validator_address": "osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4",
        "shares": "61167451190847026636236497"
      },
      "balance": {
        "denom": "uosmo",
        "amount": "60555982"
      }
    },
    {
      "delegation": {
        "delegator_address": "osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr",
        "validator_address": "osmovaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6n3re8",
        "shares": "60555984000000000000000000"
      },
      "balance": {
        "denom": "uosmo",
        "amount": "60555984"
      }
    },
    {
      "delegation": {
        "delegator_address": "osmo1vtlszxt9sdgm26vaqhkjxwhslye6c63qg23qehtzlga87kcanzkqh596sr",
        "validator_address": "osmovaloper17kht2x2ped6qytr2kklevtvmxpw7wq9r2mr7dy",
        "shares": "60555982000000000000000000"
      },
      "balance": {
        "denom": "uosmo",
        "amount": "60555982"
      }
    }
  ],
  "pagination": {
    "total": "3"
  }
}
```
2026-10-03T03:16:32Z ### osmo validators after restore
```
$ strided_new q stakeibc show-validators osmosis-test-1 -o json
2026/10/03 03:16:33 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:16:33 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:16:34 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"validators":[{"name":"val1","address":"osmovaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6n3re8","weight":"10","delegation":"60555984","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val2","address":"osmovaloper17kht2x2ped6qytr2kklevtvmxpw7wq9r2mr7dy","weight":"10","delegation":"60555982","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val3","address":"osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4","weight":"10","delegation":"60555982","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"0.990003367348911271","delegation_changes_in_progress":"0","slash_query_in_progress":false}]}
```
broadcast output (strided_new):
```
{"height":"0","txhash":"082BC686704BB3D6C1C790DC592D86700B24911D092352E45EF160B35032CD98","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:19:04Z tx 082BC686704BB3D6C1C790DC592D86700B24911D092352E45EF160B35032CD98 code=0 
2026-10-03T03:19:12Z ready: osmo val1 drained
2026-10-03T03:19:20Z CHECKPOINT PASS: nothing burned
broadcast output (strided_new):
```
{"height":"0","txhash":"1A2F42B39F66B7F2852DACFBFF2C2EF621A4FE65E42711D69E7740FDD6044190","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:21:29Z tx 1A2F42B39F66B7F2852DACFBFF2C2EF621A4FE65E42711D69E7740FDD6044190 code=0 
2026-10-03T03:21:29Z dead-window send: 2026-10-03T03:21:29Z tx 1A2F42B39F66B7F2852DACFBFF2C2EF621A4FE65E42711D69E7740FDD6044190 code=0  
2026-10-03T03:21:29Z CHECKPOINT FAIL: dead-window send fails with a timeout reason
2026-10-03T03:21:37Z ready: no flags after dead window
broadcast output (strided_new):
```
{"height":"0","txhash":"1F6E4F5D6D81A315DAF4FF2EBC34DF59826F13FBCBFCDB9B918C2B01A2DDFEB0","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:22:43Z tx 1F6E4F5D6D81A315DAF4FF2EBC34DF59826F13FBCBFCDB9B918C2B01A2DDFEB0 code=0 
2026-10-03T03:22:52Z ready: offset ack
broadcast output (strided_new):
```
{"height":"0","txhash":"7CC62390A357DEDE80D18859FC8EE4063590B5798D1CBD93AE2B61AD01C554B5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:24:57Z tx 7CC62390A357DEDE80D18859FC8EE4063590B5798D1CBD93AE2B61AD01C554B5 code=0 
2026-10-03T03:25:04Z ready: osmo drained
2026-10-03T03:25:12Z CHECKPOINT PASS: osmo total at dust
2026-10-03T03:25:18Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T03:25:27Z CHECKPOINT PASS: hub rate frozen
2026-10-03T03:25:36Z ## Phase 5: transfers to Osmosis
2026-10-03T03:25:57Z transfer-from-ica cosmoshub-test-1 WITHDRAWAL 2004996uatom
broadcast output (strided_new):
```
{"height":"0","txhash":"FEE4102A9D66BEB596D6FC698D58B656B4C7583602B67674FABB550AA9374FA4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:26:36Z tx FEE4102A9D66BEB596D6FC698D58B656B4C7583602B67674FABB550AA9374FA4 code=29 failed to execute message; message index: 0: unable to submit WITHDRAWAL ICA transfer for cosmoshub-test-1: unable to send ICA tx: cannot send packet using client (07-tendermint-0) with status Expired: client state is not active
2026-10-03T03:39:05Z ## RUN 2 ENDED in phase 5: hub validators OOMKilled (700M limit) -> every hub client expired; then OPERATOR ERROR: I deleted the hub pods to apply a 2G limit without checking the state volume (emptyDir), which wiped the hub chain. Phases 0-4 of run 2 stand. RUN 3 restarts with all fixes.
2026-10-03T03:54:03Z ## Phase 0: pre-flight
2026-10-03T03:54:08Z CHECKPOINT PASS: stride channel-0 -> cosmoshub
2026-10-03T03:54:11Z CHECKPOINT PASS: stride channel-0 client is cosmoshub
2026-10-03T03:54:13Z CHECKPOINT PASS: stride channel-1 -> osmosis
2026-10-03T03:54:17Z CHECKPOINT PASS: hub channel-1 -> osmosis
2026-10-03T03:54:21Z CHECKPOINT PASS: hub ica host allows MsgTransfer
2026-10-03T03:54:22Z CHECKPOINT PASS: osmo ica host allows MsgSend
2026-10-03T03:54:22Z CHECKPOINT PASS: REST stride reachable
2026-10-03T03:54:22Z CHECKPOINT PASS: REST cosmoshub reachable
2026-10-03T03:54:23Z CHECKPOINT PASS: REST osmosis reachable
2026-10-03T03:54:29Z tx 75F677D685023AA9FC790015B5BF0CA7CD92E30E54CA69C1B7BAD8569B9F0F7D code=0 
2026-10-03T03:54:29Z CHECKPOINT PASS: vault receives
broadcast output (osmosisd):
```
{"height":"0","txhash":"D6C46DEA89EC90D631D799D4EAB14B6CDEBD4542DCCC8F7B2813EFCD94AF4680","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T03:54:36Z tx D6C46DEA89EC90D631D799D4EAB14B6CDEBD4542DCCC8F7B2813EFCD94AF4680 code=0 
2026-10-03T03:54:36Z CHECKPOINT PASS: vault spends (multisig)
2026-10-03T03:54:43Z tx 992AC42C19F40244FAD2213D78E0CB4075E7DCF841C06A0A9C0FA9F693C54C0A code=0 
2026-10-03T03:54:43Z CHECKPOINT PASS: sweep operator spends
2026-10-03T03:54:46Z host zones not seeded yet: skipping the withdraw-address check
2026-10-03T03:54:52Z gaia v25.1.0, osmosis 28.0.0, strided 
2026-10-03T03:54:52Z ## Seed (v34)
2026-10-03T03:55:01Z CHECKPOINT PASS: admin-ms address
2026-10-03T03:55:03Z CHECKPOINT PASS: vault-ms address
2026-10-03T03:55:06Z CHECKPOINT PASS: hub-ms address
2026-10-03T03:55:06Z ### fund admin-ms
```
$ strided_old tx bank send faucet stride1mymazvsd79f9yhjq4n84dchyf6zvfmd8ed2nxc 1000000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:06 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 110745
{"height":"0","txhash":"87C1CED8EB389832C5CA49A909E674E44FC887FAD533AA2AF4A8276CDF036117","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:12Z ### fund sweep operator
```
$ strided_old tx bank send faucet stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy 100000000ustrd --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:12 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 75082
{"height":"0","txhash":"F2AED251C5184F75E6CFCC8A8E530E5ECB85DA40479F115EEF566AC20602798E","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:17Z ### fund vault-ms
```
$ osmosisd tx bank send faucet osmo1mymazvsd79f9yhjq4n84dchyf6zvfmd8jaelyx 100000000uosmo --from faucet --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 138228
{"height":"0","txhash":"89D6CD26745DD523E479281A222AF29B4288D0D636B2448927708016F2C084EE","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:22Z ### fund hub-ms
```
$ gaiad tx bank send faucet cosmos1h0dup2qw23uhgn9nxyhyze4cxzrgu8rtrcnv7d 100000000uatom --from faucet --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 171159
{"height":"0","txhash":"22B4410A0E7DF39B94B033B37DB70682D2C364ACB053B2CE252C95A176195DEA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:30Z CHECKPOINT PASS: 8 hub validators
2026-10-03T03:55:30Z CHECKPOINT PASS: 3 osmosis validators
2026-10-03T03:55:30Z ### register hub zone
```
$ strided_old tx stakeibc register-host-zone connection-0 uatom cosmos ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 channel-0 1 false --max-messages-per-ica-tx 3 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:31 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 663678
{"height":"0","txhash":"FB64A24F60B6EC869FE43E059BA7BAA661D9967ADD27BF4AEB93AA18085F3A08","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:35Z ### register osmo zone
```
$ strided_old tx stakeibc register-host-zone connection-1 uosmo osmo ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B channel-1 1 false --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:36 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 647517
{"height":"0","txhash":"39BB9B17264FC843633085B17A52C7AFF60CF8E111C43F0226D98953BD21F787","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:44Z ### add hub validators
```
$ strided_old tx stakeibc add-validators cosmoshub-test-1 /tmp/hub_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:45 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:45 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:46 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 1098741
{"height":"0","txhash":"A516142CC6DAD1069F3EEC4187A72815914FE429F8B407C6E34EC082785F0541","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:55:52Z ### add osmo validators
```
$ strided_old tx stakeibc add-validators osmosis-test-1 /tmp/osmo_vals.json --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:55:53 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 402217
{"height":"0","txhash":"5A393AD106E75789EBA9A036548BE1463971129A9766DDB391372E1EBA0141DC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:56:01Z ready: cosmoshub-test-1 delegation_ica_address
2026-10-03T03:56:03Z ready: cosmoshub-test-1 fee_ica_address
2026-10-03T03:56:06Z ready: cosmoshub-test-1 withdrawal_ica_address
2026-10-03T03:56:11Z ready: cosmoshub-test-1 redemption_ica_address
2026-10-03T03:56:17Z ready: osmosis-test-1 delegation_ica_address
2026-10-03T03:56:23Z ready: osmosis-test-1 fee_ica_address
2026-10-03T03:56:29Z ready: osmosis-test-1 withdrawal_ica_address
2026-10-03T03:56:33Z ready: osmosis-test-1 redemption_ica_address
2026-10-03T03:56:33Z ### atom to stride
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 2000000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 223377
{"height":"0","txhash":"596E63FE9EA3D5E236EE6621C35D21CD15211579DCEE8E3342506C55212E2917","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:56:42Z ### osmo to stride
```
$ osmosisd tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 184975
{"height":"0","txhash":"33FA9637D940DC6B43923A4B76A6BAA1A867B83AAEB1D5E6E8F9425A848C6030","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:56:51Z ready: atom on stride
2026-10-03T03:56:56Z ready: osmo on stride
2026-10-03T03:56:56Z ### liquid stake 1000 ATOM
```
$ strided_old tx stakeibc liquid-stake 1000000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:56:57 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 198025
{"height":"0","txhash":"F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:57:03Z ### liquid stake 300 OSMO
```
$ strided_old tx stakeibc liquid-stake 300000000 uosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:57:04 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 160689
{"height":"0","txhash":"256362526079C6ACB1A46454F6B23095C78A1AA691A1E35E703A280C0422DA87","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:58:25Z ready: hub delegated
2026-10-03T03:58:31Z ready: osmo delegated
2026-10-03T03:58:31Z ### holder base
```
$ strided_old tx bank send user1 stride1ef2axra0mrwwqacf2l33ye62qzavgtwrypqmgs 50000000stuatom,10000000ustrd,20000000ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:58:32 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 133923
{"height":"0","txhash":"4D2474FB7224E1C2D869A5192A3187D4D6A38F7FBF07F0C79E7370747D0DEF81","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:58:40Z ### vesting acct
```
$ strided_old tx vesting create-vesting-account stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 1000000ustrd 1822535920 --delayed --from faucet --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:58:41 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 94638
{"height":"0","txhash":"8938B07A77B3E84C01E864586805A80F1D85605B12993A7AC0EA281833EA694D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:58:49Z ### holder vesting
```
$ strided_old tx bank send user1 stride1lvunkuca3l200th2skqhlfmmyrx2kk5zqxzggk 30000000stuatom,5000000ustrd --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:58:50 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 97917
{"height":"0","txhash":"299B3361D406F70577910166DAD03D99ACA7C6A214CC3C9AF190BC22BEA89262","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:58:56Z ### distribution holds stATOM
```
$ strided_old tx distribution fund-community-pool 5000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:58:58 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 81694
{"height":"0","txhash":"66E35BB64578B53B10A3DF2FE9E79E628A6A63B139A6C9D8437EE5FE0C594A02","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:59:06Z ### statom to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 100000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:59:07 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:59:07 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:59:08 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138193
{"height":"0","txhash":"A3F2365D6E40A2D3991E26A128F5F38FD5F0E820E4CADB53696643D867EE32AA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:59:17Z ### statom to hub
```
$ strided_old tx ibc-transfer transfer transfer channel-0 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l 60000000stuatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:59:17 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:59:17 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:59:18 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 138240
{"height":"0","txhash":"1950828A42B5629A7FB836EFC4DC1D8BFA1B5A15E3B9A4BE2B7A687D7C08D58D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:59:24Z ### stosmo to osmosis
```
$ strided_old tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 50000000stuosmo --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 03:59:24 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 03:59:24 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 03:59:25 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 121965
{"height":"0","txhash":"0BA7345DD013BC78B751B7664923E5823A52DA81CFD87ACAB2AF0DE291F771D5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:59:34Z ready: statom on hub
2026-10-03T03:59:34Z ### statom hub->osmosis (two-hop)
```
$ gaiad tx ibc-transfer transfer transfer channel-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d 30000000ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 221128
{"height":"0","txhash":"AEAAE39E4F244308C0A20C71A162927565C625713A282818699A8F44E9F95D8C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T03:59:44Z ### fund hub fee ICA
```
$ gaiad tx bank send user1 cosmos1lrh860n54gck0dxnwacn53xje6p9lah3rc3vhv9rs8ypf70spr7qx6xwtr 3000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131751
{"height":"0","txhash":"D6CCD592B6EC8BA965820EAA780940B552C0A30928F112B7BB353AB4026A5AF3","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:00:00Z ### fund hub withdrawal ICA
```
$ gaiad tx bank send user1 cosmos1xnl77wq8tzvmwn6q8hnv2xnx5v0w5vgpxs7faq6f5nwtc0px33gqn63afp 4000000uatom --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 131755
{"height":"0","txhash":"925C4307BC8931D31B98D3334BF57AFAF90426056D0E63C04D2A905D9CD89958","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:00:23Z ### fund osmo fee ICA
```
$ osmosisd tx bank send user1 osmo1fp5c7w09r030s6szymcfg4d2a3sfh9s4raucphq7mnuhwmerzn5sazu2c4 3000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120492
{"height":"0","txhash":"D4E4CB0CDF657D4F2248CC7AFE97AAEFFD4B530D657F7AA53EF5669D214DC56F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:00:34Z ### fund osmo withdrawal ICA
```
$ osmosisd tx bank send user1 osmo1m7saah7sl73xq08x3gzpa7kw9yr9eagj4vyc2ezfaf8xw7hv5zgsznsnhk 4000000uosmo --from user1 --keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json
gas estimate: 120442
{"height":"0","txhash":"70088122F5025507741716A73F1F5E5949C041CA3AA31692D710AA3B3A757C85","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
broadcast output (gaiad):
```
{"height":"0","txhash":"A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:01:21Z tx A10DB10D1A52EBCC1A9391FFB800AC94F6BA9F25C972B83B2C99D6DDDEDA7CF5 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:01:57Z tx 0859D83FB661B03B258BDAC9EF2F3E813A8E9798CA2D95B29F8DB97F38250256 code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:02:33Z tx CE4C03CC794CB8AF4E2AE7D0120F610369B3D2E56D51D41C59658B08AFC6378D code=0 
broadcast output (gaiad):
```
{"height":"0","txhash":"02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:03:09Z tx 02E2629ECB04D72BAD7EC4389C62E5D43E5DFFD3AD09D631523F35A391711437 code=0 
2026-10-03T04:03:09Z ### multisign + broadcast transfer grant
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx sign /tmp/unsigned.json --from d1 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s1.json
  gaiad tx sign /tmp/unsigned.json --from d2 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s2.json
  gaiad tx multisign /tmp/unsigned.json hub-ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/signed.json
  gaiad tx broadcast /tmp/signed.json --chain-id cosmoshub-test-1 -o json
{"height":"0","txhash":"75666F8E7605777CE72F5BBCBA517D3BE18C0C45DC288833ABF099D6A3C28CCC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:03:47Z CHECKPOINT PASS: transfer grant present
2026-10-03T04:03:54Z ### fund st-safe
```
$ strided_old tx bank send faucet stride1tpzfseenwg4kq54sf9hdp3mkra652fvqtsuclq 10000000ustrd --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:03:55 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:03:55 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:03:56 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 90853
{"height":"0","txhash":"A922A67E29714A73317A5019E70AE79D2DE46F07D2BDA7FA709ED2D81874A430","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:04:04Z ### staketia portion into the hub zone total
```
$ strided_old tx staketia adjust-delegated-balance increase 50000000 cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p --from st-safe --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:04:06 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 201900
{"height":"0","txhash":"67DCE608A7A48D8282A8492EA032EEBA3478A2AA526F52BC2508B8C0A405A128","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:04:18Z CHECKPOINT FAIL: staketia portion booked
2026-10-03T04:04:51Z ## Seed (v34)
2026-10-03T04:04:51Z SEED_RESUME=1: skipping Part A and the first half of Part B
2026-10-03T04:05:04Z resume: HIST_TX=F62787B4B2297B2C8779363D02C33DFC27E92D64368C660F6173C0387AFA76DA,        8 hub validators,        3 osmosis validators
2026-10-03T04:05:09Z ready: osmo delegated
2026-10-03T04:05:09Z SEED_RESUME=2: skipping the holders/transfers half of Part B
2026-10-03T04:05:16Z staketia grants already present: skipping the multisig delegate and grants (rerun)
2026-10-03T04:05:24Z CHECKPOINT PASS: transfer grant present
2026-10-03T04:05:38Z CHECKPOINT PASS: staketia portion booked
2026-10-03T04:05:39Z ### rate limit proposal
```
$ strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:05:40 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 205068
{"height":"0","txhash":"296C45AAC1E0888A471F74886EF88E64017D6F0AB64B86DC0A6B2E974EBF075F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:06:14Z tx C48452FA0CC5EE2A3A704EC10C61EBDAACA774377B3C56148E806BBEA220897F code=0 
2026-10-03T04:06:16Z tx CA0009CA05B0C65469AE083B09C5E73358866D011096617CE091AA5621D4D4A1 code=0 
2026-10-03T04:06:20Z tx C48801B75D5848CB36A00083E16815E6EA074EF995F405F2531C86B69547C49E code=0 
2026-10-03T04:06:24Z tx 030A8072954A2BCEF91B9E36673E0D665105C91ACF77F737300E06926FC906C6 code=3 failed to execute message; message index: 0: 1: inactive proposal
2026-10-03T04:06:24Z vote from val1 failed: {"height":"0","txhash":"030A8072954A2BCEF91B9E36673E0D665105C91ACF77F737300E06926FC906C6","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"
2026-10-03T04:06:30Z ready: rate limit live
2026-10-03T04:06:40Z day epoch now=9: D0=1791000428 D1=1791000968 D2=1791001148 D3=1791001328 D4=1791001508 (staketia prepare epoch) upgrade target U=1791001658
2026-10-03T04:06:40Z ### hub RA redeem
```
$ strided_old tx stakeibc redeem-stake 30000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:06:42 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125350
{"height":"0","txhash":"8E0B55D15E18C5BD390BD89D0005D0E4516F2C91C55B333C8C0AB7DAFD37B8D4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:06:51Z ### osmo val3 weight 0
```
$ strided_old tx stakeibc change-validator-weight osmosis-test-1 osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4 0 --from admin --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:06:51 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:06:51 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:06:52 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 117798
{"height":"0","txhash":"A99AAAC32E7A51F97D243811D15CA50C0B9F437DE22FF865D20D042CE789E01A","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:07:00Z ### stop signing osmosis-validator-2
```
$ pause_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T04:07:52Z ready: osmo val3 jailed
2026-10-03T04:07:52Z ### resume osmosis-validator-2
```
$ resume_pod_process osmosis-validator-2 osmosisd

```
2026-10-03T04:07:53Z ### osmo RE redeem (retry)
```
$ strided_old tx stakeibc redeem-stake 120000000 osmosis-test-1 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:07:54 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 122533
{"height":"0","txhash":"15516371C1576BB89AA826551C0FBFABBA7F948AB237AF9D53322A995AF1FF38","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:19:18Z ### hub RB redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:19:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:19:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:19:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125490
{"height":"0","txhash":"494C68271FF9E32BC77CE237B9A84D591FF4F2C78E329973B4664E11902EE150","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:22:18Z ### hub RC redeem
```
$ strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:22:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:22:19 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:22:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125490
{"height":"0","txhash":"1A54CD5BE0AA9543657AB0120EFC535ABEA3B9F0FEE6AB9734A280FB0A66E276","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:22:29Z ### staketia R1
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:22:30 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 152454
{"height":"0","txhash":"F13F4481DE7F45FB9B0A0A5D485322BE484B6B205E7B3BFFA55D9DEEA4A7A67F","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:25:19Z ready: staketia unbonding record in UNBONDING_QUEUE
2026-10-03T04:25:34Z staketia record 12 native_amount=21196762
2026-10-03T04:25:34Z ### hub-ms undelegate via authz exec
```
$ kubectl --context integration -n integration exec cosmoshub-validator-0 -c validator -- sh -c set -e
  gaiad tx staking unbond cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p 21196762uatom --from hub-ms --generate-only --keyring-backend test --chain-id cosmoshub-test-1 > /tmp/unbond.json
  gaiad tx authz exec /tmp/unbond.json --from st-operator --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 390124
{"height":"0","txhash":"B2163A3F3F294B35F0E418E97C2D3C01B8C305EAC79EF934E299B730E5E3A894","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:25:46Z CHECKPOINT PASS: hub undelegate tx hash captured
2026-10-03T04:25:52Z ### staketia confirm-undelegation
```
$ strided_old tx staketia confirm-undelegation 12 B2163A3F3F294B35F0E418E97C2D3C01B8C305EAC79EF934E299B730E5E3A894 --from st-operator --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:25:52 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 227749
{"height":"0","txhash":"4DC7DCFC72B515F08F8CFAE42B1CAF9A168D58283CB320151873527F824FBE71","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:26:01Z ### hub RD redeem (queue)
```
$ strided_old tx stakeibc redeem-stake 10000000 cosmoshub-test-1 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:26:02 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 125530
{"height":"0","txhash":"C6517E7B702CCC98D31DEC98A9F9D5FA85EF6896A55952B8034CCE25FB751149","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:26:10Z ### staketia R2
```
$ strided_old tx staketia redeem-stake 20000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:26:11 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:26:11 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:26:12 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 137302
{"height":"0","txhash":"144758047728F829CE9EBE18B017049136BD9DA5F0D161C304D7E32BCC8D1A2D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:26:19Z ### staketia R3 spillover
```
$ strided_old tx staketia redeem-stake 40000000 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:26:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 308310
{"height":"0","txhash":"19F327BA93A215FF2E2E1F45BC7EF57810160DC68F66097D7C6AE886A5FE9B85","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:26:35Z ready: osmo RE in UNBONDING_RETRY_QUEUE
2026-10-03T04:26:35Z CHECKPOINT PASS: osmo RE in UNBONDING_RETRY_QUEUE (bounded wait)
2026-10-03T04:26:41Z CHECKPOINT PASS: hub RA CLAIMABLE
2026-10-03T04:26:45Z CHECKPOINT PASS: hub RB+RC EXIT_TRANSFER_QUEUE
2026-10-03T04:26:51Z CHECKPOINT PASS: hub RB+RC both EXIT_TRANSFER_QUEUE
2026-10-03T04:26:57Z CHECKPOINT PASS: hub RD UNBONDING_QUEUE
2026-10-03T04:27:12Z pre-upgrade rates: hub=1.060454643106193577 osmo=1.012017723780592943
2026-10-03T04:27:12Z CHECKPOINT PASS: state.env has HIST_TX
2026-10-03T04:27:12Z ### records at upgrade
```
$ strided_old q records list-epoch-unbonding-record -o json
2026/10/03 04:27:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:27:12 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:27:13 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"6","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"7","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"8","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"9","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"31658485","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"31658485","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791000674408205316","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.9.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"10","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"120000000","native_token_amount":"121439577","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121439577","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.10.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"11","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"12","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"13","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"14","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"21182389","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791001575305286741","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.14.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"21196762","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791001755070630609","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.15.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"42822587","native_token_amount":"45401920","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":["cosmoshub-test-1.16.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T04:27:18Z ### staketia records at upgrade
```
$ strided_old q staketia unbonding-records -o json
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:27:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"unbonding_records":[{"id":"12","status":"UNBONDING_IN_PROGRESS","st_token_amount":"20000000","native_amount":"21196762","unbonding_completion_time_seconds":"1791001797","undelegation_tx_hash":"B2163A3F3F294B35F0E418E97C2D3C01B8C305EAC79EF934E299B730E5E3A894","unbonded_token_sweep_tx_hash":""},{"id":"16","status":"ACCUMULATING_REDEMPTIONS","st_token_amount":"27177413","native_amount":"28814389","unbonding_completion_time_seconds":"0","undelegation_tx_hash":"","unbonded_token_sweep_tx_hash":""}]}
```
2026-10-03T04:27:25Z seed done at 1791001645; upgrade target U=1791001658 (now - U = -13s)
2026-10-03T04:27:26Z ## Phase 1: upgrade to v35
2026-10-03T04:27:34Z ABORT: only 4s to U (need >= 60s for the proposal to pass); pushing the height out would cross D5. Re-seed.
2026-10-03T04:27:38Z seed ended 13s before U; shifting the upgrade target one day epoch: U=1791001658 -> 1791001838 (D5=1791001688); a fresh RD is redeemed after D5
2026-10-03T04:28:30Z tx CBAEB54F9438ECC0748C0AA6E7F0712421C21B7C7C0CCCD9EE3E37EBC3513B1C code=0 
2026-10-03T04:28:30Z hub RD2 redeem (queue) after D5: CBAEB54F9438ECC0748C0AA6E7F0712421C21B7C7C0CCCD9EE3E37EBC3513B1C
2026-10-03T04:28:30Z ## Phase 1: upgrade to v35
2026-10-03T04:28:35Z upgrade height 2096 (now 1973, target time 1791001838)

Submitting proposal for v35 at height 2096...

code: 0
txhash: 6994309A0C7AD66E97DABEAC8213B6344BDE55BF50D4951E9938A752BFC77FB3

Proposal:

proposal:
  deposit_end_time: "2026-10-03T04:29:18.749387095Z"
  final_tally_result:
    abstain_count: "0"
    no_count: "0"
    no_with_veto_count: "0"
    yes_count: "0"
  id: "2"
  messages:
  - type: /cosmos.upgrade.v1beta1.MsgSoftwareUpgrade
    value:
      authority: stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
      plan:
        height: "2096"
        name: v35
        time: "0001-01-01T00:00:00Z"
  proposer: stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
  status: PROPOSAL_STATUS_VOTING_PERIOD
  submit_time: "2026-10-03T04:28:48.749387095Z"
  summary: Upgrade v35
  title: Upgrade v35
  total_deposit:
  - amount: "2000000000"
    denom: ustrd
  voting_end_time: "2026-10-03T04:29:18.749387095Z"
  voting_start_time: "2026-10-03T04:28:48.749387095Z"

Voting on proposal #2...

code: 0
txhash: 19C7A8442FC29CF8AE1FAAE0350D5C747920DE4D6287AF925E8446198A780918
code: 0
txhash: 618A61BCAEA228FB3A48343994B97D76CE74696821BFCD4C9BBFAF2BA975274D
code: 0
txhash: 38C621318C31BACE5FA2350341595FAD03BE35E2A19F35F9EBBDD67D5DD81CC1
code: 0
txhash: 736F281539A971E902B49D86361B1E3622B963E4F5827A50B308A76728382840

Vote confirmation:

tally:
  abstain_count: "0"
  no_count: "0"
  no_with_veto_count: "0"
  yes_count: "4000000000"

Proposal Status:

Proposal passed!
2026-10-03T04:30:18Z ### deposit in flight
```
$ strided_old tx stakeibc liquid-stake 10000000 uatom --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:30:20 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 163009
{"height":"0","txhash":"AF47B3D724D01B138CAF7320FB47BDACCDE594E51FEDC1D606359D1530BF4856","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:30:36Z CHECKPOINT PASS: hub deposit in TRANSFER_QUEUE
2026-10-03T04:31:53Z ready: v35 running
2026-10-03T04:32:04Z rate at the upgrade: hub 1.061651551979424899 (seed end 1.060454643106193577), osmo 1.012022081954169013 (seed end 1.012017723780592943)
2026-10-03T04:33:43Z FINDING (harness): chain stalled at 2096 after the upgrade; restarting stride-validator-1..3 processes
2026-10-03T04:34:04Z ready: chain advancing after restart
2026-10-03T04:34:04Z ### handler log lines
```
$ handler_log_lines
[90m4:29AM[0m [32mINF[0m [1mproposal tallied[0m [36mexpedited=[0mfalse [36mmodule=[0mx/gov [36mproposal=[0m2 [36mresults=[0mpassed [36mstatus=[0mPROPOSAL_STATUS_PASSED [36mtitle=[0m"Upgrade v35"
[90m4:31AM[0m [31mERR[0m [1mUPGRADE "v35" NEEDED at height: 2096: [0m [36mmodule=[0mx/upgrade
[90m4:31AM[0m [31mERR[0m [1merror in proxyAppConn.FinalizeBlock[0m [36merr=[0m"UPGRADE \"v35\" NEEDED at height: 2096: " [36mmodule=[0mstate
[90m4:31AM[0m [31mERR[0m [1mCONSENSUS FAILURE!!![0m [36merr=[0m"failed to apply block; error UPGRADE \"v35\" NEEDED at height: 2096: " [36mmodule=[0mconsensus [36mstack=[0m"goroutine 486 [running]:\nruntime/debug.Stack()\n\truntime/debug/stack.go:26 +0x5e\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine.func2()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:806 +0x46\npanic({0x4f097a0?, 0xc003a8afe0?})\n\truntime/panic.go:783 +0x132\ngithub.com/cometbft/cometbft/consensus.(*State).finalizeCommit(0xc005212388, 0x830)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1826 +0xdc5\ngithub.com/cometbft/cometbft/consensus.(*State).tryFinalizeCommit(0xc005212388, 0x830)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1725 +0x2e5\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit.func1()\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1660 +0x9c\ngithub.com/cometbft/cometbft/consensus.(*State).enterCommit(0xc005212388, 0x830, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:1698 +0xc36\ngithub.com/cometbft/cometbft/consensus.(*State).addVote(0xc005212388, 0xc0037752b0, {0xc0017b20c0, 0x28})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2402 +0x1e7f\ngithub.com/cometbft/cometbft/consensus.(*State).tryAddVote(0xc005212388, 0xc0037752b0, {0xc0017b20c0?, 0xc0081d0a00?})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:2125 +0x26\ngithub.com/cometbft/cometbft/consensus.(*State).handleMsg(0xc005212388, {{0x7dccca0, 0xc00d4b1c88}, {0xc0017b20c0, 0x28}})\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:965 +0x385\ngithub.com/cometbft/cometbft/consensus.(*State).receiveRoutine(0xc005212388, 0x0)\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:846 +0x4a7\ncreated by github.com/cometbft/cometbft/consensus.(*State).OnStart in goroutine 472\n\tgithub.com/cometbft/cometbft@v0.39.3/consensus/state.go:395 +0x107\n"
[90m4:31AM[0m [32mINF[0m running app [36margs=[0m["start","--reject-config-defaults"] [36mmodule=[0mcosmovisor [36mpath=[0m/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided
[90m4:31AM[0m [32mINF[0m [1mapplying upgrade "v35" at height: 2096[0m [36mmodule=[0mx/upgrade
[90m4:31AM[0m [32mINF[0m [1mStarting upgrade[0m [36mheight=[0m2096 [36mmodule=[0mbaseapp [36mname=[0mv35
[90m4:31AM[0m [32mINF[0m [1mStarting upgrade v35 (protocol wind-down)...[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: autopilot StakeibcActive set to false[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: wasm code upload access restricted to the gov module[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: 0 contract admin(s) moved to gov[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: host zone comdex-1 not found, skipping deprecation[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: trade route uusdc/adydx not found, skipping deletion[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: rate limit removed for stuatom on channel-1[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1kaydmnfuc8uezlk9hwyl660kg3enw3rs4j7dlhk9cwvnnljm8tqqwercxt -> stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1kaydmnfuc8uezlk9hwyl660kg3enw3rs4j7dlhk9cwvnnljm8tqqwercxt -> stride1rq5vx8e3grfc6g9k6quu4rzhs0qtdq48xpu0wxka2f6p8kk9fdgsmv7eyx removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair cosmos1lrh860n54gck0dxnwacn53xje6p9lah3rc3vhv9rs8ypf70spr7qx6xwtr -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1fp5c7w09r030s6szymcfg4d2a3sfh9s4raucphq7mnuhwmerzn5sazu2c4 -> stride178jw99dmgyaqkmn5meevmcak27qte0gnymrztv removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1xhxq62juz5gs0ajxjqxwj8gl6re5786pn7swxt4a0hrfng7jp59qk6yn0x -> stride1mgs4mq90d2t4gs8vdrg863lcjjv0cuafs9spur988td6lv2dkceqqptlep removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair osmo1xhxq62juz5gs0ajxjqxwj8gl6re5786pn7swxt4a0hrfng7jp59qk6yn0x -> stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair stride19lxulvqqr22pkjtmvrm0wh8h8v3v07a9gedzk5s3td0rll6zlsksa3k48n -> cosmos1pw42eyc059rd30xcly287m9ra5dhtl0f44kl2l7glpaexwafakzs9wu04c removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair stride1k6gmsmjl08yv9w95tqt9hgdxhn34e29maplgpzf5sjrphl2drxlsgnr4zt -> cosmos12m2l5ju4l3u2th2pfr4hjfp622zuls2xrtdm63tqhv9nznvhwjyqmzcvv4 removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair stride1przgs8xf5jes3raerk57nkzfcweqnd4hpu5tcdnr4mlsshp7ju2ssfn9zs -> osmo1k4m803rga04766vhu3t9eez4whwstdcj55xq3tfjj7cwp305cmcsv84z82 removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: whitelisted pair stride1yak5fa2ukpvhq2kmr534egsvd2prvdpugg6864ndg74fdrv5z7ksda5nhr -> osmo1z6qz43c3u2m4gsghzf9eraj5r8amlrsn277sz60ukg7nu69mungqx5f93r removed[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: cosmoshub-test-1: 0 stale in-progress flag(s) reset on channel-2[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: osmosis-test-1: 0 stale in-progress flag(s) reset on channel-8[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: 0 pending slash-path ICQ(s) deleted for haqq_11235-1[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping slash query flag reset[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: 0 pending withdrawal-balance ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: 0 pending calibration ICQ(s) deleted[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: host zone haqq_11235-1 not found, skipping delegation reconciliation[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mv35: LSM deposit cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuqlj02r/116327 not found, skipping reset[0m [36mmodule=[0mbaseapp
[90m4:31AM[0m [32mINF[0m [1mUpgrade v35 complete[0m [36mmodule=[0mbaseapp
```
2026-10-03T04:34:06Z handler error lines: none
2026-10-03T04:34:06Z CHECKPOINT PASS: no handler error
2026-10-03T04:34:21Z hub drain-refusal probe skipped: RD no longer queued
2026-10-03T04:34:21Z CHECKPOINT FAIL: drain refused while RD queued (hub)
2026-10-03T04:34:25Z CHECKPOINT PASS: liquid-stake cannot route
2026-10-03T04:34:31Z CHECKPOINT PASS: autopilot stakeibc off
2026-10-03T04:34:36Z CHECKPOINT PASS: rate limits removed
2026-10-03T04:34:43Z CHECKPOINT PASS: wasm upload gov-only
2026-10-03T04:34:48Z CHECKPOINT PASS: ica host allow-list trimmed
2026-10-03T04:34:53Z CHECKPOINT PASS: historical tx decodes
2026-10-03T04:34:57Z CHECKPOINT PASS: hub rate frozen
2026-10-03T04:35:04Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T04:35:04Z ### records after upgrade
```
$ strided_new q records list-epoch-unbonding-record -o json
2026/10/03 04:35:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:35:05 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:35:06 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"epoch_unbonding_record":[{"epoch_number":"6","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"7","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"8","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"9","host_zone_unbondings":[{"st_token_amount":"30000000","native_token_amount":"31658485","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"31658485","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791000674408205316","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.9.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"10","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"120000000","native_token_amount":"121439577","st_tokens_to_burn":"120000000","native_tokens_to_unbond":"121439577","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_RETRY_QUEUE","user_redemption_records":["osmosis-test-1.10.osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d"]}]},{"epoch_number":"11","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"12","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"13","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"14","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"21182389","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"21182389","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791001575305286741","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.14.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"15","host_zone_unbondings":[{"st_token_amount":"20000000","native_token_amount":"21196762","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"21196762","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791001755070630609","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.15.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"16","host_zone_unbondings":[{"st_token_amount":"42822587","native_token_amount":"45419545","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"45419545","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791001934085843089","status":"CLAIMABLE","user_redemption_records":["cosmoshub-test-1.16.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"17","host_zone_unbondings":[{"st_token_amount":"10000000","native_token_amount":"10614572","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"1791002113461212004","status":"EXIT_TRANSFER_QUEUE","user_redemption_records":["cosmoshub-test-1.17.cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l"]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]},{"epoch_number":"18","host_zone_unbondings":[{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uatom","host_zone_id":"cosmoshub-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]},{"st_token_amount":"0","native_token_amount":"0","st_tokens_to_burn":"0","native_tokens_to_unbond":"0","claimable_native_tokens":"0","undelegation_txs_in_progress":"0","denom":"uosmo","host_zone_id":"osmosis-test-1","unbonding_time":"0","status":"UNBONDING_QUEUE","user_redemption_records":[]}]}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T04:35:23Z ### autopilot liquid-stake memo from Hub
```
$ gaiad tx ibc-transfer transfer transfer channel-0 stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn 1000000uatom --memo {"autopilot":{"receiver":"stride15lf3jnxe8k2r72hang5cm8ymkx6che7t3xe5nn","stakeibc":{"action":"LiquidStake"}}} --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
gas estimate: 184306
{"height":"0","txhash":"504891569169732C9C9A6264929B002BE83CDE53DE2819228DC547D3813E91AA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:36:40Z CHECKPOINT PASS: autopilot route refused (no stATOM minted, ATOM refunded)
2026-10-03T04:36:40Z ### register ICA on Stride from Hub
```
$ gaiad tx interchain-accounts controller register connection-0 --from user1 --keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json
Error: rpc error: code = Unknown desc = rpc error: code = Unknown desc = failed to execute message; message index: 0: channel open init callback failed for port ID: icacontroller-cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l, channel ID: channel-8: cannot unmarshal ICS-27 interchain accounts metadata: invalid type [cosmos/ibc-go/v10@v10.3.0/modules/apps/27-interchain-accounts/types/metadata.go:62] with gas used: '90096': unknown request
Usage:
  gaiad tx interchain-accounts controller register [connection-id] [flags]

Flags:
  -a, --account-number uint         The account number of the signing account (offline mode only)
      --aux                         Generate aux signer data instead of sending a tx
  -b, --broadcast-mode string       Transaction broadcasting mode (sync|async) (default "sync")
      --chain-id string             The network chain ID
      --dry-run                     ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)
      --fee-granter string          Fee granter grants fees for the transaction
      --fee-payer string            Fee payer pays fees for the transaction instead of deducting from the signer
      --fees string                 Fees to pay along with transaction; eg: 10uatom
      --from string                 Name or address of private key with which to sign
      --gas string                  gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically. Note: "auto" option doesn't always report accurate results. Set a valid coin value to adjust the result. Can be used instead of "fees". (default 200000)
      --gas-adjustment float        adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
      --gas-prices string           Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)
      --generate-only               Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase only accessed when providing a key name)
  -h, --help                        help for register
      --keyring-backend string      Select keyring's backend (os|file|kwallet|pass|test|memory) (default "os")
      --keyring-dir string          The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                      Use a connected Ledger device
      --node string                 <host>:<port> to CometBFT rpc interface for this chain (default "tcp://localhost:26657")
      --note string                 Note to add a description to the transaction (previously --memo)
      --offline                     Offline mode (does not allow any online functionality)
      --ordering string             Channel ordering, can be one of: ORDER_ORDERED, ORDER_UNORDERED (default "ORDER_UNORDERED")
  -o, --output string               Output format (text|json) (default "json")
  -s, --sequence uint               The sequence number of the signing account (offline mode only)
      --sign-mode string            Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature
      --timeout-duration duration   TimeoutDuration is the duration the transaction will be considered valid in the mempool. The transaction's unordered nonce will be set to the time of transaction creation + the duration value passed. If the transaction is still in the mempool, and the block time has passed the time of submission + TimeoutTimestamp, the transaction will be rejected.
      --timeout-height uint         DEPRECATED: Please use --timeout-duration instead. Set a block timeout height to prevent the tx from being committed past a certain height
      --tip string                  Tip is the amount that is going to be transferred to the fee payer on the target chain. This flag is only valid when used with --aux, and is ignored if the target chain didn't enable the TipDecorator
      --unordered                   Enable unordered transaction delivery; must be used in conjunction with --timeout-duration
      --version string              Controller chain channel version
  -y, --yes                         Skip tx broadcasting prompt confirmation

Global Flags:
      --home string         directory for config and data (default "/home/validator/.gaia")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>') (default "info")
      --log_no_color        Disable colored logs
      --trace               print out full stack trace on errors

command terminated with exit code 1
```
2026-10-03T04:38:54Z TIMEOUT waiting for: hub-controlled ICA open on Stride
2026-10-03T04:38:54Z CHECKPOINT FAIL: ICA host route (channel did not open; not testable)
2026-10-03T04:38:54Z phase 1 done
2026-10-03T04:38:54Z ## Phase 2: day 0
broadcast output (strided_new):
```
{"height":"0","txhash":"5BC186FD60E5D5024BB9A84AD800CF90546202DDCB2DFEB0C9C0712EC7C98114","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:39:40Z tx 5BC186FD60E5D5024BB9A84AD800CF90546202DDCB2DFEB0C9C0712EC7C98114 code=1568 failed to execute message; message index: 0: epoch 10 record for osmosis-test-1 is UNBONDING_RETRY_QUEUE with 121439577; wait for the day epoch to submit it before draining: host zone has an unbonding record queued or retrying
2026-10-03T04:39:40Z CHECKPOINT PASS: drain refused while retry record (osmo)
2026-10-03T04:39:46Z CHECKPOINT PASS: non-admin refresh rejected
2026-10-03T04:39:53Z osmo val3 recorded delegation before refresh: 101371317
broadcast output (strided_new):
```
{"height":"0","txhash":"B5707BAA7645BCD5FA6977C3A9C1612C2A8478163FE653069BAEB28866293A71","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:40:39Z tx B5707BAA7645BCD5FA6977C3A9C1612C2A8478163FE653069BAEB28866293A71 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"4AA683BBA18273F9619F36D3F7C23E6786EE67FD983FAD5D67129E22D6ED2271","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:41:13Z tx 4AA683BBA18273F9619F36D3F7C23E6786EE67FD983FAD5D67129E22D6ED2271 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"E2A114BA66E03DD18590229728874A4F0DE085EEE5AA6B0DF35645FE62AEA680","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:41:49Z tx E2A114BA66E03DD18590229728874A4F0DE085EEE5AA6B0DF35645FE62AEA680 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"08C0590CF5689059A550F10A2FA6C0D8EE9F044517BE76DF7946B347AE61E1E9","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:42:21Z tx 08C0590CF5689059A550F10A2FA6C0D8EE9F044517BE76DF7946B347AE61E1E9 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"AD7329B4F3207DC0C524E19C30549178AA83FD64683ADA9BE2C2741293B339E0","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:42:56Z tx AD7329B4F3207DC0C524E19C30549178AA83FD64683ADA9BE2C2741293B339E0 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"DEDD3C0194421BCB0A03B5BC5C66B5C14D99B5E979BA81CEE030807BCDDE5036","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:43:31Z tx DEDD3C0194421BCB0A03B5BC5C66B5C14D99B5E979BA81CEE030807BCDDE5036 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"4DDBB928D5B75684E49E87C2C77CC4DCB954A06CF4949BEFF7F0494213CF51F4","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:44:05Z tx 4DDBB928D5B75684E49E87C2C77CC4DCB954A06CF4949BEFF7F0494213CF51F4 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"4D54038C82478504AE7DF53E6D9DDD1C11142567F4F4A71DC057AA3D85B48E14","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:44:33Z tx 4D54038C82478504AE7DF53E6D9DDD1C11142567F4F4A71DC057AA3D85B48E14 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"A311958156248D46B7759FCA69EFCABDA497BBF98991C3F86A1DBAABFD1C567D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:45:01Z tx A311958156248D46B7759FCA69EFCABDA497BBF98991C3F86A1DBAABFD1C567D code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"3E123542502C1D82FF23D93EC7DF182DC111E091F345CC07D662A3543AFD5686","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:45:37Z tx 3E123542502C1D82FF23D93EC7DF182DC111E091F345CC07D662A3543AFD5686 code=0 
broadcast output (strided_new):
```
{"height":"0","txhash":"81D53BAF7D084A7E70F6AF8D7D7DDFC9AFC94D0A9BC784D40EEB6015CC6AB411","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:46:15Z tx 81D53BAF7D084A7E70F6AF8D7D7DDFC9AFC94D0A9BC784D40EEB6015CC6AB411 code=0 
2026-10-03T04:46:32Z ready: osmo val3 recorded delegation reduced by the slash
broadcast output (strided_new):
```
{"height":"0","txhash":"2F0C60F3C16752E6B69CFD799B6F861A710BEAA8B0E9A06B7F79AC3178310F03","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T04:47:04Z tx 2F0C60F3C16752E6B69CFD799B6F861A710BEAA8B0E9A06B7F79AC3178310F03 code=0 
2026-10-03T04:47:09Z ready: no slash query in flight
2026-10-03T04:47:09Z ### osmo validators after refresh
```
$ strided_new q stakeibc show-validators osmosis-test-1 -o json
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:47:10 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"validators":[{"name":"val1","address":"osmovaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6n3re8","weight":"10","delegation":"101375823","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val2","address":"osmovaloper17kht2x2ped6qytr2kklevtvmxpw7wq9r2mr7dy","weight":"10","delegation":"101375797","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val3","address":"osmovaloper1nnurja9zt97huqvsfuartetyjx63tc5z3qt4u4","weight":"10","delegation":"100357945","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"0.990003371406121338","delegation_changes_in_progress":"0","slash_query_in_progress":false}]}
```
2026-10-03T04:47:18Z CHECKPOINT PASS: osmo rate still frozen
2026-10-03T04:47:18Z ### drift
```
$ python3 /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/measure_delegation_drift.py --chain-id cosmoshub-test-1 --chain-id osmosis-test-1
=== cosmoshub-test-1 ===
  using REST endpoint: https://cosmoshub-api.internal.stridenet.co
  host delegations: 8
=== osmosis-test-1 ===
  using REST endpoint: https://osmosis-api.internal.stridenet.co
  host delegations: 3

Done. Wrote /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/drift.json and /Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal/scripts/wind-down/drift/report.md
```
2026-10-03T04:47:20Z CHECKPOINT PASS: zero over-recorded
2026-10-03T04:47:36Z tx CF93C7AF14CBC224442F8BA399030C03A16AA5EE4C34EACDBBBBD98EDBFA6DE5 code=0 
2026-10-03T04:47:42Z ### confirm-undelegation 16
```
$ strided_new tx staketia confirm-undelegation 16 CF93C7AF14CBC224442F8BA399030C03A16AA5EE4C34EACDBBBBD98EDBFA6DE5 --from st-operator --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:47:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:47:42 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:47:43 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 190221
{"height":"0","txhash":"8FCFA60CFEAB5C4F4D3DEE0FA0FE4F0FA809EA7526061D93DE3343CF693C7921","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:47:53Z ## Phase 3: last redemption cycle
2026-10-03T04:54:35Z ready: every hub/osmo unbonding CLAIMABLE
2026-10-03T04:54:40Z ### claim cosmoshub-test-1 14
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 14 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:54:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:54:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:54:42 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156336
{"height":"0","txhash":"1B569A20B3E20C0F8A46712B1CBF5EDAA0648EEF1BF24D43537EFA28C99620B2","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:54:51Z ### claim cosmoshub-test-1 15
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 15 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:54:52 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156336
{"height":"0","txhash":"16AC2B82C02EB66937643FA785833846F03127BD21567401C9625D416A65EDD3","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:55:01Z ### claim cosmoshub-test-1 16
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 16 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:55:01 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156345
{"height":"0","txhash":"76D2121F4674F4E593FBA788D535AD98423BF286BD5B961D97F6646CD7C42623","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:55:13Z ### claim cosmoshub-test-1 17
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 17 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:55:15 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156336
{"height":"0","txhash":"B9CAB581995F4DEAC98D62E3036D9C72302C53EC8A91FDE880D1937164D13876","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:55:25Z ### claim cosmoshub-test-1 9
```
$ strided_new tx stakeibc claim-undelegated-tokens cosmoshub-test-1 9 cosmos15lf3jnxe8k2r72hang5cm8ymkx6che7tjdeg8l --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:55:27 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 156187
{"height":"0","txhash":"98E528B8B25FD1F30DEF7EE76B2F70C18C718AE10144E8AF27478AFFEB61044C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:55:40Z ### claim osmosis-test-1 10
```
$ strided_new tx stakeibc claim-undelegated-tokens osmosis-test-1 10 osmo15lf3jnxe8k2r72hang5cm8ymkx6che7t6k2c3d --from user1 --keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:55:41 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
gas estimate: 150372
{"height":"0","txhash":"7E794756852563E45A5EF579D32E9D5C2339CDA6E2DBBFA221D29EC8A0240CEF","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}
```
2026-10-03T04:56:54Z ready: zero user redemption records
2026-10-03T04:56:54Z ### deposit records
```
$ strided_new q records list-deposit-record -o json
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 04:56:55 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"deposit_record":[{"id":"159","amount":"10002841","denom":"uatom","host_zone_id":"cosmoshub-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"68","source":"STRIDE","delegation_txs_in_progress":"0"},{"id":"164","amount":"152669","denom":"uatom","host_zone_id":"cosmoshub-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"69","source":"WITHDRAWAL_ICA","delegation_txs_in_progress":"0"},{"id":"165","amount":"504","denom":"uosmo","host_zone_id":"osmosis-test-1","status":"DELEGATION_QUEUE","deposit_epoch_number":"69","source":"WITHDRAWAL_ICA","delegation_txs_in_progress":"0"}],"pagination":{"next_key":null,"total":"0"}}
```
2026-10-03T04:57:02Z CHECKPOINT PASS: stranded deposit never staked
2026-10-03T05:00:33Z CHECKPOINT PASS: hub rate frozen
2026-10-03T05:00:36Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T05:00:40Z CHECKPOINT PASS: no new epoch unbonding records
2026-10-03T05:00:41Z CHECKPOINT PASS: no reinvest/claim-rewards ICA
2026-10-03T05:00:41Z ## Phase 4: admin drain
2026-10-03T05:00:44Z ready: hub flags clear
2026-10-03T05:00:48Z ready: osmo flags clear
2026-10-03T05:00:52Z STATOM_SUPPLY=840053772
broadcast output (strided_new):
```
{"height":"0","txhash":"07F8523526A6FD4E1F003737E637B7FD8FCC9A0D68BE50C0DFA11DEDBC70AA39","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:01:35Z tx 07F8523526A6FD4E1F003737E637B7FD8FCC9A0D68BE50C0DFA11DEDBC70AA39 code=0 
2026-10-03T05:01:44Z ready: live-test ack
2026-10-03T05:01:47Z CHECKPOINT PASS: nothing burned
2026-10-03T05:02:32Z ready: hub val7 jailed
broadcast output (strided_new):
```
{"height":"0","txhash":"DBE78F67E85B19FA88234C901B10CC87374F23146F97788B19B06AA5E068D315","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:04:35Z tx DBE78F67E85B19FA88234C901B10CC87374F23146F97788B19B06AA5E068D315 code=0 
2026-10-03T05:04:50Z ready: drain acks
2026-10-03T05:04:50Z ### hub validators after --all
```
$ strided_new q stakeibc show-validators cosmoshub-test-1 -o json
2026/10/03 05:04:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pairs
2026/10/03 05:04:50 proto: duplicate proto type registered: cosmos.store.internal.kv.v1beta1.Pair
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Snapshot
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.Metadata
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotItem
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotStoreItem
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotIAVLItem
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionMeta
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.snapshots.v1.SnapshotExtensionPayload
error with code store:2 is already registered: invalid proof. Overwriting with current error...
error with code store:3 is already registered: tx parse error. Overwriting with current error...
error with code store:4 is already registered: unknown request. Overwriting with current error...
error with code store:5 is already registered: internal logic error. Overwriting with current error...
error with code store:6 is already registered: conflict. Overwriting with current error...
error with code store:7 is already registered: invalid request. Overwriting with current error...
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitInfo
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreInfo
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.v1beta1.CommitID
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.v1beta1.StoreKVPair
2026/10/03 05:04:51 proto: duplicate proto type registered: cosmos.store.v1beta1.BlockMetadata
service cosmos.poa.v1.Msg does not have cosmos.msg.v1.service proto annotation
{"validators":[{"name":"val1","address":"cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val2","address":"cosmosvaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rarvcqz","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val3","address":"cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val4","address":"cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val5","address":"cosmosvaloper1c5jnf370kaxnv009yhc3jt27f549l5u3u8rtpy","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val6","address":"cosmosvaloper1aujahjxj5ltuulsqgran68pcqg9vh5xj5e7pl5","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val7","address":"cosmosvaloper1ugz37tvvmwc7v9draq4mkpesg7cggkpyp3nsy2","weight":"10","delegation":"110236649","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false},{"name":"val8","address":"cosmosvaloper1ajkzxp8ssy8y5h58prcke2ulmuwy6g5ydfnd7z","weight":"10","delegation":"0","slash_query_progress_tracker":"0","slash_query_checkpoint":"0","shares_to_tokens_rate":"1.000000000000000000","delegation_changes_in_progress":"0","slash_query_in_progress":false}]}
```
2026-10-03T05:04:55Z CHECKPOINT PASS: val7 batch failed, others drained
broadcast output (strided_new):
```
{"height":"0","txhash":"28C47E3856F667CAF08533B075835F66A93D0D1F6633FFF67FAC53E599E9DD78","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:05:17Z tx 28C47E3856F667CAF08533B075835F66A93D0D1F6633FFF67FAC53E599E9DD78 code=0 
2026-10-03T05:05:22Z ready: val7 recorded delegation reduced by the slash
broadcast output (strided_new):
```
{"height":"0","txhash":"6E3C608B6B73A30D1388C0EC13725AC0C1B6FA57F9E80425DBE495F8075F6BFC","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:05:48Z tx 6E3C608B6B73A30D1388C0EC13725AC0C1B6FA57F9E80425DBE495F8075F6BFC code=0 
2026-10-03T05:09:51Z TIMEOUT waiting for: val7 drained
2026-10-03T05:10:47Z FINDING (code/runbook, confirmed): staketia ConfirmUndelegation subtracts record native amounts (stToken x rate) from the stakeibc zone's TotalDelegations; with rate > 1 the staketia portion went 49,703 negative, so the final hub drain batch's ack (seq 379, val7) failed in the undelegate callback and wedged the delegation channel. Same invariant as run 2. Mainnet headroom: 156.87B portion vs 14.3B pending staketia confirmations.
2026-10-03T05:10:55Z tx 8C74C99A0CD508DC71FF7647632843FCC2C19BFE33D30E05EA725D7A0FE1B496 code=0 
2026-10-03T05:11:44Z ready: hub ack 379 relayed
2026-10-03T05:12:06Z ## Phase 4: admin drain
2026-10-03T05:12:06Z ## Phase 4 (resume at injection 2 after the val7 ack recovery)
broadcast output (strided_new):
```
{"height":"0","txhash":"B63AE4738FF17EDF64FBE059B2C683DB178F08C412742F8B3BB21C48FCE7C310","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:14:03Z tx B63AE4738FF17EDF64FBE059B2C683DB178F08C412742F8B3BB21C48FCE7C310 code=0 
2026-10-03T05:20:53Z TIMEOUT waiting for: osmo delegation channel closed
2026-10-03T05:21:16Z ## Phase 4: admin drain
2026-10-03T05:21:16Z ## Phase 4 (resume at injection 3; injection 2 did not trigger: the relayer relayed the packet during its 30s graceful shutdown - proven in run 2)
broadcast output (strided_new):
```
{"height":"0","txhash":"6A7D8B9F26ED5925C243EFF991534BA96784D4DD325667C097FD43CCB8DB773C","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:22:38Z tx 6A7D8B9F26ED5925C243EFF991534BA96784D4DD325667C097FD43CCB8DB773C code=0 
2026-10-03T05:22:42Z ready: offset ack
broadcast output (strided_new):
```
{"height":"0","txhash":"B6EE86F5F92B17BD768680B8B901FFF922FB7205386AFD1D47136FA5AE886C7D","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:23:09Z tx B6EE86F5F92B17BD768680B8B901FFF922FB7205386AFD1D47136FA5AE886C7D code=0 
2026-10-03T05:23:21Z ready: osmo drained
broadcast output (strided_new):
```
{"height":"0","txhash":"8F97B6DE2F55D1B7398AD5591452A7ACC58E881F2F2A12583FDBA758F46F2570","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:24:37Z tx 8F97B6DE2F55D1B7398AD5591452A7ACC58E881F2F2A12583FDBA758F46F2570 code=17 failed to execute message; message index: 0: unable to submit unbonding ICA for cosmoshub-test-1: timeout timestamp must be in the future
2026-10-03T05:24:37Z CHECKPOINT PASS: dead-window send fails
2026-10-03T05:24:42Z ready: no flags after dead window
2026-10-03T05:24:45Z CHECKPOINT PASS: hub total at dust
2026-10-03T05:24:49Z CHECKPOINT PASS: osmo total at dust
2026-10-03T05:24:52Z CHECKPOINT PASS: hub rate frozen
2026-10-03T05:24:56Z CHECKPOINT PASS: osmo rate frozen
2026-10-03T05:24:57Z ## Phase 5: transfers to Osmosis
2026-10-03T05:25:08Z transfer-from-ica cosmoshub-test-1 WITHDRAWAL 8480261uatom
broadcast output (strided_new):
```
{"height":"0","txhash":"2B04713CD22EAB8B50F2D4A05921A2D9DA536CB23CF506968DD2A64D931278BA","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:25:31Z tx 2B04713CD22EAB8B50F2D4A05921A2D9DA536CB23CF506968DD2A64D931278BA code=0 
2026-10-03T05:25:37Z transfer-from-ica osmosis-test-1 WITHDRAWAL 1853uosmo
broadcast output (strided_new):
```
{"height":"0","txhash":"393F4AF07F09752E9547D0F6F4345342D1B6B4478ED08ED178D81B9EDBEAFAC3","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:26:01Z tx 393F4AF07F09752E9547D0F6F4345342D1B6B4478ED08ED178D81B9EDBEAFAC3 code=0 
2026-10-03T05:26:09Z transfer-from-ica cosmoshub-test-1 FEE 16963uatom
broadcast output (strided_new):
```
{"height":"0","txhash":"D3ECE95A9EAD233F521338027458A489841B75F0A0FCAA1709DBE7573FDBD995","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:26:35Z tx D3ECE95A9EAD233F521338027458A489841B75F0A0FCAA1709DBE7573FDBD995 code=0 
2026-10-03T05:26:42Z transfer-from-ica osmosis-test-1 FEE 105uosmo
broadcast output (strided_new):
```
{"height":"0","txhash":"5ABAA184CDCA142822038644DA39A307CADCE8F60640758396F545DC3E2662A9","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:27:12Z tx 5ABAA184CDCA142822038644DA39A307CADCE8F60640758396F545DC3E2662A9 code=11 failed to execute message; message index: 0: unable to submit FEE ICA transfer for osmosis-test-1: unable to send ICA tx: failed to retrieve active channel on connection connection-1 for port icacontroller-osmosis-test-1.FEE: no active channel for this owner
2026-10-03T05:27:26Z FINDING (runbook): osmosis FEE ICA channel was closed (a v34 fee ICA timed out during an earlier stall), so MsgTransferFromIca FEE failed with 'no active channel'. Pre-transfer checklist: every ICA the transfer uses must have an OPEN channel; restore first.
broadcast output (strided_new):
```
{"height":"0","txhash":"DB92FD9B6069A2D613D205D3CF5144DF01FCC1F4DC3E43093949EB64C753B872","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:27:51Z tx DB92FD9B6069A2D613D205D3CF5144DF01FCC1F4DC3E43093949EB64C753B872 code=0 
2026-10-03T05:28:02Z ready: osmosis FEE ICA reopened
2026-10-03T05:28:02Z ## Phase 5: transfers to Osmosis
2026-10-03T05:28:11Z cosmoshub-test-1 WITHDRAWAL ICA is empty, nothing to transfer
2026-10-03T05:28:17Z osmosis-test-1 WITHDRAWAL ICA is empty, nothing to transfer
2026-10-03T05:28:23Z cosmoshub-test-1 FEE ICA is empty, nothing to transfer
2026-10-03T05:28:30Z transfer-from-ica osmosis-test-1 FEE 105uosmo
broadcast output (strided_new):
```
{"height":"0","txhash":"7257EC65CCD761AD40CEC416F93961679712B5821104FED5E880E9E37AA7EE79","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:28:53Z tx 7257EC65CCD761AD40CEC416F93961679712B5821104FED5E880E9E37AA7EE79 code=0 
2026-10-03T05:33:56Z TIMEOUT waiting for: hub tokens landed in vault as ATOM-on-Osmosis
2026-10-03T05:35:07Z ## Phase 5: transfers to Osmosis
2026-10-03T05:35:16Z cosmoshub-test-1 WITHDRAWAL ICA is empty, nothing to transfer
2026-10-03T05:35:22Z osmosis-test-1 WITHDRAWAL ICA is empty, nothing to transfer
2026-10-03T05:35:31Z cosmoshub-test-1 FEE ICA is empty, nothing to transfer
2026-10-03T05:35:37Z osmosis-test-1 FEE ICA is empty, nothing to transfer
2026-10-03T05:35:40Z ready: hub tokens landed in vault as ATOM-on-Osmosis
2026-10-03T05:35:42Z ready: osmo bank-send form landed
broadcast output (strided_new):
```
{"height":"0","txhash":"57AC8C3C778E8D28A0F7B82EFA95F96CDBB492CCCCF9EBE390A6B2CFBC73D498","codespace":"","code":0,"data":"","raw_log":"","logs":[],"info":"","gas_wanted":"0","gas_used":"0","tx":null,"timestamp":"","events":[]}

```
2026-10-03T05:36:18Z tx 57AC8C3C778E8D28A0F7B82EFA95F96CDBB492CCCCF9EBE390A6B2CFBC73D498 code=0 
2026-10-03T05:36:18Z sleeping 140s: the ICA-wrapped transfer's inner timeout is 2 x the 60s WindDownTransferTimeout (120s) plus margin
2026-10-03T05:43:45Z TIMEOUT waiting for: timed-out transfer refunded to the withdrawal ICA
2026-10-03T05:46:55Z ## Client recovery: hub 07-tendermint-1 (osmosis) expired during the phase-5 relayer pause
2026-10-03T05:47:00Z substitutes: hub[07-tendermint-2] osmosis[07-tendermint-1] (osmosis one unused)
2026-10-03T05:47:11Z tx FEE9B04B81F305E20B956068DE10B482D925ACDF2F236B7CBA3502259AB1DB4A code=0 
2026-10-03T05:49:03Z CHECKPOINT PASS (manual): timed-out transfer refunded to the hub withdrawal ICA (1000000uatom) after recovering the hub osmosis client via gov MsgRecoverClient (proposal #1 passed); FINDING (harness): a relayer scale-down/up cycle takes minutes and expires 204s clients - injections must keep pauses short
