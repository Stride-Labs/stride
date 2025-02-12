import {
  DeliverTxResponse,
  SigningStargateClient,
  Event,
} from "@cosmjs/stargate";
import { coinsFromString, EncodeObject, StrideClient } from "stridejs";
import { expect } from "vitest";
import { Chain, CosmosClient } from "./main.test";
import { IbcResponse } from "stridejs";
import { newTransferMsg } from "./msgs";
import { TRANSFER_CHANNEL } from "./main.test";
import { getTxIbcResponses } from "stridejs";

export function isCosmosClient(client: any): client is CosmosClient {
  return (
    "address" in client &&
    "client" in client &&
    client.client instanceof SigningStargateClient
  );
}

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
 * @param client The Stride or cosmos client
 * @param msgs The message or messages array
 * @param signer The address of the signer. Only required for cosmos client
 */
export async function submitTxAndExpectSuccess(
  client: StrideClient | CosmosClient,
  msgs: EncodeObject | EncodeObject[],
): Promise<
  DeliverTxResponse & {
    ibcResponses: Array<Promise<IbcResponse>>;
  }
> {
  msgs = Array.isArray(msgs) ? msgs : [msgs];

  if (isCosmosClient(client)) {
    const tx = await client.client.signAndBroadcast(client.address, msgs, 2);

    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    return {
      ...tx,
      ibcResponses: getTxIbcResponses(client.client, tx, 30_000, 50),
    };
  } else {
    const tx = await client.signAndBroadcast(msgs, 2);

    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    return tx;
  }
}

/**
 * Executes an IBC transfer between chains
 * @param stridejs The Stride client instance
 * @param signingClient The signing client (either Stride or Cosmos client)
 * @param sourceChain The source chain for the transfer
 * @param destinationChain The destination chain for the transfer
 * @param coins The amount and denomination of coins to transfer as a string (e.g. "1000ustrd")
 * @param sender The address of the sender
 * @param receiver The address of the receiver
 */
export async function transfer({
  stridejs,
  signingClient,
  sourceChain,
  destinationChain,
  coins,
  sender,
  receiver,
}: {
  stridejs: StrideClient;
  signingClient: StrideClient | CosmosClient;
  sourceChain: Chain;
  destinationChain: Chain;
  sender: string;
  receiver: string;
  coins: string;
}) {
  const msg = newTransferMsg({
    stridejs: stridejs,
    channelId: TRANSFER_CHANNEL[sourceChain][destinationChain]!,
    coins: coins,
    sender: sender,
    receiver: receiver,
  });

  const tx = await submitTxAndExpectSuccess(signingClient, msg);

  const isStrideClient = "signingStargateClient" in signingClient;

  let ibcAck = isStrideClient
    ? await tx.ibcResponses[0]
    : await getTxIbcResponses(signingClient.client, tx, 30_000, 50)[0];

  expect(ibcAck.type).toBe("ack");
  expect(ibcAck.tx.code).toBe(0);
}

export function getValueFromEvents(
  events: readonly Event[],
  key: string,
): string {
  for (const e of events) {
    for (const a of e.attributes) {
      if (`${e.type}.${a.key}` === key) {
        return String(a.value);
      }
    }
  }

  throw new Error(`Event ${key} isn't in ${JSON.stringify(events)}`);
}
