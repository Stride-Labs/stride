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
