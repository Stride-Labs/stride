import {
    QueryClient,
    setupAuthExtension,
    setupBankExtension,
    setupStakingExtension,
    setupTxExtension,
    SigningStargateClient,
} from "@cosmjs/stargate";
import { Comet38Client } from "@cosmjs/tendermint-rpc";
import {
    DirectSecp256k1HdWallet,
    GasPrice,
    sleep,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
    ATOM_DENOM_ON_STRIDE,
    GAIA_CHAIN_ID,
    GAIA_RPC_ENDPOINT,
    STRIDE_RPC_ENDPOINT,
    TRANSFER_CHANNEL,
    UATOM,
    USTRD,
} from "./consts";
import { CosmosClient } from "./types";
import {
    ibcTransfer,
    waitForChain,
    waitForIbc,
} from "./utils";
import { StrideClient } from "stridejs";

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
            // setup signer for Gaia
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
        }
    }

    console.log("waiting for stride to start...");
    await waitForChain(strideAccounts.user, USTRD);

    console.log("waiting for gaia to start...");
    await waitForChain(gaiaAccounts.user, UATOM);

    console.log("waiting for stride-gaia ibc...");
    await waitForIbc(
        strideAccounts.user,
        TRANSFER_CHANNEL.STRIDE.GAIA!,
        USTRD,
        "cosmos",
    );
}, 45_000);

describe("Core Tests", () => {
    test("IBC Transfer", async () => {
        const stridejs = strideAccounts.user;
        const gaiajs = gaiaAccounts.user;
        const transferAmount = 50000000;

        // Get initial balances - only track what we can reliably query
        const { balance: { amount: strideInitialStrideBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
                address: stridejs.address,
                denom: USTRD,
            });

        const { balance: { amount: strideInitialAtomBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
                address: stridejs.address,
                denom: ATOM_DENOM_ON_STRIDE,
            });

        const gaiaInitialAtomBalance = await gaiajs.query.bank.balance(gaiajs.address, UATOM);

        console.log("Initial balances:");
        console.log(`Stride USTRD: ${strideInitialStrideBalance}`);
        console.log(`Stride ATOM: ${strideInitialAtomBalance}`);
        console.log(`Gaia ATOM: ${gaiaInitialAtomBalance.amount}`);

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

        // Wait a bit for transfers to complete
        await sleep(5000);

        // Get final balances
        const { balance: { amount: strideFinalStrideBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
                address: stridejs.address,
                denom: USTRD,
            });

        const { balance: { amount: strideFinalAtomBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
                address: stridejs.address,
                denom: ATOM_DENOM_ON_STRIDE,
            });

        const gaiaFinalAtomBalance = await gaiajs.query.bank.balance(gaiajs.address, UATOM);

        console.log("Final balances:");
        console.log(`Stride USTRD: ${strideFinalStrideBalance}`);
        console.log(`Stride ATOM: ${strideFinalAtomBalance}`);
        console.log(`Gaia ATOM: ${gaiaFinalAtomBalance.amount}`);

        // Calculate and verify balance changes
        const strideBalanceDiff = BigInt(strideInitialStrideBalance) - BigInt(strideFinalStrideBalance);
        const strideAtomBalanceDiff = BigInt(strideFinalAtomBalance) - BigInt(strideInitialAtomBalance);
        const gaiaAtomBalanceDiff = BigInt(gaiaInitialAtomBalance.amount) - BigInt(gaiaFinalAtomBalance.amount);

        console.log("Balance differences:");
        console.log(`Stride USTRD diff: ${strideBalanceDiff}`);
        console.log(`Stride ATOM diff: ${strideAtomBalanceDiff}`);
        console.log(`Gaia ATOM diff: ${gaiaAtomBalanceDiff}`);

        // Verify the transfers worked (accounting for gas fees)
        // USTRD should decrease on Stride by at least the transfer amount (plus gas fees)
        expect(strideBalanceDiff).toBeGreaterThanOrEqual(BigInt(transferAmount));
        expect(strideBalanceDiff).toBeLessThan(BigInt(transferAmount + 1000000)); // increased gas fee limit

        // ATOM should increase on Stride by exactly the transfer amount
        expect(strideAtomBalanceDiff).toBe(BigInt(transferAmount));

        // ATOM should decrease on Gaia by at least the transfer amount (plus gas fees)
        expect(gaiaAtomBalanceDiff).toBeGreaterThanOrEqual(BigInt(transferAmount));
        expect(gaiaAtomBalanceDiff).toBeLessThan(BigInt(transferAmount + 1000000)); // increased gas fee limit
    }, 120_000);
});
