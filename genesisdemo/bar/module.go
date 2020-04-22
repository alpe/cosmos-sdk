package bar

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/genesisdemo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/jsonpb"
)

const Name = "bar"

type Module struct {
	State string
}

func NewModule() *Module {
	return &Module{}
}

func (a *Module) FromGenesis1(ctx sdk.Context, opts genesisdemo.Options, params genesisdemo.GenesisParams) error {
	var data GenesisState
	unmarshaler := jsonpb.Unmarshaler{}
	raw, ok := opts[Name]
	if !ok {
		return errors.Wrapf(genesisdemo.ErrEmpty, "key %s", Name)
	}
	if err := unmarshaler.Unmarshal(bytes.NewReader(raw), &data); err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	// persist
	a.State = data.Any
	return nil
}

type GenesisSource interface {
	GetBar() GenesisState
}

func (a *Module) FromGenesis2(ctx sdk.Context, data GenesisSource, params genesisdemo.GenesisParams) error {
	// persist
	a.State = data.GetBar().Any
	return nil
}
