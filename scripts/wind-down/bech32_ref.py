"""Bech32 reference implementation (BIP-173), vendored for the wind-down scripts.

Copyright (c) 2017 Pieter Wuille. Licensed under the MIT License
(https://github.com/sipa/bech32/blob/master/ref/python/segwit_addr.py). The functions named
bech32_* and convertbits are the reference code unchanged; encode/decode are thin helpers for
cosmos-style addresses, which carry raw address bytes and no witness version.
"""

CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def bech32_polymod(values: list[int]) -> int:
    """Internal function that computes the Bech32 checksum."""
    generator = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    chk = 1
    for value in values:
        top = chk >> 25
        chk = (chk & 0x1FFFFFF) << 5 ^ value
        for i in range(5):
            chk ^= generator[i] if ((top >> i) & 1) else 0
    return chk


def bech32_hrp_expand(hrp: str) -> list[int]:
    """Expand the HRP into values for checksum computation."""
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]


def bech32_verify_checksum(hrp: str, data: list[int]) -> bool:
    """Verify a checksum given HRP and converted data characters."""
    return bech32_polymod(bech32_hrp_expand(hrp) + data) == 1


def bech32_create_checksum(hrp: str, data: list[int]) -> list[int]:
    """Compute the checksum values given HRP and data."""
    values = bech32_hrp_expand(hrp) + data
    polymod = bech32_polymod(values + [0, 0, 0, 0, 0, 0]) ^ 1
    return [(polymod >> 5 * (5 - i)) & 31 for i in range(6)]


def bech32_encode(hrp: str, data: list[int]) -> str:
    """Compute a Bech32 string given HRP and data values."""
    combined = data + bech32_create_checksum(hrp, data)
    return hrp + "1" + "".join([CHARSET[d] for d in combined])


def bech32_decode(bech: str) -> tuple[str | None, list[int] | None]:
    """Validate a Bech32 string, and determine HRP and data."""
    if (any(ord(x) < 33 or ord(x) > 126 for x in bech)) or (bech.lower() != bech and bech.upper() != bech):
        return (None, None)
    bech = bech.lower()
    pos = bech.rfind("1")
    if pos < 1 or pos + 7 > len(bech) or len(bech) > 90:
        return (None, None)
    if not all(x in CHARSET for x in bech[pos + 1 :]):
        return (None, None)
    hrp = bech[:pos]
    data = [CHARSET.find(x) for x in bech[pos + 1 :]]
    if not bech32_verify_checksum(hrp, data):
        return (None, None)
    return (hrp, data[:-6])


def convertbits(data: list[int] | bytes, frombits: int, tobits: int, pad: bool = True) -> list[int] | None:
    """General power-of-2 base conversion."""
    acc = 0
    bits = 0
    ret = []
    maxv = (1 << tobits) - 1
    max_acc = (1 << (frombits + tobits - 1)) - 1
    for value in data:
        if value < 0 or (value >> frombits):
            return None
        acc = ((acc << frombits) | value) & max_acc
        bits += frombits
        while bits >= tobits:
            bits -= tobits
            ret.append((acc >> bits) & maxv)
    if pad:
        if bits:
            ret.append((acc << (tobits - bits)) & maxv)
    elif bits >= frombits or ((acc << (tobits - bits)) & maxv):
        return None
    return ret


def encode(hrp: str, address: bytes) -> str:
    """Bech32-encode raw cosmos address bytes under the given prefix."""
    data = convertbits(address, 8, 5)
    if data is None:
        raise ValueError(f"cannot convert {len(address)} address bytes to base32")
    return bech32_encode(hrp, data)


def decode(bech: str) -> tuple[str, bytes]:
    """Decode a cosmos bech32 address into (prefix, raw bytes); raises ValueError when invalid."""
    hrp, data = bech32_decode(bech)
    if hrp is None or data is None:
        raise ValueError(f"invalid bech32 string {bech!r}")
    converted = convertbits(data, 5, 8, False)
    if converted is None:
        raise ValueError(f"invalid bech32 payload in {bech!r}")
    return hrp, bytes(converted)
