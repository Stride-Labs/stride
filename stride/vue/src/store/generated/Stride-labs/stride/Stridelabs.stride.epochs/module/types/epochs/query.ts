/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";
import { EpochInfo } from "../epochs/genesis";

export const protobufPackage = "Stridelabs.stride.epochs";

export interface QueryEpochsInfoRequest {
  pagination: PageRequest | undefined;
}

export interface QueryEpochsInfoResponse {
  epochs: EpochInfo[];
  pagination: PageResponse | undefined;
}

export interface QueryCurrentEpochRequest {
  identifier: string;
}

export interface QueryCurrentEpochResponse {
  current_epoch: number;
}

const baseQueryEpochsInfoRequest: object = {};

export const QueryEpochsInfoRequest = {
  encode(
    message: QueryEpochsInfoRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryEpochsInfoRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryEpochsInfoRequest } as QueryEpochsInfoRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryEpochsInfoRequest {
    const message = { ...baseQueryEpochsInfoRequest } as QueryEpochsInfoRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryEpochsInfoRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryEpochsInfoRequest>
  ): QueryEpochsInfoRequest {
    const message = { ...baseQueryEpochsInfoRequest } as QueryEpochsInfoRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryEpochsInfoResponse: object = {};

export const QueryEpochsInfoResponse = {
  encode(
    message: QueryEpochsInfoResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.epochs) {
      EpochInfo.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryEpochsInfoResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryEpochsInfoResponse,
    } as QueryEpochsInfoResponse;
    message.epochs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.epochs.push(EpochInfo.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryEpochsInfoResponse {
    const message = {
      ...baseQueryEpochsInfoResponse,
    } as QueryEpochsInfoResponse;
    message.epochs = [];
    if (object.epochs !== undefined && object.epochs !== null) {
      for (const e of object.epochs) {
        message.epochs.push(EpochInfo.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryEpochsInfoResponse): unknown {
    const obj: any = {};
    if (message.epochs) {
      obj.epochs = message.epochs.map((e) =>
        e ? EpochInfo.toJSON(e) : undefined
      );
    } else {
      obj.epochs = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryEpochsInfoResponse>
  ): QueryEpochsInfoResponse {
    const message = {
      ...baseQueryEpochsInfoResponse,
    } as QueryEpochsInfoResponse;
    message.epochs = [];
    if (object.epochs !== undefined && object.epochs !== null) {
      for (const e of object.epochs) {
        message.epochs.push(EpochInfo.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryCurrentEpochRequest: object = { identifier: "" };

export const QueryCurrentEpochRequest = {
  encode(
    message: QueryCurrentEpochRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.identifier !== "") {
      writer.uint32(10).string(message.identifier);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryCurrentEpochRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryCurrentEpochRequest,
    } as QueryCurrentEpochRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.identifier = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryCurrentEpochRequest {
    const message = {
      ...baseQueryCurrentEpochRequest,
    } as QueryCurrentEpochRequest;
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = String(object.identifier);
    } else {
      message.identifier = "";
    }
    return message;
  },

  toJSON(message: QueryCurrentEpochRequest): unknown {
    const obj: any = {};
    message.identifier !== undefined && (obj.identifier = message.identifier);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryCurrentEpochRequest>
  ): QueryCurrentEpochRequest {
    const message = {
      ...baseQueryCurrentEpochRequest,
    } as QueryCurrentEpochRequest;
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = object.identifier;
    } else {
      message.identifier = "";
    }
    return message;
  },
};

const baseQueryCurrentEpochResponse: object = { current_epoch: 0 };

export const QueryCurrentEpochResponse = {
  encode(
    message: QueryCurrentEpochResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.current_epoch !== 0) {
      writer.uint32(8).int64(message.current_epoch);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryCurrentEpochResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryCurrentEpochResponse,
    } as QueryCurrentEpochResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.current_epoch = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryCurrentEpochResponse {
    const message = {
      ...baseQueryCurrentEpochResponse,
    } as QueryCurrentEpochResponse;
    if (object.current_epoch !== undefined && object.current_epoch !== null) {
      message.current_epoch = Number(object.current_epoch);
    } else {
      message.current_epoch = 0;
    }
    return message;
  },

  toJSON(message: QueryCurrentEpochResponse): unknown {
    const obj: any = {};
    message.current_epoch !== undefined &&
      (obj.current_epoch = message.current_epoch);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryCurrentEpochResponse>
  ): QueryCurrentEpochResponse {
    const message = {
      ...baseQueryCurrentEpochResponse,
    } as QueryCurrentEpochResponse;
    if (object.current_epoch !== undefined && object.current_epoch !== null) {
      message.current_epoch = object.current_epoch;
    } else {
      message.current_epoch = 0;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** EpochInfos provide running epochInfos */
  EpochInfos(request: QueryEpochsInfoRequest): Promise<QueryEpochsInfoResponse>;
  /** CurrentEpoch provide current epoch of specified identifier */
  CurrentEpoch(
    request: QueryCurrentEpochRequest
  ): Promise<QueryCurrentEpochResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  EpochInfos(
    request: QueryEpochsInfoRequest
  ): Promise<QueryEpochsInfoResponse> {
    const data = QueryEpochsInfoRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.epochs.Query",
      "EpochInfos",
      data
    );
    return promise.then((data) =>
      QueryEpochsInfoResponse.decode(new Reader(data))
    );
  }

  CurrentEpoch(
    request: QueryCurrentEpochRequest
  ): Promise<QueryCurrentEpochResponse> {
    const data = QueryCurrentEpochRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.epochs.Query",
      "CurrentEpoch",
      data
    );
    return promise.then((data) =>
      QueryCurrentEpochResponse.decode(new Reader(data))
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
