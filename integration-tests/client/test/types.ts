import {
  SigningStargateClient,
  QueryClient,
  AuthExtension,
  BankExtension,
  StakingExtension,
  TxExtension,
  IbcExtension,
} from "@cosmjs/stargate";

export type Chain = "STRIDE" | "GAIA" | "OSMO";

export type CosmosClient = {
  address: string;
  denom: string;
  client: SigningStargateClient;
  query: QueryClient & AuthExtension & BankExtension & StakingExtension & IbcExtension & TxExtension;
};
