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

## State

The `interchainquery` module keeps `Query` objects and modifies the information from query to query, as defined in `proto/interchainquery/v1/genesis.proto`

### InterchainQuery information type

`Query` has information types that pertain to the query itself. `Query` keeps the following:

1. `id` keeps the query identification string.
2. `connection_id` keeps the id of the channel or connection between the controller and host chain.
3. `chain_id` keeps the id of the queried chain.
4. `query_type` keeps the type of interchain query
5. `request` keeps an bytecode encoded version of the interchain query
6. `period` TODO
7. `last_height` keeps the blockheight of the last block before the query was made
8. `callback_id` keeps the function that will be called by the interchain query
9. `ttl` TODO
10. `height` keeps the height at which the ICQ query should execute on the host zone. This is often `0`, meaning the query should execute at the latest height on the host zone.

`DataPoint` has information types that pertain to the data that is queried. `DataPoint` keeps the following:

1. `id` keeps the identification string of the datapoint
2. `remote_height` keeps the block height of the queried chain
3. `local_height` keeps the block height of the querying chain
4. `value` keeps the bytecode value of the data retrieved by the Query

## Events

The `interchainquery` module emits an event at the end of every 3 `stride_epoch`s (e.g. 15 minutes on local testnet).

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
				// TODO: add height to request type
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

`interchainquery` has a `Msg` service that passes messages between chains. 

```protobuf
service Msg {
  // SubmitQueryResponse defines a method for submiting query responses.
  rpc SubmitQueryResponse(MsgSubmitQueryResponse) returns (MsgSubmitQueryResponseResponse)
}
```

