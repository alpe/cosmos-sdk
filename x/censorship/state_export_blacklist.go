package censorship

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/types"
)

type StateSnapshot struct {
	height int64
	slots  []blacklistEntry
}
type StateExportBlacklist struct {
	stateExport chan<- StateSnapshot
	store       *BlacklistStore
}

func NewStateExportBlacklistStore(store *BlacklistStore, stateExport chan<- StateSnapshot) *StateExportBlacklist {
	return &StateExportBlacklist{
		store:       store,
		stateExport: stateExport,
	}
}

func (s *StateExportBlacklist) Add(slot, height int64, validator sdk.ValAddress) error {
	if err := s.store.Add(slot, height, validator); err != nil {
		return err
	}
	return s.export(height)
}

func (s *StateExportBlacklist) Remove(slot, height int64, validator sdk.ValAddress) error {
	if err := s.store.Remove(slot, height, validator); err != nil {
		return err
	}
	return s.export(height)
}

func (s *StateExportBlacklist) Blacklisted(vote types.Vote) bool {
	return s.Blacklisted(vote)
}

func (s *StateExportBlacklist) export(height int64) error {
	if s.stateExport == nil {
		return nil
	}
	slotsCopy := make([]blacklistEntry, len(s.store.slots))
	copy(slotsCopy, s.store.slots)
	s.stateExport <- StateSnapshot{height: height, slots: slotsCopy}
	return nil
}
