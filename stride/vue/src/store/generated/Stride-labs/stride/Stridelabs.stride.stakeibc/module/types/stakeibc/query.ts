/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../stakeibc/params";
import { Validator } from "../stakeibc/validator";
import { Delegation } from "../stakeibc/delegation";
import { MinValidatorRequirements } from "../stakeibc/min_validator_requirements";
import { ICAAccount } from "../stakeibc/ica_account";
import { HostZone } from "../stakeibc/host_zone";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";

export const protobufPackage = "Stridelabs.stride.stakeibc";

/** QueryInterchainAccountFromAddressRequest is the request type for the Query/InterchainAccountAddress RPC */
export interface QueryInterchainAccountFromAddressRequest {
  owner: string;
  connection_id: string;
}

/** QueryInterchainAccountFromAddressResponse the response type for the Query/InterchainAccountAddress RPC */
export interface QueryInterchainAccountFromAddressResponse {
  interchain_account_address: string;
}

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetValidatorRequest {}

export interface QueryGetValidatorResponse {
  Validator: Validator | undefined;
}

export interface QueryGetDelegationRequest {}

export interface QueryGetDelegationResponse {
  Delegation: Delegation | undefined;
}

export interface QueryGetMinValidatorRequirementsRequest {}

export interface QueryGetMinValidatorRequirementsResponse {
  MinValidatorRequirements: MinValidatorRequirements | undefined;
}

export interface QueryGetICAAccountRequest {}

export interface QueryGetICAAccountResponse {
  ICAAccount: ICAAccount | undefined;
}

export interface QueryGetHostZoneRequest {
  id: number;
}

export interface QueryGetHostZoneResponse {
  HostZone: HostZone | undefined;
}

export interface QueryAllHostZoneRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllHostZoneResponse {
  HostZone: HostZone[];
  pagination: PageResponse | undefined;
}

const baseQueryInterchainAccountFromAddressRequest: object = {
  owner: "",
  connection_id: "",
};

