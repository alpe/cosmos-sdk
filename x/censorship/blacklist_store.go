package censorship

import (
	"bytes"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

const HighestSlotNumber = 9

type blacklistEntry struct {
	addr              types.Address
	lastUpdatedHeight int64
}

type BlacklistStore struct {
	logger      log.Logger
	mu          sync.Mutex
	slots       []blacklistEntry
	stateExport []chan<- []blacklistEntry
}

func NewBlacklistStore(logger log.Logger, stateExport ...chan<- []blacklistEntry) *BlacklistStore {
	return &BlacklistStore{
		logger:      logger,
		slots:       make([]blacklistEntry, HighestSlotNumber+1),
		stateExport: stateExport,
	}
}

func (b *BlacklistStore) Add(slot, height int64, addr sdk.ValAddress) error {
	if slot > HighestSlotNumber || slot < 0 {
		return errors.Errorf("unsupported slot number: %d", slot)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.slots[slot].lastUpdatedHeight >= height {
		b.logger.Info("Slot already taken", "slot", slot, "addr", addr.String())
		return nil
	}
	var newAddr types.Address
	if err := newAddr.Unmarshal(addr.Bytes()); err != nil {
		return errors.Wrap(err, "failed to encode into hex format")
	}
	b.logger.Info("Adding to censorship", "slot", slot, "addr", addr.String())

	b.slots[slot] = blacklistEntry{
		addr:              newAddr,
		lastUpdatedHeight: height,
	}
	return nil
}

func (b *BlacklistStore) Remove(slot, height int64, addr sdk.ValAddress) error {
	if slot > HighestSlotNumber || slot < 0 {
		return errors.Errorf("unsupported slot number: %d", slot)
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.slots[slot].lastUpdatedHeight >= height || b.slots[slot].addr == nil {
		return nil
	}
	b.logger.Info("Removing from censorship", "slot", slot, "addr", addr.String())

	b.slots[slot] = blacklistEntry{
		addr:              nil,
		lastUpdatedHeight: 0,
	}
	return nil
}

func (b *BlacklistStore) Blacklisted(vote types.Vote) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	voter := vote.ValidatorAddress
	var result bool
	for _, s := range b.slots {
		if s.addr != nil && s.lastUpdatedHeight < vote.Height && bytes.Equal(voter, s.addr) {
			result = true
			break
		}
	}
	b.logger.Debug("Checking voter", "addr", voter.String(), "result", result)
	return result
}

type NoopBlacklist struct{}

func (n NoopBlacklist) Add(slot, height int64, validator sdk.ValAddress) error    { return nil }
func (n NoopBlacklist) Remove(slot, height int64, validator sdk.ValAddress) error { return nil }
func (b NoopBlacklist) Blacklisted(types.Vote) bool {
	return false
}
