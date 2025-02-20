import { IbcClient, Link } from "@confio/relayer";
import { OfflineSigner } from "@cosmjs/proto-signing";
import {
  DeliverTxResponse,
  GasPrice,
  SigningStargateClient,
} from "@cosmjs/stargate";
import {
  coinsFromString,
  convertBech32Prefix,
  cosmos,
  EncodeObject,
  getTxIbcResponses,
  getValueFromEvents,
  IbcResponse,
  sleep,
  StrideClient,
} from "stridejs";
import { ModuleAccount } from "stridejs/dist/types/codegen/cosmos/auth/v1beta1/auth";
import { expect } from "vitest";
import { TRANSFER_CHANNEL } from "./consts";
import { newTransferMsg } from "./msgs";
import { Chain, CosmosClient } from "./types";

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
 * @param {StrideClient | CosmosClient} client The client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForChain(
  client: StrideClient | CosmosClient,
  denom: string,
): Promise<void> {
  // the best way to ensure a chain is up is to successfully send a tx

  const msg = cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
    fromAddress: client.address,
    toAddress: client.address,
    amount: coinsFromString(`1${denom}`),
  });

  while (true) {
    let tx: DeliverTxResponse;

    try {
      if (client instanceof StrideClient) {
        tx = await client.signAndBroadcast([msg], 2);
      } else if (isCosmosClient(client)) {
        tx = await client.client.signAndBroadcast(client.address, [msg], 2);
      } else {
        throw new Error(`unknown client ${client}`);
      }

      if (tx.code === 0) {
        break;
      }
    } catch (e) {
      // signAndBroadcast might throw if the RPC is not up yet
      console.log(e);
    }
  }
}

/**
 * Waits for the chain to start by continuously sending transactions until .
 *
 * @param {StrideClient | CosmosClient} client The client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForIbc(
  client: StrideClient | CosmosClient,
  channelId: string,
  denom: string,
  receiverPrefix: string,
): Promise<void> {
  // the best way to ensure ibc is up is to successfully transfer funds

  const msg = newTransferMsg({
    channelId,
    coin: `1${denom}`,
    sender: client.address,
    receiver: convertBech32Prefix(client.address, receiverPrefix),
  });

  while (true) {
    let ibcAck: IbcResponse;

    try {
      if (client instanceof StrideClient) {
        const tx = await client.signAndBroadcast([msg], 2);
        if (tx.code === 0) {
          break;
        }
        ibcAck = await tx.ibcResponses[0];
      } else if (isCosmosClient(client)) {
        const tx = await client.client.signAndBroadcast(
          client.address,
          [msg],
          2,
        );
        if (tx.code === 0) {
          break;
        }

        ibcAck = await getTxIbcResponses(client.client, tx, 30_000, 50)[0];
      } else {
        throw new Error(`unknown client ${client}`);
      }

      expect(ibcAck.type).toBe("ack");
      expect(ibcAck.tx.code).toBe(0);

      expect(
        getValueFromEvents(ibcAck.tx.events, "fungible_token_packet.success"),
      ).toBe("\u0001");
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
  sourceChain: Chain;
  destinationChain: Chain;
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

  let ibcAck = isStrideClient
    ? await tx.ibcResponses[0]
    : await getTxIbcResponses(client.client, tx, 30_000, 50)[0];

  expect(ibcAck.type).toBe("ack");
  expect(ibcAck.tx.code).toBe(0);

  expect(
    getValueFromEvents(ibcAck.tx.events, "fungible_token_packet.success"),
  ).toBe("\u0001");
}

export async function moduleAddress(
  client: StrideClient,
  name: string,
): Promise<string> {
  return (
    (
      await client.query.cosmos.auth.v1beta1.moduleAccountByName({
        name,
      })
    ).account as ModuleAccount
  ).baseAccount?.address!;
}

export async function getIbcConnection(
  signerA: OfflineSigner,
  signerB: OfflineSigner,
  rpcA: string,
  rpcB: string,
  gasPriceA: GasPrice,
  gasPriceB: GasPrice,
  connIdA: string,
  connIdB: string,
): Promise<Link> {
  const [{ address: addressA }] = await signerA.getAccounts();
  const [{ address: addressB }] = await signerB.getAccounts();

  // Create IBC Client for chain A
  const clientA = await IbcClient.connectWithSigner(rpcA, signerA, addressA, {
    gasPrice: gasPriceA,
    estimatedBlockTime: 1000,
    estimatedIndexerTime: 500,
  });

  // Create IBC Client for chain B
  const clientB = await IbcClient.connectWithSigner(rpcB, signerB, addressB, {
    gasPrice: gasPriceB,
    estimatedBlockTime: 1000,
    estimatedIndexerTime: 500,
  });

  // Connect to the existing connection for the 2 clients
  return await Link.createWithExistingConnections(
    clientA,
    clientB,
    connIdA,
    connIdB,
  );
}

export async function loopRelayer(
  signerA: OfflineSigner,
  signerB: OfflineSigner,
  rpcA: string,
  rpcB: string,
  gasPriceA: GasPrice,
  gasPriceB: GasPrice,
  connIdA: string,
  connIdB: string,
) {
  const conn = await getIbcConnection(
    signerA,
    signerB,
    rpcA,
    rpcB,
    gasPriceA,
    gasPriceB,
    connIdA,
    connIdB,
  );

  while (true) {
    try {
      await conn.relayAll();

      await Promise.all([conn.updateClient("A"), conn.updateClient("B")]);
    } catch (e) {
      console.warn(`<loopRelayer>: caught error:`, e.message);
    }
    await sleep(500);
  }
}
