import { DeliverTxResponse, SigningStargateClient, StdFee } from "@cosmjs/stargate";
import {
  coinsFromString,
  convertBech32Prefix,
  cosmos,
  EncodeObject,
  getTxIbcResponses,
  getValueFromEvents,
  IbcResponse,
  stride,
  StrideClient,
} from "stridejs";
import { ModuleAccount } from "stridejs/dist/types/codegen/cosmos/auth/v1beta1/auth";
import { ibc } from "stridejs";
import { expect } from "vitest";
import { DEFAULT_FEE, DEFAULT_GAS, REMOVED, TRANSFER_CHANNEL, TRANSFER_PORT, USTRD } from "./consts";
import { newTransferMsg } from "./msgs";
import { Chain, CosmosClient } from "./types";
import { sleep } from "stridejs";
import { QueryGetHostZoneResponse } from "stridejs/dist/types/codegen/stride/stakeibc/query";
import { DepositRecord_Status } from "stridejs/dist/types/codegen/stride/records/records";

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

/**
 * Waits for the chain to start by continuously sending transactions until .
 *
 * @param {StrideClient | CosmosClient} client The client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForChain(client: StrideClient | CosmosClient, denom: string): Promise<void> {
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

  let ibcAck = isStrideClient ? await tx.ibcResponses[0] : await getTxIbcResponses(client.client, tx, 30_000, 50)[0];

  expect(ibcAck.type).toBe("ack");
  expect(ibcAck.tx.code).toBe(0);

  expect(getValueFromEvents(ibcAck.tx.events, "fungible_token_packet.success")).toBe("\u0001");
}

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
 * Wait for a balance to change (increase from initial value)
 */
export async function waitForBalanceChange({
  client,
  address,
  denom,
  initialBalance,
  minChange = 0,
  maxAttempts = 60,
}: {
  client: StrideClient | CosmosClient;
  address: string;
  denom: string;
  initialBalance?: bigint;
  minChange?: number;
  maxAttempts?: number;
}): Promise<bigint> {
  let attempts = 0;
  let prevBalance = initialBalance === undefined ? await getBalance({ client, address, denom }) : initialBalance;

  while (attempts < maxAttempts) {
    const currBalance = await getBalance({ client, address, denom });
    if (bigIntAbs(currBalance - prevBalance) >= BigInt(minChange)) {
      return currBalance;
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for balance change for ${denom} at ${address}`);
}

/**
 * Wait for a delegation to occur on the host zone
 * @param client The cosmos client of the host zone
 * @param delegator The delegator's address
 * @param minChange The minimum change before returning a success
 * @param maxAttempts The max number of attempts to try, each spaced by a second
 */
export async function waitForDelegationChange({
  client,
  delegator,
  minChange = 0,
  maxAttempts = 60,
}: {
  client: StrideClient | CosmosClient;
  delegator: string;
  minChange?: number;
  maxAttempts?: number;
}): Promise<bigint> {
  let attempts = 0;
  let prevBalance = await getDelegatedBalance({ client, delegator });

  while (attempts < maxAttempts) {
    const currBalance = await getDelegatedBalance({ client, delegator });
    if (bigIntAbs(currBalance - prevBalance) >= BigInt(minChange)) {
      return currBalance;
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for delegated balance change at ${delegator}`);
}

/**
 * Wait for the deposit record to change to status DELEGATION_QUEUE
 */
export async function waitForDepositRecordStatus({
  client,
  depositRecordId,
  status,
  maxAttempts = 60,
}: {
  client: StrideClient;
  depositRecordId: bigint;
  status: any;
  maxAttempts?: number;
}): Promise<void> {
  let attempts = 0;

  while (attempts < maxAttempts) {
    // If we're checking that the record was removed, query all records and check that the ID is not found
    if (status === REMOVED) {
      const { depositRecord } = await client.query.stride.records.depositRecordAll();
      if (depositRecord.filter((record) => record.id == depositRecordId).length == 0) {
        return;
      }
    } else {
      // Otherwise, if we're checking for a status, query the record by ID and check the status
      const { depositRecord } = await client.query.stride.records.depositRecord({ id: depositRecordId });
      if (depositRecord.status === status) {
        return;
      }
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for delegation record to change to status: ${status.toString()}`);
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
