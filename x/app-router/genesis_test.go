package app_router_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v6/app/apptesting"
	router "github.com/Stride-Labs/stride/v6/x/app-router"
	"github.com/Stride-Labs/stride/v6/x/app-router/types"
)

func TestGenesis(t *testing.T) {
	expectedGenesisState := types.GenesisState{
		Params: types.Params{Active: true},
	}

	s := apptesting.SetupSuitelessTestHelper()
	router.InitGenesis(s.Ctx, s.App.RouterKeeper, expectedGenesisState)

	actualGenesisState := router.ExportGenesis(s.Ctx, s.App.RouterKeeper)
	require.NotNil(t, actualGenesisState)
	require.Equal(t, expectedGenesisState.Params, actualGenesisState.Params)
}
