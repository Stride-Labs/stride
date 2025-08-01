import { DeliverTxResponse, getTxIbcResponses, getValueFromEvents, IbcResponse, sleep, StrideClient } from "stridejs";
import { expect } from "vitest";
import { newTransferMsg } from "./msgs";
import { CosmosClient } from "./types";
import { DEFAULT_FEE, DEFAULT_GAS, TRANSFER_CHANNEL, USTRD } from "./consts";
import { EncodeObject } from "@cosmjs/proto-signing";
import { isCosmosClient } from "./utils";
import { StdFee } from "@cosmjs/stargate";

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
export async function ibcTransfer({
  client,
  sourceChain,
  destinationChain,
  coin,
  sender,
  receiver,
  memo = "",
}: {
  client: StrideClient | CosmosClient;
  sourceChain: string;
  destinationChain: string;
  sender: string;
  receiver: string;
  coin: string;
  memo?: string;
}) {
  const msg = newTransferMsg({
    channelId: TRANSFER_CHANNEL[sourceChain][destinationChain]!,
    coin,
    sender,
    receiver,
    memo,
  });

  const tx = await submitTxAndExpectSuccess(client, msg);

  const isStrideClient = "signingStargateClient" in client;

  let ibcAck = isStrideClient ? await tx.ibcResponses[0] : await getTxIbcResponses(client.client, tx, 30_000, 50)[0];

  expect(ibcAck.type).toBe("ack");
  expect(ibcAck.tx.code).toBe(0);

  expect(getValueFromEvents(ibcAck.tx.events, "fungible_token_packet.success")).toBe("\u0001");
}

/**
 * Submit a transaction and wait for it to be broadcasted and executed.
 *
 * @param client The Stride or cosmos client
 * @param msgs The message or messages array
 * @param signer The address of the signer. Only required for cosmos client
 * @param fee Optional fee to use for the transaction
 */
export async function submitTxAndExpectSuccess(
  client: StrideClient | CosmosClient,
  msgs: EncodeObject | EncodeObject[],
  fee?: StdFee,
): Promise<
  DeliverTxResponse & {
    ibcResponses: Array<Promise<IbcResponse>>;
  }
> {
  msgs = Array.isArray(msgs) ? msgs : [msgs];

  const feeDenom = isCosmosClient(client) ? client.denom : USTRD;
  const defaultFee: StdFee = {
    amount: [{ amount: DEFAULT_FEE.toString(), denom: feeDenom }],
    gas: DEFAULT_GAS,
  };
  const txFee = fee || defaultFee;

  if (isCosmosClient(client)) {
    const tx = await client.client.signAndBroadcast(client.address, msgs, txFee);

    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);
    sleep(1500);

    return {
      ...tx,
      ibcResponses: getTxIbcResponses(client.client, tx, 30_000, 50),
    };
  } else {
    const tx = await client.signAndBroadcast(msgs, txFee);

    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);
    sleep(1500);

    return tx;
  }
}
