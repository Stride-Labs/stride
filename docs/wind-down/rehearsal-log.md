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
