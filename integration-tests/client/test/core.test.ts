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
  CONNECTION_ID,
  REMOVED,
} from "./consts";
import { CosmosClient } from "./types";
import {
  ibcTransfer,
  waitForChain,
  submitTxAndExpectSuccess,
  waitForBalanceChange,
  getBalance,
  assertOpenTransferChannel,
  assertICAChannelsOpen,
  waitForDepositRecordStatus,
  getDelegatedBalance,
  waitForDelegationChange,
} from "./utils";
import { StrideClient } from "stridejs";
import { Decimal } from "decimal.js";
import { getAllDepositRecords, getHostZone } from "./queries";
import { DepositRecord_Status } from "stridejs/dist/types/codegen/stride/records/records";

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

  console.log("registering host zones...");
  const { hostZone } = await strideAccounts.admin.query.stride.stakeibc.hostZoneAll({});
  const gaiaHostZoneNotRegistered = hostZone.find((hz) => hz.chainId === GAIA_CHAIN_ID) === undefined;

  if (gaiaHostZoneNotRegistered) {
    const gaiaRegisterHostZoneMsg = stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
      creator: strideAccounts.admin.address,
      connectionId: CONNECTION_ID.STRIDE.GAIA!,
      bech32prefix: "cosmos",
      hostDenom: UATOM,
      ibcDenom: ATOM_DENOM_ON_STRIDE,
      transferChannelId: TRANSFER_CHANNEL.STRIDE.GAIA!,
      unbondingPeriod: BigInt(1),
      minRedemptionRate: "0.9",
      maxRedemptionRate: "1.5",
      lsmLiquidStakeEnabled: true,
      communityPoolTreasuryAddress: "",
      maxMessagesPerIcaTx: BigInt(2),
    });

    const { validators: gaiaValidators } = await gaiaAccounts.user.query.staking.validators("BOND_STATUS_BONDED");
    const gaiaAddValidatorsMsg = stride.stakeibc.MessageComposer.withTypeUrl.addValidators({
      creator: strideAccounts.admin.address,
      hostZone: GAIA_CHAIN_ID,
      validators: gaiaValidators.map((val) => ({
        name: val.description.moniker,
        address: val.operatorAddress,
        weight: 10n,
        delegation: "0", // ignored
        slashQueryProgressTracker: "0", // ignored
        slashQueryCheckpoint: "0", // ignored
        sharesToTokensRate: "0", // ignored
        delegationChangesInProgress: 0n, // ignored
        slashQueryInProgress: false, // ignored
      })),
    });

    await submitTxAndExpectSuccess(strideAccounts.admin, [gaiaRegisterHostZoneMsg, gaiaAddValidatorsMsg]);
  }

  console.log("waiting for ICA channels...");
  await assertICAChannelsOpen(strideAccounts.admin, GAIA_CHAIN_ID);
}, 45_000);

