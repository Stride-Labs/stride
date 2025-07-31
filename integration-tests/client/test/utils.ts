import { SigningStargateClient } from "@cosmjs/stargate";
import { CosmosClient } from "./types";

/**
 * Returns the absolute value of a bigint
 * @param i The value to take the absolute value of
 */
export function bigIntAbs(i: bigint): bigint {
  return i < BigInt(0) ? -i : i;
}

/**
 * Returns true if a client is a cosmos client (rather than a Stride client)
 * @param client The Stride or Cosmos client
 */
export function isCosmosClient(client: any): client is CosmosClient {
  return "address" in client && "client" in client && client.client instanceof SigningStargateClient;
}
