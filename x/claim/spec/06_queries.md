<!--
order: 6
-->

# Queries

## GRPC queries

Claim module provides below GRPC queries to query claim status

```protobuf
service Query {
  rpc DistributorAccountBalance(QueryDistributorAccountBalanceRequest) returns (QueryDistributorAccountBalanceResponse) {}
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse) {}
  rpc ClaimRecord(QueryClaimRecordRequest) returns (QueryClaimRecordResponse) {}
  rpc ClaimableForAction(QueryClaimableForActionRequest) returns (QueryClaimableForActionResponse) {}
  rpc TotalClaimable(QueryTotalClaimableRequest) returns (QueryTotalClaimableResponse) {}
}
```

## CLI commands

For the following commands, you can change `$(strided keys show -a {your key name})` with the address directly.

Query the claim record for a given address

```sh
strided query claim claim-record $(strided keys show -a {your key name})
```

Query the claimable amount that would be earned if a specific action is completed right now.

```sh

strided query claim claimable-for-action $(strided keys show -a {your key name}) ActionAddLiquidity
```

Query the total claimable amount that would be earned if all remaining actions were completed right now.

```sh
strided query claim total-claimable $(strided keys show -a {your key name}) ActionAddLiquidity
```
