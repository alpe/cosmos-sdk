package censorship

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

const NotInitialized = "not initialized"

type Slot struct {
	Position          int
	AddressHex        string
	AddressBech32     string
	LastUpdatedHeight int64
}

type Response struct {
	LastModifiedHeight, currentHeight int64
	Slots                             []Slot
}

func NewHttpHandler(stateIn <-chan StateSnapshot, logger log.Logger) http.Handler {
	var mu sync.RWMutex
	out := []byte(NotInitialized)
	go func() {
		for s := range stateIn {
			r := Response{
				LastModifiedHeight: s.height,
				Slots:              make([]Slot, len(s.slots)),
			}
			for i, v := range s.slots {
				bech32, hexAddr := "empty", "empty"
				if v.addr != nil {
					bech32 = sdk.ValAddress(v.addr).String()
					hexAddr = strings.ToUpper(hex.EncodeToString(v.addr.Bytes()))
				}
				r.Slots[i] = Slot{
					Position:          i,
					AddressBech32:     bech32,
					AddressHex:        hexAddr,
					LastUpdatedHeight: v.lastUpdatedHeight,
				}
			}
			var newOut bytes.Buffer
			encoder := json.NewEncoder(&newOut)
			encoder.SetIndent(" ", " ")
			if err := encoder.Encode(&r); err != nil {
				logger.Error("failed to encode response", "cause", err)
				newOut.WriteString("internal error")
			}
			mu.Lock()
			out = newOut.Bytes()
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		io.Copy(w, bytes.NewReader(out))
	})
}
