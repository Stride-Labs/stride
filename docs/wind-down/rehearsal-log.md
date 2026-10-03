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
