// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package sync

import (
	"time"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"

	"testing"
)

const (
	BlockInterval        = 5
	AcceptedNetWorkDelay = 2
	DynastySize          = 6
)

var (
	stor, _ = storage.NewMemoryStorage()
)

var (
	// must be order by address.hash
	MockDynasty = []string{
		"n1LQxBdAtxcfjUazHeK94raKdxRsNpujUyU",
		"n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr",
		"n1SRGKRFrF6DHK4Ym4MoXbbUHYkV5W2MZPw",
		"n1TRySsvYmAU8ChPZyYyvrPpDYJ1Z5DFoxo",
		"n1aoyV8M2g79pFXxdZEK9GfU7fzuJcCN75X",
		"n1beo9QAjhhJX6tjpjHyinoorbqdi6UKAEb",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "n1UZtMgi94oE913L2Sa2C9XwvAzNTQ82v64",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1coJhpn8QXvKFogVG93wx49eCQ6aPQHSAN",
				Value:   "10000000000000000000000",
			},
		},
	}
}

type mockManager struct{}

func (m mockManager) NewAccount([]byte) (*core.Address, error) { return nil, nil }
func (m mockManager) Accounts() []*core.Address                { return nil }

func (m mockManager) Unlock(addr *core.Address, passphrase []byte, expire time.Duration) error {
	return nil
}
func (m mockManager) Lock(addr *core.Address) error { return nil }

func (m mockManager) SignBlock(addr *core.Address, block *core.Block) error  { return nil }
func (m mockManager) SignTransaction(*core.Address, *core.Transaction) error { return nil }
func (m mockManager) SignTransactionWithPassphrase(*core.Address, *core.Transaction, []byte) error {
	return nil
}

func (m mockManager) Update(*core.Address, []byte, []byte) error   { return nil }
func (m mockManager) Load([]byte, []byte) (*core.Address, error)   { return nil, nil }
func (m mockManager) Import([]byte, []byte) (*core.Address, error) { return nil, nil }
func (m mockManager) Delete(*core.Address, []byte) error           { return nil }

var (
	received = []byte{}
)

type mockNetService struct{}

func (n mockNetService) Start() error { return nil }
func (n mockNetService) Stop()        {}

func (n mockNetService) Node() *net.Node { return nil }

func (n mockNetService) Sync(net.Serializable) error { return nil }

func (n mockNetService) Register(...*net.Subscriber)   {}
func (n mockNetService) Deregister(...*net.Subscriber) {}

func (n mockNetService) Broadcast(name string, msg net.Serializable, priority int) {}
func (n mockNetService) Relay(name string, msg net.Serializable, priority int)     {}
func (n mockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n mockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n mockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n mockNetService) ClosePeer(peerID string, reason error) {}

func (n mockNetService) BroadcastNetworkID([]byte) {}

func (n mockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

type mockNeb struct {
	config    *nebletpb.Config
	chain     *core.BlockChain
	ns        net.Service
	am        core.AccountManager
	genesis   *corepb.Genesis
	storage   storage.Storage
	consensus core.Consensus
	emitter   *core.EventEmitter
	nvm       core.Engine
}

func (n *mockNeb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *mockNeb) Storage() storage.Storage {
	return n.storage
}

func (n *mockNeb) EventEmitter() *core.EventEmitter {
	return n.emitter
}

func (n *mockNeb) Consensus() core.Consensus {
	return n.consensus
}

func (n *mockNeb) BlockChain() *core.BlockChain {
	return n.chain
}

func (n *mockNeb) NetService() net.Service {
	return n.ns
}

func (n *mockNeb) AccountManager() core.AccountManager {
	return n.am
}

func (n *mockNeb) Nvm() core.Engine {
	return n.nvm
}

func (n *mockNeb) StartPprof(string) error {
	return nil
}

func (n *mockNeb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

func (n *mockNeb) StartActiveSync() {}

func testNeb(t *testing.T) *mockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	consensus := dpos.NewDpos()
	var ns mockNetService
	neb := &mockNeb{
		genesis:   MockGenesisConf(),
		storage:   storage,
		emitter:   eventEmitter,
		consensus: consensus,
		ns:        ns,
		config: &nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    MockGenesisConf().Meta.ChainId,
				Keydir:     "keydir",
				Coinbase:   "n1K4rWU3YrhZmU1GHHYqnES8CcypTYQa9oJ",
				Miner:      "n1SRGKRFrF6DHK4Ym4MoXbbUHYkV5W2MZPw",
				Passphrase: "passphrase",
			},
		},
	}
	neb.am = account.NewManager(neb)
	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	chain.BlockPool().RegisterInNetwork(ns)
	neb.chain = chain
	assert.Nil(t, consensus.Setup(neb))
	assert.Nil(t, chain.Setup(neb))
	return neb
}

func GetUnlockAddress(t *testing.T, am core.AccountManager, addr string) *core.Address {
	address, err := core.AddressParse(addr)
	assert.Nil(t, err)
	assert.Nil(t, am.Unlock(address, []byte("passphrase"), time.Second*60*60*24*365))
	return address
}

func TestChunk_generateChunkMeta(t *testing.T) {
	neb := testNeb(t)
	chain := neb.chain
	ck := NewChunk(chain)
	am := neb.AccountManager()

	blocks := []*core.Block{}
	for i := 0; i < 96; i++ {
		coinbase := GetUnlockAddress(t, am, MockDynasty[(i+1)%DynastySize])
		consensusState, err := chain.TailBlock().NextConsensusState(BlockInterval)
		assert.Nil(t, err)
		block, err := chain.NewBlock(coinbase)
		assert.Nil(t, err)
		block.LoadConsensusState(consensusState)
		assert.Nil(t, block.Seal())
		assert.Nil(t, am.SignBlock(coinbase, block))
		assert.Nil(t, chain.BlockPool().Push(block))
		assert.Nil(t, chain.SetTailBlock(block))
		blocks = append(blocks, block)
	}

	meta, err := ck.generateChunkHeaders(blocks[1].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)
	chunks, err := ck.generateChunkData(meta.ChunkHeaders[1])
	assert.Nil(t, err)
	assert.Equal(t, len(chunks.Blocks), core.ChunkSize)
	for i := 0; i < core.ChunkSize; i++ {
		index := core.ChunkSize + 2 + i
		assert.Equal(t, int(chunks.Blocks[i].Height), index)
	}

	meta, err = ck.generateChunkHeaders(chain.GenesisBlock().Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[31].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[31].Height()), 33)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	meta, err = ck.generateChunkHeaders(blocks[62].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[62].Height()), 64)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	neb2 := testNeb(t)
	chain2 := neb2.BlockChain()
	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	for _, header := range meta.ChunkHeaders {
		chunk, err := ck.generateChunkData(header)
		assert.Nil(t, err)
		for _, v := range chunk.Blocks {
			block := new(core.Block)
			assert.Nil(t, block.FromProto(v))
			assert.Nil(t, chain2.BlockPool().Push(block))
		}
	}
	tail := blocks[len(blocks)-1]
	bytes, err := chain2.Storage().Get(tail.Hash())
	assert.Nil(t, err)
	pbBlock := new(corepb.Block)
	checkBlock := new(core.Block)
	assert.Nil(t, proto.Unmarshal(bytes, pbBlock))
	assert.Nil(t, checkBlock.FromProto(pbBlock))
	assert.Equal(t, checkBlock.Hash(), tail.Hash())
}
