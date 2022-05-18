/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "stride.mint.v1beta1";

/** Minter represents the minting state. */
export interface Minter {
  /** current epoch provisions */
  epoch_provisions: string;
}

export interface DistributionProportions {
  /**
   * staking defines the proportion of the minted minted_denom that is to be
   * allocated as staking rewards.
   */
  staking: string;
  /**
   * pool_incentives defines the proportion of the minted minted_denom that is
   * to be allocated as pool incentives.
   */
  pool_incentives: string;
  /**
   * participation_rewards defines the proportion of the minted minted_denom that is
   * to be allocated to participation rewards address.
   */
  participation_rewards: string;
  /**
   * community_pool defines the proportion of the minted minted_denom that is
   * to be allocated to the community pool.
   */
  community_pool: string;
}

/** Params holds parameters for the mint module. */
export interface Params {
  /** type of coin to mint */
  mint_denom: string;
  /** epoch provisions from the first epoch */
  genesis_epoch_provisions: string;
  /** mint epoch identifier */
  epoch_identifier: string;
  /** number of epochs take to reduce rewards */
  reduction_period_in_epochs: number;
  /** reduction multiplier to execute on each period */
  reduction_factor: string;
  /** distribution_proportions defines the proportion of the minted denom */
  distribution_proportions: DistributionProportions | undefined;
  /** start epoch to distribute minting rewards */
  minting_rewards_distribution_start_epoch: number;
}

const baseMinter: object = { epoch_provisions: "" };

