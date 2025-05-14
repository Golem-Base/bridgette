package logparser

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type L1StandardBridgeETHDepositInitiated bindings.L1StandardBridgeETHDepositInitiated

var ethToken = common.HexToAddress("0x0")

func (e *L1StandardBridgeETHDepositInitiated) DepositMatchingHash() common.Hash {
	return crypto.Keccak256Hash(ethToken.Bytes(), e.From.Bytes(), e.Amount.Bytes(), e.ExtraData)
}

func ParseL1StandardBridgeETHDepositInitiatedEvent(log *types.Log) (*L1StandardBridgeETHDepositInitiated, error) {
	contractAbi, err := bindings.L1StandardBridgeMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get L1StandardBridge ABI: %w", err)
	}

	event := new(L1StandardBridgeETHDepositInitiated)
	err = contractAbi.UnpackIntoInterface(event, "ETHDepositInitiated", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack log data: %w", err)
	}

	// The first two topics are the from and to addresses
	if len(log.Topics) != 3 {
		return nil, fmt.Errorf("invalid number of topics: got %d, want 3", len(log.Topics))
	}

	event.From = common.BytesToAddress(log.Topics[1].Bytes())
	event.To = common.BytesToAddress(log.Topics[2].Bytes())

	return event, nil
}
