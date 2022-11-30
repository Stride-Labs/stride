<!--
order: 2
-->

# State

### Claim Records

```protobuf
// A Claim Records is the metadata of claim data per address
message ClaimRecord {
  // address of claim user
  string address = 1 [ (gogoproto.moretags) = "yaml:\"address\"" ];

  // weight that represent the portion from total allocations
  double weight = 2;

  // true if action is completed
  // index of bool in array refers to action enum #
  repeated bool action_completed = 3 [
    (gogoproto.moretags) = "yaml:\"action_completed\""
  ];
}
```
When a user get airdrop for his/her action, claim record is created to prevent duplicated actions on future actions.

### State

```protobuf
message GenesisState {
  // params defines all the parameters of the module.
  Params params = 2 [
    (gogoproto.moretags) = "yaml:\"params\"",
    (gogoproto.nullable) = false
  ];

  // list of claim records, one for every airdrop recipient
  repeated ClaimRecord claim_records = 3 [
    (gogoproto.moretags) = "yaml:\"claim_records\"",
    (gogoproto.nullable) = false
  ];
}
```

Claim module's state consists of `params`, and `claim_records`.
