import { Secp256k1HdWallet } from "@cosmjs/amino";
import { DirectSecp256k1HdWallet, Registry } from "@cosmjs/proto-signing";
import { Event, GasPrice, SigningStargateClient } from "@cosmjs/stargate";
import { fromSeconds } from "@cosmjs/tendermint-rpc";
import {
  cosmosProtoRegistry,
  ibcProtoRegistry,
  osmosis,
  osmosisProtoRegistry,
} from "osmojs";
import {
  coinFromString,
  coinsFromString,
  convertBech32Prefix,
  decToString,
  getTxIbcResponses,
  ibcDenom,
  sleep,
  StrideClient,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  addConcentratedLiquidityPositionMsg,
  newConcentratedLiquidityPoolMsg,
  newGammPoolMsg,
  newRegisterTokenPriceQueryMsg,
} from "./msgs";
import {
  getValueFromEvents,
  submitTxAndExpectSuccess,
  transfer,
  waitForChain,
} from "./utils";
import { ProposalStatus, VoteOption } from "osmojs/cosmos/gov/v1beta1/gov";
import { ModuleAccount } from "osmojs/cosmos/auth/v1beta1/auth";

const STRIDE_RPC_ENDPOINT = "http://stride-rpc.internal.stridenet.co";
const GAIA_RPC_ENDPOINT = "http://cosmoshub-rpc.internal.stridenet.co";
const OSMO_RPC_ENDPOINT = "http://osmosis-rpc.internal.stridenet.co";

export const USTRD = "ustrd";
export const UATOM = "uatom";
export const UOSMO = "uosmo";

export type Chain = "STRIDE" | "GAIA" | "OSMO";

export const TRANSFER_CHANNEL: Record<Chain, Partial<Record<Chain, string>>> = {
  STRIDE: { GAIA: "channel-0", OSMO: "channel-1" },
  GAIA: { STRIDE: "channel-0" },
  OSMO: { STRIDE: "channel-0" },
};

let accounts: {
  user: StrideClient; // a normal account loaded with 100 STRD
  admin: StrideClient; // the stride admin account loaded with 1000 STRD
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

export type CosmosClient = {
  address: string;
  client: SigningStargateClient;
};

let gaiaAccounts: {
  user: CosmosClient; // a normal account loaded with 100 ATOM
};

let osmoAccounts: {
  user: CosmosClient; // a normal account loaded with 1,000,000 OSMO
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

// init accounts and wait for chain to start
beforeAll(async () => {
  console.log("setting up accounts...");
  // @ts-expect-error
  // init accounts as an empty object, then add the accounts in the loop
  accounts = {};
  for (const { name, mnemonic } of mnemonics) {
    // setup signer
    //
    // IMPORTANT: We're using Secp256k1HdWallet from @cosmjs/amino because sending amino txs tests both amino and direct.
    // That's because the tx contains the direct encoding anyway, and also attaches a signature on the amino encoding.
    // The mempool then converts from direct to amino to verify the signature.
    // Therefore if the signature verification passes, we can be sure that both amino and direct are working properly.
    const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "stride",
    });

    // get signer address
    const [{ address }] = await signer.getAccounts();

    accounts[name] = await StrideClient.create(
      STRIDE_RPC_ENDPOINT,
      signer,
      address,
      {
        gasPrice: GasPrice.fromString(`0.025${USTRD}`),
        broadcastPollIntervalMs: 50,
        resolveIbcResponsesCheckIntervalMs: 50,
      },
    );

    if (name === "user") {
      const gaiaSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);

      const [{ address: gaiaAddress }] = await gaiaSigner.getAccounts();

      gaiaAccounts = {
        user: {
          address: gaiaAddress,
          client: await SigningStargateClient.connectWithSigner(
            GAIA_RPC_ENDPOINT,
            gaiaSigner,
            {
              gasPrice: GasPrice.fromString("1.0uatom"),
              broadcastPollIntervalMs: 50,
            },
          ),
        },
      };

      const osmoSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
        prefix: "osmo",
      });

      const [{ address: osmoAddress }] = await osmoSigner.getAccounts();

      osmoAccounts = {
        user: {
          address: osmoAddress,
          client: await SigningStargateClient.connectWithSigner(
            OSMO_RPC_ENDPOINT,
            osmoSigner,
            {
              gasPrice: GasPrice.fromString("1.0uosmo"),
              broadcastPollIntervalMs: 50,
              registry: new Registry([
                ...osmosisProtoRegistry,
                ...cosmosProtoRegistry,
                ...ibcProtoRegistry,
              ]),
            },
          ),
        },
      };
    }
  }
  console.log("waiting for stride to start...");
  await waitForChain(accounts.user, USTRD);

  console.log("waiting for gaia to start...");
  await waitForChain(gaiaAccounts.user, UATOM);

  console.log("waiting for osmosis to start...");
  await waitForChain(osmoAccounts.user, UOSMO);
});

