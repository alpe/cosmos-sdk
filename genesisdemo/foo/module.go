package foo

import (
	"github.com/cosmos/cosmos-sdk/genesisdemo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

const Name = "foo"

type Module struct {
	State []string
}

func NewModule() *Module {
	return &Module{}
}

func (a *Module) FromGenesis1(ctx sdk.Context, opts genesisdemo.Options, params genesisdemo.GenesisParams) error {
	f := opts.Stream(Name)
	for {
		var data GenesisState
		if err := f(&data); err != nil {
			if genesisdemo.ErrEmpty.Is(err) {
				break
			}
			return errors.Wrap(err, "stream")
		}
		a.State = append(a.State, data.Other)
	}
	return nil
}

type GenesisSource interface {
	GetFoo() []GenesisState
}

func (a *Module) FromGenesis2(ctx sdk.Context, data GenesisSource, param genesisdemo.GenesisParams) error {
	for _, s := range data.GetFoo() {
		a.State = append(a.State, s.Other)
	}
	return nil
}
