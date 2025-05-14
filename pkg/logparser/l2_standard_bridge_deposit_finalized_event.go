package logparser

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type L2StandardBridgeDepositFinalized bindings.L2StandardBridgeDepositFinalized

func (e *L2StandardBridgeDepositFinalized) DepositMatchingHash() common.Hash {
	return crypto.Keccak256Hash(e.L1Token.Bytes(), e.From.Bytes(), e.Amount.Bytes(), e.ExtraData)
}

func ParseL2StandardBridgeDepositFinalizedEvent(log *types.Log) (*L2StandardBridgeDepositFinalized, error) {
	contractAbi, err := bindings.L2StandardBridgeMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get L2StandardBridge ABI: %w", err)
	}

	event := new(L2StandardBridgeDepositFinalized)
	err = contractAbi.UnpackIntoInterface(event, "DepositFinalized", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack log data: %w", err)
	}

	// The first three topics are the l1Token, l2Token, and from addresses
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf("invalid number of topics: got %d, want 4", len(log.Topics))
	}

	event.L1Token = common.BytesToAddress(log.Topics[1].Bytes())
	event.L2Token = common.BytesToAddress(log.Topics[2].Bytes())
	event.From = common.BytesToAddress(log.Topics[3].Bytes())

	return event, nil
}
