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
5. **[Hooks](#hooks)**  
6. **[Queries](#queries)**  
7. **[Future Improvements](#future-improvements)**

## Concepts

Nearly all of Stride's functionality is built using interchain accounts (ICAs), which are a new functionality in Cosmos, and a critical component of IBC. ICAs allow accounts on Zone A to be controlled by Zone B. ICAs communicate with one another using Interchain Queries (ICQs), which involve Zone A querying Zone B for relevant information. 

Two Zones communicate via a connection, or channel. All communications between the Controller Zone (the chain that is querying) and the Host Zone (the chain that is being queried) is done through a dedicated IBC channel between the two chains, which is opened the first time the two chains interact.


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
10. `height` keeps the number of blocks delay between the ICQ and the callback function being called. This is often `0`, meaning the callback function should be called immediately.

`DataPoint` has information types that pertain to the data that is queried. `DataPoint` keeps the following:

1. `id` keeps the identification string of the datapoint
2. `remote_height` keeps the block height of the queried chain
3. `local_height` keeps the block height of the querying chain
4. `value` keeps the bytecode value of the data retrieved by the Query

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