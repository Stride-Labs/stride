import { stride, StrideClient } from "stridejs";
import { CosmosClient } from "./types";
import { Tendermint37Client } from "@cosmjs/tendermint-rpc";
import { QueryClient } from "@cosmjs/stargate";
import { ModuleAccount } from "stridejs/dist/types/codegen/cosmos/auth/v1beta1/auth";
import { HostZone } from "stridejs/dist/types/codegen/stride/stakeibc/host_zone";
import { expect } from "vitest";
import {
  DepositRecord,
  EpochUnbondingRecord,
  HostZoneUnbonding,
  UserRedemptionRecord,
} from "stridejs/dist/types/codegen/stride/records/records";

/**
 * Queryes a module account address from the account name
 * @param client Stride client
 * @param name Module account name
 * @returns The module account address
 */
export async function moduleAddress(client: StrideClient, name: string): Promise<string> {
  return (
    (
      await client.query.cosmos.auth.v1beta1.moduleAccountByName({
        name,
      })
    ).account as ModuleAccount
  ).baseAccount?.address!;
}

/**
 * Returns the balance of an account
 * @param client The stride or cosmos client
 * @param address The address to query
 * @param denom The denom
 * @returns The balance as a bigint
 */
export async function getBalance({
  client,
  denom,
  address,
}: {
  client: StrideClient | CosmosClient;
  denom: string;
  address?: string;
}): Promise<bigint> {
  if (address == undefined) {
    address = client.address;
  }

  if (client instanceof StrideClient) {
    const { balance: { amount } = { amount: "0" } } = await client.query.cosmos.bank.v1beta1.balance({
      address,
      denom,
    });
    return BigInt(amount);
  } else {
    const balance = await client.query.bank.balance(address, denom);
    return BigInt(balance.amount);
  }
}

/**
 * Returns the total delegated balance from the host zone's delegation ICA account
 * @param stridejs The stride client
 * @param hostjs The host client
 * @param chainId The host chain ID
 * @returns The total delegated balance across all validators
 */
export async function getDelegatedBalance({
  stridejs,
  hostjs,
  chainId,
}: {
  stridejs: StrideClient;
  hostjs: CosmosClient;
  chainId: string;
}): Promise<bigint> {
  const {
    hostZone: { delegationIcaAddress },
  } = await stridejs.query.stride.stakeibc.hostZone({
    chainId: chainId,
  });
  const { delegationResponses } = await hostjs.query.staking.delegatorDelegations(delegationIcaAddress);

  const totalDelegations = delegationResponses.reduce((sum, response) => {
    return sum + BigInt(response.balance.amount);
  }, BigInt(0));

  return totalDelegations;
}

/**
 * Returns the balance of the redemption ICA account
 * @param stridejs The stride client
 * @param hostjs The host client
 * @param chainId The host chain ID
 * @returns The balance of the redemption account
 */
export async function getRedemptionAccountBalance({
  stridejs,
  hostjs,
  chainId,
}: {
  stridejs: StrideClient;
  hostjs: CosmosClient;
  chainId: string;
}): Promise<bigint> {
  const {
    hostZone: { redemptionIcaAddress, hostDenom },
  } = await stridejs.query.stride.stakeibc.hostZone({
    chainId: chainId,
  });
  return await getBalance({ client: hostjs, denom: hostDenom, address: redemptionIcaAddress });
}

/**
 * Returns the "total delegations" value on the host zone struct
 * This is the stride chain's understanding of the delegated balance
 * @param client The stride client
 * @param chainId The host zone chain ID
 */
export async function getHostZoneTotalDelegations({
  client,
  chainId,
}: {
  client: StrideClient;
  chainId: string;
}): Promise<bigint> {
  const {
    hostZone: { totalDelegations },
  } = await client.query.stride.stakeibc.hostZone({
    chainId: chainId,
  });
  return BigInt(totalDelegations);
}

/**
 * Queries a host zone at a specified height
 * @param client The stride client
 * @param chainId The chain ID of the host zone
 * @param blockHeight The block height to query
 * @returns
 */
export async function getHostZone({
  client,
  chainId,
  blockHeight = 0,
}: {
  client: StrideClient;
  chainId: string;
  blockHeight?: number;
}): Promise<HostZone> {
  const tmClient = await Tendermint37Client.connect(client.rpcEndpoint);
  const queryClient = new QueryClient(tmClient);

  const response = await queryClient.queryAbci(
    "/stride.stakeibc.Query/HostZone",
    stride.stakeibc.QueryGetHostZoneRequest.encode({ chainId }).finish(),
    blockHeight,
  );
  return stride.stakeibc.QueryGetHostZoneResponse.decode(response.value).hostZone;
}

/**
 * Queries all deposit records at a specified height
 * @param client The stride client
 * @param chainId The chain ID of the host zone
 * @param blockHeight The block height to query
 * @returns
 */
export async function getAllDepositRecords(
  client: StrideClient,
  chainId: string,
  blockHeight: number = 0,
): Promise<DepositRecord[]> {
  const tmClient = await Tendermint37Client.connect(client.rpcEndpoint);
  const queryClient = new QueryClient(tmClient);

  const response = await queryClient.queryAbci(
    "/stride.records.Query/DepositRecordByHost",
    stride.records.QueryDepositRecordByHostRequest.encode({ hostZoneId: chainId }).finish(),
    blockHeight,
  );
  return stride.records.QueryDepositRecordByHostResponse.decode(response.value).depositRecord;
}

