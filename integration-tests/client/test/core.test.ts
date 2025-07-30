import {
  QueryClient,
  setupAuthExtension,
  setupBankExtension,
  setupIbcExtension,
  setupStakingExtension,
  setupTxExtension,
  SigningStargateClient,
} from "@cosmjs/stargate";
import { Comet38Client } from "@cosmjs/tendermint-rpc";
import { DirectSecp256k1HdWallet, GasPrice, ibcDenom, sleep, stride } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  STRIDE_RPC_ENDPOINT,
  USTRD,
  DEFAULT_FEE,
  REMOVED,
  DEFAULT_TRANSFER_CHANNEL_ID,
  DEFAULT_CONNECTION_ID,
  CHAIN_CONFIGS,
  TRANSFER_PORT,
  STRIDE_CHAIN_NAME,
  toStToken,
} from "./consts";
import { CosmosClient } from "./types";
import { ibcTransfer, submitTxAndExpectSuccess } from "./txs";
import { waitForChain, assertICAChannelsOpen, assertOpenTransferChannel } from "./startup";
import { waitForBalanceChange, waitForHostZoneTotalDelegationsChange, waitForDepositRecordStatus } from "./polling";
import { getBalance, getDelegatedBalance, getLatestDepositRecord } from "./queries";
import { StrideClient } from "stridejs";
import { Decimal } from "decimal.js";
import { getHostZone } from "./queries";
import { newRegisterHostZoneMsg, newValidator } from "./msgs";

const HOST_CHAIN_NAME = "cosmoshub";
const HOST_CONFIG = CHAIN_CONFIGS[HOST_CHAIN_NAME];
const HOST_CHAIN_ID = HOST_CONFIG.chainId;
const HOST_DENOM = HOST_CONFIG.hostDenom;
const ST_DENOM = toStToken(HOST_DENOM);

const HOST_DENOM_ON_STRIDE = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: HOST_CONFIG.transferChannelId,
    },
  ],
  HOST_DENOM,
);

const STRD_DENOM_ON_HOST = ibcDenom(
  [
    {
      incomingPortId: TRANSFER_PORT,
      incomingChannelId: HOST_CONFIG.transferChannelId,
    },
  ],
  USTRD,
);

