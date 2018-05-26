package node

import (
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/fs"
	"github.com/jephir/tradeblocks/web"
	"log"
	"net/http"
	"sync"
)

type peerMap map[string]struct{} // "IP:port" address

type blockHashMap map[string]struct{}

// Node represents a node in the TradeBlocks network
type Node struct {
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
	storage := fs.NewBlockStorage(store, dir)
	server := web.NewServer(store)
	c := &http.Client{}
	n = &Node{
		store:             store,
		storage:           storage,
		client:            c,
		server:            server,
		peers:             make(peerMap),
		seenAccountBlocks: make(blockHashMap),
	}
	err = storage.Load()
	if err != nil {
		return
	}
	store.AccountChangeListener = n.accountChangeHandler()
	return
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

		// Send block to account listeners
		if err := n.server.BroadcastAccountBlock(b); err != nil {
			log.Printf("node: couldn't broadcast account block: %s", err.Error())
		}

		// Broadcast block if not seen before
		if _, found := n.seenAccountBlocks[hash]; !found {
			n.seenAccountBlocks[hash] = struct{}{}
			for address := range n.peers {
				c := web.NewClient(address)
				r, err := c.NewPostAccountRequest(b)
				if err != nil {
					log.Print(err)
					continue
				}
				//ss, _ := app.SerializeAccountBlock(b)
				//log.Printf("node: sending %s to %s: %s", hash, address, ss)
				res, err := n.client.Do(r)
				if err != nil {
					log.Print(err)
					continue
				}
				if res.StatusCode != http.StatusOK {
					log.Printf("node: unexpected status code %d", res.StatusCode)
					continue
				}
			}
		}
	}
}

// Sync flushes all unbroadcasted blocks to known peers
func (n *Node) Sync() error {
	return nil
}
