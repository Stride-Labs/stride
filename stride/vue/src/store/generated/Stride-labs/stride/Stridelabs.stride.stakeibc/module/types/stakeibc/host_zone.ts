/* eslint-disable */
import { Validator } from "../stakeibc/validator";
import { ICAAccount } from "../stakeibc/ica_account";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** next id: 10 */
export interface HostZone {
  chainId: string;
  connectionId: string;
  validators: Validator[];
  blacklistedValidators: Validator[];
  withdrawalAccount: ICAAccount | undefined;
  feeAccount: ICAAccount | undefined;
  delegationAccount: ICAAccount | undefined;
  LocalDenom: string;
  BaseDenom: string;
}

const baseHostZone: object = {
  chainId: "",
  connectionId: "",
  LocalDenom: "",
  BaseDenom: "",
};

export const HostZone = {
  encode(message: HostZone, writer: Writer = Writer.create()): Writer {
    if (message.chainId !== "") {
      writer.uint32(10).string(message.chainId);
    }
    if (message.connectionId !== "") {
      writer.uint32(18).string(message.connectionId);
    }
    for (const v of message.validators) {
      Validator.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.blacklistedValidators) {
      Validator.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    if (message.withdrawalAccount !== undefined) {
      ICAAccount.encode(
        message.withdrawalAccount,
        writer.uint32(42).fork()
      ).ldelim();
    }
    if (message.feeAccount !== undefined) {
      ICAAccount.encode(message.feeAccount, writer.uint32(50).fork()).ldelim();
    }
    if (message.delegationAccount !== undefined) {
      ICAAccount.encode(
        message.delegationAccount,
        writer.uint32(58).fork()
      ).ldelim();
    }
    if (message.LocalDenom !== "") {
      writer.uint32(66).string(message.LocalDenom);
    }
    if (message.BaseDenom !== "") {
      writer.uint32(74).string(message.BaseDenom);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): HostZone {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.blacklistedValidators = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chainId = reader.string();
          break;
        case 2:
          message.connectionId = reader.string();
          break;
        case 3:
          message.validators.push(Validator.decode(reader, reader.uint32()));
          break;
        case 4:
          message.blacklistedValidators.push(
            Validator.decode(reader, reader.uint32())
          );
          break;
        case 5:
          message.withdrawalAccount = ICAAccount.decode(
            reader,
            reader.uint32()
          );
          break;
        case 6:
          message.feeAccount = ICAAccount.decode(reader, reader.uint32());
          break;
        case 7:
          message.delegationAccount = ICAAccount.decode(
            reader,
            reader.uint32()
          );
          break;
        case 8:
          message.LocalDenom = reader.string();
          break;
        case 9:
          message.BaseDenom = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): HostZone {
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.blacklistedValidators = [];
    if (object.chainId !== undefined && object.chainId !== null) {
      message.chainId = String(object.chainId);
    } else {
      message.chainId = "";
    }
    if (object.connectionId !== undefined && object.connectionId !== null) {
      message.connectionId = String(object.connectionId);
    } else {
      message.connectionId = "";
    }
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromJSON(e));
      }
    }
    if (
      object.blacklistedValidators !== undefined &&
      object.blacklistedValidators !== null
    ) {
      for (const e of object.blacklistedValidators) {
        message.blacklistedValidators.push(Validator.fromJSON(e));
      }
    }
    if (
      object.withdrawalAccount !== undefined &&
      object.withdrawalAccount !== null
    ) {
      message.withdrawalAccount = ICAAccount.fromJSON(object.withdrawalAccount);
    } else {
      message.withdrawalAccount = undefined;
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      message.feeAccount = ICAAccount.fromJSON(object.feeAccount);
    } else {
      message.feeAccount = undefined;
    }
    if (
      object.delegationAccount !== undefined &&
      object.delegationAccount !== null
    ) {
      message.delegationAccount = ICAAccount.fromJSON(object.delegationAccount);
    } else {
      message.delegationAccount = undefined;
    }
    if (object.LocalDenom !== undefined && object.LocalDenom !== null) {
      message.LocalDenom = String(object.LocalDenom);
    } else {
      message.LocalDenom = "";
    }
    if (object.BaseDenom !== undefined && object.BaseDenom !== null) {
      message.BaseDenom = String(object.BaseDenom);
    } else {
      message.BaseDenom = "";
    }
    return message;
  },

  toJSON(message: HostZone): unknown {
    const obj: any = {};
    message.chainId !== undefined && (obj.chainId = message.chainId);
    message.connectionId !== undefined &&
      (obj.connectionId = message.connectionId);
    if (message.validators) {
      obj.validators = message.validators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.validators = [];
    }
    if (message.blacklistedValidators) {
      obj.blacklistedValidators = message.blacklistedValidators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.blacklistedValidators = [];
    }
    message.withdrawalAccount !== undefined &&
      (obj.withdrawalAccount = message.withdrawalAccount
        ? ICAAccount.toJSON(message.withdrawalAccount)
        : undefined);
    message.feeAccount !== undefined &&
      (obj.feeAccount = message.feeAccount
        ? ICAAccount.toJSON(message.feeAccount)
        : undefined);
    message.delegationAccount !== undefined &&
      (obj.delegationAccount = message.delegationAccount
        ? ICAAccount.toJSON(message.delegationAccount)
        : undefined);
    message.LocalDenom !== undefined && (obj.LocalDenom = message.LocalDenom);
    message.BaseDenom !== undefined && (obj.BaseDenom = message.BaseDenom);
    return obj;
  },

  fromPartial(object: DeepPartial<HostZone>): HostZone {
    const message = { ...baseHostZone } as HostZone;
    message.validators = [];
    message.blacklistedValidators = [];
    if (object.chainId !== undefined && object.chainId !== null) {
      message.chainId = object.chainId;
    } else {
      message.chainId = "";
    }
    if (object.connectionId !== undefined && object.connectionId !== null) {
      message.connectionId = object.connectionId;
    } else {
      message.connectionId = "";
    }
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromPartial(e));
      }
    }
    if (
      object.blacklistedValidators !== undefined &&
      object.blacklistedValidators !== null
    ) {
      for (const e of object.blacklistedValidators) {
        message.blacklistedValidators.push(Validator.fromPartial(e));
      }
    }
    if (
      object.withdrawalAccount !== undefined &&
      object.withdrawalAccount !== null
    ) {
      message.withdrawalAccount = ICAAccount.fromPartial(
        object.withdrawalAccount
      );
    } else {
      message.withdrawalAccount = undefined;
    }
    if (object.feeAccount !== undefined && object.feeAccount !== null) {
      message.feeAccount = ICAAccount.fromPartial(object.feeAccount);
    } else {
      message.feeAccount = undefined;
    }
    if (
      object.delegationAccount !== undefined &&
      object.delegationAccount !== null
    ) {
      message.delegationAccount = ICAAccount.fromPartial(
        object.delegationAccount
      );
    } else {
      message.delegationAccount = undefined;
    }
    if (object.LocalDenom !== undefined && object.LocalDenom !== null) {
      message.LocalDenom = object.LocalDenom;
    } else {
      message.LocalDenom = "";
    }
    if (object.BaseDenom !== undefined && object.BaseDenom !== null) {
      message.BaseDenom = object.BaseDenom;
    } else {
      message.BaseDenom = "";
    }
    return message;
  },
};

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
