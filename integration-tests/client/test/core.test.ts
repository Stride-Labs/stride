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
import { DirectSecp256k1HdWallet, GasPrice, sleep, stride } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  ATOM_DENOM_ON_STRIDE,
  GAIA_CHAIN_ID,
  GAIA_RPC_ENDPOINT,
  STRIDE_RPC_ENDPOINT,
  TRANSFER_CHANNEL,
  UATOM,
  USTRD,
  STRD_DENOM_ON_GAIA,
  DEFAULT_FEE,
  STATOM,
} from "./consts";
import { CosmosClient } from "./types";
import {
  ibcTransfer,
  waitForChain,
  submitTxAndExpectSuccess,
  waitForBalanceChange,
  getBalance,
  assertOpenTransferChannel,
} from "./utils";
import { StrideClient } from "stridejs";
import { getHostZone } from "./queries";

// Initialize accounts
let strideAccounts: {
  user: StrideClient;
  admin: StrideClient;
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

let gaiaAccounts: {
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
  gaiaAccounts = {};

  for (const { name, mnemonic } of mnemonics) {
    // setup signer for Stride
    const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "stride",
    });

    const [{ address }] = await signer.getAccounts();

    strideAccounts[name] = await StrideClient.create(STRIDE_RPC_ENDPOINT, signer, address, {
      gasPrice: GasPrice.fromString(`0.025${USTRD}`),
      broadcastPollIntervalMs: 50,
      resolveIbcResponsesCheckIntervalMs: 50,
    });

    if (name === "user" || name === "val1") {
      // setup signer for Gaia
      const gaiaSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);
      const [{ address: gaiaAddress }] = await gaiaSigner.getAccounts();

      gaiaAccounts[name] = {
        address: gaiaAddress,
        denom: UATOM,
        client: await SigningStargateClient.connectWithSigner(GAIA_RPC_ENDPOINT, gaiaSigner, {
          gasPrice: GasPrice.fromString(`1.0${UATOM}`),
          broadcastPollIntervalMs: 50,
        }),
        query: QueryClient.withExtensions(
          await Comet38Client.connect(GAIA_RPC_ENDPOINT),
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

  console.log("waiting for gaia to start...");
  await waitForChain(gaiaAccounts.user, UATOM);

  console.log("waiting for stride-gaia ibc...");
  await assertOpenTransferChannel(strideAccounts.user, TRANSFER_CHANNEL.STRIDE.GAIA!);
  await assertOpenTransferChannel(gaiaAccounts.user, TRANSFER_CHANNEL.GAIA.STRIDE!);
}, 45_000);

describe("Core Tests", () => {
  test.skip("IBC Transfer", async () => {
    const stridejs = strideAccounts.user;
    const gaiajs = gaiaAccounts.user;
    const transferAmount = 50000000;

    // Get initial balances
    // We'll send STRD from Stride -> Gaia
    const strideInitialStrdBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: USTRD,
    });

    const gaiaInitialStrdBalance = await getBalance({
      client: gaiajs,
      address: gaiajs.address,
      denom: STRD_DENOM_ON_GAIA,
    });

    // As well as ATOM from Gaia -> Stride
    const gaiaInitialAtomBalance = await getBalance({
      client: gaiajs,
      address: gaiajs.address,
      denom: UATOM,
    });

    const strideInitialAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    console.log("Initial balances:");
    console.log(`Stride USTRD: ${strideInitialStrdBalance}`);
    console.log(`Stride ATOM: ${strideInitialAtomBalance}`);
    console.log(`Gaia STRD: ${gaiaInitialStrdBalance}`);
    console.log(`Gaia ATOM: ${gaiaInitialAtomBalance}`);

    // Perform IBC transfers
    console.log("Transferring USTRD from Stride to Gaia...");
    await ibcTransfer({
      client: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "GAIA",
      coin: `${transferAmount}${USTRD}`,
      sender: stridejs.address,
      receiver: gaiajs.address,
    });

    console.log("Transferring ATOM from Gaia to Stride...");
    await ibcTransfer({
      client: gaiajs,
      sourceChain: "GAIA",
      destinationChain: "STRIDE",
      coin: `${transferAmount}${UATOM}`,
      sender: gaiajs.address,
      receiver: stridejs.address,
    });

    // Wait a bit for transfers to complete.
    await sleep(5000);

    // Get final balances
    const strideFinalStrdBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: USTRD,
    });

    const gaiaFinalStrdBalance = await getBalance({
      client: gaiajs,
      address: gaiajs.address,
      denom: STRD_DENOM_ON_GAIA,
    });

    const gaiaFinalAtomBalance = await getBalance({
      client: gaiajs,
      address: gaiajs.address,
      denom: UATOM,
    });

    const strideFinalAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    // Calculate and verify balance changes
    const strideStrdBalanceDiff = BigInt(strideFinalStrdBalance) - BigInt(strideInitialStrdBalance);
    const strideAtomBalanceDiff = BigInt(strideFinalAtomBalance) - BigInt(strideInitialAtomBalance);
    const gaiaStrdBalanceDiff = BigInt(gaiaFinalStrdBalance) - BigInt(gaiaInitialStrdBalance);
    const gaiaAtomBalanceDiff = BigInt(gaiaFinalAtomBalance) - BigInt(gaiaInitialAtomBalance);

    // Verify the transfers worked
    // STRD sent out from Stride → negative balance change + fee
    // STRD received on Gaia → positive balance change
    expect(strideStrdBalanceDiff).to.equal(BigInt(-(transferAmount + DEFAULT_FEE)), "Stride STRD balance change");
    expect(gaiaStrdBalanceDiff).to.equal(BigInt(transferAmount), "Gaia STRD balance change");

    // ATOM sent out from Gaia → negative balance change + fee
    // ATOM received on Stride → positive balance change
    expect(gaiaAtomBalanceDiff).to.equal(BigInt(-(transferAmount + DEFAULT_FEE)), "Gaia ATOM balance change");
    expect(strideAtomBalanceDiff).to.equal(BigInt(transferAmount), "Stride ATOM balance change");
  }, 120_000);

  test("Liquid Stake Mint and Transfer", async () => {
    const stridejs = strideAccounts.user;
    const gaiajs = gaiaAccounts.user;
    const stakeAmount = 10000000;

    // Get initial balances
    const strideInitialAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    const strideInitialStAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: STATOM,
    });

    // Get delegation address and assert it exists
    const {
      hostZone: { delegationIcaAddress },
    } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: GAIA_CHAIN_ID,
    });

    // Get initial delegation ICA balance
    const delegationInitialBalance = await getBalance({
      client: gaiajs,
      address: delegationIcaAddress,
      denom: UATOM,
    });

    console.log("Initial balances:");
    console.log(`Stride ATOM: ${strideInitialAtomBalance}`);
    console.log(`Stride stATOM: ${strideInitialStAtomBalance}`);
    console.log(`Delegation ICA ATOM: ${delegationInitialBalance}`);

    // Perform liquid staking
    console.log("Liquid staking...");
    const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
      creator: stridejs.address,
      amount: String(stakeAmount),
      hostDenom: UATOM,
    });

    const tx = await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);

    // Wait for stTokens to be minted
    const strideFinalStAtomBalance = await waitForBalanceChange({
      client: stridejs,
      address: stridejs.address,
      denom: "stuatom",
    });

    // Get the redemption rate at the time of the liquid stake
    const {
      hostZone: { redemptionRate },
    } = await getHostZone(stridejs, GAIA_CHAIN_ID, tx.height);
    console.log("Redemption Rate", redemptionRate);

    // Get final ATOM balance on Stride
    const strideFinalAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    // Wait for tokens to be transferred to the delegation account
    const delegationFinalBalance = await waitForBalanceChange({
      client: gaiajs,
      address: delegationIcaAddress,
      denom: UATOM,
    });

    console.log("Final balances:");
    console.log(`Stride ATOM: ${strideFinalAtomBalance}`);
    console.log(`Stride stATOM: ${strideFinalStAtomBalance}`);
    console.log(`Delegation ICA ATOM: ${delegationFinalBalance}`);

    // Calculate balance differences (final - initial for consistent sign tracking)
    const strideAtomBalanceDiff = BigInt(strideFinalAtomBalance) - BigInt(strideInitialAtomBalance);
    const strideStAtomBalanceDiff = BigInt(strideFinalStAtomBalance) - BigInt(strideInitialStAtomBalance);
    const delegationBalanceDiff = BigInt(delegationFinalBalance) - BigInt(delegationInitialBalance);

    console.log("Balance differences:");
    console.log(`Stride ATOM diff: ${strideAtomBalanceDiff}`);
    console.log(`Stride stATOM diff: ${strideStAtomBalanceDiff}`);
    console.log(`Delegation ICA diff: ${delegationBalanceDiff}`);

    // Verify balance changes
    // ATOM should decrease (sent for staking)
    // stATOM should increase (minted)
    // Delegation ICA should receive tokens
    expect(strideAtomBalanceDiff).to.equal(BigInt(-stakeAmount), "User ATOM balance change on Stride");
    expect(strideStAtomBalanceDiff).to.equal(BigInt(stakeAmount), "User stATOM balance change on Stride");
    expect(delegationBalanceDiff).to.equal(BigInt(stakeAmount), "Delegation account balance change");
  }, 180_000); // 3 minutes timeout
});
