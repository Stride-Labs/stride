import { StrideClient } from "stridejs";
import { Tendermint37Client } from "@cosmjs/tendermint-rpc";
import { QueryGetHostZoneRequest, QueryGetHostZoneResponse } from "stridejs/dist/types/codegen/stride/stakeibc/query";
import { QueryClient } from "@cosmjs/stargate";

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
    QueryGetHostZoneRequest.encode({ chainId }).finish(),
    blockHeight,
  );
  return QueryGetHostZoneResponse.decode(response.value);
}
