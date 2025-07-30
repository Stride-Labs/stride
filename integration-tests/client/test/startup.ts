import { DeliverTxResponse } from "@cosmjs/stargate";
import {
  coinsFromString,
  convertBech32Prefix,
  cosmos,
  getTxIbcResponses,
  getValueFromEvents,
  IbcResponse,
  StrideClient,
} from "stridejs";
import { ibc } from "stridejs";
import { expect } from "vitest";
import { TRANSFER_PORT } from "./consts";
import { newTransferMsg } from "./msgs";
import { CosmosClient } from "./types";
import { sleep } from "stridejs";
import { isCosmosClient } from "./utils";

/**
 * Waits for the chain to start by continuously sending transactions until it succeeds
 * The best way to ensure a chain is up is to successfully send a tx
 *
 * @param {StrideClient | CosmosClient} client The client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForChain(client: StrideClient | CosmosClient, denom: string): Promise<void> {
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
    await sleep(1000);
  }
}

// TODO: Deprecate in favor of assertOpenTransferChannel
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
        const tx = await client.client.signAndBroadcast(client.address, [msg], 2);
        if (tx.code === 0) {
          break;
        }

        ibcAck = await getTxIbcResponses(client.client, tx, 30_000, 50)[0];
      } else {
        throw new Error(`unknown client ${client}`);
      }

      expect(ibcAck.type).toBe("ack");
      expect(ibcAck.tx.code).toBe(0);

      expect(getValueFromEvents(ibcAck.tx.events, "fungible_token_packet.success")).toBe("\u0001");
    } catch (e) {
      // signAndBroadcast might throw if the RPC is not up yet
      console.log(e);
    }
  }
}

/**
 * Waits for the IBC channels to be open on either side
 *
 * @param {StrideClient | CosmosClient} client The client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function assertOpenTransferChannel(client: StrideClient | CosmosClient, channelId: string): Promise<void> {
  const stateOpen = ibc.core.channel.v1.State.STATE_OPEN;
  if (client instanceof StrideClient) {
    const { channel } = await client.query.ibc.core.channel.v1.channel({
      channelId: channelId,
      portId: TRANSFER_PORT,
    });
    expect(channel?.state).to.equal(stateOpen, `Stride transfer ${channelId} should be open`);
  } else {
    const channel = await client.query.ibc.channel.channel(TRANSFER_PORT, channelId);
    expect(channel?.channel?.state).to.equal(stateOpen, `Host transfer ${channelId} should be open`);
  }
}

/**
 * Waits for the ICA addresses to be flushed out on the host zone to confirm they're all opened
 * @param client Stride client
 * @param chainId The host chain id
 */
export async function assertICAChannelsOpen(
  stridejs: StrideClient,
  chainId: string,
  maxAttempts: number = 60,
): Promise<void> {
  let attempts = 0;

  while (attempts < maxAttempts) {
    const { hostZone } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: chainId,
    });

    if (
      hostZone.delegationIcaAddress != "" &&
      hostZone.withdrawalIcaAddress != "" &&
      hostZone.redemptionIcaAddress != "" &&
      hostZone.feeIcaAddress != ""
    ) {
      return;
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error("Timed out waiting for ICA channel opening");
}
