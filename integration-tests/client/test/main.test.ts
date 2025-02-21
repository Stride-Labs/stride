import { Registry } from "@cosmjs/proto-signing";
import {
  QueryClient,
  setupAuthExtension,
  setupBankExtension,
  setupStakingExtension,
  setupTxExtension,
  SigningStargateClient,
} from "@cosmjs/stargate";
import { Comet38Client, fromSeconds } from "@cosmjs/tendermint-rpc";
import {
  cosmosProtoRegistry,
  ibcProtoRegistry,
  osmosisProtoRegistry,
} from "osmojs";
import { ProposalStatus, VoteOption } from "osmojs/cosmos/gov/v1beta1/gov";
import {
  coinsFromString,
  convertBech32Prefix,
  cosmos,
  decToString,
  DirectSecp256k1HdWallet,
  EncodeObject,
  GasPrice,
  getValueFromEvents,
  sleep,
  stride,
  StrideClient,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  ATOM_DENOM_ON_OSMOSIS,
  ATOM_DENOM_ON_STRIDE,
  CONNECTION_ID,
  GAIA_CHAIN_ID,
  GAIA_RPC_ENDPOINT,
  OSMO_CHAIN_ID,
  OSMO_DENOM_ON_STRIDE,
  OSMO_RPC_ENDPOINT,
  STRD_DENOM_ON_OSMOSIS,
  STRIDE_RPC_ENDPOINT,
  TRANSFER_CHANNEL,
  UATOM,
  UOSMO,
  USTRD,
} from "./consts";
import {
  addConcentratedLiquidityPositionMsg,
  newConcentratedLiquidityPoolMsg,
  newGammPoolMsg,
  newRegisterTokenPriceQueryMsg,
} from "./msgs";
import { CosmosClient } from "./types";
import {
  ibcTransfer,
  moduleAddress,
  submitTxAndExpectSuccess,
  waitForChain,
  waitForIbc,
} from "./utils";

let strideAccounts: {
  user: StrideClient; // a normal account loaded with 100 STRD
  admin: StrideClient; // the stride admin account loaded with 1000 STRD
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
};

let gaiaAccounts: {
  user: CosmosClient; // a normal account loaded with 100 ATOM
  val1: CosmosClient;
};

