package node

import (
	"fmt"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/fs"
	"github.com/jephir/tradeblocks/web"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type peerMap map[string]struct{} // "IP:port" address

type blockHashMap map[string]struct{}

// Node represents a node in the TradeBlocks network
type Node struct {
	dir     string
	store   *app.BlockStore
	storage *fs.BlockStorage
	client  *http.Client
	server  *web.Server

	mu                sync.Mutex
	peers             peerMap
	seenAccountBlocks blockHashMap
}

// NewNode creates a new node that bootstraps from the specified URL. An error is returned if boostrapping fails.
func NewNode(dir string) (n *Node, err error) {
	store := app.NewBlockStore()
	storage := fs.NewBlockStorage(store, blocksDir(dir))
	server := web.NewServer(store)
	c := &http.Client{}
	n = &Node{
		dir:               dir,
		store:             store,
		storage:           storage,
		client:            c,
		server:            server,
		peers:             make(peerMap),
		seenAccountBlocks: make(blockHashMap),
	}
	err = n.initStorage()
	if err != nil {
		return
	}
	store.AccountChangeListener = n.accountChangeHandler()
	return
}

func (n *Node) initStorage() error {
	if err := os.MkdirAll(blocksDir(n.dir), 0700); err != nil {
		return err
	}
	return n.storage.Load()
}

// Bootstrap registers with the specified server and downloads all blocks
func (n *Node) Bootstrap(hostURL, bootstrapURL string) error {
	client := web.NewClient(bootstrapURL)

	// Create get all blocks request and register this server
	r, err := client.NewGetBlocksRequest()
	if err != nil {
		return err
	}
	r.Header.Add("TradeBlocks-Register", hostURL)

	// Execute get all blocks request
	res, err := n.client.Do(r)
	if err != nil {
		return err
	}

	// Decode response
	blocks, err := client.DecodeGetBlocksResponse(res)
	if err != nil {
		return err
	}

	// Add blocks to store
	// TODO order incoming blocks and validate them first
	for hash, block := range blocks {
		n.store.AccountBlocks[hash] = block
		n.seenAccountBlocks[hash] = struct{}{}
	}
	return nil
}

func (n *Node) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if addr := r.Header.Get("TradeBlocks-Register"); addr != "" {
		n.addPeer(addr)
	}
	n.server.ServeHTTP(rw, r)
}

// AddBlock adds the specified block to this node
func (n *Node) AddBlock(b tradeblocks.Block) error {
	if b, ok := b.(*tradeblocks.AccountBlock); ok {
		_, err := n.store.AddBlock(b)
		return err
	}
	return fmt.Errorf("node: unsupported block type")
}

// GetAccountBlock returns the account block with the specified hash; otherwise nil
func (n *Node) GetAccountBlock(h string) *tradeblocks.AccountBlock {
	b, err := n.store.GetBlock(h)
	if err != nil {
		panic(err)
	}
	return b
}

func (n *Node) addPeer(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// TODO do connection check before adding peer
	n.peers[address] = struct{}{}
}

func (n *Node) accountChangeHandler() app.AccountChangeListener {
	return func(hash string, b *tradeblocks.AccountBlock) {
		n.mu.Lock()
		defer n.mu.Unlock()

		if err := n.storage.SaveBlock(hash, b); err != nil {
			log.Printf("node: couldn't save block '%s': %s", hash, err.Error())
		}

		// Send block to account listeners
		if err := n.server.BroadcastAccountBlock(b); err != nil {
			log.Printf("node: couldn't broadcast account block: %s", err.Error())
		}

		// Broadcast block if not seen before
		if _, found := n.seenAccountBlocks[hash]; !found {
			n.seenAccountBlocks[hash] = struct{}{}
			for address := range n.peers {
				if err := n.broadcastAccountBlock(address, b); err != nil {
					log.Println(err.Error())
				}
			}
		}
	}
}

func (n *Node) broadcastAccountBlock(address string, b *tradeblocks.AccountBlock) error {
	c := web.NewClient(address)
	r, err := c.NewPostAccountRequest(b)
	if err != nil {
		return err
	}
	//ss, _ := app.SerializeAccountBlock(b)
	//log.Printf("node: sending %s to %s: %s", hash, address, ss)
	res, err := n.client.Do(r)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("node: unexpected status code %d", res.StatusCode)
	}
	return nil
}

func (n *Node) broadcastVote(address string, b *tradeblocks.AccountBlock) error {
	// v := &tradeblocks.VoteBlock{
	// 	Account: b.Account,
	// 	Link: b.Hash(),
	// 	Order: 0,
	// 	Signature: "",
	// }
	return nil
}

// Sync flushes all unbroadcasted blocks to known peers
func (n *Node) Sync() error {
	return nil
}

func blocksDir(dir string) string {
	return filepath.Join(dir, "blocks")
}
