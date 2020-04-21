package genesisdemo

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

const ModuleName = "genesisdemo"

var (
	ErrInput = sdkerrors.Register(ModuleName, 101, "value is empty")
	ErrState = sdkerrors.Register(ModuleName, 102, "value is empty")
	ErrEmpty = sdkerrors.Register(ModuleName, 103, "value is empty")
)
