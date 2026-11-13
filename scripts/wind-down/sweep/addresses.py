"""Address helpers: bech32 validation and the derivations the chain uses (module accounts, ICS-20 escrows, ibc/ denoms).
The vendored bech32 reference implementation lives one directory up."""

import hashlib
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # for bech32_ref, shared with coverage_check.py

import bech32_ref  # noqa: E402

from sweep import config  # noqa: E402

ESCROW_ADDRESS_VERSION = "ics20-1"


def address_bytes(address: str) -> bytes:
    """The raw bytes behind a bech32 address, or b"" when the string is not valid bech32."""
    _, data = bech32_ref.bech32_decode(address)
    if data is None:
        return b""
    converted = bech32_ref.convertbits(data, 5, 8, False)
    return bytes(converted) if converted is not None else b""


def is_stride_address(address: str) -> bool:
    """A valid `stride1…` address of exactly 20 bytes (the chain's first skip rule)."""
    hrp, _ = bech32_ref.bech32_decode(address)
    return hrp == config.BECH32_PREFIX and len(address_bytes(address=address)) == config.ADDRESS_LENGTH_BYTES


def module_address(name: str) -> str:
    """SDK authtypes.NewModuleAddress: sha256(name)[:20], bech32 stride."""
    return bech32_ref.encode(config.BECH32_PREFIX, hashlib.sha256(name.encode()).digest()[: config.ADDRESS_LENGTH_BYTES])


def escrow_address(channel_id: str) -> str:
    """ibc-go transfertypes.GetEscrowAddress: sha256("ics20-1\\0" + port/channel)[:20], bech32 stride."""
    pre_image = ESCROW_ADDRESS_VERSION.encode() + b"\x00" + f"{config.TRANSFER_PORT}/{channel_id}".encode()
    return bech32_ref.encode(config.BECH32_PREFIX, hashlib.sha256(pre_image).digest()[: config.ADDRESS_LENGTH_BYTES])


def ibc_denom(path: str) -> str:
    """The on-chain denom of an IBC voucher from its full trace path, e.g. `transfer/channel-0/uatom`."""
    return config.IBC_PREFIX + hashlib.sha256(path.encode()).hexdigest().upper()
