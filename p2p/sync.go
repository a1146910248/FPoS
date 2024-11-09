package p2p

import (
	. "FPoS/types"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Layer2Node) setupTopics() error {
	txTopic, err := n.pubsub.Join("l2_transactions")
	if err != nil {
		return err
	}
	n.txTopic = txTopic

	blockTopic, err := n.pubsub.Join("l2_blocks")
	if err != nil {
		return err
	}
	n.blockTopic = blockTopic

	stateTopic, err := n.pubsub.Join("l2_state")
	if err != nil {
		return err
	}
	n.stateTopic = stateTopic

	go n.handleTxMessages()
	go n.handleBlockMessages()
	go n.handleStateMessages()

	return nil
}

func (n *Layer2Node) handleTxMessages() {
	sub, err := n.txTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}

		var tx Transaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			if n.validateTransaction(tx) {
				n.txPool.Store(tx.Hash, tx)
				n.BroadcastTransaction(tx)
			}
		}
	}
}

func (n *Layer2Node) handleBlockMessages() {
	sub, err := n.blockTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}

		var block Block
		if err := json.Unmarshal(msg.Data, &block); err == nil {
			if n.validateBlock(block) {
				n.processNewBlock(block)
			}
		}
	}
}

func (n *Layer2Node) handleStateMessages() {
	sub, err := n.stateTopic.Subscribe()
	if err != nil {
		return
	}

	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			continue
		}

		// 尝试解析为节点信息
		var peerInfo peer.AddrInfo
		if err := json.Unmarshal(msg.Data, &peerInfo); err == nil {
			// 收到新节点信息，尝试连接
			if peerInfo.ID != n.host.ID() && n.host.Network().Connectedness(peerInfo.ID) != network.Connected {
				if err := n.host.Connect(n.ctx, peerInfo); err == nil {
					fmt.Printf("Connected to broadcasted peer: %s\n", peerInfo.ID)
				}
			}
			continue
		}

		// 如果不是节点信息，则尝试解析为状态更新
		var state struct {
			StateRoot string `json:"stateRoot"`
			Height    uint64 `json:"height"`
		}
		if err := json.Unmarshal(msg.Data, &state); err == nil {
			n.updateState(state.StateRoot, state.Height)
		}
	}
}
