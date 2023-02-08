package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

func TestGenesisState_Validate(t *testing.T) {
    for _, tc := range []struct {
    		desc          string
    		genState      *types.GenesisState
    		valid bool
    } {
        {
            desc:     "default is valid",
            genState: types.DefaultGenesis(),
            valid:    true,
        },
        {
            desc:     "valid genesis state",
            genState: &types.GenesisState{
            	
                AuctionPoolList: []types.AuctionPool{
	{
		Id: 0,
	},
	{
		Id: 1,
	},
},
AuctionPoolCount: 2,
// this line is used by starport scaffolding # types/genesis/validField
            },
            valid:    true,
        },
        {
	desc:     "duplicated auctionPool",
	genState: &types.GenesisState{
		AuctionPoolList: []types.AuctionPool{
			{
				Id: 0,
			},
			{
				Id: 0,
			},
		},
	},
	valid:    false,
},
{
	desc:     "invalid auctionPool count",
	genState: &types.GenesisState{
		AuctionPoolList: []types.AuctionPool{
			{
				Id: 1,
			},
		},
		AuctionPoolCount: 0,
	},
	valid:    false,
},
// this line is used by starport scaffolding # types/genesis/testcase
    } {
        t.Run(tc.desc, func(t *testing.T) {
            err := tc.genState.Validate()
            if tc.valid {
                require.NoError(t, err)
            } else {
                require.Error(t, err)
            }
        })
    }
}