// time variables in seconds
const now = () => Math.floor(Date.now() / 1000);
const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;

describe("x/airdrop", () => {
  test("MsgCreateAirdrop", async () => {
    const stridejs = accounts.admin;

    const nowSec = now();
    const airdropId = String(nowSec);

    const tx = await stridejs.signAndBroadcast(
      [
        stridejs.types.stride.airdrop.MessageComposer.withTypeUrl.createAirdrop(
          {
            admin: stridejs.address,
            airdropId: airdropId,
            rewardDenom: USTRD,
            distributionStartDate: fromSeconds(now()),
            distributionEndDate: fromSeconds(nowSec + 3 * day),
            clawbackDate: fromSeconds(nowSec + 4 * day),
            claimTypeDeadlineDate: fromSeconds(nowSec + 2 * day),
            earlyClaimPenalty: decToString(0.5),
            allocatorAddress: stridejs.address,
            distributorAddress: stridejs.address,
            linkerAddress: stridejs.address,
          },
        ),
      ],
      2,
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const { id, earlyClaimPenalty } =
      await stridejs.query.stride.airdrop.airdrop({
        id: airdropId,
      });

    expect(id).toBe(airdropId);
    expect(earlyClaimPenalty).toBe("0.5");
  });
});

describe("ibc", () => {
  test("MsgTransfer", async () => {
    const stridejs = accounts.user;

    await transfer({
      stridejs: stridejs,
      signingClient: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "GAIA",
      coins: `1${USTRD}`,
      sender: stridejs.address,
      receiver: convertBech32Prefix(stridejs.address, "cosmos"),
    });

    const tx = await stridejs.signAndBroadcast(
      [
        stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
          {
            sourcePort: "transfer",
            sourceChannel: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
            token: coinFromString(`1${USTRD}`),
            sender: stridejs.address,
            receiver: convertBech32Prefix(stridejs.address, "cosmos"),
            timeoutHeight: {
              revisionNumber: 0n,
              revisionHeight: 0n,
            },
            timeoutTimestamp: BigInt(
              `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
            ),
            memo: "",
          },
        ),
      ],
      2,
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const ibcAck = await tx.ibcResponses[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);
  }, 30_000);
});

describe("x/stakeibc", () => {
  // skip due to amino bullshit
  test.skip("Registration", async () => {
    const stridejs = accounts.admin;

    const msg =
      stridejs.types.stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone(
        {
          creator: stridejs.address,
          bech32prefix: "cosmos",
          hostDenom: "uatom",
          ibcDenom: ibcDenom(
            [{ incomingPortId: "transfer", incomingChannelId: "channel-0" }],
            "uatom",
          ),
          connectionId: "connection-0",
          transferChannelId: "channel-0",
          unbondingPeriod: BigInt(1),
          minRedemptionRate: "0.9",
          maxRedemptionRate: "1.5",
          lsmLiquidStakeEnabled: true,
          communityPoolTreasuryAddress:
            "cosmos1kl8d29eadt93rfxmkf2q8msxwylaax9dxzr5lj",
          maxMessagesPerIcaTx: BigInt(2),
        },
      );

    await submitTxAndExpectSuccess(stridejs, [msg]);
    console.log(stridejs.query.stride.stakeibc.hostZoneAll());
  });
});

describe("buyback and burn", () => {
  test("gamm pool price", async () => {
    const stridejs = accounts.user;
    const osmojs = osmoAccounts.user;

    await transfer({
      stridejs: stridejs,
      signingClient: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "OSMO",
      coins: `1000000${USTRD}`,
      sender: stridejs.address,
      receiver: osmojs.address,
    });

    const osmoDenomOnStride = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
        },
      ],
      UOSMO,
    );

    const strdOnOsmoDenom = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
        },
      ],
      USTRD,
    );

    const poolMsg = newGammPoolMsg({
      sender: osmojs.address,
      tokens: [`10${UOSMO}`, `2${strdOnOsmoDenom}`],
      weights: [1, 1],
    });
    const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

    const osmoStrdPoolId = BigInt(
      poolTx.events.find((e) => e.type === "pool_created")?.attributes[0]
        .value!,
    );

    const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
      adminClient: accounts.admin,
      baseDenom: USTRD,
      quoteDenom: osmoDenomOnStride,
      baseDenomOnOsmosis: strdOnOsmoDenom,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(accounts.admin, registerTokenPriceMsg);

    while (true) {
      const {
        tokenPrice: {
          baseDenom,
          quoteDenom,
          baseDenomDecimals,
          quoteDenomDecimals,
          osmosisBaseDenom,
          osmosisQuoteDenom,
          osmosisPoolId,
          spotPrice,
          lastRequestTime,
          lastResponseTime,
          queryInProgress,
        },
      } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: osmoDenomOnStride,
        poolId: osmoStrdPoolId,
      });
      if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
        expect(Number(spotPrice)).toBe(5);

        // Verify base denom matches
        expect(baseDenom).toBe(USTRD);
        expect(osmosisBaseDenom).toBe(strdOnOsmoDenom);

        // Verify quote denom matches
        expect(quoteDenom).toBe(osmoDenomOnStride);
        expect(osmosisQuoteDenom).toBe(UOSMO);

        // Verify pool ID
        expect(osmosisPoolId).toBe(osmoStrdPoolId);

        // Verify decimals
        expect(baseDenomDecimals).toBe(6n); // STRD has 6 decimals
        expect(quoteDenomDecimals).toBe(6n); // OSMO has 6 decimals

        // Verify query metadata
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastResponseTime.toISOString()).toBeTruthy();
        expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(
          true,
        );

        break;
      }
      await sleep(500);
    }
  });

  test("concentrated liquidity pool price", async () => {
    const stridejs = accounts.user;
    const osmojs = osmoAccounts.user;

    await transfer({
      stridejs: stridejs,
      signingClient: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "OSMO",
      coins: `1000000${USTRD}`,
      sender: stridejs.address,
      receiver: osmojs.address,
    });

    const osmoDenomOnStride = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
        },
      ],
      UOSMO,
    );

    const strdOnOsmoDenom = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
        },
      ],
      USTRD,
    );

    const poolMsg = newConcentratedLiquidityPoolMsg({
      sender: osmojs.address,
      denom0: strdOnOsmoDenom,
    });
    const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

    const osmoStrdPoolId = BigInt(
      poolTx.events.find((e) => e.type === "pool_created")?.attributes[0]
        .value!,
    );

    const addLiquidityMsg = addConcentratedLiquidityPositionMsg({
      sender: osmojs.address,
      poolId: osmoStrdPoolId,
      tokensProvided: coinsFromString(`5${strdOnOsmoDenom},10uosmo`),
      tokenMinAmount0: "5",
      tokenMinAmount1: "10",
    });
    await submitTxAndExpectSuccess(osmojs, addLiquidityMsg);

    const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
      adminClient: accounts.admin,
      baseDenom: USTRD,
      quoteDenom: osmoDenomOnStride,
      baseDenomOnOsmosis: strdOnOsmoDenom,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(accounts.admin, registerTokenPriceMsg);

    while (true) {
      const {
        tokenPrice: {
          baseDenom,
          quoteDenom,
          baseDenomDecimals,
          quoteDenomDecimals,
          osmosisBaseDenom,
          osmosisQuoteDenom,
          osmosisPoolId,
          spotPrice,
          lastRequestTime,
          lastResponseTime,
          queryInProgress,
        },
      } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: osmoDenomOnStride,
        poolId: osmoStrdPoolId,
      });
      if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
        expect(Number(spotPrice)).toBe(5);

        // Verify base denom matches
        expect(baseDenom).toBe(USTRD);
        expect(osmosisBaseDenom).toBe(strdOnOsmoDenom);

        // Verify quote denom matches
        expect(quoteDenom).toBe(osmoDenomOnStride);
        expect(osmosisQuoteDenom).toBe(UOSMO);

        // Verify pool ID
        expect(osmosisPoolId).toBe(osmoStrdPoolId);

        // Verify decimals
        expect(baseDenomDecimals).toBe(6n); // STRD has 6 decimals
        expect(quoteDenomDecimals).toBe(6n); // OSMO has 6 decimals

        // Verify query metadata
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastResponseTime.toISOString()).toBeTruthy();
        expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(
          true,
        );

        break;
      }
      await sleep(500);
    }
  });

  test("icqoracle happy path", async () => {
    // - Transfer STRD to Osmosis
    // - Transfer ATOM to Osmosis
    // - Create STRD/OSMO pool
    // - Create ATOM/OSMO pool
    // - Add TokenPrice(base=STRD, quote=OSMO)
    // - Add TokenPrice(base=ATOM, quote=OSMO)
    // - Query for price of ATOM in STRD

    const stridejs = accounts.user;
    const gaiajs = gaiaAccounts.user;
    const osmojs = osmoAccounts.user;

    // Transfer STRD to Osmosis
    let strideTx = await stridejs.signAndBroadcast([
      stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
        {
          sourcePort: "transfer",
          sourceChannel: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
          token: coinFromString(`1000000${USTRD}`),
          sender: stridejs.address,
          receiver: osmojs.address,
          timeoutHeight: {
            revisionNumber: 0n,
            revisionHeight: 0n,
          },
          timeoutTimestamp: BigInt(
            `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
          ),
          memo: "",
        },
      ),
    ]);
    if (strideTx.code !== 0) {
      console.error(strideTx.rawLog);
    }
    expect(strideTx.code).toBe(0);

    let ibcAck = await strideTx.ibcResponses[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);

    // Transfer ATOM to Osmosis
    let tx = await gaiajs.client.signAndBroadcast(
      gaiajs.address,
      [
        stridejs.types.ibc.applications.transfer.v1.MessageComposer.withTypeUrl.transfer(
          {
            sourcePort: "transfer",
            sourceChannel: TRANSFER_CHANNEL["GAIA"]["STRIDE"]!,
            token: coinFromString("1000000uatom"),
            sender: gaiajs.address,
            receiver: stridejs.address, // needs to be valid but ignored by pfm
            timeoutHeight: {
              revisionNumber: 0n,
              revisionHeight: 0n,
            },
            timeoutTimestamp: BigInt(
              `${Math.floor(Date.now() / 1000) + 3 * 60}000000000`, // 3 minutes
            ),
            memo: JSON.stringify({
              forward: {
                receiver: osmojs.address,
                port: "transfer",
                channel: TRANSFER_CHANNEL["STRIDE"]["OSMO"],
              },
            }),
          },
        ),
      ],
      "auto",
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    ibcAck = await getTxIbcResponses(gaiajs.client, tx, 30_000, 50)[0];
    expect(ibcAck.type).toBe("ack");
    expect(ibcAck.tx.code).toBe(0);

    // Create STRD/OSMO pool
    const strdDenomOnOsmosis = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
        },
      ],
      USTRD,
    );

    tx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.MessageComposer.withTypeUrl.createConcentratedPool(
          {
            sender: osmojs.address,
            denom0: strdDenomOnOsmosis,
            denom1: "uosmo",
            tickSpacing: 100n,
            spreadFactor: "0.001",
          },
        ),
      ],
      "auto",
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const osmoStrdPoolId = BigInt(
      tx.events.find((e) => e.type === "pool_created")?.attributes[0].value!,
    );

    tx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        osmosis.concentratedliquidity.v1beta1.MessageComposer.withTypeUrl.createPosition(
          {
            poolId: osmoStrdPoolId,
            sender: osmojs.address,
            lowerTick: -108000000n,
            upperTick: 342000000n,
            tokensProvided: coinsFromString(`5${strdDenomOnOsmosis},10uosmo`),
            tokenMinAmount0: "5",
            tokenMinAmount1: "10",
          },
        ),
      ],
      "auto",
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    // Create ATOM/OSMO pool
    const atomDenomOnOsmosis = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
        },
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["OSMO"]["STRIDE"]!,
        },
      ],
      "uatom",
    );

    tx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        osmosis.gamm.poolmodels.balancer.v1beta1.MessageComposer.withTypeUrl.createBalancerPool(
          {
            sender: osmojs.address,
            poolAssets: [
              {
                token: coinFromString(`10uosmo`),
                weight: "1",
              },
              {
                token: coinFromString(`2${atomDenomOnOsmosis}`),
                weight: "1",
              },
            ],
            futurePoolGovernor: "",
            poolParams: {
              swapFee: "0.001",
              exitFee: "0",
            },
          },
        ),
      ],
      "auto",
    );
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const osmoAtomPoolId = BigInt(
      tx.events.find((e) => e.type === "pool_created")?.attributes[0].value!,
    );

    // Add TokenPrice(base=STRD, quote=OSMO)
    const osmoDenomOnStride = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["OSMO"]!,
        },
      ],
      "uosmo",
    );
    strideTx = await accounts.admin.signAndBroadcast(
      [
        stridejs.types.stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery(
          {
            admin: accounts.admin.address,
            baseDenom: USTRD,
            quoteDenom: osmoDenomOnStride,
            baseDenomDecimals: 6n,
            quoteDenomDecimals: 6n,
            osmosisBaseDenom: strdDenomOnOsmosis,
            osmosisQuoteDenom: "uosmo",
            osmosisPoolId: osmoStrdPoolId,
          },
        ),
      ],
      2,
    );
    if (strideTx.code !== 0) {
      console.error(strideTx.rawLog);
    }
    expect(strideTx.code).toBe(0);

    // Add TokenPrice(base=ATOM, quote=OSMO)
    const atomDenomOnStride = ibcDenom(
      [
        {
          incomingPortId: "transfer",
          incomingChannelId: TRANSFER_CHANNEL["STRIDE"]["GAIA"]!,
        },
      ],
      "uatom",
    );
    strideTx = await accounts.admin.signAndBroadcast(
      [
        stridejs.types.stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery(
          {
            admin: accounts.admin.address,
            baseDenom: atomDenomOnStride,
            quoteDenom: osmoDenomOnStride,
            baseDenomDecimals: 6n,
            quoteDenomDecimals: 6n,
            osmosisBaseDenom: atomDenomOnOsmosis,
            osmosisQuoteDenom: "uosmo",
            osmosisPoolId: osmoAtomPoolId,
          },
        ),
      ],
      2,
    );
    if (strideTx.code !== 0) {
      console.error(strideTx.rawLog);
    }
    expect(strideTx.code).toBe(0);

    // Wait for both TokenPrices to be updated
    while (true) {
      const { tokenPrice } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: osmoDenomOnStride,
        poolId: osmoStrdPoolId,
      });
      if (
        tokenPrice.lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z"
      ) {
        expect(Number(tokenPrice.spotPrice)).toBe(2);
        break;
      }
      await sleep(500);
    }
    while (true) {
      const { tokenPrice } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: atomDenomOnStride,
        quoteDenom: osmoDenomOnStride,
        poolId: osmoAtomPoolId,
      });
      if (
        tokenPrice.lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z"
      ) {
        expect(Number(tokenPrice.spotPrice)).toBe(5);
        break;
      }
      await sleep(500);
    }

    // Query for price of ATOM in STRD
    const { price } =
      await stridejs.query.stride.icqoracle.tokenPriceForQuoteDenom({
        baseDenom: atomDenomOnStride,
        quoteDenom: USTRD,
      });

    // Price should be 2.5:
    //
    // ATOM/OSMO pool is 2/10 => 1 ATOM is 5 OSMO
    // STRD/OSMO pool is 5/10 => 1 STRD is 2 OSMO
    // =>
    // 2.5 STRD is 5 OSMO
    // =>
    // 1 ATOM is 2.5 STRD
    expect(Number(price)).toBe(2.5);
  }, 240_000);

  // skip due to amino bullshit
  test.skip("update params", async () => {
    const stridejs = accounts.user;

    const { params } = await stridejs.query.stride.icqoracle.params({});
    params.priceExpirationTimeoutSec += 1n;

    const govAccount =
      await stridejs.query.cosmos.auth.v1beta1.moduleAccountByName({
        name: "gov",
      });
    const govAddress = (govAccount.account as ModuleAccount).baseAccount
      ?.address!;

    const tx = await stridejs.signAndBroadcast([
      stridejs.types.cosmos.gov.v1.MessageComposer.withTypeUrl.submitProposal({
        messages: [
          stridejs.types.stride.icqoracle.MsgUpdateParams.toProtoMsg({
            authority: govAddress,
            params,
          }),
        ],
        initialDeposit: coinsFromString(`10000000${USTRD}`),
        proposer: stridejs.address,
        metadata: "Update icqoracle params",
        title: "Update icqoracle params",
        summary: "Update icqoracle params",
      }),
    ]);
    if (tx.code !== 0) {
      console.error(tx.rawLog);
    }
    expect(tx.code).toBe(0);

    const proposalId = BigInt(
      getValueFromEvents(tx.events, "submit_proposal.proposal_id"),
    );

    const txs = await Promise.all([
      accounts.val1.signAndBroadcast(
        [
          stridejs.types.cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: accounts.val1.address,
            option: VoteOption.VOTE_OPTION_YES,
            metadata: "",
          }),
        ],
        2,
      ),
      accounts.val2.signAndBroadcast(
        [
          stridejs.types.cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: accounts.val2.address,
            option: VoteOption.VOTE_OPTION_YES,
            metadata: "",
          }),
        ],
        2,
      ),
      accounts.val3.signAndBroadcast(
        [
          stridejs.types.cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: accounts.val3.address,
            option: VoteOption.VOTE_OPTION_YES,
            metadata: "",
          }),
        ],
        2,
      ),
    ]);

    for (const tx of txs) {
      if (tx.code !== 0) {
        console.error(tx.rawLog);
      }
      expect(tx.code).toBe(0);
    }

    while (true) {
      const { proposal } = await stridejs.query.cosmos.gov.v1.proposal({
        proposalId,
      });

      if (proposal?.status !== ProposalStatus.PROPOSAL_STATUS_VOTING_PERIOD) {
        expect(proposal?.status).toBe(ProposalStatus.PROPOSAL_STATUS_PASSED);
        break;
      }

      await sleep(500);
    }

    const { params: newParams } = await stridejs.query.stride.icqoracle.params(
      {},
    );
    expect(newParams).toStrictEqual(params);
  }, 60_000);

  // TODO test unwrapIBCDenom via stridejs.query.stride.icqoracle.tokenPrices
});
