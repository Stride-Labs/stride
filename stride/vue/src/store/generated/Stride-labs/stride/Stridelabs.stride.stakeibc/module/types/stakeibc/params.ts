/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** Params defines the parameters for the module. */
export interface Params {
  /** define epoch lengths, in blocks */
  sweeping_rewards_interval: number;
  invest_deposits_interval: number;
  calc_exchange_rate_interval: number;
  stride_fee: number;
  /**
   * fee_address_weights stores which addresses to
   * send the Stride fee too, as well as what portion
   * of the fee each address is entitled to
   */
  zone_fee_address: { [key: string]: string };
}

export interface Params_ZoneFeeAddressEntry {
  key: string;
  value: string;
}

const baseParams: object = {
  sweeping_rewards_interval: 0,
  invest_deposits_interval: 0,
  calc_exchange_rate_interval: 0,
  stride_fee: 0,
};

export const Params = {
  encode(message: Params, writer: Writer = Writer.create()): Writer {
    if (message.sweeping_rewards_interval !== 0) {
      writer.uint32(8).uint64(message.sweeping_rewards_interval);
    }
    if (message.invest_deposits_interval !== 0) {
      writer.uint32(16).uint64(message.invest_deposits_interval);
    }
    if (message.calc_exchange_rate_interval !== 0) {
      writer.uint32(24).uint64(message.calc_exchange_rate_interval);
    }
    if (message.stride_fee !== 0) {
      writer.uint32(33).double(message.stride_fee);
    }
    Object.entries(message.zone_fee_address).forEach(([key, value]) => {
      Params_ZoneFeeAddressEntry.encode(
        { key: key as any, value },
        writer.uint32(42).fork()
      ).ldelim();
    });
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Params {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseParams } as Params;
    message.zone_fee_address = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sweeping_rewards_interval = longToNumber(
            reader.uint64() as Long
          );
          break;
        case 2:
          message.invest_deposits_interval = longToNumber(
            reader.uint64() as Long
          );
          break;
        case 3:
          message.calc_exchange_rate_interval = longToNumber(
            reader.uint64() as Long
          );
          break;
        case 4:
          message.stride_fee = reader.double();
          break;
        case 5:
          const entry5 = Params_ZoneFeeAddressEntry.decode(
            reader,
            reader.uint32()
          );
          if (entry5.value !== undefined) {
            message.zone_fee_address[entry5.key] = entry5.value;
          }
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
    message.zone_fee_address = {};
    if (
      object.sweeping_rewards_interval !== undefined &&
      object.sweeping_rewards_interval !== null
    ) {
      message.sweeping_rewards_interval = Number(
        object.sweeping_rewards_interval
      );
    } else {
      message.sweeping_rewards_interval = 0;
    }
    if (
      object.invest_deposits_interval !== undefined &&
      object.invest_deposits_interval !== null
    ) {
      message.invest_deposits_interval = Number(
        object.invest_deposits_interval
      );
    } else {
      message.invest_deposits_interval = 0;
    }
    if (
      object.calc_exchange_rate_interval !== undefined &&
      object.calc_exchange_rate_interval !== null
    ) {
      message.calc_exchange_rate_interval = Number(
        object.calc_exchange_rate_interval
      );
    } else {
      message.calc_exchange_rate_interval = 0;
    }
    if (object.stride_fee !== undefined && object.stride_fee !== null) {
      message.stride_fee = Number(object.stride_fee);
    } else {
      message.stride_fee = 0;
    }
    if (
      object.zone_fee_address !== undefined &&
      object.zone_fee_address !== null
    ) {
      Object.entries(object.zone_fee_address).forEach(([key, value]) => {
        message.zone_fee_address[key] = String(value);
      });
    }
    return message;
  },

  toJSON(message: Params): unknown {
    const obj: any = {};
    message.sweeping_rewards_interval !== undefined &&
      (obj.sweeping_rewards_interval = message.sweeping_rewards_interval);
    message.invest_deposits_interval !== undefined &&
      (obj.invest_deposits_interval = message.invest_deposits_interval);
    message.calc_exchange_rate_interval !== undefined &&
      (obj.calc_exchange_rate_interval = message.calc_exchange_rate_interval);
    message.stride_fee !== undefined && (obj.stride_fee = message.stride_fee);
    obj.zone_fee_address = {};
    if (message.zone_fee_address) {
      Object.entries(message.zone_fee_address).forEach(([k, v]) => {
        obj.zone_fee_address[k] = v;
      });
    }
    return obj;
  },

  fromPartial(object: DeepPartial<Params>): Params {
    const message = { ...baseParams } as Params;
    message.zone_fee_address = {};
    if (
      object.sweeping_rewards_interval !== undefined &&
      object.sweeping_rewards_interval !== null
    ) {
      message.sweeping_rewards_interval = object.sweeping_rewards_interval;
    } else {
      message.sweeping_rewards_interval = 0;
    }
    if (
      object.invest_deposits_interval !== undefined &&
      object.invest_deposits_interval !== null
    ) {
      message.invest_deposits_interval = object.invest_deposits_interval;
    } else {
      message.invest_deposits_interval = 0;
    }
    if (
      object.calc_exchange_rate_interval !== undefined &&
      object.calc_exchange_rate_interval !== null
    ) {
      message.calc_exchange_rate_interval = object.calc_exchange_rate_interval;
    } else {
      message.calc_exchange_rate_interval = 0;
    }
    if (object.stride_fee !== undefined && object.stride_fee !== null) {
      message.stride_fee = object.stride_fee;
    } else {
      message.stride_fee = 0;
    }
    if (
      object.zone_fee_address !== undefined &&
      object.zone_fee_address !== null
    ) {
      Object.entries(object.zone_fee_address).forEach(([key, value]) => {
        if (value !== undefined) {
          message.zone_fee_address[key] = String(value);
        }
      });
    }
    return message;
  },
};

const baseParams_ZoneFeeAddressEntry: object = { key: "", value: "" };

export const Params_ZoneFeeAddressEntry = {
  encode(
    message: Params_ZoneFeeAddressEntry,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): Params_ZoneFeeAddressEntry {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseParams_ZoneFeeAddressEntry,
    } as Params_ZoneFeeAddressEntry;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Params_ZoneFeeAddressEntry {
    const message = {
      ...baseParams_ZoneFeeAddressEntry,
    } as Params_ZoneFeeAddressEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = String(object.value);
    } else {
      message.value = "";
    }
    return message;
  },

  toJSON(message: Params_ZoneFeeAddressEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(
    object: DeepPartial<Params_ZoneFeeAddressEntry>
  ): Params_ZoneFeeAddressEntry {
    const message = {
      ...baseParams_ZoneFeeAddressEntry,
    } as Params_ZoneFeeAddressEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = "";
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
