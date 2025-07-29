import { stride, StrideClient } from "stridejs";
import { CosmosClient } from "./types";
import { Tendermint37Client } from "@cosmjs/tendermint-rpc";
import { QueryClient } from "@cosmjs/stargate";
import { ModuleAccount } from "stridejs/dist/types/codegen/cosmos/auth/v1beta1/auth";

type QueryGetHostZoneResponse = ReturnType<typeof stride.stakeibc.QueryGetHostZoneResponse.decode>;
type QueryDepositRecordByHostResponse = ReturnType<typeof stride.records.QueryDepositRecordByHostResponse.decode>;

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
export async function getHostZone(
  client: StrideClient,
  chainId: string,
  blockHeight: number = 0,
): Promise<QueryGetHostZoneResponse> {
  const tmClient = await Tendermint37Client.connect(client.rpcEndpoint);
  const queryClient = new QueryClient(tmClient);

  const response = await queryClient.queryAbci(
    "/stride.stakeibc.Query/HostZone",
    stride.stakeibc.QueryGetHostZoneRequest.encode({ chainId }).finish(),
    blockHeight,
  );
  return stride.stakeibc.QueryGetHostZoneResponse.decode(response.value);
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
): Promise<QueryDepositRecordByHostResponse> {
  const tmClient = await Tendermint37Client.connect(client.rpcEndpoint);
  const queryClient = new QueryClient(tmClient);

  const response = await queryClient.queryAbci(
    "/stride.records.Query/DepositRecordByHost",
    stride.records.QueryDepositRecordByHostRequest.encode({ hostZoneId: chainId }).finish(),
    blockHeight,
  );
  return stride.records.QueryDepositRecordByHostResponse.decode(response.value);
}
