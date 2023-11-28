---
title: "ICA Oracle"
excerpt: ""
category: 6392913957c533007128548e
---

# ICA Oracle Module

## Overview
The `icaoracle` facilities trustless data publication to cosmwasm outposts on adjacent chains. The full ICA Oracle solution consists of two components: the `icaoracle` module deployed on the _source_ chain (described here), as well as a corresponding cosmwasm oracle contract deployed on the _destination_ chain. The contract features a standard key-value store, that accepts push messages from this module via interchain accounts. The data sent is referred to as a `Metric`. For Stride, the primary application of this module is to enable integrations to trustlessly retrieve the redemption rate (internal exchange rate) of stTokens.

Some key features of this module are:
* **Trustless**: The solution uses interchain accounts which allows the data to be transmitted and retrieved with the same trust properties as IBC 
* **Easy adoption to other chains**: Since the destination chain's oracle is deployed as a contract, it can be easily added to any chain with cosmwasm. There is no need for those chains to upgrade.
* **Generic key-value store**: While the immediate use case of this module is for the redemption rate, the oracle uses a generic key-value format that should work with any metric
* **High throughput/low latency**: ICAs are triggered as soon as possible, which eliminates a hop compared to interchain queries
* **Support for multiple oracles**: Each metric can be simultaneously pushed to multiple CW outposts

### Setup 
Before data can be transmitted, there are a few setup steps required:
1. The contract must be stored on the cosmwasm chain (destination chain)
2. A connection must exist between the source and destination chain
3. The oracle must be added to the source chain using the `add-oracle` transaction. This transaction will begin the registration on the source chain and create an interchain account on the destination chain. The interchain account will be responsible for instantiating the contract and posting metrics.
4. After the oracle is added, the `instantiate-oracle` transaction must be submitted which will submit an interchain account message (`MsgInstantiateContract`) to instantiate the oracle contract with the interchain account's address as the contract admin.

### Pushing Metrics
After an oracle is registered, metrics can be posted on-chain using the `QueueMetricUpdate` function. This will queue the data so that it can be pushed to each registered oracle. In the `EndBlocker` after the metric is queued, an interchain account message (`MsgExecuteContract{MsgPostMetric}`) will be submitted to post the value to the oracle.

## Diagrams
### Setup
![alt text](https://github.com/Stride-Labs/stride/blob/main/x/icaoracle/docs/setup.png?raw=true)
### Pushing Metrics
![alt text](https://github.com/Stride-Labs/stride/blob/main/x/icaoracle/docs/pushing.png?raw=true)
### Metric Status
![alt text](https://github.com/Stride-Labs/stride/blob/main/x/icaoracle/docs/metric-status.png?raw=true)

## Implementation
### State
```go
Oracle
  ChainId string
  ConnectionId string
  ChannelId string
  PortId string
  ICAAddress string
  ContractAddress string
  Active bool

Metric
  Key string
  Value string
  MetricType string
  UpdateTime int64 
  BlockHeight int64 
  Attributes string
  DestinationOracle string
  Status (enum: QUEUED/IN_PROGRESS)
```

### Keeper functions
#### Oracles
```go
// Stores/updates an oracle object in the store
func SetOracle(oracle types.Oracle) 

// Grabs and returns an oracle object from the store using the chain-id
func GetOracle(chainId string) (oracle types.Oracle, found bool) 

// Returns all oracles
func GetAllOracles() []types.Oracle 

// Removes an oracle from the store
func RemoveOracle(chainId string) 

// Toggle whether an oracle is active
func ToggleOracle(chainId string, active bool) error 

// Grab's an oracle from it's connectionId
func GetOracleFromConnectionId(connectionId string) (oracle types.Oracle, found bool) 

// Checks if the oracle ICA channel is open
func IsOracleICAChannelOpen(oracle types.Oracle) bool 
```

#### Metrics
```go
// Stores a metric in the main metric store and then either
// adds the metric to the queue or removes it from the queue
// depending on the status of the metric
func SetMetric(metric types.Metric) 

// Gets a specifc metric from the store
func GetMetric(metricId string) (metric types.Metric, found bool) 

// Returns all metrics from the store
func GetAllMetrics() (metrics []types.Metric) 

// Removes a metric from the store
func RemoveMetric(metricId string) 

// Updates the status of a metric which will consequently move it either
// in or out of the queue
func UpdateMetricStatus(, metric types.Metric, status types.MetricStatus) 

// Adds a metric to the queue, which acts as an index for all metrics
// that should be submitted to it's relevant oracle
func addMetricToQueue(metricKey []byte)

// Removes a metric from the queue
func removeMetricFromQueue(, metricKey []byte) 

// Returns all metrics from the index queue
func GetAllQueuedMetrics() (metrics []types.Metric) 
```

### Transactions
```go
// Adds a new oracle
AddOracle(connectionId string)

// Instantiates the oracle's CW contract
InstantiateOracle(oracleChainId string, contractCodeId uint64)

// Restore's a closed ICA channel for a given oracle
RestoreOracleICA(oracleChainId string)

// Toggle's whether an oracle is active and should receive metric updates
ToggleOracle(oracleChainId string, active bool) [Governance]

// Removes an oracle completely
RemoveOracle(oracleChainId string) [Governance]
```

### Queries
```go
// Query a specific oracle
//   /Stride-Labs/stride/icaoracle/oracle/{chain_id}
Oracle(oracleChainId string)

// Query all oracles
//   /Stride-Labs/stride/icaoracle/oracles
AllOracles()

// Query metrics with optional filters
//
// Ex:
// - /Stride-Labs/stride/icaoracle/metrics
// - /Stride-Labs/stride/icaoracle/metrics?metric_key=X
// - /Stride-Labs/stride/icaoracle/metrics?oracle_chain_id=Y
Metrics(metricKey, oracleChainId string)
```

### Business Logic
```go
// Queues an metric update across each active oracle
// One metric record is created for each oracle, in status QUEUED
// This is called by the modules that want to publish metrics
func QueueMetricUpdate(key, value, metricType, attributes string) 

// For each queued metric, submit an ICA to each oracle, and then flag the metric as IN_PROGRESS
// This is called each block in the EndBlocker
func PostAllQueuedMetrics() 
```

### ICA Callbacks
```go
// Callback after an oracle is instantiated
func InstantiateOracleCallback()

// Callback after a metric is published
func UpdateOracleCallback()
```
