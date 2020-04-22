package examples

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/genesisdemo/bar"
	"github.com/cosmos/cosmos-sdk/genesisdemo/foo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExample2(t *testing.T) {
	var genesis struct {
		AppState json.RawMessage `json:"app_state"`
	}
	err := json.Unmarshal([]byte(Genesis), &genesis)
	require.NoError(t, err)

	var appState AppGenesisState
	unmarsh := jsonpb.Unmarshaler{}
	err = unmarsh.Unmarshal(bytes.NewReader(genesis.AppState), &appState)
	require.NoError(t, err)
	var ctx sdk.Context
	module1 := bar.NewModule()
	require.NoError(t, module1.FromGenesis2(ctx, &appState, param))

	module2 := foo.NewModule()
	require.NoError(t, module2.FromGenesis2(ctx, &appState, param))

	assert.Equal(t, "example", module1.State)
	assert.Equal(t, []string{"any", "data"}, module2.State)
}