/**
 * Fetches the latest deposit record with a given status
 * @param client The stride client
 * @param chainId The host zone chain ID
 * @param status The deposit record status to filter for
 * @param blockHeight The block height to search at
 * @returns The newest deposit record for the host zone that matches the criteria
 */
export async function getLatestDepositRecord({
  client,
  chainId,
  status,
  blockHeight = 0,
}: {
  client: StrideClient;
  chainId: string;
  status: any;
  blockHeight?: number;
}): Promise<DepositRecord> {
  const allDepositRecords = await getAllDepositRecords(client, chainId, blockHeight);
  const filteredRecords = allDepositRecords
    .filter((record) => record.status === status) // filter by status
    .sort((a, b) => Number(b.id - a.id)); // sort by ID

  expect(filteredRecords.length).to.be.greaterThan(0, `No deposit records with status: ${status.toString()}`);
  return filteredRecords[0];
}

/**
 * Queries all epoch unbonding records at a specified height
 * @param client The stride client
 * @param chainId The chain ID of the host zone
 * @param blockHeight The block height to query
 * @returns
 */
export async function getHostZoneUnbondingRecords({
  client,
  chainId,
  blockHeight = 0,
}: {
  client: StrideClient;
  chainId: string;
  blockHeight?: number;
}): Promise<Record<number, HostZoneUnbonding>> {
  const tmClient = await Tendermint37Client.connect(client.rpcEndpoint);
  const queryClient = new QueryClient(tmClient);

  const response = await queryClient.queryAbci(
    "/stride.records.Query/EpochUnbondingRecordAll",
    stride.records.QueryAllEpochUnbondingRecordRequest.encode({}).finish(),
    blockHeight,
  );

  const epochUnbondingRecords = stride.records.QueryAllEpochUnbondingRecordResponse.decode(
    response.value,
  ).epochUnbondingRecord;

  const hostZoneUnbondingRecords: Record<number, HostZoneUnbonding> = {};
  epochUnbondingRecords.forEach((epochUnbondingRecord) => {
    epochUnbondingRecord.hostZoneUnbondings.forEach((hostZoneUnbonding) => {
      if (hostZoneUnbonding.hostZoneId === chainId) {
        hostZoneUnbondingRecords[Number(epochUnbondingRecord.epochNumber)] = hostZoneUnbonding;
      }
    });
  });

  return hostZoneUnbondingRecords;
}

/**
 * Queries a specific host zone unbonding record
 * @param client The stride client
 * @param chainId The chain ID of the host zone
 * @param epochNumber The epoch unbonding record epoch number
 * @returns
 */
export async function getHostZoneUnbondingRecord({
  client,
  chainId,
  epochNumber,
}: {
  client: StrideClient;
  chainId: string;
  epochNumber: bigint;
}): Promise<HostZoneUnbonding> {
  const hostZoneUnbondings = (await client.query.stride.records.epochUnbondingRecord({ epochNumber: epochNumber }))
    .epochUnbondingRecord.hostZoneUnbondings;
  return hostZoneUnbondings.filter((record) => record.hostZoneId == chainId)[0];
}

/**
 * Fetches the latest host zone unbonding record with a given status
 * @param client The stride client
 * @param chainId The host zone chain ID
 * @param status The host zone unbonding status to filter for
 * @param blockHeight The block height to search at
 * @returns The newest deposit record for the host zone that matches the criteria
 */
export async function getLatestHostZoneUnbondingRecord({
  client,
  chainId,
  status,
  blockHeight = 0,
}: {
  client: StrideClient;
  chainId: string;
  status: any;
  blockHeight?: number;
}): Promise<{ epochNumber: bigint; hostZoneUnbonding: HostZoneUnbonding }> {
  const allHostZoneUnbondingRecords = await getHostZoneUnbondingRecords({ client, chainId, blockHeight });
  const filteredRecords = Object.entries(allHostZoneUnbondingRecords)
    .map(([epochNumber, record]) => ({ epochNumber: Number(epochNumber), ...record })) // add epochNumber to struct
    .filter((record) => record.status === status) // filter by status
    .sort((a, b) => b.epochNumber - a.epochNumber); // sort by epoch number

  expect(filteredRecords.length).to.be.greaterThan(0, `No unbonding records with status: ${status.toString()}`);
  const hostZoneUnbondingRecord = filteredRecords[0];

  return { epochNumber: BigInt(hostZoneUnbondingRecord.epochNumber), hostZoneUnbonding: hostZoneUnbondingRecord };
}

/**
 * Fetches a user redemption record
 * @param client The stride client
 * @param chainId The host zone chain ID
 * @param epochNumber The epoch number for the record
 * @param receiver The host address of where the redeemed tokens should go
 * @returns The user redemption record
 */
export async function getUserRedemptionRecord({
  client,
  chainId,
  epochNumber,
  receiver,
}: {
  client: StrideClient;
  chainId: string;
  epochNumber: bigint;
  receiver: string;
}): Promise<UserRedemptionRecord> {
  return (await client.query.stride.records.userRedemptionRecord({ id: `${chainId}.${epochNumber}.${receiver}` }))
    .userRedemptionRecord;
}
