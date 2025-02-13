import {
  SigningStargateClient,
  QueryClient,
  AuthExtension,
  BankExtension,
  StakingExtension,
  TxExtension,
} from "@cosmjs/stargate";

export type Chain = "STRIDE" | "GAIA" | "OSMO";

export type CosmosClient = {
  address: string;
  client: SigningStargateClient;
  query: QueryClient &
    AuthExtension &
    BankExtension &
    StakingExtension &
    TxExtension;
};
