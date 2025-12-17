package utils_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v31/utils"
)

func (s *UtilsTestSuite) TestCreateModuleAccount() {
	baseWithAddr := func(addr sdk.AccAddress) sdk.AccountI {
		acc := authtypes.ProtoBaseAccount()
		err := acc.SetAddress(addr)
		s.Require().NoError(err)
		return acc
	}
	userAccViaSeqnum := func(addr sdk.AccAddress) sdk.AccountI {
		base := baseWithAddr(addr)
		err := base.SetSequence(2)
		s.Require().NoError(err)
		return base
	}
	userAccViaPubkey := func(addr sdk.AccAddress) sdk.AccountI {
		base := baseWithAddr(addr)
		err := base.SetPubKey(secp256k1.GenPrivKey().PubKey())
		s.Require().NoError(err)
		return base
	}
	defaultModuleAccAddr := address.Module("dummy module", []byte{1})
	testcases := map[string]struct {
		priorAccounts []sdk.AccountI
		moduleAccAddr sdk.AccAddress
		expErr        bool
	}{
		"no prior acc": {
			priorAccounts: []sdk.AccountI{},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        false,
		},
		"prior empty acc at addr": {
			priorAccounts: []sdk.AccountI{baseWithAddr(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        false,
		},
		"prior user acc at addr (sequence)": {
			priorAccounts: []sdk.AccountI{userAccViaSeqnum(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        true,
		},
		"prior user acc at addr (pubkey)": {
			priorAccounts: []sdk.AccountI{userAccViaPubkey(defaultModuleAccAddr)},
			moduleAccAddr: defaultModuleAccAddr,
			expErr:        true,
		},
	}
	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for _, priorAcc := range tc.priorAccounts {
				s.App.AccountKeeper.SetAccount(s.Ctx, priorAcc)
			}
			err := utils.CreateModuleAccount(s.Ctx, s.App.AccountKeeper, tc.moduleAccAddr)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