export const Minter = {
  encode(message: Minter, writer: Writer = Writer.create()): Writer {
    if (message.epoch_provisions !== "") {
      writer.uint32(10).string(message.epoch_provisions);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Minter {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMinter } as Minter;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.epoch_provisions = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Minter {
    const message = { ...baseMinter } as Minter;
    if (
      object.epoch_provisions !== undefined &&
      object.epoch_provisions !== null
    ) {
      message.epoch_provisions = String(object.epoch_provisions);
    } else {
      message.epoch_provisions = "";
    }
    return message;
  },

  toJSON(message: Minter): unknown {
    const obj: any = {};
    message.epoch_provisions !== undefined &&
      (obj.epoch_provisions = message.epoch_provisions);
    return obj;
  },

  fromPartial(object: DeepPartial<Minter>): Minter {
    const message = { ...baseMinter } as Minter;
    if (
      object.epoch_provisions !== undefined &&
      object.epoch_provisions !== null
    ) {
      message.epoch_provisions = object.epoch_provisions;
    } else {
      message.epoch_provisions = "";
    }
    return message;
  },
};

const baseDistributionProportions: object = {
  staking: "",
  pool_incentives: "",
  participation_rewards: "",
  community_pool: "",
};

export const DistributionProportions = {
  encode(
    message: DistributionProportions,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.staking !== "") {
      writer.uint32(10).string(message.staking);
    }
    if (message.pool_incentives !== "") {
      writer.uint32(18).string(message.pool_incentives);
    }
    if (message.participation_rewards !== "") {
      writer.uint32(26).string(message.participation_rewards);
    }
    if (message.community_pool !== "") {
      writer.uint32(34).string(message.community_pool);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): DistributionProportions {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDistributionProportions,
    } as DistributionProportions;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.staking = reader.string();
          break;
        case 2:
          message.pool_incentives = reader.string();
          break;
        case 3:
          message.participation_rewards = reader.string();
          break;
        case 4:
          message.community_pool = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DistributionProportions {
    const message = {
      ...baseDistributionProportions,
    } as DistributionProportions;
    if (object.staking !== undefined && object.staking !== null) {
      message.staking = String(object.staking);
    } else {
      message.staking = "";
    }
    if (
      object.pool_incentives !== undefined &&
      object.pool_incentives !== null
    ) {
      message.pool_incentives = String(object.pool_incentives);
    } else {
      message.pool_incentives = "";
    }
    if (
      object.participation_rewards !== undefined &&
      object.participation_rewards !== null
    ) {
      message.participation_rewards = String(object.participation_rewards);
    } else {
      message.participation_rewards = "";
    }
    if (object.community_pool !== undefined && object.community_pool !== null) {
      message.community_pool = String(object.community_pool);
    } else {
      message.community_pool = "";
    }
    return message;
  },

  toJSON(message: DistributionProportions): unknown {
    const obj: any = {};
    message.staking !== undefined && (obj.staking = message.staking);
    message.pool_incentives !== undefined &&
      (obj.pool_incentives = message.pool_incentives);
    message.participation_rewards !== undefined &&
      (obj.participation_rewards = message.participation_rewards);
    message.community_pool !== undefined &&
      (obj.community_pool = message.community_pool);
    return obj;
  },

  fromPartial(
    object: DeepPartial<DistributionProportions>
  ): DistributionProportions {
    const message = {
      ...baseDistributionProportions,
    } as DistributionProportions;
    if (object.staking !== undefined && object.staking !== null) {
      message.staking = object.staking;
    } else {
      message.staking = "";
    }
    if (
      object.pool_incentives !== undefined &&
      object.pool_incentives !== null
    ) {
      message.pool_incentives = object.pool_incentives;
    } else {
      message.pool_incentives = "";
    }
    if (
      object.participation_rewards !== undefined &&
      object.participation_rewards !== null
    ) {
      message.participation_rewards = object.participation_rewards;
    } else {
      message.participation_rewards = "";
    }
    if (object.community_pool !== undefined && object.community_pool !== null) {
      message.community_pool = object.community_pool;
    } else {
      message.community_pool = "";
    }
    return message;
  },
};

const baseParams: object = {
  mint_denom: "",
  genesis_epoch_provisions: "",
  epoch_identifier: "",
  reduction_period_in_epochs: 0,
  reduction_factor: "",
  minting_rewards_distribution_start_epoch: 0,
};

export const Params = {
  encode(message: Params, writer: Writer = Writer.create()): Writer {
    if (message.mint_denom !== "") {
      writer.uint32(10).string(message.mint_denom);
    }
    if (message.genesis_epoch_provisions !== "") {
      writer.uint32(18).string(message.genesis_epoch_provisions);
    }
    if (message.epoch_identifier !== "") {
      writer.uint32(26).string(message.epoch_identifier);
    }
    if (message.reduction_period_in_epochs !== 0) {
      writer.uint32(32).int64(message.reduction_period_in_epochs);
    }
    if (message.reduction_factor !== "") {
      writer.uint32(42).string(message.reduction_factor);
    }
    if (message.distribution_proportions !== undefined) {
      DistributionProportions.encode(
        message.distribution_proportions,
        writer.uint32(50).fork()
      ).ldelim();
    }
    if (message.minting_rewards_distribution_start_epoch !== 0) {
      writer.uint32(56).int64(message.minting_rewards_distribution_start_epoch);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Params {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseParams } as Params;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.mint_denom = reader.string();
          break;
        case 2:
          message.genesis_epoch_provisions = reader.string();
          break;
        case 3:
          message.epoch_identifier = reader.string();
          break;
        case 4:
          message.reduction_period_in_epochs = longToNumber(
            reader.int64() as Long
          );
          break;
        case 5:
          message.reduction_factor = reader.string();
          break;
        case 6:
          message.distribution_proportions = DistributionProportions.decode(
            reader,
            reader.uint32()
          );
          break;
        case 7:
          message.minting_rewards_distribution_start_epoch = longToNumber(
            reader.int64() as Long
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Params {
    const message = { ...baseParams } as Params;
    if (object.mint_denom !== undefined && object.mint_denom !== null) {
      message.mint_denom = String(object.mint_denom);
    } else {
      message.mint_denom = "";
    }
    if (
      object.genesis_epoch_provisions !== undefined &&
      object.genesis_epoch_provisions !== null
    ) {
      message.genesis_epoch_provisions = String(
        object.genesis_epoch_provisions
      );
    } else {
      message.genesis_epoch_provisions = "";
    }
    if (
      object.epoch_identifier !== undefined &&
      object.epoch_identifier !== null
    ) {
      message.epoch_identifier = String(object.epoch_identifier);
    } else {
      message.epoch_identifier = "";
    }
    if (
      object.reduction_period_in_epochs !== undefined &&
      object.reduction_period_in_epochs !== null
    ) {
      message.reduction_period_in_epochs = Number(
        object.reduction_period_in_epochs
      );
    } else {
      message.reduction_period_in_epochs = 0;
    }
    if (
      object.reduction_factor !== undefined &&
      object.reduction_factor !== null
    ) {
      message.reduction_factor = String(object.reduction_factor);
    } else {
      message.reduction_factor = "";
    }
    if (
      object.distribution_proportions !== undefined &&
      object.distribution_proportions !== null
    ) {
      message.distribution_proportions = DistributionProportions.fromJSON(
        object.distribution_proportions
      );
    } else {
      message.distribution_proportions = undefined;
    }
    if (
      object.minting_rewards_distribution_start_epoch !== undefined &&
      object.minting_rewards_distribution_start_epoch !== null
    ) {
      message.minting_rewards_distribution_start_epoch = Number(
        object.minting_rewards_distribution_start_epoch
      );
    } else {
      message.minting_rewards_distribution_start_epoch = 0;
    }
    return message;
  },

  toJSON(message: Params): unknown {
    const obj: any = {};
    message.mint_denom !== undefined && (obj.mint_denom = message.mint_denom);
    message.genesis_epoch_provisions !== undefined &&
      (obj.genesis_epoch_provisions = message.genesis_epoch_provisions);
    message.epoch_identifier !== undefined &&
      (obj.epoch_identifier = message.epoch_identifier);
    message.reduction_period_in_epochs !== undefined &&
      (obj.reduction_period_in_epochs = message.reduction_period_in_epochs);
    message.reduction_factor !== undefined &&
      (obj.reduction_factor = message.reduction_factor);
    message.distribution_proportions !== undefined &&
      (obj.distribution_proportions = message.distribution_proportions
        ? DistributionProportions.toJSON(message.distribution_proportions)
        : undefined);
    message.minting_rewards_distribution_start_epoch !== undefined &&
      (obj.minting_rewards_distribution_start_epoch =
        message.minting_rewards_distribution_start_epoch);
    return obj;
  },

  fromPartial(object: DeepPartial<Params>): Params {
    const message = { ...baseParams } as Params;
    if (object.mint_denom !== undefined && object.mint_denom !== null) {
      message.mint_denom = object.mint_denom;
    } else {
      message.mint_denom = "";
    }
    if (
      object.genesis_epoch_provisions !== undefined &&
      object.genesis_epoch_provisions !== null
    ) {
      message.genesis_epoch_provisions = object.genesis_epoch_provisions;
    } else {
      message.genesis_epoch_provisions = "";
    }
    if (
      object.epoch_identifier !== undefined &&
      object.epoch_identifier !== null
    ) {
      message.epoch_identifier = object.epoch_identifier;
    } else {
      message.epoch_identifier = "";
    }
    if (
      object.reduction_period_in_epochs !== undefined &&
      object.reduction_period_in_epochs !== null
    ) {
      message.reduction_period_in_epochs = object.reduction_period_in_epochs;
    } else {
      message.reduction_period_in_epochs = 0;
    }
    if (
      object.reduction_factor !== undefined &&
      object.reduction_factor !== null
    ) {
      message.reduction_factor = object.reduction_factor;
    } else {
      message.reduction_factor = "";
    }
    if (
      object.distribution_proportions !== undefined &&
      object.distribution_proportions !== null
    ) {
      message.distribution_proportions = DistributionProportions.fromPartial(
        object.distribution_proportions
      );
    } else {
      message.distribution_proportions = undefined;
    }
    if (
      object.minting_rewards_distribution_start_epoch !== undefined &&
      object.minting_rewards_distribution_start_epoch !== null
    ) {
      message.minting_rewards_distribution_start_epoch =
        object.minting_rewards_distribution_start_epoch;
    } else {
      message.minting_rewards_distribution_start_epoch = 0;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