let osmoAccounts: {
  user: CosmosClient; // a normal account loaded with 1,000,000 OSMO
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

// init accounts and wait for chain to start
beforeAll(async () => {
  console.log("setting up accounts...");

  // init {,gaia,osmo}Accounts as an empty object, then add the accounts in the loop
  // @ts-expect-error
  strideAccounts = {};
  // @ts-expect-error
  gaiaAccounts = {};
  // @ts-expect-error
  osmoAccounts = {};
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

    strideAccounts[name] = await StrideClient.create(
      STRIDE_RPC_ENDPOINT,
      signer,
      address,
      {
        gasPrice: GasPrice.fromString(`0.025${USTRD}`),
        broadcastPollIntervalMs: 50,
        resolveIbcResponsesCheckIntervalMs: 50,
      },
    );

    if (name === "user" || name === "val1") {
      const gaiaSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);

      const [{ address: gaiaAddress }] = await gaiaSigner.getAccounts();

      gaiaAccounts[name] = {
        address: gaiaAddress,
        client: await SigningStargateClient.connectWithSigner(
          GAIA_RPC_ENDPOINT,
          gaiaSigner,
          {
            gasPrice: GasPrice.fromString(`1.0${UATOM}`),
            broadcastPollIntervalMs: 50,
          },
        ),
        query: QueryClient.withExtensions(
          await Comet38Client.connect(GAIA_RPC_ENDPOINT),
          setupAuthExtension,
          setupBankExtension,
          setupStakingExtension,
          setupTxExtension,
        ),
      };

      const osmoSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
        prefix: "osmo",
      });

      const [{ address: osmoAddress }] = await osmoSigner.getAccounts();

      osmoAccounts[name] = {
        address: osmoAddress,
        client: await SigningStargateClient.connectWithSigner(
          OSMO_RPC_ENDPOINT,
          osmoSigner,
          {
            gasPrice: GasPrice.fromString(`1.0${UOSMO}`),
            broadcastPollIntervalMs: 50,
            registry: new Registry([
              ...osmosisProtoRegistry,
              ...cosmosProtoRegistry,
              ...ibcProtoRegistry,
            ]),
          },
        ),
        query: QueryClient.withExtensions(
          await Comet38Client.connect(OSMO_RPC_ENDPOINT),
          setupAuthExtension,
          setupBankExtension,
          setupStakingExtension,
          setupTxExtension,
        ),
      };
    }
  }
  console.log("waiting for stride to start...");
  await waitForChain(strideAccounts.user, USTRD);

  console.log("waiting for gaia to start...");
  await waitForChain(gaiaAccounts.user, UATOM);

  console.log("waiting for osmosis to start...");
  await waitForChain(osmoAccounts.user, UOSMO);

  console.log("waiting for stride-gaia ibc...");
  await waitForIbc(
    strideAccounts.user,
    TRANSFER_CHANNEL.STRIDE.GAIA!,
    USTRD,
    "cosmos",
  );

  console.log("waiting for stride-osmosis ibc...");
  await waitForIbc(
    strideAccounts.user,
    TRANSFER_CHANNEL.STRIDE.OSMO!,
    USTRD,
    "osmo",
  );

  console.log("registering host zones...");

  const registerHostZonesMsgs: EncodeObject[] = [];
  const { hostZone } =
    await strideAccounts.admin.query.stride.stakeibc.hostZoneAll({});

  const gaiaHostZoneNotRegistered =
    hostZone.find((hz) => hz.chainId === GAIA_CHAIN_ID) === undefined;
  const osmoHostZoneNotRegistered =
    hostZone.find((hz) => hz.chainId === OSMO_CHAIN_ID) === undefined;

  if (gaiaHostZoneNotRegistered) {
    const gaiaRegisterHostZoneMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
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
        communityPoolTreasuryAddress:
          "cosmos1kl8d29eadt93rfxmkf2q8msxwylaax9dxzr5lj", // TODO fix magic string?
        maxMessagesPerIcaTx: BigInt(2),
      });

    const { validators: gaiaValidators } =
      await gaiaAccounts.user.query.staking.validators("BOND_STATUS_BONDED");
    const gaiaAddValidatorsMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.addValidators({
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

    registerHostZonesMsgs.push(gaiaRegisterHostZoneMsg, gaiaAddValidatorsMsg);
  }

  if (osmoHostZoneNotRegistered) {
    const osmoRegisterHostZoneMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
        creator: strideAccounts.admin.address,
        connectionId: CONNECTION_ID.STRIDE.OSMO!,
        bech32prefix: "osmo",
        hostDenom: UOSMO,
        ibcDenom: OSMO_DENOM_ON_STRIDE,
        transferChannelId: TRANSFER_CHANNEL.STRIDE.OSMO!,
        unbondingPeriod: BigInt(1),
        minRedemptionRate: "0.9",
        maxRedemptionRate: "1.5",
        lsmLiquidStakeEnabled: true,
        communityPoolTreasuryAddress: convertBech32Prefix(
          "cosmos1kl8d29eadt93rfxmkf2q8msxwylaax9dxzr5lj", // TODO fix magic string?
          "osmo",
        ),
        maxMessagesPerIcaTx: BigInt(2),
      });

    const { validators: osmoValidators } =
      await gaiaAccounts.user.query.staking.validators("BOND_STATUS_BONDED");
    const osmoAddValidatorsMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.addValidators({
        creator: strideAccounts.admin.address,
        hostZone: OSMO_CHAIN_ID,
        validators: osmoValidators.map((val) => ({
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

    registerHostZonesMsgs.push(osmoRegisterHostZoneMsg, osmoAddValidatorsMsg);
  }

  if (registerHostZonesMsgs.length > 0) {
    await submitTxAndExpectSuccess(strideAccounts.admin, registerHostZonesMsgs);
  }
}, 45_000);

describe("x/airdrop", () => {
  // time variables in seconds
  const now = () => Math.floor(Date.now() / 1000);
  const minute = 60;
  const hour = 60 * minute;
  const day = 24 * hour;

  test("MsgCreateAirdrop", async () => {
    const stridejs = strideAccounts.admin;

    const nowSec = now();
    const airdropId = String(nowSec);

    const tx = await stridejs.signAndBroadcast(
      [
        stride.airdrop.MessageComposer.withTypeUrl.createAirdrop({
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
        }),
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
    const stridejs = strideAccounts.user;

    await ibcTransfer({
      client: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "GAIA",
      coin: `1${USTRD}`,
      sender: stridejs.address,
      receiver: convertBech32Prefix(stridejs.address, "cosmos"),
    });
  }, 30_000);
});

describe("x/stakeibc", () => {
  // skip due to amino bullshit
  test.skip("Registration", async () => {
    const stridejs = strideAccounts.admin;

    const msg = stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
      creator: stridejs.address,
      bech32prefix: "cosmos",
      hostDenom: UATOM,
      ibcDenom: ATOM_DENOM_ON_STRIDE,
      connectionId: CONNECTION_ID.STRIDE.GAIA!,
      transferChannelId: TRANSFER_CHANNEL.STRIDE.GAIA!,
      unbondingPeriod: BigInt(1),
      minRedemptionRate: "0.9",
      maxRedemptionRate: "1.5",
      lsmLiquidStakeEnabled: true,
      communityPoolTreasuryAddress:
        "cosmos1kl8d29eadt93rfxmkf2q8msxwylaax9dxzr5lj", // TODO fix magic string?
      maxMessagesPerIcaTx: BigInt(2),
    });

    await submitTxAndExpectSuccess(stridejs, [msg]);
    console.log(stridejs.query.stride.stakeibc.hostZoneAll());
  });
});

describe("buyback and burn", () => {
  test("gamm pool price", async () => {
    const stridejs = strideAccounts.user;
    const osmojs = osmoAccounts.user;

    await ibcTransfer({
      client: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "OSMO",
      coin: `1000000${USTRD}`,
      sender: stridejs.address,
      receiver: osmojs.address,
    });

    const poolMsg = newGammPoolMsg({
      sender: osmojs.address,
      tokens: [`10${UOSMO}`, `2${STRD_DENOM_ON_OSMOSIS}`],
      weights: [1, 1],
    });
    const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

    const osmoStrdPoolId = BigInt(
      getValueFromEvents(poolTx.events, "pool_created.pool_id"),
    );

    const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
      admin: strideAccounts.admin.address,
      baseDenom: USTRD,
      quoteDenom: OSMO_DENOM_ON_STRIDE,
      baseDenomOnOsmosis: STRD_DENOM_ON_OSMOSIS,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);

    while (true) {
      const {
        tokenPrice: {
          baseDenom,
          quoteDenom,
          osmosisBaseDenom,
          osmosisQuoteDenom,
          osmosisPoolId,
          spotPrice,
          lastRequestTime,
          lastResponseTime,
        },
      } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: OSMO_DENOM_ON_STRIDE,
        poolId: osmoStrdPoolId,
      });
      if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
        expect(Number(spotPrice)).toBe(5);

        // Verify base denom matches
        expect(baseDenom).toBe(USTRD);
        expect(osmosisBaseDenom).toBe(STRD_DENOM_ON_OSMOSIS);

        // Verify quote denom matches
        expect(quoteDenom).toBe(OSMO_DENOM_ON_STRIDE);
        expect(osmosisQuoteDenom).toBe(UOSMO);

        // Verify pool ID
        expect(osmosisPoolId).toBe(osmoStrdPoolId);

        // Verify query metadata
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(
          true,
        );

        break;
      }
      await sleep(500);
    }
  });

  test("concentrated liquidity pool price", async () => {
    const stridejs = strideAccounts.user;
    const osmojs = osmoAccounts.user;

    await ibcTransfer({
      client: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "OSMO",
      coin: `1000000${USTRD}`,
      sender: stridejs.address,
      receiver: osmojs.address,
    });

    const poolMsg = newConcentratedLiquidityPoolMsg({
      sender: osmojs.address,
      denom0: STRD_DENOM_ON_OSMOSIS,
    });
    const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

    const osmoStrdPoolId = BigInt(
      getValueFromEvents(poolTx.events, "pool_created.pool_id"),
    );

    const addLiquidityMsg = addConcentratedLiquidityPositionMsg({
      sender: osmojs.address,
      poolId: osmoStrdPoolId,
      tokensProvided: coinsFromString(`5${STRD_DENOM_ON_OSMOSIS},10${UOSMO}`),
      tokenMinAmount0: "5",
      tokenMinAmount1: "10",
    });
    await submitTxAndExpectSuccess(osmojs, addLiquidityMsg);

    const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
      admin: strideAccounts.admin.address,
      baseDenom: USTRD,
      quoteDenom: OSMO_DENOM_ON_STRIDE,
      baseDenomOnOsmosis: STRD_DENOM_ON_OSMOSIS,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);

    while (true) {
      const {
        tokenPrice: {
          baseDenom,
          quoteDenom,
          osmosisBaseDenom,
          osmosisQuoteDenom,
          osmosisPoolId,
          spotPrice,
          lastRequestTime,
          lastResponseTime,
        },
      } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: OSMO_DENOM_ON_STRIDE,
        poolId: osmoStrdPoolId,
      });
      if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
        expect(Number(spotPrice)).toBe(2);

        // Verify base denom matches
        expect(baseDenom).toBe(USTRD);
        expect(osmosisBaseDenom).toBe(STRD_DENOM_ON_OSMOSIS);

        // Verify quote denom matches
        expect(quoteDenom).toBe(OSMO_DENOM_ON_STRIDE);
        expect(osmosisQuoteDenom).toBe(UOSMO);

        // Verify pool ID
        expect(osmosisPoolId).toBe(osmoStrdPoolId);

        // Verify query metadata
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
        expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(
          true,
        );

        break;
      }
      await sleep(500);
    }
  });

  test(
    "happy path",
    async () => {
      // - Transfer STRD to Osmosis
      // - Transfer ATOM to Osmosis
      // - Create STRD/OSMO pool
      // - Create ATOM/OSMO pool
      // - Add TokenPrice(base=STRD, quote=OSMO)
      // - Add TokenPrice(base=ATOM, quote=OSMO)
      // - Query for price of ATOM in STRD
      // - Send 10 ATOM to reward_collector account on Stride
      // - Wait for rewards to get swept from reward_collector to x/auction
      // - Create ATOM auction
      // - Buy ATOM with STRD off auction
      // - Verify STRD was burned by x/strdburner and ATOM was sent to user

      const stridejs = strideAccounts.user;
      const gaiajs = gaiaAccounts.val1;
      const osmojs = osmoAccounts.user;

      let price: string = "0";
      try {
        ({ price } =
          await stridejs.query.stride.icqoracle.tokenPriceForQuoteDenom({
            baseDenom: ATOM_DENOM_ON_STRIDE,
            quoteDenom: USTRD,
          }));
      } catch (error) {}

      // Price should be 2.5. Can skip some steps if already true.
      if (Number(price) !== 2.5) {
        console.log("Transfer STRD to Osmosis");
        await ibcTransfer({
          client: stridejs,
          sourceChain: "STRIDE",
          destinationChain: "OSMO",
          coin: `1000000${USTRD}`,
          sender: stridejs.address,
          receiver: osmojs.address,
        });

        console.log("Transfer ATOM to Osmosis");
        await ibcTransfer({
          client: gaiajs,
          sourceChain: "GAIA",
          destinationChain: "STRIDE",
          coin: `1000000${UATOM}`,
          sender: gaiajs.address,
          receiver: stridejs.address, // needs to be valid but ignored by pfm
          memo: JSON.stringify({
            forward: {
              receiver: osmojs.address,
              port: "transfer",
              channel: TRANSFER_CHANNEL.STRIDE.OSMO,
            },
          }),
        });

        console.log("Create STRD/OSMO pool");
        const createClPoolTx = await submitTxAndExpectSuccess(
          osmojs,
          newConcentratedLiquidityPoolMsg({
            sender: osmojs.address,
            denom0: STRD_DENOM_ON_OSMOSIS,
          }),
        );

        const osmoStrdPoolId = BigInt(
          getValueFromEvents(createClPoolTx.events, "pool_created.pool_id"),
        );

        await submitTxAndExpectSuccess(
          osmojs,
          addConcentratedLiquidityPositionMsg({
            poolId: osmoStrdPoolId,
            sender: osmojs.address,
            tokensProvided: coinsFromString(
              `5${STRD_DENOM_ON_OSMOSIS},10${UOSMO}`,
            ),
            tokenMinAmount0: "5",
            tokenMinAmount1: "10",
          }),
        );

        console.log("Create ATOM/OSMO pool");
        const createGammPoolTx = await submitTxAndExpectSuccess(
          osmojs,
          newGammPoolMsg({
            sender: osmojs.address,
            tokens: [`10${UOSMO}`, `2${ATOM_DENOM_ON_OSMOSIS}`],
            weights: [1, 1],
          }),
        );

        const osmoAtomPoolId = BigInt(
          getValueFromEvents(createGammPoolTx.events, "pool_created.pool_id"),
        );

        console.log("Add TokenPrice(base=STRD, quote=OSMO)");
        await submitTxAndExpectSuccess(
          strideAccounts.admin,
          stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
            admin: strideAccounts.admin.address,
            baseDenom: USTRD,
            quoteDenom: OSMO_DENOM_ON_STRIDE,
            osmosisBaseDenom: STRD_DENOM_ON_OSMOSIS,
            osmosisQuoteDenom: UOSMO,
            osmosisPoolId: osmoStrdPoolId,
          }),
        );

        console.log("Add TokenPrice(base=ATOM, quote=OSMO)");
        await submitTxAndExpectSuccess(
          strideAccounts.admin,
          stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
            admin: strideAccounts.admin.address,
            baseDenom: ATOM_DENOM_ON_STRIDE,
            quoteDenom: OSMO_DENOM_ON_STRIDE,
            osmosisBaseDenom: ATOM_DENOM_ON_OSMOSIS,
            osmosisQuoteDenom: UOSMO,
            osmosisPoolId: osmoAtomPoolId,
          }),
        );

        console.log("Wait for both TokenPrices to be updated");
        while (true) {
          const { tokenPrice } =
            await stridejs.query.stride.icqoracle.tokenPrice({
              baseDenom: USTRD,
              quoteDenom: OSMO_DENOM_ON_STRIDE,
              poolId: osmoStrdPoolId,
            });
          if (
            tokenPrice.lastResponseTime.toISOString() !=
            "0001-01-01T00:00:00.000Z"
          ) {
            expect(Number(tokenPrice.spotPrice)).toBe(2);
            break;
          }
          await sleep(500);
        }
        while (true) {
          const { tokenPrice } =
            await stridejs.query.stride.icqoracle.tokenPrice({
              baseDenom: ATOM_DENOM_ON_STRIDE,
              quoteDenom: OSMO_DENOM_ON_STRIDE,
              poolId: osmoAtomPoolId,
            });
          if (
            tokenPrice.lastResponseTime.toISOString() !=
            "0001-01-01T00:00:00.000Z"
          ) {
            expect(Number(tokenPrice.spotPrice)).toBe(5);
            break;
          }
          await sleep(500);
        }

        console.log("Query for price of ATOM in STRD");
        ({ price } =
          await stridejs.query.stride.icqoracle.tokenPriceForQuoteDenom({
            baseDenom: ATOM_DENOM_ON_STRIDE,
            quoteDenom: USTRD,
          }));

        // Price should be 2.5:
        //
        // TODO: Tind a better way to test this.
        // This will fail if other tests set the price to be something different
        //
        // ATOM/OSMO pool is 2/10 => 1 ATOM is 5 OSMO
        // STRD/OSMO pool is 5/10 => 1 STRD is 2 OSMO
        // =>
        // 2.5 STRD is 5 OSMO
        // =>
        // 1 ATOM is 2.5 STRD
        expect(Number(price)).toBe(2.5);
      }

      const rewardAmount = 10_000000;
      const rewardCollectorAddress = await moduleAddress(
        stridejs,
        "reward_collector",
      );
      console.log("Send 10 ATOM to reward_collector account on Stride");
      await ibcTransfer({
        client: gaiajs,
        sourceChain: "GAIA",
        destinationChain: "STRIDE",
        coin: `${rewardAmount}${UATOM}`,
        sender: gaiajs.address,
        receiver: rewardCollectorAddress,
      });

      console.log(
        "Wait for funds to get swept from reward_collector to auction",
      );
      const auctionAddress = await moduleAddress(stridejs, "auction");
      console.log({ rewardCollectorAddress });
      console.log({ auctionAddress });

      let auctionAtomBalance: string;
      while (true) {
        ({ balance: { amount: auctionAtomBalance } = { amount: "0" } } =
          await stridejs.query.cosmos.bank.v1beta1.balance({
            address: auctionAddress,
            denom: ATOM_DENOM_ON_STRIDE,
          }));

        if (BigInt(auctionAtomBalance) > 0n) {
          break;
        }

        const {
          balance: { amount: rewardCollectorAtomBalance } = { amount: "0" },
        } = await stridejs.query.cosmos.bank.v1beta1.balance({
          address: rewardCollectorAddress,
          denom: ATOM_DENOM_ON_STRIDE,
        });

        console.log({ rewardCollectorAtomBalance });
        console.log({ auctionAtomBalance });
        await sleep(500);
      }

      console.log("Create ATOM auction");
      const auctionName = "ATOM" + Math.random();
      const { address: strdburnerAddress } =
        await stridejs.query.stride.strdburner.strdBurnerAddress({});

      await submitTxAndExpectSuccess(
        strideAccounts.admin,
        stride.auction.MessageComposer.withTypeUrl.createAuction({
          admin: strideAccounts.admin.address,
          auctionName,
          auctionType: stride.auction.AuctionType.AUCTION_TYPE_FCFS,
          sellingDenom: ATOM_DENOM_ON_STRIDE,
          paymentDenom: USTRD,
          enabled: true,
          minPriceMultiplier: "0.95",
          minBidAmount: "1",
          beneficiary: strdburnerAddress,
        }),
      );

      console.log(
        "Buy ATOM with STRD off auction and verify STRD was burned and ATOM was sent to user",
      );
      const { totalBurned: totalBurnedBefore } =
        await stridejs.query.stride.strdburner.totalStrdBurned({});
      const { balance: { amount: userAtomBalanceBefore } = { amount: "0" } } =
        await stridejs.query.cosmos.bank.v1beta1.balance({
          address: strideAccounts.user.address,
          denom: ATOM_DENOM_ON_STRIDE,
        });

      // price is 2.5 and BigInt doesn't support fractions so we'll multiply by 10 for a price of 25
      // and then divide by 10
      const atomsToBuy = BigInt(auctionAtomBalance) / 100n;
      const strdToPay = (BigInt(Number(price) * 10) * atomsToBuy) / 10n;

      await submitTxAndExpectSuccess(
        strideAccounts.user,
        stride.auction.MessageComposer.withTypeUrl.placeBid({
          bidder: strideAccounts.user.address,
          auctionName,
          sellingTokenAmount: String(atomsToBuy),
          paymentTokenAmount: String(strdToPay),
        }),
      );

      const { totalBurned: totalBurnedAfter } =
        await stridejs.query.stride.strdburner.totalStrdBurned({});
      const { balance: { amount: userAtomBalanceAfter } = { amount: "0" } } =
        await stridejs.query.cosmos.bank.v1beta1.balance({
          address: strideAccounts.user.address,
          denom: ATOM_DENOM_ON_STRIDE,
        });

      expect(BigInt(userAtomBalanceAfter)).toBe(
        BigInt(userAtomBalanceBefore) + atomsToBuy,
      );
      expect(BigInt(totalBurnedAfter)).toBe(
        BigInt(totalBurnedBefore) + strdToPay,
      );
    },
    20 * 60 * 1000 /* 20min */,
  );

  // skip due to amino bullshit
  test.skip("update params", async () => {
    const stridejs = strideAccounts.user;

    const { params } = await stridejs.query.stride.icqoracle.params({});
    params.priceExpirationTimeoutSec += 1n;

    const govAddress = await moduleAddress(stridejs, "gov");

    const tx = await stridejs.signAndBroadcast([
      cosmos.gov.v1.MessageComposer.withTypeUrl.submitProposal({
        messages: [
          stride.icqoracle.MsgUpdateParams.toProtoMsg({
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
      strideAccounts.val1.signAndBroadcast(
        [
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: strideAccounts.val1.address,
            option: VoteOption.VOTE_OPTION_YES,
            metadata: "",
          }),
        ],
        2,
      ),
      strideAccounts.val2.signAndBroadcast(
        [
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: strideAccounts.val2.address,
            option: VoteOption.VOTE_OPTION_YES,
            metadata: "",
          }),
        ],
        2,
      ),
      strideAccounts.val3.signAndBroadcast(
        [
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
            proposalId: proposalId,
            voter: strideAccounts.val3.address,
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

  test("staking rewards funneled to x/auction", async () => {
    const stridejs = strideAccounts.admin;
    const gaiajs = gaiaAccounts.user;

    const auctionAddress = await moduleAddress(stridejs, "auction");
    const { balance: { amount: auctionBalanceBefore } = { amount: "0" } } =
      await stridejs.query.cosmos.bank.v1beta1.balance({
        address: auctionAddress,
        denom: ATOM_DENOM_ON_STRIDE,
      });

    const stakeAmount = 10_000_000;
    const rewardAmount = 10_000;
    const feeAmount = 1_000;

    // Liquid stake 10 ATOM
    await ibcTransfer({
      client: gaiajs,
      sourceChain: "GAIA",
      destinationChain: "STRIDE",
      coin: `${stakeAmount}${UATOM}`,
      sender: gaiajs.address,
      receiver: stridejs.address,
    });

    const liquidStakeMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
        creator: stridejs.address,
        amount: String(stakeAmount),
        hostDenom: UATOM,
      });

    await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
    sleep(1000);

    // Check st tokens
    const { balance: { amount: stAtomBalance } = { amount: "0" } } =
      await stridejs.query.cosmos.bank.v1beta1.balance({
        address: stridejs.address,
        denom: "stuatom",
      });
    expect(BigInt(stAtomBalance)).toBeGreaterThan(0n);

    // Send 10% of stake to fee address
    // If we send more, you risk tripping some rate limits
    const {
      hostZone: { withdrawalIcaAddress },
    } = await stridejs.query.stride.stakeibc.hostZone({
      chainId: GAIA_CHAIN_ID,
    });

    await submitTxAndExpectSuccess(gaiajs, [
      cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
        fromAddress: gaiajs.address,
        toAddress: withdrawalIcaAddress,
        amount: coinsFromString(`${rewardAmount}${UATOM}`),
      }),
    ]);

    // Wait for funds to get swept from fee address on gaia into x/auction
    console.log("Waiting for funds to land in auction account");
    while (true) {
      const { balance: { amount: auctionBalanceAfter } = { amount: "0" } } =
        await stridejs.query.cosmos.bank.v1beta1.balance({
          address: auctionAddress,
          denom: ATOM_DENOM_ON_STRIDE,
        });

      if (
        BigInt(auctionBalanceAfter) >=
        BigInt(auctionBalanceBefore) + BigInt(feeAmount)
      ) {
        break;
      }

      await sleep(500);
    }
  }, 240_000);

  test("unwrapIBCDenom", async () => {
    const stridejs = strideAccounts.admin;
    const gaiajs = gaiaAccounts.user;
    const osmojs = osmoAccounts.user;

    // Transfer ATOM & OSMO to Stride to register their ibc denoms on Stride's ibc transfer app
    await ibcTransfer({
      client: gaiajs,
      sourceChain: "GAIA",
      destinationChain: "STRIDE",
      coin: `1${UATOM}`,
      sender: gaiajs.address,
      receiver: stridejs.address,
    });

    await ibcTransfer({
      client: osmojs,
      sourceChain: "OSMO",
      destinationChain: "STRIDE",
      coin: `1${UOSMO}`,
      sender: osmojs.address,
      receiver: stridejs.address,
    });

    const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
      admin: strideAccounts.admin.address,
      baseDenom: ATOM_DENOM_ON_STRIDE,
      quoteDenom: OSMO_DENOM_ON_STRIDE,
      baseDenomOnOsmosis: ATOM_DENOM_ON_OSMOSIS,
      quoteDenomOnOsmosis: UOSMO,
      poolId: 1n, // not important for thie TokenPrice to work for the test to work
    });
    await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);

    const { baseDenomUnwrapped, quoteDenomUnwrapped } =
      await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: ATOM_DENOM_ON_STRIDE,
        quoteDenom: OSMO_DENOM_ON_STRIDE,
        poolId: 1n,
      });

    expect(baseDenomUnwrapped).toBe(UATOM);
    expect(quoteDenomUnwrapped).toBe(UOSMO);
  });
});
