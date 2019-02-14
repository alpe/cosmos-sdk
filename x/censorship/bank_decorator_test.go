package censorship

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

func TestMsgSendAdd(t *testing.T) {
	logger := log.TestingLogger()
	store := NewBlacklistStore(logger)
	d := Decorator(func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		return sdk.Result{Code: 0}
	}, store, logger)

	// when
	ctx := sdk.NewContext(nil, abci.Header{}, true, logger)
	ctx = ctx.WithString("memo", "v/add/9/77C5A1F576A8A077D7815FD30349B89EA137FD70").WithBlockHeight(1)
	d(ctx, bank.NewMsgSend(sdk.AccAddress([]byte("from")), ourAccount, sdk.Coins{sdk.NewInt64Coin("stake", 10)}))

	// then
	expHexAdd, _ := hex.DecodeString("77C5A1F576A8A077D7815FD30349B89EA137FD70")
	assert.True(t, store.Blacklisted(types.Vote{ValidatorAddress: types.Address(expHexAdd), Height: 2}))
}

func TestMsgSendRemove(t *testing.T) {
	logger := log.TestingLogger()
	expHexAdd, _ := hex.DecodeString("77C5A1F576A8A077D7815FD30349B89EA137FD70")

	store := NewBlacklistStore(logger)
	store.Add(1, 1, expHexAdd)
	d := Decorator(func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		return sdk.Result{Code: 0}
	}, store, logger)

	// when
	ctx := sdk.NewContext(nil, abci.Header{}, true, logger)
	ctx = ctx.WithString("memo", "v/rm/1/77C5A1F576A8A077D7815FD30349B89EA137FD70").WithBlockHeight(2)
	d(ctx, bank.NewMsgSend(sdk.AccAddress([]byte("from")), ourAccount, sdk.Coins{sdk.NewInt64Coin("stake", 10)}))

	// then
	assert.True(t, store.Blacklisted(types.Vote{ValidatorAddress: types.Address(expHexAdd)}))
}

func TestDoNotCensorInTheSameBlock(t *testing.T) {
	logger := log.TestingLogger()
	expHexAdd, _ := hex.DecodeString("77C5A1F576A8A077D7815FD30349B89EA137FD70")

	store := NewBlacklistStore(logger)
	// when
	store.Add(0, 1, expHexAdd)
	// then
	assert.False(t, store.Blacklisted(types.Vote{ValidatorAddress: types.Address(expHexAdd), Height: 1}))
}

func TestParseMemo(t *testing.T) {
	expHexAdd, _ := hex.DecodeString("77C5A1F576A8A077D7815FD30349B89EA137FD70")
	specs := map[string]struct {
		src    string
		expErr bool
		op     string
		slot   int64
		target sdk.ValAddress
	}{
		"ok add-slot:0":         {"v/add/0/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 0, expHexAdd},
		"ok add-slot:1":         {"v/add/1/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 1, expHexAdd},
		"ok add-slot:2":         {"v/add/2/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 2, expHexAdd},
		"ok add-slot:3":         {"v/add/3/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 3, expHexAdd},
		"ok add-slot:4":         {"v/add/4/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 4, expHexAdd},
		"ok add-slot:5":         {"v/add/5/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 5, expHexAdd},
		"ok add-slot:6":         {"v/add/6/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 6, expHexAdd},
		"ok add-slot:7":         {"v/add/7/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 7, expHexAdd},
		"ok add-slot:8":         {"v/add/8/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 8, expHexAdd},
		"ok add-slot:9":         {"v/add/9/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "add", 9, expHexAdd},
		"ok rm-slot:9":          {"v/rm/9/77C5A1F576A8A077D7815FD30349B89EA137FD70", false, "rm", 9, expHexAdd},
		"ok with random suffix": {"v/rm/9/77C5A1F576A8A077D7815FD30349B89EA137FD70 3434jajdfajsf&^&*(()", false, "rm", 9, expHexAdd},

		"not adding us":      {"v/add/0/C39B97C2AC4777F9704A297DC17ED629A9AE7FE9", true, "", 0, nil},
		"not numeric slot":   {"v/add/a/77C5A1F576A8A077D7815FD30349B89EA137FD70", true, "", 0, nil},
		"exceeding max slot": {"v/add/a/77C5A1F576A8A077D7815FD30349B89EA137FD70", true, "", 0, nil},
		"wrong prefix":       {"x/add/0/77C5A1F576A8A077D7815FD30349B89EA137FD70", true, "", 0, nil},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			op, slot, target, err := parseMemo(spec.src)
			if spec.expErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, spec.op, op)
			assert.Equal(t, spec.slot, slot)
			assert.Equal(t, spec.target, target)
		})
	}

}
