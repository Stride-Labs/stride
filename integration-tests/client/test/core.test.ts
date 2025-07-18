import {
    StrideClient
} from "stridejs";
import { CosmosClient } from "./types";

let strideAccounts: {
    user: StrideClient; // a normal account loaded with 100 STRD
    admin: StrideClient; // the stride admin account loaded with 1000 STRD
    val1: StrideClient;
    val2: StrideClient;
    val3: StrideClient;
};

let gaiaAccounts: {
    user: CosmosClient; // a normal account loaded with 1000 ATOM
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