// Initialize accounts
let strideAccounts: {
  user: StrideClient;
  admin: StrideClient;
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

let hostAccounts: {
  user: CosmosClient;
  val1: CosmosClient;
};

const mnemonics: {
  name: "user" | "admin" | "val1" | "val2" | "val3";
  mnemonic: string;
}[] = [
  {
    name: "user",
    mnemonic:
      "brief play describe burden half aim soccer carbon hope wait output play vacuum joke energy crucial output mimic cruise brother document rail anger leaf",
  },
  {
    name: "admin",
    mnemonic:
      "tone cause tribe this switch near host damage idle fragile antique tail soda alien depth write wool they rapid unfold body scan pledge soft",
  },
  {
    name: "val1",
    mnemonic:
      "close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly",
  },
  {
    name: "val2",
    mnemonic:
      "turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner",
  },
  {
    name: "val3",
    mnemonic:
      "tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat",
  },
];

// Initialize accounts and wait for the chain to start
beforeAll(async () => {
  console.log("setting up accounts...");

  // @ts-expect-error
  strideAccounts = {};
  // @ts-expect-error
  hostAccounts = {};

  for (const { name, mnemonic } of mnemonics) {
    // setup signer for Stride
    const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: STRIDE_CHAIN_NAME,
    });

    const [{ address }] = await signer.getAccounts();

    strideAccounts[name] = await StrideClient.create(STRIDE_RPC_ENDPOINT, signer, address, {
      gasPrice: GasPrice.fromString(`0.025${USTRD}`),
      broadcastPollIntervalMs: 50,
      resolveIbcResponsesCheckIntervalMs: 50,
    });

    if (name === "user" || name === "val1") {
      // setup signer for host zone
      const hostSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);
      const [{ address: hostAddress }] = await hostSigner.getAccounts();

      hostAccounts[name] = {
        address: hostAddress,
        denom: HOST_DENOM,
        client: await SigningStargateClient.connectWithSigner(HOST_CONFIG.rpcEndpoint, hostSigner, {
          gasPrice: GasPrice.fromString(`1.0${HOST_DENOM}`),
          broadcastPollIntervalMs: 50,
        }),
        query: QueryClient.withExtensions(
          await Comet38Client.connect(HOST_CONFIG.rpcEndpoint),
          setupAuthExtension,
          setupBankExtension,
          setupStakingExtension,
          setupIbcExtension,
          setupTxExtension,
        ),
      };
    }
  }

  console.log("waiting for stride to start...");
  await waitForChain(strideAccounts.user, USTRD);

  console.log("waiting for host to start...");
  await waitForChain(hostAccounts.user, HOST_DENOM);

  console.log("waiting for stride-host ibc...");
  await assertOpenTransferChannel(strideAccounts.user, DEFAULT_TRANSFER_CHANNEL_ID);
  await assertOpenTransferChannel(hostAccounts.user, DEFAULT_TRANSFER_CHANNEL_ID);

  console.log("registering host zones...");
  const { hostZone } = await strideAccounts.admin.query.stride.stakeibc.hostZoneAll({});
  const hostZoneNotRegistered = hostZone.find((hz) => hz.chainId === HOST_CHAIN_ID) === undefined;

  if (hostZoneNotRegistered) {
    const registerHostZoneMsg = newRegisterHostZoneMsg({
      sender: strideAccounts.admin.address,
      connectionId: DEFAULT_CONNECTION_ID,
      transferChannelId: DEFAULT_TRANSFER_CHANNEL_ID,
      hostDenom: HOST_DENOM,
      bechPrefix: HOST_CONFIG.bechPrefix,
    });

    const { validators } = await hostAccounts.user.query.staking.validators("BOND_STATUS_BONDED");
    const addValidatorsMsg = stride.stakeibc.MessageComposer.withTypeUrl.addValidators({
      creator: strideAccounts.admin.address,
      hostZone: HOST_CHAIN_ID,
      validators: validators.map((val) =>
        newValidator({
          name: val.description.moniker,
          address: val.operatorAddress,
          weight: 10n,
        }),
      ),
    });

    await submitTxAndExpectSuccess(strideAccounts.admin, [registerHostZoneMsg, addValidatorsMsg]);
    await sleep(2000);
  }

  console.log("waiting for ICA channels...");
  await assertICAChannelsOpen(strideAccounts.admin, HOST_CHAIN_ID);
}, 45_000);