export const QueryInterchainAccountFromAddressRequest = {
  encode(
    message: QueryInterchainAccountFromAddressRequest,
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

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryInterchainAccountFromAddressRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryInterchainAccountFromAddressRequest,
    } as QueryInterchainAccountFromAddressRequest;
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

  fromJSON(object: any): QueryInterchainAccountFromAddressRequest {
    const message = {
      ...baseQueryInterchainAccountFromAddressRequest,
    } as QueryInterchainAccountFromAddressRequest;
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

  toJSON(message: QueryInterchainAccountFromAddressRequest): unknown {
    const obj: any = {};
    message.owner !== undefined && (obj.owner = message.owner);
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryInterchainAccountFromAddressRequest>
  ): QueryInterchainAccountFromAddressRequest {
    const message = {
      ...baseQueryInterchainAccountFromAddressRequest,
    } as QueryInterchainAccountFromAddressRequest;
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

const baseQueryInterchainAccountFromAddressResponse: object = {
  interchain_account_address: "",
};

export const QueryInterchainAccountFromAddressResponse = {
  encode(
    message: QueryInterchainAccountFromAddressResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.interchain_account_address !== "") {
      writer.uint32(10).string(message.interchain_account_address);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryInterchainAccountFromAddressResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryInterchainAccountFromAddressResponse,
    } as QueryInterchainAccountFromAddressResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.interchain_account_address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryInterchainAccountFromAddressResponse {
    const message = {
      ...baseQueryInterchainAccountFromAddressResponse,
    } as QueryInterchainAccountFromAddressResponse;
    if (
      object.interchain_account_address !== undefined &&
      object.interchain_account_address !== null
    ) {
      message.interchain_account_address = String(
        object.interchain_account_address
      );
    } else {
      message.interchain_account_address = "";
    }
    return message;
  },

  toJSON(message: QueryInterchainAccountFromAddressResponse): unknown {
    const obj: any = {};
    message.interchain_account_address !== undefined &&
      (obj.interchain_account_address = message.interchain_account_address);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryInterchainAccountFromAddressResponse>
  ): QueryInterchainAccountFromAddressResponse {
    const message = {
      ...baseQueryInterchainAccountFromAddressResponse,
    } as QueryInterchainAccountFromAddressResponse;
    if (
      object.interchain_account_address !== undefined &&
      object.interchain_account_address !== null
    ) {
      message.interchain_account_address = object.interchain_account_address;
    } else {
      message.interchain_account_address = "";
    }
    return message;
  },
};

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
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

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryGetValidatorRequest: object = {};

export const QueryGetValidatorRequest = {
  encode(
    _: QueryGetValidatorRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValidatorRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
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

  fromJSON(_: any): QueryGetValidatorRequest {
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
    return message;
  },

  toJSON(_: QueryGetValidatorRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetValidatorRequest>
  ): QueryGetValidatorRequest {
    const message = {
      ...baseQueryGetValidatorRequest,
    } as QueryGetValidatorRequest;
    return message;
  },
};

const baseQueryGetValidatorResponse: object = {};

export const QueryGetValidatorResponse = {
  encode(
    message: QueryGetValidatorResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.Validator !== undefined) {
      Validator.encode(message.Validator, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValidatorResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Validator = Validator.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetValidatorResponse {
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    if (object.Validator !== undefined && object.Validator !== null) {
      message.Validator = Validator.fromJSON(object.Validator);
    } else {
      message.Validator = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetValidatorResponse): unknown {
    const obj: any = {};
    message.Validator !== undefined &&
      (obj.Validator = message.Validator
        ? Validator.toJSON(message.Validator)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetValidatorResponse>
  ): QueryGetValidatorResponse {
    const message = {
      ...baseQueryGetValidatorResponse,
    } as QueryGetValidatorResponse;
    if (object.Validator !== undefined && object.Validator !== null) {
      message.Validator = Validator.fromPartial(object.Validator);
    } else {
      message.Validator = undefined;
    }
    return message;
  },
};

const baseQueryGetDelegationRequest: object = {};

export const QueryGetDelegationRequest = {
  encode(
    _: QueryGetDelegationRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDelegationRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
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

  fromJSON(_: any): QueryGetDelegationRequest {
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
    return message;
  },

  toJSON(_: QueryGetDelegationRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetDelegationRequest>
  ): QueryGetDelegationRequest {
    const message = {
      ...baseQueryGetDelegationRequest,
    } as QueryGetDelegationRequest;
    return message;
  },
};

const baseQueryGetDelegationResponse: object = {};

export const QueryGetDelegationResponse = {
  encode(
    message: QueryGetDelegationResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.Delegation !== undefined) {
      Delegation.encode(message.Delegation, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDelegationResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Delegation = Delegation.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDelegationResponse {
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    if (object.Delegation !== undefined && object.Delegation !== null) {
      message.Delegation = Delegation.fromJSON(object.Delegation);
    } else {
      message.Delegation = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetDelegationResponse): unknown {
    const obj: any = {};
    message.Delegation !== undefined &&
      (obj.Delegation = message.Delegation
        ? Delegation.toJSON(message.Delegation)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDelegationResponse>
  ): QueryGetDelegationResponse {
    const message = {
      ...baseQueryGetDelegationResponse,
    } as QueryGetDelegationResponse;
    if (object.Delegation !== undefined && object.Delegation !== null) {
      message.Delegation = Delegation.fromPartial(object.Delegation);
    } else {
      message.Delegation = undefined;
    }
    return message;
  },
};

const baseQueryGetMinValidatorRequirementsRequest: object = {};

export const QueryGetMinValidatorRequirementsRequest = {
  encode(
    _: QueryGetMinValidatorRequirementsRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetMinValidatorRequirementsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
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

  fromJSON(_: any): QueryGetMinValidatorRequirementsRequest {
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
    return message;
  },

  toJSON(_: QueryGetMinValidatorRequirementsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetMinValidatorRequirementsRequest>
  ): QueryGetMinValidatorRequirementsRequest {
    const message = {
      ...baseQueryGetMinValidatorRequirementsRequest,
    } as QueryGetMinValidatorRequirementsRequest;
    return message;
  },
};

const baseQueryGetMinValidatorRequirementsResponse: object = {};

export const QueryGetMinValidatorRequirementsResponse = {
  encode(
    message: QueryGetMinValidatorRequirementsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.MinValidatorRequirements !== undefined) {
      MinValidatorRequirements.encode(
        message.MinValidatorRequirements,
        writer.uint32(10).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetMinValidatorRequirementsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.MinValidatorRequirements = MinValidatorRequirements.decode(
            reader,
            reader.uint32()
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetMinValidatorRequirementsResponse {
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    if (
      object.MinValidatorRequirements !== undefined &&
      object.MinValidatorRequirements !== null
    ) {
      message.MinValidatorRequirements = MinValidatorRequirements.fromJSON(
        object.MinValidatorRequirements
      );
    } else {
      message.MinValidatorRequirements = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetMinValidatorRequirementsResponse): unknown {
    const obj: any = {};
    message.MinValidatorRequirements !== undefined &&
      (obj.MinValidatorRequirements = message.MinValidatorRequirements
        ? MinValidatorRequirements.toJSON(message.MinValidatorRequirements)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetMinValidatorRequirementsResponse>
  ): QueryGetMinValidatorRequirementsResponse {
    const message = {
      ...baseQueryGetMinValidatorRequirementsResponse,
    } as QueryGetMinValidatorRequirementsResponse;
    if (
      object.MinValidatorRequirements !== undefined &&
      object.MinValidatorRequirements !== null
    ) {
      message.MinValidatorRequirements = MinValidatorRequirements.fromPartial(
        object.MinValidatorRequirements
      );
    } else {
      message.MinValidatorRequirements = undefined;
    }
    return message;
  },
};

const baseQueryGetICAAccountRequest: object = {};

export const QueryGetICAAccountRequest = {
  encode(
    _: QueryGetICAAccountRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetICAAccountRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetICAAccountRequest,
    } as QueryGetICAAccountRequest;
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

  fromJSON(_: any): QueryGetICAAccountRequest {
    const message = {
      ...baseQueryGetICAAccountRequest,
    } as QueryGetICAAccountRequest;
    return message;
  },

  toJSON(_: QueryGetICAAccountRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryGetICAAccountRequest>
  ): QueryGetICAAccountRequest {
    const message = {
      ...baseQueryGetICAAccountRequest,
    } as QueryGetICAAccountRequest;
    return message;
  },
};

const baseQueryGetICAAccountResponse: object = {};

export const QueryGetICAAccountResponse = {
  encode(
    message: QueryGetICAAccountResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.ICAAccount !== undefined) {
      ICAAccount.encode(message.ICAAccount, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetICAAccountResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetICAAccountResponse,
    } as QueryGetICAAccountResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ICAAccount = ICAAccount.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetICAAccountResponse {
    const message = {
      ...baseQueryGetICAAccountResponse,
    } as QueryGetICAAccountResponse;
    if (object.ICAAccount !== undefined && object.ICAAccount !== null) {
      message.ICAAccount = ICAAccount.fromJSON(object.ICAAccount);
    } else {
      message.ICAAccount = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetICAAccountResponse): unknown {
    const obj: any = {};
    message.ICAAccount !== undefined &&
      (obj.ICAAccount = message.ICAAccount
        ? ICAAccount.toJSON(message.ICAAccount)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetICAAccountResponse>
  ): QueryGetICAAccountResponse {
    const message = {
      ...baseQueryGetICAAccountResponse,
    } as QueryGetICAAccountResponse;
    if (object.ICAAccount !== undefined && object.ICAAccount !== null) {
      message.ICAAccount = ICAAccount.fromPartial(object.ICAAccount);
    } else {
      message.ICAAccount = undefined;
    }
    return message;
  },
};

const baseQueryGetHostZoneRequest: object = { id: 0 };

export const QueryGetHostZoneRequest = {
  encode(
    message: QueryGetHostZoneRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetHostZoneRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetHostZoneRequest,
    } as QueryGetHostZoneRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetHostZoneRequest {
    const message = {
      ...baseQueryGetHostZoneRequest,
    } as QueryGetHostZoneRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: QueryGetHostZoneRequest): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetHostZoneRequest>
  ): QueryGetHostZoneRequest {
    const message = {
      ...baseQueryGetHostZoneRequest,
    } as QueryGetHostZoneRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseQueryGetHostZoneResponse: object = {};

export const QueryGetHostZoneResponse = {
  encode(
    message: QueryGetHostZoneResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.HostZone !== undefined) {
      HostZone.encode(message.HostZone, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetHostZoneResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetHostZoneResponse,
    } as QueryGetHostZoneResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.HostZone = HostZone.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetHostZoneResponse {
    const message = {
      ...baseQueryGetHostZoneResponse,
    } as QueryGetHostZoneResponse;
    if (object.HostZone !== undefined && object.HostZone !== null) {
      message.HostZone = HostZone.fromJSON(object.HostZone);
    } else {
      message.HostZone = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetHostZoneResponse): unknown {
    const obj: any = {};
    message.HostZone !== undefined &&
      (obj.HostZone = message.HostZone
        ? HostZone.toJSON(message.HostZone)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetHostZoneResponse>
  ): QueryGetHostZoneResponse {
    const message = {
      ...baseQueryGetHostZoneResponse,
    } as QueryGetHostZoneResponse;
    if (object.HostZone !== undefined && object.HostZone !== null) {
      message.HostZone = HostZone.fromPartial(object.HostZone);
    } else {
      message.HostZone = undefined;
    }
    return message;
  },
};

const baseQueryAllHostZoneRequest: object = {};

export const QueryAllHostZoneRequest = {
  encode(
    message: QueryAllHostZoneRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllHostZoneRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllHostZoneRequest,
    } as QueryAllHostZoneRequest;
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

  fromJSON(object: any): QueryAllHostZoneRequest {
    const message = {
      ...baseQueryAllHostZoneRequest,
    } as QueryAllHostZoneRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllHostZoneRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllHostZoneRequest>
  ): QueryAllHostZoneRequest {
    const message = {
      ...baseQueryAllHostZoneRequest,
    } as QueryAllHostZoneRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllHostZoneResponse: object = {};

export const QueryAllHostZoneResponse = {
  encode(
    message: QueryAllHostZoneResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.HostZone) {
      HostZone.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllHostZoneResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllHostZoneResponse,
    } as QueryAllHostZoneResponse;
    message.HostZone = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.HostZone.push(HostZone.decode(reader, reader.uint32()));
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

  fromJSON(object: any): QueryAllHostZoneResponse {
    const message = {
      ...baseQueryAllHostZoneResponse,
    } as QueryAllHostZoneResponse;
    message.HostZone = [];
    if (object.HostZone !== undefined && object.HostZone !== null) {
      for (const e of object.HostZone) {
        message.HostZone.push(HostZone.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllHostZoneResponse): unknown {
    const obj: any = {};
    if (message.HostZone) {
      obj.HostZone = message.HostZone.map((e) =>
        e ? HostZone.toJSON(e) : undefined
      );
    } else {
      obj.HostZone = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllHostZoneResponse>
  ): QueryAllHostZoneResponse {
    const message = {
      ...baseQueryAllHostZoneResponse,
    } as QueryAllHostZoneResponse;
    message.HostZone = [];
    if (object.HostZone !== undefined && object.HostZone !== null) {
      for (const e of object.HostZone) {
        message.HostZone.push(HostZone.fromPartial(e));
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

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a Validator by index. */
  Validator(
    request: QueryGetValidatorRequest
  ): Promise<QueryGetValidatorResponse>;
  /** Queries a Delegation by index. */
  Delegation(
    request: QueryGetDelegationRequest
  ): Promise<QueryGetDelegationResponse>;
  /** Queries a MinValidatorRequirements by index. */
  MinValidatorRequirements(
    request: QueryGetMinValidatorRequirementsRequest
  ): Promise<QueryGetMinValidatorRequirementsResponse>;
  /** Queries a ICAAccount by index. */
  ICAAccount(
    request: QueryGetICAAccountRequest
  ): Promise<QueryGetICAAccountResponse>;
  /** Queries a HostZone by id. */
  HostZone(request: QueryGetHostZoneRequest): Promise<QueryGetHostZoneResponse>;
  /** Queries a list of HostZone items. */
  HostZoneAll(
    request: QueryAllHostZoneRequest
  ): Promise<QueryAllHostZoneResponse>;
  /** QueryInterchainAccountFromAddress returns the interchain account for given owner address on a given connection pair */
  InterchainAccountFromAddress(
    request: QueryInterchainAccountFromAddressRequest
  ): Promise<QueryInterchainAccountFromAddressResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  Validator(
    request: QueryGetValidatorRequest
  ): Promise<QueryGetValidatorResponse> {
    const data = QueryGetValidatorRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Validator",
      data
    );
    return promise.then((data) =>
      QueryGetValidatorResponse.decode(new Reader(data))
    );
  }

  Delegation(
    request: QueryGetDelegationRequest
  ): Promise<QueryGetDelegationResponse> {
    const data = QueryGetDelegationRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "Delegation",
      data
    );
    return promise.then((data) =>
      QueryGetDelegationResponse.decode(new Reader(data))
    );
  }

  MinValidatorRequirements(
    request: QueryGetMinValidatorRequirementsRequest
  ): Promise<QueryGetMinValidatorRequirementsResponse> {
    const data = QueryGetMinValidatorRequirementsRequest.encode(
      request
    ).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "MinValidatorRequirements",
      data
    );
    return promise.then((data) =>
      QueryGetMinValidatorRequirementsResponse.decode(new Reader(data))
    );
  }

  ICAAccount(
    request: QueryGetICAAccountRequest
  ): Promise<QueryGetICAAccountResponse> {
    const data = QueryGetICAAccountRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "ICAAccount",
      data
    );
    return promise.then((data) =>
      QueryGetICAAccountResponse.decode(new Reader(data))
    );
  }

  HostZone(
    request: QueryGetHostZoneRequest
  ): Promise<QueryGetHostZoneResponse> {
    const data = QueryGetHostZoneRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "HostZone",
      data
    );
    return promise.then((data) =>
      QueryGetHostZoneResponse.decode(new Reader(data))
    );
  }

  HostZoneAll(
    request: QueryAllHostZoneRequest
  ): Promise<QueryAllHostZoneResponse> {
    const data = QueryAllHostZoneRequest.encode(request).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "HostZoneAll",
      data
    );
    return promise.then((data) =>
      QueryAllHostZoneResponse.decode(new Reader(data))
    );
  }

  InterchainAccountFromAddress(
    request: QueryInterchainAccountFromAddressRequest
  ): Promise<QueryInterchainAccountFromAddressResponse> {
    const data = QueryInterchainAccountFromAddressRequest.encode(
      request
    ).finish();
    const promise = this.rpc.request(
      "Stridelabs.stride.stakeibc.Query",
      "InterchainAccountFromAddress",
      data
    );
    return promise.then((data) =>
      QueryInterchainAccountFromAddressResponse.decode(new Reader(data))
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
