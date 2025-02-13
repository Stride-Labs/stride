import { DirectSecp256k1HdWallet, Registry } from "@cosmjs/proto-signing";
import {
  GasPrice,
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
import { ModuleAccount } from "osmojs/cosmos/auth/v1beta1/auth";
import { ProposalStatus, VoteOption } from "osmojs/cosmos/gov/v1beta1/gov";
import {
  coinsFromString,
  convertBech32Prefix,
  cosmos,
  decToString,
  EncodeObject,
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
  submitTxAndExpectSuccess,
  waitForChain,
  waitForIbc,
} from "./utils";

let accounts: {
  user: StrideClient; // a normal account loaded with 100 STRD
  admin: StrideClient; // the stride admin account loaded with 1000 STRD
  val1: StrideClient;
  val2: StrideClient;
  val3: StrideClient;
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

  console.log("waiting for stride-gaia ibc...");
  await waitForIbc(
    accounts.user,
    TRANSFER_CHANNEL.STRIDE.GAIA!,
    USTRD,
    "cosmos",
  );

  console.log("waiting for stride-osmosis ibc...");
  await waitForIbc(accounts.user, TRANSFER_CHANNEL.STRIDE.OSMO!, USTRD, "osmo");

  console.log("registering host zones...");

  const registerHostZonesMsgs: EncodeObject[] = [];
  const { hostZone } = await accounts.admin.query.stride.stakeibc.hostZoneAll(
    {},
  );

  const gaiaHostZoneNotRegistered =
    hostZone.find((hz) => hz.chainId === GAIA_CHAIN_ID) === undefined;
  const osmoHostZoneNotRegistered =
    hostZone.find((hz) => hz.chainId === OSMO_CHAIN_ID) === undefined;

  if (gaiaHostZoneNotRegistered) {
    const gaiaRegisterHostZoneMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.registerHostZone({
        creator: accounts.admin.address,
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
        creator: accounts.admin.address,
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
        creator: accounts.admin.address,
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
        creator: accounts.admin.address,
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
    await submitTxAndExpectSuccess(accounts.admin, registerHostZonesMsgs);
  }
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
    const stridejs = accounts.user;

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
    const stridejs = accounts.admin;

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
    const stridejs = accounts.user;
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
      admin: accounts.admin.address,
      baseDenom: USTRD,
      quoteDenom: OSMO_DENOM_ON_STRIDE,
      baseDenomOnOsmosis: STRD_DENOM_ON_OSMOSIS,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(accounts.admin, registerTokenPriceMsg);

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
          queryInProgress,
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
    const stridejs = accounts.user;
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
      admin: accounts.admin.address,
      baseDenom: USTRD,
      quoteDenom: OSMO_DENOM_ON_STRIDE,
      baseDenomOnOsmosis: STRD_DENOM_ON_OSMOSIS,
      quoteDenomOnOsmosis: UOSMO,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(accounts.admin, registerTokenPriceMsg);

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
          queryInProgress,
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
    await ibcTransfer({
      client: stridejs,
      sourceChain: "STRIDE",
      destinationChain: "OSMO",
      coin: `1000000${USTRD}`,
      sender: stridejs.address,
      receiver: osmojs.address,
    });

    // Transfer ATOM to Osmosis
    await ibcTransfer({
      client: stridejs,
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
    // Create STRD/OSMO pool

    let osmoTx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        newConcentratedLiquidityPoolMsg({
          sender: osmojs.address,
          denom0: STRD_DENOM_ON_OSMOSIS,
        }),
      ],
      2,
    );
    if (osmoTx.code !== 0) {
      console.error(osmoTx.rawLog);
    }
    expect(osmoTx.code).toBe(0);

    const osmoStrdPoolId = BigInt(
      getValueFromEvents(osmoTx.events, "pool_created.pool_id"),
    );

    osmoTx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        addConcentratedLiquidityPositionMsg({
          poolId: osmoStrdPoolId,
          sender: osmojs.address,
          tokensProvided: coinsFromString(
            `5${STRD_DENOM_ON_OSMOSIS},10${UOSMO}`,
          ),
          tokenMinAmount0: "5",
          tokenMinAmount1: "10",
        }),
      ],
      2,
    );
    if (osmoTx.code !== 0) {
      console.error(osmoTx.rawLog);
    }
    expect(osmoTx.code).toBe(0);

    // Create ATOM/OSMO pool
    osmoTx = await osmojs.client.signAndBroadcast(
      osmojs.address,
      [
        newGammPoolMsg({
          sender: osmojs.address,
          tokens: [`10${UOSMO}`, `2${ATOM_DENOM_ON_OSMOSIS}`],
          weights: [1, 1],
        }),
      ],
      2,
    );
    if (osmoTx.code !== 0) {
      console.error(osmoTx.rawLog);
    }
    expect(osmoTx.code).toBe(0);

    const osmoAtomPoolId = BigInt(
      getValueFromEvents(osmoTx.events, "pool_created.pool_id"),
    );

    // Add TokenPrice(base=STRD, quote=OSMO)
    let strideTx = await accounts.admin.signAndBroadcast(
      [
        stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
          admin: accounts.admin.address,
          baseDenom: USTRD,
          quoteDenom: OSMO_DENOM_ON_STRIDE,
          osmosisBaseDenom: STRD_DENOM_ON_OSMOSIS,
          osmosisQuoteDenom: UOSMO,
          osmosisPoolId: osmoStrdPoolId,
        }),
      ],
      2,
    );
    if (strideTx.code !== 0) {
      console.error(strideTx.rawLog);
    }
    expect(strideTx.code).toBe(0);

    // Add TokenPrice(base=ATOM, quote=OSMO)
    strideTx = await accounts.admin.signAndBroadcast(
      [
        stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
          admin: accounts.admin.address,
          baseDenom: ATOM_DENOM_ON_STRIDE,
          quoteDenom: OSMO_DENOM_ON_STRIDE,
          osmosisBaseDenom: ATOM_DENOM_ON_OSMOSIS,
          osmosisQuoteDenom: UOSMO,
          osmosisPoolId: osmoAtomPoolId,
        }),
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
        quoteDenom: OSMO_DENOM_ON_STRIDE,
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
        baseDenom: ATOM_DENOM_ON_STRIDE,
        quoteDenom: OSMO_DENOM_ON_STRIDE,
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
        baseDenom: ATOM_DENOM_ON_STRIDE,
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
      accounts.val1.signAndBroadcast(
        [
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
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
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
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
          cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
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

  test("auction + strdburner", async () => {
    const stridejs = accounts.user;
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
      admin: accounts.admin.address,
      baseDenom: OSMO_DENOM_ON_STRIDE,
      quoteDenom: USTRD,
      baseDenomOnOsmosis: UOSMO,
      quoteDenomOnOsmosis: STRD_DENOM_ON_OSMOSIS,
      poolId: osmoStrdPoolId,
    });
    await submitTxAndExpectSuccess(accounts.admin, registerTokenPriceMsg);

    while (true) {
      const {
        tokenPrice: { spotPrice, lastResponseTime },
      } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: USTRD,
        quoteDenom: OSMO_DENOM_ON_STRIDE,
        poolId: osmoStrdPoolId,
      });
      if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
        expect(Number(spotPrice)).toBe(5);
        break;
      }
      await sleep(500);
    }

    const auctionName = String(Math.random());

    const strdburnerAddress = (
      (
        await stridejs.query.cosmos.auth.v1beta1.moduleAccountByName({
          name: "strdburner",
        })
      ).account as ModuleAccount
    ).baseAccount?.address!;

    const tx = await accounts.admin.signAndBroadcast(
      [
        stride.auction.MessageComposer.withTypeUrl.createAuction({
          admin: accounts.admin.address,
          auctionName: auctionName,
          auctionType: 1, //AuctionType.AUCTION_TYPE_FCFS,
          sellingDenom: OSMO_DENOM_ON_STRIDE,
          paymentDenom: USTRD,
          enabled: true,
          minPriceMultiplier: "0.95",
          minBidAmount: "1",
          beneficiary: strdburnerAddress,
        }),
      ],
      2,
    );

    // const { balance: balanceBefore } =
    //   await stridejs.query.cosmos.bank.v1beta1.balance({
    //     address: strdburnerAddress,
    //     denom: USTRD,
    //   });

    // expect(balanceBefore?.amount).toBe("0");

    // const { totalBurned: totalBurnedBefore } =
    //   await stridejs.query.stride.strdburner.totalStrdBurned({});

    // const amount = 100;

    // const tx = await stridejs.signAndBroadcast([
    //   cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
    //     fromAddress: stridejs.address,
    //     toAddress: strdburnerAddress,
    //     amount: coinsFromString(`${amount}${USTRD}`),
    //   }),
    // ]);
    // if (tx.code !== 0) {
    //   console.error(tx.rawLog);
    // }
    // expect(tx.code).toBe(0);

    // const { balance: balanceAfter } =
    //   await stridejs.query.cosmos.bank.v1beta1.balance({
    //     address: strdburnerAddress,
    //     denom: USTRD,
    //   });

    // expect(balanceAfter?.amount).toBe("0");

    // const { totalBurned: totalBurnedAfter } =
    //   await stridejs.query.stride.strdburner.totalStrdBurned({});

    // expect(BigInt(totalBurnedAfter) - BigInt(totalBurnedBefore)).toBe(amount);
  });

  test.only("registration, liquid stake and collect rewards", async () => {
    const stridejs = accounts.admin;
    const gaiajs = gaiaAccounts.user;

    const amount = String(1_000_000);

    await ibcTransfer({
      client: gaiajs,
      sourceChain: "GAIA",
      destinationChain: "STRIDE",
      coin: `${amount}${UATOM}`,
      sender: gaiajs.address,
      receiver: stridejs.address,
    });

    const liquidStakeMsg =
      stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
        creator: stridejs.address,
        amount,
        hostDenom: UATOM,
      });

    await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);

    // Wait for 2 STRIDE_DAY_EPOCH to pass (140s)
    await sleep(2 * 140_000);

    // Check Rewards
    const auctionAddress = (
      (
        await stridejs.query.cosmos.auth.v1beta1.moduleAccountByName({
          name: "auction",
        })
      ).account as ModuleAccount
    ).baseAccount?.address!;

    const { balances } = await stridejs.query.cosmos.bank.v1beta1.allBalances({
      address: auctionAddress,
    });

    console.log(balances);
  }, 360_000); // Set timeout to 4 minutes to account for epoch waits

  // TODO test unwrapIBCDenom via stridejs.query.stride.icqoracle.tokenPrices
});
