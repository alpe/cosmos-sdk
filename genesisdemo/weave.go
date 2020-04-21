package genesisdemo

import (
	"bytes"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Options are the app options
// Each extension can look up it's key and parse the json as desired
type Options map[string]json.RawMessage

// ReadOptions reads the values stored under a given key,
// and parses the json into the given obj.
// Returns an error if it cannot parse.
// Noop and no error if key is missing
func (o Options) ReadOptions(key string, obj interface{}) error {
	msg := o[key]
	if len(msg) == 0 {
		return nil
	}
	return json.Unmarshal(msg, obj)
}

// Stream expects an array of json elements and allows to process them sequentially
// this helps when one needs to parse a large json without having any memory leaks.
// Returns ErrEmpty on empty key or when there are no more elements.
// Returns ErrState when the stream has finished/encountered a Decode error.mi
func (o Options) Stream(key string) func(obj interface{}) error {
	msg := o[key]
	dec := json.NewDecoder(bytes.NewReader(msg))
	initialized := false
	closed := false

	return func(obj interface{}) error {
		if !initialized {
			if len(msg) == 0 {
				return errors.Wrap(ErrEmpty, "data")
			}

			// read opening bracket
			if _, err := dec.Token(); err != nil {
				return errors.Wrapf(ErrInput, "opening bracket %s", err)
			}

			initialized = true
		}
		if closed {
			return errors.Wrap(ErrState, "closed")
		}

		if dec.More() {
			if err := dec.Decode(obj); err != nil {
				return errors.Wrapf(ErrInput, "decode %s", err)
			}
			return nil
		}

		closed = true
		// read closing bracket
		if _, err := dec.Token(); err != nil {
			return errors.Wrapf(ErrInput, "closing bracket %s", err)
		}

		return errors.Wrap(ErrEmpty, "end")
	}
}

// GenesisParams represents parameters set in genesis that could be useful
// for some of the extensions.
type GenesisParams struct {
	Validators []abci.ValidatorUpdate
}

// FromInitChain initialises GenesisParams using abci.RequestInitChain
// data.
func FromInitChain(req abci.RequestInitChain) GenesisParams {
	return GenesisParams{
		Validators: req.Validators,
	}
}

// Initializer implementations are used to initialize
// extensions from genesis file contents
type Initializer interface {
	FromGenesis(ctx sdk.Context, opts Options, params GenesisParams) error
}
