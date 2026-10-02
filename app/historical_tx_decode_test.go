package app_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
)

// historicalTx is one entry of app/testdata/historical_txs.json: a real signed mainnet
// transaction whose message type no longer has a handler after the wind-down removals.
type historicalTx struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Height    int64  `json:"height"`
	TypeURL   string `json:"type_url"`
	AminoName string `json:"amino_name"` // the name registered in the module's RegisterCodec
	TxBase64  string `json:"tx_base64"`
}

type HistoricalTxDecodeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *HistoricalTxDecodeTestSuite) SetupTest() {
	s.Setup()
}

func TestHistoricalTxDecodeTestSuite(t *testing.T) {
	suite.Run(t, new(HistoricalTxDecodeTestSuite))
}

// Removing a message's rpc removes its handler; it must not remove the ability to decode
// the chain's history. The message types stay registered in the interface registry and
// with amino, and this test is what fails if a registration is dropped by mistake (the
// first dry run of these removals broke `strided q tx` on old hashes exactly that way).
func (s *HistoricalTxDecodeTestSuite) TestHistoricalTxsStillDecode() {
	raw, err := os.ReadFile("testdata/historical_txs.json")
	s.Require().NoError(err, "fixture must exist")

	var fixtures []historicalTx
	s.Require().NoError(json.Unmarshal(raw, &fixtures))
	s.Require().NotEmpty(fixtures)

	for _, fixture := range fixtures {
		s.Run(fixture.Name, func() {
			txBytes, err := base64.StdEncoding.DecodeString(fixture.TxBase64)
			s.Require().NoError(err)

			// Protobuf decode through the app's tx decoder (what `strided q tx` uses)
			tx, err := s.App.TxDecode(txBytes)
			s.Require().NoError(err, "tx %s must still decode", fixture.Hash)

			msgs := tx.GetMsgs()
			s.Require().Len(msgs, 1)
			s.Require().Equal(fixture.TypeURL, sdk.MsgTypeURL(msgs[0]))

			// Legacy amino JSON rendering must still carry the registered name. NoError alone
			// proves nothing: go-amino's MarshalJSON on a concrete type that lost its
			// RegisterAminoMsg line just omits the {"type": ...} wrapper and returns no error,
			// so the name check is the assertion that fails when a registration is dropped
			aminoJson, err := s.App.LegacyAmino().MarshalJSON(msgs[0])
			s.Require().NoError(err, "amino JSON for %s", fixture.TypeURL)
			s.Require().Contains(string(aminoJson), fmt.Sprintf(`"type":"%s"`, fixture.AminoName),
				"amino JSON for %s must carry its registered name", fixture.TypeURL)
		})
	}
}
