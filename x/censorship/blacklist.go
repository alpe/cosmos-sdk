package censorship

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/types"
)

type Blacklist interface {
	Blacklisted(types.Vote) bool
	Add(slot, height int64, validator sdk.ValAddress) error
	Remove(slot, height int64, validator sdk.ValAddress) error
}
