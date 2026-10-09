"""Address helpers: the derivations the chain uses for module accounts and transfer escrows."""

import unittest

from sweep import addresses, config


class AddressTests(unittest.TestCase):
    def test_module_address_matches_the_sdk(self) -> None:
        # authtypes.NewModuleAddress("distribution"), as printed by `strided q auth module-account distribution`
        self.assertEqual(addresses.module_address(name="distribution"), "stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan")

    def test_escrow_address_is_twenty_bytes_under_the_stride_prefix(self) -> None:
        escrow = addresses.escrow_address(channel_id="channel-5")
        self.assertTrue(escrow.startswith("stride1"))
        self.assertEqual(len(addresses.address_bytes(address=escrow)), config.ADDRESS_LENGTH_BYTES)

    def test_ibc_denom_hashes_the_trace_path(self) -> None:
        self.assertEqual(
            addresses.ibc_denom(path="transfer/channel-0/uatom"),
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
        )

    def test_is_stride_address_rejects_other_prefixes_and_bad_checksums(self) -> None:
        self.assertTrue(addresses.is_stride_address(address=config.SWEEP_OPERATOR))
        self.assertFalse(addresses.is_stride_address(address="osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"))
        self.assertFalse(addresses.is_stride_address(address="stride1yz3mp7c21q3qrq6krk5f8ed5qhjyhgyw2k6vgm"))
        self.assertEqual(addresses.address_bytes(address="not-bech32"), b"")


if __name__ == "__main__":
    unittest.main()
