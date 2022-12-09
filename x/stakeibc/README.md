---
title: "The StakeIBC Module"
excerpt: ""
category: 62c5c5ff03a5bf069004def2
---

# The StakeIBC Module

The StakeIBC Module contains Stride's main app logic:
- it exposes core liquid staking entry points to the user (liquid staking and redeeming)
- it executes automated beginBlocker and endBlocker logic to stake funds on relevant host zones using Interchain Accounts  
- it handles registering new host zones and adjusting host zone validator sets and weights
- it defines Stride's core data structures (e.g. hostZone)
- it defines all the callbacks used when issuing Interchain Account logic 

## Params
```
DepositInterval (default uint64 = 1)
DelegateInterval (default uint64 = 1)
ReinvestInterval (default uint64 = 1)
RewardsInterval (default uint64 = 1)
RedemptionRateInterval (default uint64 = 1)
StrideCommission (default uint64 = 10)
ValidatorRebalancingThreshold (default uint64 = 100)
ICATimeoutNanos(default uint64 = 600000000000)
BufferSize (default uint64 = 5)  
IbcTimeoutBlocks (default uint64 = 300)
FeeTransferTimeoutNanos (default uint64 = 1800000000000
SafetyMinRedemptionRateThreshold (default uint64 = 90)     
SafetyMaxRedemptionRateThreshold (default uint64 = 150)         
MaxStakeICACallsPerEpoch (default uint64 = 100)
IBCTransferTimeoutNanos (default uint64 = 1800000000000)
SafetyNumValidators (default uint64 = 35)
```

## Keeper functions

- `LiquidStake()`
- `RedeemStake()`
- `ClaimUndelegatedTokens()`
- `RebalanceValidators()`
- `AddValidator()`
- `ChangeValidatorWeight()`
- `DeleteValidator()`
- `RegisterHostZone()`
- `ClearBalance()`
- `RestoreInterchainAccount()`
- `UpdateValidatorSharesExchRate()`

## State

Callbacks
- `SplitDelegation`  
- `DelegateCallback` 
- `ClaimCallback`    
- `ReinvestCallback` 
- `UndelegateCallback`   
- `RedemptionCallback`   
- `Rebalancing`  
- `RebalanceCallback`    

HostZone
- `HostZone`
- `ICAAccount`
- `MinValidatorRequirements`

Host Zone Validators
- `Validator`
- `ValidatorExchangeRate`

Misc
- `GenesisState`
- `EpochTracker`
- `Delegation`

Governance
- `AddValidatorProposal`

## Queries

- `QueryInterchainAccountFromAddressRequest`
- `QueryInterchainAccountFromAddressResponse`
- `QueryParamsRequest`
- `QueryParamsResponse`
- `QueryGetValidatorsRequest`
- `QueryGetValidatorsResponse`
- `QueryGetICAAccountRequest`
- `QueryGetICAAccountResponse`
- `QueryGetHostZoneRequest`
- `QueryGetHostZoneResponse`
- `QueryAllHostZoneRequest`
- `QueryAllHostZoneResponse`
- `QueryModuleAddressRequest`
- `QueryModuleAddressResponse`
- `QueryGetEpochTrackerRequest`
- `QueryGetEpochTrackerResponse`
- `QueryAllEpochTrackerRequest`
- `QueryAllEpochTrackerResponse`

## Events

`stakeibc` module emits the following events:

Type: Attribute Key &rarr; Attribute Value
--------------------------------------------------
registerHostZone: module &rarr; stakeibc
registerHostZone: connectionId &rarr; connectionId
registerHostZone: chainId &rarr; chainId
submitHostZoneUnbonding: hostZone &rarr;  chainId
submitHostZoneUnbonding: newAmountUnbonding &rarr; totalAmtToUnbond
stakeExistingDepositsOnHostZone: hostZone &rarr; chainId
stakeExistingDepositsOnHostZone: newAmountStaked &rarr; amount
onAckPacket (IBC): module &rarr;  moduleName
onAckPacket (IBC): ack &rarr; ackInfo