describe("Core Tests", () => {
  test("IBC Transfer", async () => {
    const stridejs = strideAccounts.user;
    const gaiajs = gaiaAccounts.user;
    const transferAmount = BigInt(50000000);

    // Get initial balances
    // We'll send STRD from Stride -> Gaia
    const strideInitialStrdBalance = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const gaiaInitialStrdBalance = await getBalance({
      client: gaiajs,
      denom: STRD_DENOM_ON_GAIA,
    });

    // As well as ATOM from Gaia -> Stride
    const gaiaInitialAtomBalance = await getBalance({
      client: gaiajs,
      denom: UATOM,
    });

    const strideInitialAtomBalance = await getBalance({
      client: stridejs,
      denom: ATOM_DENOM_ON_STRIDE,
    });

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
    console.log("Waiting for transfers to complete...");
    await sleep(5000);

    // Get final balances
    const strideFinalStrdBalance = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const gaiaFinalStrdBalance = await getBalance({
      client: gaiajs,
      denom: STRD_DENOM_ON_GAIA,
    });

    const gaiaFinalAtomBalance = await getBalance({
      client: gaiajs,
      denom: UATOM,
    });

    const strideFinalAtomBalance = await getBalance({
      client: stridejs,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    // Calculate and verify balance changes
    const strideStrdBalanceDiff = strideFinalStrdBalance - strideInitialStrdBalance;
    const strideAtomBalanceDiff = strideFinalAtomBalance - strideInitialAtomBalance;
    const gaiaStrdBalanceDiff = gaiaFinalStrdBalance - gaiaInitialStrdBalance;
    const gaiaAtomBalanceDiff = gaiaFinalAtomBalance - gaiaInitialAtomBalance;

    // Verify the transfers worked
    // STRD sent out from Stride → negative balance change + fee
    // STRD received on Gaia → positive balance change
    expect(strideStrdBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Stride STRD balance change");
    expect(gaiaStrdBalanceDiff).to.equal(transferAmount, "Gaia STRD balance change");

    // ATOM sent out from Gaia → negative balance change + fee
    // ATOM received on Stride → positive balance change
    expect(gaiaAtomBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Gaia ATOM balance change");
    expect(strideAtomBalanceDiff).to.equal(transferAmount, "Stride ATOM balance change");
  }, 120_000);

  test("Liquid Stake", async () => {
    const stridejs = strideAccounts.user;
    const gaiajs = gaiaAccounts.user;
    const stakeAmount = 10000000;

    // Get initial balances
    let strideInitialAtomBalance = await getBalance({
      client: stridejs,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    const strideInitialStAtomBalance = await getBalance({
      client: stridejs,
      denom: STATOM,
    });

    // Get the initial delegated balance
    const {
      hostZone: { delegationIcaAddress },
    } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: GAIA_CHAIN_ID,
    });
    const initialDelegatedBalance = await getDelegatedBalance({
      client: gaiajs,
      delegator: delegationIcaAddress,
    });

    // Ensure there's enough ATOM to liquid stake, if not transfer
    if (strideInitialAtomBalance == BigInt(0)) {
      console.log("Transferring ATOM from Gaia to Stride...");
      await ibcTransfer({
        client: gaiajs,
        sourceChain: "GAIA",
        destinationChain: "STRIDE",
        coin: `${stakeAmount}${UATOM}`,
        sender: gaiajs.address,
        receiver: stridejs.address,
      });

      strideInitialAtomBalance = await waitForBalanceChange({
        initialBalance: strideInitialAtomBalance,
        client: stridejs,
        address: stridejs.address,
        denom: ATOM_DENOM_ON_STRIDE,
      });
    }

    // Perform liquid staking
    console.log("Liquid staking...");
    const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
      creator: stridejs.address,
      amount: String(stakeAmount),
      hostDenom: UATOM,
    });

    const tx = await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
    await sleep(2000); // sleep to make sure block finalized

    // Get final ATOM and stATOM balances
    const strideFinalStAtomBalance = await getBalance({ client: stridejs, address: stridejs.address, denom: STATOM });
    const strideFinalAtomBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: ATOM_DENOM_ON_STRIDE,
    });

    // Get the redemption rate at the time of the liquid stake
    const {
      hostZone: { redemptionRate },
    } = await getHostZone(stridejs, GAIA_CHAIN_ID, tx.height);

    // Confirm the balance changes
    // ATOM should decrease (sent for staking)
    // stATOM should increase (minted)
    // Delegation ICA should receive tokens
    const strideAtomBalanceDiff = strideFinalAtomBalance - strideInitialAtomBalance;
    const strideStAtomBalanceDiff = strideFinalStAtomBalance - strideInitialStAtomBalance;
    const expectedStAtomAmount = BigInt(
      Decimal(stakeAmount.toString()).div(Decimal(redemptionRate)).floor().toString(),
    );
    expect(strideStAtomBalanceDiff).to.equal(expectedStAtomAmount, "User stATOM balance change on Stride");
    expect(strideAtomBalanceDiff).to.equal(BigInt(-stakeAmount), "User ATOM balance change on Stride");

    // Get the deposit record that was used for the liquid stake
    // We grab the latest TRANSFER_QUEUE record
    const { depositRecord: allDepositRecords } = await getAllDepositRecords(stridejs, GAIA_CHAIN_ID, tx.height);
    const transferRecords = allDepositRecords
      .filter((record) => record.status === stride.records.DepositRecord_Status.TRANSFER_QUEUE)
      .sort((a, b) => Number(b.id - a.id));
    expect(transferRecords.length).to.be.greaterThan(0, "No transfer queue deposit records");

    const depositRecord = transferRecords[0];
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

    const updatedDelegatedBalance = await waitForDelegationChange({ client: gaiajs, delegator: delegationIcaAddress });

    // Confirm at least our staked amount was delegated (it could be more if there was reinvestment)
    expect(updatedDelegatedBalance - initialDelegatedBalance >= BigInt(stakeAmount)).to.be.true;

    // Confirm the host zone updated
    const {
      hostZone: { totalDelegations },
    } = await stridejs.query.stride.stakeibc.hostZone({ chainId: GAIA_CHAIN_ID });
    expect(BigInt(totalDelegations)).to.equal(updatedDelegatedBalance, "Updated delegated balance");
  }, 180_000); // 3 minutes timeout
});
