/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { Any } from "../google/protobuf/any";

export const protobufPackage = "Stridelabs.stride.stakeibc";

export interface MsgLiquidStake {
  creator: string;
  amount: number;
  denom: string;
}

export interface MsgLiquidStakeResponse {}

/**
 * TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
 * MsgRegisterAccount defines the payload for Msg/RegisterAccount
 */
export interface MsgRegisterAccount {
  owner: string;
  connection_id: string;
}

/**
 * TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
 * MsgRegisterAccountResponse defines the response for Msg/RegisterAccount
 */
export interface MsgRegisterAccountResponse {}

/**
 * TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
 * MsgSubmitTx defines the payload for Msg/SubmitTx
 */
export interface MsgSubmitTx {
  owner: string;
  connection_id: string;
  msg: Any | undefined;
}

/**
 * TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
 * MsgSubmitTxResponse defines the response for Msg/SubmitTx
 */
export interface MsgSubmitTxResponse {}

export interface MsgRegisterHostZone {
  connection_id: string;
  base_denom: string;
  local_denom: string;
  creator: string;
}

/** TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs) */
export interface MsgRegisterHostZoneResponse {}

const baseMsgLiquidStake: object = { creator: "", amount: 0, denom: "" };

export const MsgLiquidStake = {
  encode(message: MsgLiquidStake, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.amount !== 0) {
      writer.uint32(16).int32(message.amount);
    }
    if (message.denom !== "") {
      writer.uint32(26).string(message.denom);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgLiquidStake {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgLiquidStake } as MsgLiquidStake;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.amount = reader.int32();
          break;
        case 3:
          message.denom = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgLiquidStake {
    const message = { ...baseMsgLiquidStake } as MsgLiquidStake;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
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
    return message;
  },

  toJSON(message: MsgLiquidStake): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.amount !== undefined && (obj.amount = message.amount);
    message.denom !== undefined && (obj.denom = message.denom);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgLiquidStake>): MsgLiquidStake {
    const message = { ...baseMsgLiquidStake } as MsgLiquidStake;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
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
    return message;
  },
};

const baseMsgLiquidStakeResponse: object = {};

export const MsgLiquidStakeResponse = {
  encode(_: MsgLiquidStakeResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgLiquidStakeResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgLiquidStakeResponse } as MsgLiquidStakeResponse;
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

  fromJSON(_: any): MsgLiquidStakeResponse {
    const message = { ...baseMsgLiquidStakeResponse } as MsgLiquidStakeResponse;
    return message;
  },

  toJSON(_: MsgLiquidStakeResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgLiquidStakeResponse>): MsgLiquidStakeResponse {
    const message = { ...baseMsgLiquidStakeResponse } as MsgLiquidStakeResponse;
    return message;
  },
};

const baseMsgRegisterAccount: object = { owner: "", connection_id: "" };

