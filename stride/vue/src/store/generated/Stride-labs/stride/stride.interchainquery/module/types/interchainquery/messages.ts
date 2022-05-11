/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "stride.interchainquery";

/** MsgSubmitQueryResponse represents a message type to fulfil a query request. */
export interface MsgSubmitQueryResponse {
  chain_id: string;
  query_id: string;
  result: Uint8Array;
  height: number;
  /**
   * TODO(TEST-15): Add this type annotation back after installing cosmos_proto
   * string from_address = 5 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
   */
  from_address: string;
}

/**
 * MsgSubmitQueryResponseResponse defines the MsgSubmitQueryResponse response
 * type.
 */
export interface MsgSubmitQueryResponseResponse {}

const baseMsgSubmitQueryResponse: object = {
  chain_id: "",
  query_id: "",
  height: 0,
  from_address: "",
};

export const MsgSubmitQueryResponse = {
  encode(
    message: MsgSubmitQueryResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.chain_id !== "") {
      writer.uint32(10).string(message.chain_id);
    }
    if (message.query_id !== "") {
      writer.uint32(18).string(message.query_id);
    }
    if (message.result.length !== 0) {
      writer.uint32(26).bytes(message.result);
    }
    if (message.height !== 0) {
      writer.uint32(32).int64(message.height);
    }
    if (message.from_address !== "") {
      writer.uint32(42).string(message.from_address);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitQueryResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSubmitQueryResponse } as MsgSubmitQueryResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chain_id = reader.string();
          break;
        case 2:
          message.query_id = reader.string();
          break;
        case 3:
          message.result = reader.bytes();
          break;
        case 4:
          message.height = longToNumber(reader.int64() as Long);
          break;
        case 5:
          message.from_address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSubmitQueryResponse {
    const message = { ...baseMsgSubmitQueryResponse } as MsgSubmitQueryResponse;
    if (object.chain_id !== undefined && object.chain_id !== null) {
      message.chain_id = String(object.chain_id);
    } else {
      message.chain_id = "";
    }
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = String(object.query_id);
    } else {
      message.query_id = "";
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = bytesFromBase64(object.result);
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    if (object.from_address !== undefined && object.from_address !== null) {
      message.from_address = String(object.from_address);
    } else {
      message.from_address = "";
    }
    return message;
  },

  toJSON(message: MsgSubmitQueryResponse): unknown {
    const obj: any = {};
    message.chain_id !== undefined && (obj.chain_id = message.chain_id);
    message.query_id !== undefined && (obj.query_id = message.query_id);
    message.result !== undefined &&
      (obj.result = base64FromBytes(
        message.result !== undefined ? message.result : new Uint8Array()
      ));
    message.height !== undefined && (obj.height = message.height);
    message.from_address !== undefined &&
      (obj.from_address = message.from_address);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgSubmitQueryResponse>
  ): MsgSubmitQueryResponse {
    const message = { ...baseMsgSubmitQueryResponse } as MsgSubmitQueryResponse;
    if (object.chain_id !== undefined && object.chain_id !== null) {
      message.chain_id = object.chain_id;
    } else {
      message.chain_id = "";
    }
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = object.query_id;
    } else {
      message.query_id = "";
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = object.result;
    } else {
      message.result = new Uint8Array();
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    if (object.from_address !== undefined && object.from_address !== null) {
      message.from_address = object.from_address;
    } else {
      message.from_address = "";
    }
    return message;
  },
};

const baseMsgSubmitQueryResponseResponse: object = {};

export const MsgSubmitQueryResponseResponse = {
  encode(
    _: MsgSubmitQueryResponseResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgSubmitQueryResponseResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgSubmitQueryResponseResponse,
    } as MsgSubmitQueryResponseResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgSubmitQueryResponseResponse {
    const message = {
      ...baseMsgSubmitQueryResponseResponse,
    } as MsgSubmitQueryResponseResponse;
    return message;
  },

  toJSON(_: MsgSubmitQueryResponseResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgSubmitQueryResponseResponse>
  ): MsgSubmitQueryResponseResponse {
    const message = {
      ...baseMsgSubmitQueryResponseResponse,
    } as MsgSubmitQueryResponseResponse;
    return message;
  },
};

/** Msg defines the interchainquery Msg service. */
export interface Msg {
  /** SubmitQueryResponse defines a method for submit query responses. */
  SubmitQueryResponse(
    request: MsgSubmitQueryResponse
  ): Promise<MsgSubmitQueryResponseResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  SubmitQueryResponse(
    request: MsgSubmitQueryResponse
  ): Promise<MsgSubmitQueryResponseResponse> {
    const data = MsgSubmitQueryResponse.encode(request).finish();
    const promise = this.rpc.request(
      "stride.interchainquery.Msg",
      "SubmitQueryResponse",
      data
    );
    return promise.then((data) =>
      MsgSubmitQueryResponseResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (let i = 0; i < arr.byteLength; ++i) {
    bin.push(String.fromCharCode(arr[i]));
  }
  return btoa(bin.join(""));
}

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