describe("Core Tests", () => {
  test("IBC Transfer", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const transferAmount = BigInt(50000000);

    // Get initial balances
    // We'll send STRD from Stride -> Host Zone
    const initialStrdBalanceOnStride = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const initialStrdBalanceOnHost = await getBalance({
      client: hostjs,
      denom: STRD_DENOM_ON_HOST,
    });

    // As well as host denom (e.g. ATOM) from Host chain -> Stride
    const initialHostBalanceOnHost = await getBalance({
      client: hostjs,
      denom: HOST_DENOM,
    });

    const initialHostBalanceOnStride = await getBalance({
      client: stridejs,
      denom: HOST_DENOM_ON_STRIDE,
    });

    // Perform IBC transfers
    console.log("Transferring USTRD from Stride to host zone...");
    await ibcTransfer({
      client: stridejs,
      sourceChain: STRIDE_CHAIN_NAME,
      destinationChain: HOST_CHAIN_NAME,
      coin: `${transferAmount}${USTRD}`,
      sender: stridejs.address,
      receiver: hostjs.address,
    });

    console.log("Transferring native host token from host zone to Stride...");
    await ibcTransfer({
      client: hostjs,
      sourceChain: HOST_CHAIN_NAME,
      destinationChain: STRIDE_CHAIN_NAME,
      coin: `${transferAmount}${HOST_DENOM}`,
      sender: hostjs.address,
      receiver: stridejs.address,
    });

    // Wait a bit for transfers to complete.
    console.log("Waiting for transfers to complete...");
    await sleep(5000);

    // Get final balances
    const finalStrdBalanceOnStride = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const finalStrdBalanceOnHost = await getBalance({
      client: hostjs,
      denom: STRD_DENOM_ON_HOST,
    });

    const finalHostBalanceOnHost = await getBalance({
      client: hostjs,
      denom: HOST_DENOM,
    });

    const finalHostBalanceOnStride = await getBalance({
      client: stridejs,
      denom: HOST_DENOM_ON_STRIDE,
    });

    // Calculate and verify balance changes
    const strideStrdBalanceDiff = finalStrdBalanceOnStride - initialStrdBalanceOnStride;
    const strideHostBalanceDiff = finalHostBalanceOnStride - initialHostBalanceOnStride;
    const hostStrdBalanceDiff = finalStrdBalanceOnHost - initialStrdBalanceOnHost;
    const hostHostBalanceDiff = finalHostBalanceOnHost - initialHostBalanceOnHost;

    // Verify the transfers worked
    // STRD sent out from Stride → negative balance change + fee
    // STRD received on host → positive balance change
    expect(strideStrdBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Stride STRD balance change");
    expect(hostStrdBalanceDiff).to.equal(transferAmount, "Host STRD balance change");

    // Host denom sent out from host zone → negative balance change + fee
    // Host denom received on Stride → positive balance change
    expect(hostHostBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Host native balance change");
    expect(strideHostBalanceDiff).to.equal(transferAmount, "Stride host denom balance change");
  }, 120_000);

  test("Liquid Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;

    // Get initial balances on Stride
    let initialUserNativeBalance = await getBalance({
      client: stridejs,
      denom: HOST_DENOM_ON_STRIDE,
    });

    const initialUserStBalance = await getBalance({
      client: stridejs,
      denom: ST_DENOM,
    });

    // Get the initial delegated balance
    const {
      hostZone: { delegationIcaAddress },
    } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: HOST_CHAIN_ID,
    });
    const initialDelegatedBalance = await getDelegatedBalance({
      client: hostjs,
      delegator: delegationIcaAddress,
    });

    // Ensure there's enough native host tokens to liquid stake, if not transfer
    if (initialUserNativeBalance == BigInt(0)) {
      console.log("Transferring host zone token to Stride...");
      await ibcTransfer({
        client: hostjs,
        sourceChain: HOST_CHAIN_NAME,
        destinationChain: STRIDE_CHAIN_NAME,
        coin: `${stakeAmount}${HOST_DENOM}`,
        sender: hostjs.address,
        receiver: stridejs.address,
      });

      initialUserNativeBalance = await waitForBalanceChange({
        initialBalance: initialUserNativeBalance,
        client: stridejs,
        address: stridejs.address,
        denom: HOST_DENOM_ON_STRIDE,
      });
    }

    // Perform liquid staking
    console.log("Liquid staking...");
    const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
      creator: stridejs.address,
      amount: String(stakeAmount),
      hostDenom: HOST_DENOM,
    });

    const tx = await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
    await sleep(2000); // sleep to make sure block finalized

    // Get final native and st balances
    const finalUserStBalance = await getBalance({ client: stridejs, address: stridejs.address, denom: ST_DENOM });
    const finalUserNativeBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: HOST_DENOM_ON_STRIDE,
    });

    // Get the redemption rate at the time of the liquid stake
    const { redemptionRate } = await getHostZone({ client: stridejs, chainId: HOST_CHAIN_ID, blockHeight: tx.height });

    // Confirm the balance changes
    // Native balance should decrease (sent for staking)
    // StBalance should increase (minted)
    const nativeBalanceDiff = finalUserNativeBalance - initialUserNativeBalance;
    const stBalaanceDiff = finalUserStBalance - initialUserStBalance;
    const expectedStBalanceAmount = BigInt(
      Decimal(stakeAmount.toString()).div(Decimal(redemptionRate)).floor().toString(),
    );
    expect(stBalaanceDiff).to.equal(expectedStBalanceAmount, "User st balance change on Stride");
    expect(nativeBalanceDiff).to.equal(BigInt(-stakeAmount), "User native balance change on Stride");

    // Get the deposit record that was used for the liquid stake
    // We grab the latest TRANSFER_QUEUE record
    const depositRecord = await getLatestDepositRecord({
      client: stridejs,
      chainId: HOST_CHAIN_ID,
      blockHeight: tx.height,
      status: stride.records.DepositRecord_Status.TRANSFER_QUEUE,
    });
    const depositRecordId = depositRecord.id;
    expect(BigInt(depositRecord.amount) >= BigInt(stakeAmount)).to.be.true;

    // Wait for the the transfer to complete by checking for when the deposit record
    // changes to state DELEGATION_QUEUE
    console.log("Waiting for transfer to complete...");
    await waitForDepositRecordStatus({
      client: stridejs,
      depositRecordId,
      status: stride.records.DepositRecord_Status.DELEGATION_QUEUE,
    });

    // Wait for delegation to complete by checking the ICA account's delegations
    console.log("Waiting for delegation to complete...");
    await waitForDepositRecordStatus({
      client: stridejs,
      depositRecordId,
      status: REMOVED,
    });

    // Confirm at least our staked amount was delegated (it could be more if there was reinvestment)
    const updatedDelegatedBalance = await getDelegatedBalance({ client: hostjs, delegator: delegationIcaAddress });
    expect(updatedDelegatedBalance - initialDelegatedBalance >= BigInt(stakeAmount)).to.be.true;

    // Confirm the host zone updated
    const {
      hostZone: { totalDelegations: hostZoneTotalDelegations },
    } = await stridejs.query.stride.stakeibc.hostZone({ chainId: HOST_CHAIN_ID });
    expect(BigInt(hostZoneTotalDelegations)).to.equal(updatedDelegatedBalance, "Updated delegated balance");
  }, 180_000); // 3 minutes timeout

  test.only("Redeem Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;
    const redeemAmount = 1000000;

    // Get initial stBalance and native balances
    const initialUserNativeBalanceOnStride = await getBalance({
      client: stridejs,
      denom: HOST_DENOM_ON_STRIDE,
    });

    const initialUserNativeBalanceOnHost = await getBalance({
      client: hostjs,
      denom: HOST_DENOM,
    });

    const initialUserStBalance = await getBalance({
      client: stridejs,
      denom: ST_DENOM,
    });

    // Get the initial delegated balance
    const {
      hostZone: { totalDelegations: initialDelegatedBalance },
    } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: HOST_CHAIN_ID,
    });

    // If there isn't enough staked to cover the redemption, we need to liquid stake
    if (BigInt(initialDelegatedBalance) < BigInt(redeemAmount)) {
      console.log("No active delegations on host zone");

      // If there's not enough native token on stride to liquid stake, we need to transfer
      if (initialUserNativeBalanceOnStride == BigInt(0)) {
        console.log("Transfering native tokens to Stride...");
        await ibcTransfer({
          client: hostjs,
          sourceChain: HOST_CHAIN_NAME,
          destinationChain: STRIDE_CHAIN_NAME,
          coin: `${stakeAmount}${HOST_DENOM}`,
          sender: hostjs.address,
          receiver: stridejs.address,
        });

        await waitForBalanceChange({
          initialBalance: initialUserNativeBalanceOnStride,
          client: stridejs,
          address: stridejs.address,
          denom: HOST_DENOM_ON_STRIDE,
        });
      }

      // Then once we know we have native tokens, we can liquid stake
      console.log("Liquid staking...");
      const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
        creator: stridejs.address,
        amount: String(stakeAmount),
        hostDenom: HOST_DENOM,
      });

      await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
      await sleep(2000); // sleep to make sure block finalized

      // Then wait for there to be enough delegated to process the redemption
      console.log("Waiting for delegation on host zone...");
      await waitForHostZoneTotalDelegationsChange({
        client: stridejs,
        chainId: HOST_CHAIN_ID,
        minDelegation: redeemAmount,
      });
    }

    // Submit redeem stake tx
  }, 600_000); // 10 minutes timeout
});
