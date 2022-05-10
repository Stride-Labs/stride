/* eslint-disable */
import { Timestamp } from "../google/protobuf/timestamp";
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Duration } from "../google/protobuf/duration";

export const protobufPackage = "Stridelabs.stride.epochs";

export interface EpochInfo {
  identifier: string;
  start_time: Date | undefined;
  duration: Duration | undefined;
  current_epoch: number;
  current_epoch_start_time: Date | undefined;
  epoch_counting_started: boolean;
  current_epoch_start_height: number;
}

/** GenesisState defines the epochs module's genesis state. */
export interface GenesisState {
  epochs: EpochInfo[];
}

const baseEpochInfo: object = {
  identifier: "",
  current_epoch: 0,
  epoch_counting_started: false,
  current_epoch_start_height: 0,
};

export const EpochInfo = {
  encode(message: EpochInfo, writer: Writer = Writer.create()): Writer {
    if (message.identifier !== "") {
      writer.uint32(10).string(message.identifier);
    }
    if (message.start_time !== undefined) {
      Timestamp.encode(
        toTimestamp(message.start_time),
        writer.uint32(18).fork()
      ).ldelim();
    }
    if (message.duration !== undefined) {
      Duration.encode(message.duration, writer.uint32(26).fork()).ldelim();
    }
    if (message.current_epoch !== 0) {
      writer.uint32(32).int64(message.current_epoch);
    }
    if (message.current_epoch_start_time !== undefined) {
      Timestamp.encode(
        toTimestamp(message.current_epoch_start_time),
        writer.uint32(42).fork()
      ).ldelim();
    }
    if (message.epoch_counting_started === true) {
      writer.uint32(48).bool(message.epoch_counting_started);
    }
    if (message.current_epoch_start_height !== 0) {
      writer.uint32(56).int64(message.current_epoch_start_height);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): EpochInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseEpochInfo } as EpochInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.identifier = reader.string();
          break;
        case 2:
          message.start_time = fromTimestamp(
            Timestamp.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.duration = Duration.decode(reader, reader.uint32());
          break;
        case 4:
          message.current_epoch = longToNumber(reader.int64() as Long);
          break;
        case 5:
          message.current_epoch_start_time = fromTimestamp(
            Timestamp.decode(reader, reader.uint32())
          );
          break;
        case 6:
          message.epoch_counting_started = reader.bool();
          break;
        case 7:
          message.current_epoch_start_height = longToNumber(
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

  fromJSON(object: any): EpochInfo {
    const message = { ...baseEpochInfo } as EpochInfo;
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = String(object.identifier);
    } else {
      message.identifier = "";
    }
    if (object.start_time !== undefined && object.start_time !== null) {
      message.start_time = fromJsonTimestamp(object.start_time);
    } else {
      message.start_time = undefined;
    }
    if (object.duration !== undefined && object.duration !== null) {
      message.duration = Duration.fromJSON(object.duration);
    } else {
      message.duration = undefined;
    }
    if (object.current_epoch !== undefined && object.current_epoch !== null) {
      message.current_epoch = Number(object.current_epoch);
    } else {
      message.current_epoch = 0;
    }
    if (
      object.current_epoch_start_time !== undefined &&
      object.current_epoch_start_time !== null
    ) {
      message.current_epoch_start_time = fromJsonTimestamp(
        object.current_epoch_start_time
      );
    } else {
      message.current_epoch_start_time = undefined;
    }
    if (
      object.epoch_counting_started !== undefined &&
      object.epoch_counting_started !== null
    ) {
      message.epoch_counting_started = Boolean(object.epoch_counting_started);
    } else {
      message.epoch_counting_started = false;
    }
    if (
      object.current_epoch_start_height !== undefined &&
      object.current_epoch_start_height !== null
    ) {
      message.current_epoch_start_height = Number(
        object.current_epoch_start_height
      );
    } else {
      message.current_epoch_start_height = 0;
    }
    return message;
  },

  toJSON(message: EpochInfo): unknown {
    const obj: any = {};
    message.identifier !== undefined && (obj.identifier = message.identifier);
    message.start_time !== undefined &&
      (obj.start_time =
        message.start_time !== undefined
          ? message.start_time.toISOString()
          : null);
    message.duration !== undefined &&
      (obj.duration = message.duration
        ? Duration.toJSON(message.duration)
        : undefined);
    message.current_epoch !== undefined &&
      (obj.current_epoch = message.current_epoch);
    message.current_epoch_start_time !== undefined &&
      (obj.current_epoch_start_time =
        message.current_epoch_start_time !== undefined
          ? message.current_epoch_start_time.toISOString()
          : null);
    message.epoch_counting_started !== undefined &&
      (obj.epoch_counting_started = message.epoch_counting_started);
    message.current_epoch_start_height !== undefined &&
      (obj.current_epoch_start_height = message.current_epoch_start_height);
    return obj;
  },

  fromPartial(object: DeepPartial<EpochInfo>): EpochInfo {
    const message = { ...baseEpochInfo } as EpochInfo;
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = object.identifier;
    } else {
      message.identifier = "";
    }
    if (object.start_time !== undefined && object.start_time !== null) {
      message.start_time = object.start_time;
    } else {
      message.start_time = undefined;
    }
    if (object.duration !== undefined && object.duration !== null) {
      message.duration = Duration.fromPartial(object.duration);
    } else {
      message.duration = undefined;
    }
    if (object.current_epoch !== undefined && object.current_epoch !== null) {
      message.current_epoch = object.current_epoch;
    } else {
      message.current_epoch = 0;
    }
    if (
      object.current_epoch_start_time !== undefined &&
      object.current_epoch_start_time !== null
    ) {
      message.current_epoch_start_time = object.current_epoch_start_time;
    } else {
      message.current_epoch_start_time = undefined;
    }
    if (
      object.epoch_counting_started !== undefined &&
      object.epoch_counting_started !== null
    ) {
      message.epoch_counting_started = object.epoch_counting_started;
    } else {
      message.epoch_counting_started = false;
    }
    if (
      object.current_epoch_start_height !== undefined &&
      object.current_epoch_start_height !== null
    ) {
      message.current_epoch_start_height = object.current_epoch_start_height;
    } else {
      message.current_epoch_start_height = 0;
    }
    return message;
  },
};

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    for (const v of message.epochs) {
      EpochInfo.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.epochs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.epochs.push(EpochInfo.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.epochs = [];
    if (object.epochs !== undefined && object.epochs !== null) {
      for (const e of object.epochs) {
        message.epochs.push(EpochInfo.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    if (message.epochs) {
      obj.epochs = message.epochs.map((e) =>
        e ? EpochInfo.toJSON(e) : undefined
      );
    } else {
      obj.epochs = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.epochs = [];
    if (object.epochs !== undefined && object.epochs !== null) {
      for (const e of object.epochs) {
        message.epochs.push(EpochInfo.fromPartial(e));
      }
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

function toTimestamp(date: Date): Timestamp {
  const seconds = date.getTime() / 1_000;
  const nanos = (date.getTime() % 1_000) * 1_000_000;
  return { seconds, nanos };
}

function fromTimestamp(t: Timestamp): Date {
  let millis = t.seconds * 1_000;
  millis += t.nanos / 1_000_000;
  return new Date(millis);
}

function fromJsonTimestamp(o: any): Date {
  if (o instanceof Date) {
    return o;
  } else if (typeof o === "string") {
    return new Date(o);
  } else {
    return fromTimestamp(Timestamp.fromJSON(o));
  }
}

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
