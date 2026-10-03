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