export const MsgRegisterAccount = {
  encode(
    message: MsgRegisterAccount,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.owner !== "") {
      writer.uint32(10).string(message.owner);
    }
    if (message.connection_id !== "") {
      writer.uint32(18).string(message.connection_id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgRegisterAccount {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgRegisterAccount } as MsgRegisterAccount;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.owner = reader.string();
          break;
        case 2:
          message.connection_id = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgRegisterAccount {
    const message = { ...baseMsgRegisterAccount } as MsgRegisterAccount;
    if (object.owner !== undefined && object.owner !== null) {
      message.owner = String(object.owner);
    } else {
      message.owner = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = String(object.connection_id);
    } else {
      message.connection_id = "";
    }
    return message;
  },

  toJSON(message: MsgRegisterAccount): unknown {
    const obj: any = {};
    message.owner !== undefined && (obj.owner = message.owner);
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgRegisterAccount>): MsgRegisterAccount {
    const message = { ...baseMsgRegisterAccount } as MsgRegisterAccount;
    if (object.owner !== undefined && object.owner !== null) {
      message.owner = object.owner;
    } else {
      message.owner = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = object.connection_id;
    } else {
      message.connection_id = "";
    }
    return message;
  },
};

const baseMsgRegisterAccountResponse: object = {};

export const MsgRegisterAccountResponse = {
  encode(
    _: MsgRegisterAccountResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgRegisterAccountResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgRegisterAccountResponse,
    } as MsgRegisterAccountResponse;
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

  fromJSON(_: any): MsgRegisterAccountResponse {
    const message = {
      ...baseMsgRegisterAccountResponse,
    } as MsgRegisterAccountResponse;
    return message;
  },

  toJSON(_: MsgRegisterAccountResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgRegisterAccountResponse>
  ): MsgRegisterAccountResponse {
    const message = {
      ...baseMsgRegisterAccountResponse,
    } as MsgRegisterAccountResponse;
    return message;
  },
};

const baseMsgSubmitTx: object = { owner: "", connection_id: "" };

export const MsgSubmitTx = {
  encode(message: MsgSubmitTx, writer: Writer = Writer.create()): Writer {
    if (message.owner !== "") {
      writer.uint32(10).string(message.owner);
    }
    if (message.connection_id !== "") {
      writer.uint32(18).string(message.connection_id);
    }
    if (message.msg !== undefined) {
      Any.encode(message.msg, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitTx {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSubmitTx } as MsgSubmitTx;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.owner = reader.string();
          break;
        case 2:
          message.connection_id = reader.string();
          break;
        case 3:
          message.msg = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSubmitTx {
    const message = { ...baseMsgSubmitTx } as MsgSubmitTx;
    if (object.owner !== undefined && object.owner !== null) {
      message.owner = String(object.owner);
    } else {
      message.owner = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = String(object.connection_id);
    } else {
      message.connection_id = "";
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromJSON(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },

  toJSON(message: MsgSubmitTx): unknown {
    const obj: any = {};
    message.owner !== undefined && (obj.owner = message.owner);
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    message.msg !== undefined &&
      (obj.msg = message.msg ? Any.toJSON(message.msg) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgSubmitTx>): MsgSubmitTx {
    const message = { ...baseMsgSubmitTx } as MsgSubmitTx;
    if (object.owner !== undefined && object.owner !== null) {
      message.owner = object.owner;
    } else {
      message.owner = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = object.connection_id;
    } else {
      message.connection_id = "";
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromPartial(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },
};

const baseMsgSubmitTxResponse: object = {};

export const MsgSubmitTxResponse = {
  encode(_: MsgSubmitTxResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitTxResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSubmitTxResponse } as MsgSubmitTxResponse;
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

  fromJSON(_: any): MsgSubmitTxResponse {
    const message = { ...baseMsgSubmitTxResponse } as MsgSubmitTxResponse;
    return message;
  },

  toJSON(_: MsgSubmitTxResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgSubmitTxResponse>): MsgSubmitTxResponse {
    const message = { ...baseMsgSubmitTxResponse } as MsgSubmitTxResponse;
    return message;
  },
};

const baseMsgRegisterHostZone: object = {
  connection_id: "",
  base_denom: "",
  local_denom: "",
  creator: "",
};

export const MsgRegisterHostZone = {
  encode(
    message: MsgRegisterHostZone,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.connection_id !== "") {
      writer.uint32(18).string(message.connection_id);
    }
    if (message.base_denom !== "") {
      writer.uint32(34).string(message.base_denom);
    }
    if (message.local_denom !== "") {
      writer.uint32(42).string(message.local_denom);
    }
    if (message.creator !== "") {
      writer.uint32(50).string(message.creator);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgRegisterHostZone {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgRegisterHostZone } as MsgRegisterHostZone;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          message.connection_id = reader.string();
          break;
        case 4:
          message.base_denom = reader.string();
          break;
        case 5:
          message.local_denom = reader.string();
          break;
        case 6:
          message.creator = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgRegisterHostZone {
    const message = { ...baseMsgRegisterHostZone } as MsgRegisterHostZone;
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = String(object.connection_id);
    } else {
      message.connection_id = "";
    }
    if (object.base_denom !== undefined && object.base_denom !== null) {
      message.base_denom = String(object.base_denom);
    } else {
      message.base_denom = "";
    }
    if (object.local_denom !== undefined && object.local_denom !== null) {
      message.local_denom = String(object.local_denom);
    } else {
      message.local_denom = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    return message;
  },

  toJSON(message: MsgRegisterHostZone): unknown {
    const obj: any = {};
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    message.base_denom !== undefined && (obj.base_denom = message.base_denom);
    message.local_denom !== undefined &&
      (obj.local_denom = message.local_denom);
    message.creator !== undefined && (obj.creator = message.creator);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgRegisterHostZone>): MsgRegisterHostZone {
    const message = { ...baseMsgRegisterHostZone } as MsgRegisterHostZone;
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = object.connection_id;
    } else {
      message.connection_id = "";
    }
    if (object.base_denom !== undefined && object.base_denom !== null) {
      message.base_denom = object.base_denom;
    } else {
      message.base_denom = "";
    }
    if (object.local_denom !== undefined && object.local_denom !== null) {
      message.local_denom = object.local_denom;
    } else {
      message.local_denom = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    return message;
  },
};

const baseMsgRegisterHostZoneResponse: object = {};

export const MsgRegisterHostZoneResponse = {
  encode(
    _: MsgRegisterHostZoneResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgRegisterHostZoneResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgRegisterHostZoneResponse,
    } as MsgRegisterHostZoneResponse;
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

  fromJSON(_: any): MsgRegisterHostZoneResponse {
    const message = {
      ...baseMsgRegisterHostZoneResponse,
    } as MsgRegisterHostZoneResponse;
    return message;
  },

  toJSON(_: MsgRegisterHostZoneResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgRegisterHostZoneResponse>
  ): MsgRegisterHostZoneResponse {
    const message = {
      ...baseMsgRegisterHostZoneResponse,
    } as MsgRegisterHostZoneResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  LiquidStake(request: MsgLiquidStake): Promise<MsgLiquidStakeResponse>;
  /**
   * TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
   * Register defines a rpc handler for MsgRegisterAccount
   */
  RegisterAccount(
    request: MsgRegisterAccount
  ): Promise<MsgRegisterAccountResponse>;
  /** TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs) */
  SubmitTx(request: MsgSubmitTx): Promise<MsgSubmitTxResponse>;
  /** TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs) */
  RegisterHostZone(
    request: MsgRegisterHostZone
  ): Promise<MsgRegisterHostZoneResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  LiquidStake(request: MsgLiquidStake): Promise<MsgLiquidStakeResponse> {
    const data = MsgLiquidStake.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Msg",
      "LiquidStake",
      data
    );
    return promise.then((data) =>
      MsgLiquidStakeResponse.decode(new Reader(data))
    );
  }

  RegisterAccount(
    request: MsgRegisterAccount
  ): Promise<MsgRegisterAccountResponse> {
    const data = MsgRegisterAccount.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Msg",
      "RegisterAccount",
      data
    );
    return promise.then((data) =>
      MsgRegisterAccountResponse.decode(new Reader(data))
    );
  }

  SubmitTx(request: MsgSubmitTx): Promise<MsgSubmitTxResponse> {
    const data = MsgSubmitTx.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Msg",
      "SubmitTx",
      data
    );
    return promise.then((data) => MsgSubmitTxResponse.decode(new Reader(data)));
  }

  RegisterHostZone(
    request: MsgRegisterHostZone
  ): Promise<MsgRegisterHostZoneResponse> {
    const data = MsgRegisterHostZone.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Msg",
      "RegisterHostZone",
      data
    );
    return promise.then((data) =>
      MsgRegisterHostZoneResponse.decode(new Reader(data))
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
