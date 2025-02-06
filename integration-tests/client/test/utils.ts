import { SigningStargateClient } from "@cosmjs/stargate";
import { coinsFromString, EncodeObject, StrideClient } from "stridejs";
import { expect } from "vitest";
import { CosmosClient, isCosmosClient } from "./main.test";

/**
 * Waits for the chain to start by continuously sending transactions until .
 *
 * @param {StrideClient} client The Stride client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForChain(
  client: StrideClient | CosmosClient,
  denom: string,
): Promise<void> {
  // the best way to ensure a chain is up is to successfully send a tx

  while (true) {
    try {
      if (client instanceof StrideClient) {
        const tx = await client.signAndBroadcast(
          [
            client.types.cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
              fromAddress: client.address,
              toAddress: client.address,
              amount: coinsFromString(`1${denom}`),
            }),
          ],
          2,
        );

        if (tx.code === 0) {
          break;
        }
      } else if (isCosmosClient(client)) {
        const tx = await client.client.sendTokens(
          client.address,
          client.address,
          coinsFromString(`1${denom}`),
          2,
        );

        if (tx.code === 0) {
          break;
        }
      } else {
        throw new Error(`unknown client ${client}`);
      }
    } catch (e) {
      // signAndBroadcast might throw if the RPC is not up yet
      console.log(e);
    }
  }
}

/**
 * Submit a transaction and wait for it to be broadcasted and executed.
 *
 * @param {StrideClient} stridejs The Stride client instance.
 * @param {any[]} msgs The messages array.
 */
export async function submitTxAndExpectSuccess(
  stridejs: StrideClient,
  msgs: EncodeObject[],
): Promise<void> {
  const tx = await stridejs.signAndBroadcast(msgs, 2);

  if (tx.code !== 0) {
    console.error(tx.rawLog);
  }
  expect(tx.code).toBe(0);
}
