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

//go:embed fixtures/l2/000000000001059695-0001-0000.json
var depositFinalizedFixture []byte

func TestParseL2StandardBridgeDepositFinalizedEvent(t *testing.T) {
	// Parse the log from the fixture
	var log types.Log
	err := json.Unmarshal(depositFinalizedFixture, &log)
	require.NoError(t, err)

	// Parse the event
	event, err := logparser.ParseL2StandardBridgeDepositFinalizedEvent(&log)
	require.NoError(t, err)

	// Assert the expected values
	assert.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000000"), event.L1Token)
	assert.Equal(t, common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000"), event.L2Token)
	assert.Equal(t, common.HexToAddress("0x9192c90ffb804d224b0988b1dbfc1d0be199c257"), event.From)
	assert.Equal(t, common.HexToAddress("0x9192c90ffb804d224b0988b1dbfc1d0be199c257"), event.To)
	assert.Equal(t, "5000000000000000000000", event.Amount.String()) // 5000 ETH in wei

	// The fixture shows extraData is empty
	expectedExtraData := make([]byte, 0)
	assert.Equal(t, expectedExtraData, event.ExtraData)
}
