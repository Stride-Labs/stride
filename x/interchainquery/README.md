---
title: "Interchainquery"
excerpt: ""
category: 6392913957c533007128548e
---

<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# Interchain Query

## Abstract

Stride uses interchain queries and interchain accounts to perform multichain liquid staking. The `interchainquery` module creates a framework that allows other modules to query other appchains using IBC. The `interchainquery` module is used to make bank balance ICQ queries to withdrawal account every N. The callback triggers ICA bank sends for 90% of the rewards to the delegation account and 10% to the stride hostzone revenue account. The ICA bank send logic is in x/stakeibc/keeper/callbacks.go.

## Contents

1. **[Concepts](#concepts)**
2. **[State](#state)**
3. **[Events](#events)**
4. **[Keeper](#keeper)**
5. **[Msgs](#msgs)**  
6. **[Queries](#queries)**

## State

The `interchainquery` module keeps `Query` objects and modifies the information from query to query, as defined in `proto/interchainquery/v1/genesis.proto`

### InterchainQuery information type

`Query` has information types that pertain to the query itself. `Query` keeps the following:

1. `id`: query identification string.
2. `connection_id`: id of the connection between the controller and host chain.
3. `chain_id`: id of the queried chain.
4. `query_type`: type of interchain query (e.g. bank store query)
5. `request_data`: serialized request information (e.g. the address with which to query)
6. `callback_module`: name of the module that will handle the callback
7. `callback_id`: ID for the function that will be called after the response is returned
8. `callback_data`: optional serialized data associated with the callback
9. `timeout_policy`: specifies how to handle a timeout (fail the query, retry the query, or execute the callback with a timeout)
10. `timeout_duration`: the relative time from the current block with which the query should timeout
11. `timeout_timestamp`: the absolute time at which the query times out
12. `request_sent`: boolean indicating whether the query event has been emitted (and can be identified by a relayer)
13. `submission_height`: the light client hight of the queried chain at the time of query submission


`DataPoint` has information types that pertain to the data that is queried. `DataPoint` keeps the following:

1. `id` keeps the identification string of the datapoint
2. `remote_height` keeps the block height of the queried chain
3. `local_height` keeps the block height of the querying chain
4. `value` keeps the bytecode value of the data retrieved by the Query

## Events

The `interchainquery` module emits an event at the end of every `stride_epoch`s (e.g. 15 minutes on local testnet).

The purpose of this event is to send interchainqueries that query data about staking rewards, which Stride uses to reinvest (aka autocompound) staking rewards.

```go
   event := sdk.NewEvent(
    sdk.EventTypeMessage,
    sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
    sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
    sdk.NewAttribute(types.AttributeKeyQueryId, queryInfo.Id),
    sdk.NewAttribute(types.AttributeKeyChainId, queryInfo.ChainId),
    sdk.NewAttribute(types.AttributeKeyConnectionId, queryInfo.ConnectionId),
    sdk.NewAttribute(types.AttributeKeyType, queryInfo.QueryType),
    sdk.NewAttribute(types.AttributeKeyHeight, "0"),
    sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(queryInfo.Request)),
   )
```

## Keeper

### Keeper Functions

`interchainquery/keeper/` module provides utility functions to manage ICQs

```go
// GetQuery returns query
GetQuery(ctx sdk.Context, id string) (types.Query, bool)
// SetQuery set query info
SetQuery(ctx sdk.Context, query types.Query)
// DeleteQuery delete query info
DeleteQuery(ctx sdk.Context, id string)
// IterateQueries iterate through queries
IterateQueries(ctx sdk.Context, fn func(index int64, queryInfo types.Query) (stop bool))
// AllQueries returns every queryInfo in the store
AllQueries(ctx sdk.Context) []types.Query
```

## Msgs

```protobuf
// SubmitQueryResponse is used to return the query response back to Stride
message MsgSubmitQueryResponse {
  string chain_id = 1;
  string query_id = 2;
  bytes result = 3;
  tendermint.crypto.ProofOps proof_ops = 4;
  int64 height = 5;
  string from_address = 6;
}
```

## Queries

```protobuf
// Query PendingQueries lists all queries that have been requested (i.e. emitted)
//  but have not had a response submitted yet
message QueryPendingQueriesRequest {}
```
