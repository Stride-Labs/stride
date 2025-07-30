import { stride, StrideClient } from "stridejs";
import { CosmosClient } from "./types";
import { Tendermint37Client } from "@cosmjs/tendermint-rpc";
import { QueryClient } from "@cosmjs/stargate";
import { ModuleAccount } from "stridejs/dist/types/codegen/cosmos/auth/v1beta1/auth";
import { HostZone } from "stridejs/dist/types/codegen/stride/stakeibc/host_zone";
import { expect } from "vitest";
import { DepositRecord } from "stridejs/dist/types/codegen/stride/records/records";

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
 * Returns the balance of an account
 * @param client The stride or cosmos client
 * @param delegator The address of the delegator
 * @returns The total delegated balance across all validators
 */
export async function getDelegatedBalance({
  client,
  delegator,
}: {
  client: StrideClient | CosmosClient;
  delegator: string;
}): Promise<bigint> {
  const { delegationResponses } =
    client instanceof StrideClient
      ? await client.query.cosmos.staking.v1beta1.delegatorDelegations({
          delegatorAddr: delegator,
        })
      : await client.query.staking.delegatorDelegations(delegator);

  return delegationResponses.reduce((sum, response) => {
    return sum + BigInt(response.balance.amount);
  }, BigInt(0));
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
