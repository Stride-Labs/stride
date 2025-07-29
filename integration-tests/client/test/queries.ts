import { stride, StrideClient } from "stridejs";
import { Tendermint37Client } from "@cosmjs/tendermint-rpc";
import { QueryClient } from "@cosmjs/stargate";

type QueryGetHostZoneResponse = ReturnType<typeof stride.stakeibc.QueryGetHostZoneResponse.decode>;
type QueryDepositRecordByHostResponse = ReturnType<typeof stride.records.QueryDepositRecordByHostResponse.decode>;

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
