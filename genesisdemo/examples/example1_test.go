package examples

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/genesisdemo"
	"github.com/cosmos/cosmos-sdk/genesisdemo/bar"
	"github.com/cosmos/cosmos-sdk/genesisdemo/foo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var param genesisdemo.GenesisParams

func TestExample1(t *testing.T) {
	var genesis struct {
		AppState genesisdemo.Options `json:"app_state"`
	}
	err := json.Unmarshal([]byte(Genesis), &genesis)
	require.NoError(t, err)
	var ctx sdk.Context
	module1 := bar.NewModule()
	module2 := foo.NewModule()
	for _, m := range []genesisdemo.Initializer{module1, module2} {
		require.NoError(t, m.FromGenesis1(ctx, genesis.AppState, param))
	}
	assert.Equal(t, "example", module1.State)
	assert.Equal(t, []string{"any", "data"}, module2.State)
}
