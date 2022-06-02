/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

export interface DepositRecord {
  id: number;
  /** TODO do we care that amount is int32? should we change this to uint64? */
  amount: number;
  denom: string;
  hostZoneId: number;
  sender: string;
  purpose: DepositRecord_Purpose;
}

export enum DepositRecord_Purpose {
  RECEIPT = 0,
  TRANSACTION = 1,
  UNRECOGNIZED = -1,
}

export function depositRecord_PurposeFromJSON(
  object: any
): DepositRecord_Purpose {
  switch (object) {
    case 0:
    case "RECEIPT":
      return DepositRecord_Purpose.RECEIPT;
    case 1:
    case "TRANSACTION":
      return DepositRecord_Purpose.TRANSACTION;
    case -1:
    case "UNRECOGNIZED":
    default:
      return DepositRecord_Purpose.UNRECOGNIZED;
  }
}

export function depositRecord_PurposeToJSON(
  object: DepositRecord_Purpose
): string {
  switch (object) {
    case DepositRecord_Purpose.RECEIPT:
      return "RECEIPT";
    case DepositRecord_Purpose.TRANSACTION:
      return "TRANSACTION";
    default:
      return "UNKNOWN";
  }
}

const baseDepositRecord: object = {
  id: 0,
  amount: 0,
  denom: "",
  hostZoneId: 0,
  sender: "",
  purpose: 0,
};

export const DepositRecord = {
  encode(message: DepositRecord, writer: Writer = Writer.create()): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    if (message.amount !== 0) {
      writer.uint32(16).int32(message.amount);
    }
    if (message.denom !== "") {
      writer.uint32(26).string(message.denom);
    }
    if (message.hostZoneId !== 0) {
      writer.uint32(32).uint64(message.hostZoneId);
    }
    if (message.sender !== "") {
      writer.uint32(42).string(message.sender);
    }
    if (message.purpose !== 0) {
      writer.uint32(48).int32(message.purpose);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): DepositRecord {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDepositRecord } as DepositRecord;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.amount = reader.int32();
          break;
        case 3:
          message.denom = reader.string();
          break;
        case 4:
          message.hostZoneId = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.sender = reader.string();
          break;
        case 6:
          message.purpose = reader.int32() as any;
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DepositRecord {
    const message = { ...baseDepositRecord } as DepositRecord;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = Number(object.amount);
    } else {
      message.amount = 0;
    }
    if (object.denom !== undefined && object.denom !== null) {
      message.denom = String(object.denom);
    } else {
      message.denom = "";
    }
    if (object.hostZoneId !== undefined && object.hostZoneId !== null) {
      message.hostZoneId = Number(object.hostZoneId);
    } else {
      message.hostZoneId = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender);
    } else {
      message.sender = "";
    }
    if (object.purpose !== undefined && object.purpose !== null) {
      message.purpose = depositRecord_PurposeFromJSON(object.purpose);
    } else {
      message.purpose = 0;
    }
    return message;
  },

  toJSON(message: DepositRecord): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.amount !== undefined && (obj.amount = message.amount);
    message.denom !== undefined && (obj.denom = message.denom);
    message.hostZoneId !== undefined && (obj.hostZoneId = message.hostZoneId);
    message.sender !== undefined && (obj.sender = message.sender);
    message.purpose !== undefined &&
      (obj.purpose = depositRecord_PurposeToJSON(message.purpose));
    return obj;
  },

  fromPartial(object: DeepPartial<DepositRecord>): DepositRecord {
    const message = { ...baseDepositRecord } as DepositRecord;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.amount !== undefined && object.amount !== null) {
      message.amount = object.amount;
    } else {
      message.amount = 0;
    }
    if (object.denom !== undefined && object.denom !== null) {
      message.denom = object.denom;
    } else {
      message.denom = "";
    }
    if (object.hostZoneId !== undefined && object.hostZoneId !== null) {
      message.hostZoneId = object.hostZoneId;
    } else {
      message.hostZoneId = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender;
    } else {
      message.sender = "";
    }
    if (object.purpose !== undefined && object.purpose !== null) {
      message.purpose = object.purpose;
    } else {
      message.purpose = 0;
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
