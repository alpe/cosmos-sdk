package censorship

import (
	"bytes"
	"encoding/hex"
	"math/rand"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"io"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
)

func TestServerWithoutState(t *testing.T) {
	logger := log.TestingLogger()
	c := make(chan StateSnapshot, 0)

	s := httptest.NewServer(NewHttpHandler(c, logger))
	defer s.Close()

	// when
	resp, err := s.Client().Get(s.URL)
	require.NoError(t, err)

	// then
	require.Equal(t, 200, resp.StatusCode)
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	require.Equal(t, NotInitialized, buf.String())

}

func TestServerWithState(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])

	logger := log.TestingLogger()
	c := make(chan StateSnapshot, 2)
	blacklist := NewStateExportBlacklistStore(NewBlacklistStore(logger), c)
	blacklist.Add(0, 2, sdk.ValAddress(pub.Address()))
	x, _ := hex.DecodeString("8792FA5FCD3B8654ADE1930F1A841F545F6B538297ECC7B436A18CF5D60E619B")
	copy(pub[:], x)
	blacklist.Add(0, 4, sdk.ValAddress(pub.Address()))

	s := httptest.NewServer(NewHttpHandler(c, logger))
	defer s.Close()

	time.Sleep(100 * time.Millisecond) // wait for server to consyme
	// when
	resp, err := s.Client().Get(s.URL)
	require.NoError(t, err)

	// then
	require.Equal(t, 200, resp.StatusCode)
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	expResult := `
{
  "LastModifiedHeight": 4,
  "Slots": [
    {
      "Position": 0,
      "AddressHex": "43B71C59245FA038C9EDD1B26A072440BF0343C1",
      "AddressBech32": "cosmosvaloper1gwm3ckfyt7sr3j0d6xex5peygzlsxs7pgsl449",
      "LastUpdatedHeight": 4
    },
    {
      "Position": 1,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 2,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 3,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 4,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 5,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 6,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 7,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 8,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    },
    {
      "Position": 9,
      "AddressHex": "empty",
      "AddressBech32": "empty",
      "LastUpdatedHeight": 0
    }
  ]
}
`
	require.JSONEq(t, expResult, buf.String(), buf.String())

}
