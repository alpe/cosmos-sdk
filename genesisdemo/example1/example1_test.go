package example1

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/genesisdemo"
	"github.com/cosmos/cosmos-sdk/genesisdemo/bar"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var param genesisdemo.GenesisParams

type genesis struct {
	AppState genesisdemo.Options `json:"app_state"`
}

func TestExample1(t *testing.T) {
	var genesis genesis
	err := json.Unmarshal([]byte(genesisdemo.Genesis), &genesis)
	require.NoError(t, err)
	var ctx sdk.Context
	module1 := bar.NewModule()
	for _, m := range []genesisdemo.Initializer{module1} {
		require.NoError(t, m.FromGenesis(ctx, genesis.AppState, param))
	}
	assert.Equal(t, "example", module1.State)
}
