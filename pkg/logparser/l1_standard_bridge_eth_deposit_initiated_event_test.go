package logparser_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/Golem-Base/bridgette/pkg/logparser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed fixtures/l1/000000000003831667-0017-0031.json
var ethDepositInitiatedFixture []byte

func TestParseL1StandardBridgeETHDepositInitiatedEvent(t *testing.T) {
	// Parse the log from the fixture
	var log types.Log
	err := json.Unmarshal(ethDepositInitiatedFixture, &log)
	require.NoError(t, err)

	// Parse the event
	event, err := logparser.ParseL1StandardBridgeETHDepositInitiatedEvent(&log)
	require.NoError(t, err)

	// Assert the expected values
	assert.Equal(t, common.HexToAddress("0x9192c90ffb804d224b0988b1dbfc1d0be199c257"), event.From)
	assert.Equal(t, common.HexToAddress("0x9192c90ffb804d224b0988b1dbfc1d0be199c257"), event.To)
	assert.Equal(t, "5000000000000000000000", event.Amount.String()) // 5000 ETH in wei

	// The fixture shows extraData is empty
	expectedExtraData := make([]byte, 0)
	assert.Equal(t, expectedExtraData, event.ExtraData)
}
