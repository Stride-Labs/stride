<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# Interchain Query

## Abstract
Stride uses interchain queries and interchain accounts to perform multichain liquid staking. The `interchainquery` module creates a framework that allows other modules to query other appchains using IBC. 

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

Two Zones communicate via a connection, or channel. All communications between the Controller Zone (the chain that is controlling) and the Host Zone is done through a dedicated IBC channel between the two chains, which is opened the first time the two chains interact.


## State

The `interchainquery` module keeps `Query` objects and modifies the information from query to query, as defined in `proto/interchainquery/v1/genesis.proto`

### InterchainQuery information type

Query keeps the following:

1. `id` keeps the query identification string.
2. `connection_id` keeps 



## Keeper
### abci.go
Gives endblocker of Interchainquery module. No beginblocker – why?
### keeper.go 
The `keeper.go` file maintains the collections of registered zones, as well as the fu