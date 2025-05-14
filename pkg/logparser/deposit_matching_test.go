package logparser_test

import (
	"encoding/json"
	"testing"

	"github.com/Golem-Base/bridgette/pkg/logparser"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestDepositMatching(t *testing.T) {

	var l1Event *logparser.L1StandardBridgeETHDepositInitiated
	{
		// Parse the log from the fixture
		var log types.Log
		err := json.Unmarshal(ethDepositInitiatedFixture, &log)
		require.NoError(t, err)

		// Parse the event
		l1Event, err = logparser.ParseL1StandardBridgeETHDepositInitiatedEvent(&log)
		require.NoError(t, err)
	}
	var l2Event *logparser.L2StandardBridgeDepositFinalized
	{
		// Parse the log from the fixture
		var log types.Log
		err := json.Unmarshal(depositFinalizedFixture, &log)
		require.NoError(t, err)

		// Parse the event
		l2Event, err = logparser.ParseL2StandardBridgeDepositFinalizedEvent(&log)
		require.NoError(t, err)
	}

	require.Equal(t, l1Event.DepositMatchingHash(), l2Event.DepositMatchingHash())

}
