/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** Params defines the parameters for the module. */
export interface Params {
  /** define epoch lengths, in stride_epochs */
  rewards_interval: number;
  deposit_interval: number;
  exchange_rate_interval: number;
  stride_commission: number;
  /**
   * zone_com_address stores which addresses to
   * send the Stride commission too, as well as what portion
   * of the fee each address is entitled to
   */
  zone_com_address: { [key: string]: string };
}

export interface Params_ZoneComAddressEntry {
  key: string;
  value: string;
}

const baseParams: object = {
  rewards_interval: 0,
  deposit_interval: 0,
  exchange_rate_interval: 0,
  stride_commission: 0,
};

export const Params = {
  encode(message: Params, writer: Writer = Writer.create()): Writer {
    if (message.rewards_interval !== 0) {
      writer.uint32(8).uint64(message.rewards_interval);
    }
    if (message.deposit_interval !== 0) {
      writer.uint32(16).uint64(message.deposit_interval);
    }
    if (message.exchange_rate_interval !== 0) {
      writer.uint32(24).uint64(message.exchange_rate_interval);
    }
    if (message.stride_commission !== 0) {
      writer.uint32(32).uint64(message.stride_commission);
    }
    Object.entries(message.zone_com_address).forEach(([key, value]) => {
      Params_ZoneComAddressEntry.encode(
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
    message.zone_com_address = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rewards_interval = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.deposit_interval = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.exchange_rate_interval = longToNumber(
            reader.uint64() as Long
          );
          break;
        case 4:
          message.stride_commission = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          const entry5 = Params_ZoneComAddressEntry.decode(
            reader,
            reader.uint32()
          );
          if (entry5.value !== undefined) {
            message.zone_com_address[entry5.key] = entry5.value;
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
    message.zone_com_address = {};
    if (
      object.rewards_interval !== undefined &&
      object.rewards_interval !== null
    ) {
      message.rewards_interval = Number(object.rewards_interval);
    } else {
      message.rewards_interval = 0;
    }
    if (
      object.deposit_interval !== undefined &&
      object.deposit_interval !== null
    ) {
      message.deposit_interval = Number(object.deposit_interval);
    } else {
      message.deposit_interval = 0;
    }
    if (
      object.exchange_rate_interval !== undefined &&
      object.exchange_rate_interval !== null
    ) {
      message.exchange_rate_interval = Number(object.exchange_rate_interval);
    } else {
      message.exchange_rate_interval = 0;
    }
    if (
      object.stride_commission !== undefined &&
      object.stride_commission !== null
    ) {
      message.stride_commission = Number(object.stride_commission);
    } else {
      message.stride_commission = 0;
    }
    if (
      object.zone_com_address !== undefined &&
      object.zone_com_address !== null
    ) {
      Object.entries(object.zone_com_address).forEach(([key, value]) => {
        message.zone_com_address[key] = String(value);
      });
    }
    return message;
  },

  toJSON(message: Params): unknown {
    const obj: any = {};
    message.rewards_interval !== undefined &&
      (obj.rewards_interval = message.rewards_interval);
    message.deposit_interval !== undefined &&
      (obj.deposit_interval = message.deposit_interval);
    message.exchange_rate_interval !== undefined &&
      (obj.exchange_rate_interval = message.exchange_rate_interval);
    message.stride_commission !== undefined &&
      (obj.stride_commission = message.stride_commission);
    obj.zone_com_address = {};
    if (message.zone_com_address) {
      Object.entries(message.zone_com_address).forEach(([k, v]) => {
        obj.zone_com_address[k] = v;
      });
    }
    return obj;
  },

  fromPartial(object: DeepPartial<Params>): Params {
    const message = { ...baseParams } as Params;
    message.zone_com_address = {};
    if (
      object.rewards_interval !== undefined &&
      object.rewards_interval !== null
    ) {
      message.rewards_interval = object.rewards_interval;
    } else {
      message.rewards_interval = 0;
    }
    if (
      object.deposit_interval !== undefined &&
      object.deposit_interval !== null
    ) {
      message.deposit_interval = object.deposit_interval;
    } else {
      message.deposit_interval = 0;
    }
    if (
      object.exchange_rate_interval !== undefined &&
      object.exchange_rate_interval !== null
    ) {
      message.exchange_rate_interval = object.exchange_rate_interval;
    } else {
      message.exchange_rate_interval = 0;
    }
    if (
      object.stride_commission !== undefined &&
      object.stride_commission !== null
    ) {
      message.stride_commission = object.stride_commission;
    } else {
      message.stride_commission = 0;
    }
    if (
      object.zone_com_address !== undefined &&
      object.zone_com_address !== null
    ) {
      Object.entries(object.zone_com_address).forEach(([key, value]) => {
        if (value !== undefined) {
          message.zone_com_address[key] = String(value);
        }
      });
    }
    return message;
  },
};

const baseParams_ZoneComAddressEntry: object = { key: "", value: "" };

export const Params_ZoneComAddressEntry = {
  encode(
    message: Params_ZoneComAddressEntry,
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
  ): Params_ZoneComAddressEntry {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseParams_ZoneComAddressEntry,
    } as Params_ZoneComAddressEntry;
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

  fromJSON(object: any): Params_ZoneComAddressEntry {
    const message = {
      ...baseParams_ZoneComAddressEntry,
    } as Params_ZoneComAddressEntry;
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

  toJSON(message: Params_ZoneComAddressEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(
    object: DeepPartial<Params_ZoneComAddressEntry>
  ): Params_ZoneComAddressEntry {
    const message = {
      ...baseParams_ZoneComAddressEntry,
    } as Params_ZoneComAddressEntry;
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
