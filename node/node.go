package node

import (
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
	store   *app.BlockStore2
	storage *fs.BlockStorage
	client  *http.Client
	server  *web.Server

	mu                sync.Mutex
	peers             peerMap
	seenAccountBlocks blockHashMap
}

// NewNode creates a new node that bootstraps from the specified URL. An error is returned if boostrapping fails.
func NewNode(dir string) (n *Node, err error) {
	store := app.NewBlockStore2()
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
	server.BlockHandler = n.handleBlock
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
	accounts, err := n.bootstrapAccounts(hostURL, client)
	if err != nil {
		return err
	}
	swaps, err := n.bootstrapSwaps(hostURL, client)
	if err != nil {
		return err
	}
	orders, err := n.bootstrapOrders(hostURL, client)
	if err != nil {
		return err
	}
	var sequence int
	for {
		var found bool
		if b := accountWithSequence(accounts, sequence); b != nil {
			if err := n.store.AddAccountBlock(b.AccountBlock); err != nil {
				return err
			}
			found = true
		}
		if b := swapWithSequence(swaps, sequence); b != nil {
			if err := n.store.AddSwapBlock(b.SwapBlock); err != nil {
				return err
			}
			found = true
		}
		if b := orderWithSequence(orders, sequence); b != nil {
			if err := n.store.AddOrderBlock(b.OrderBlock); err != nil {
				return err
			}
			found = true
		}
		if !found {
			break
		}
		sequence++
	}
	return n.storage.Save()
}

func accountWithSequence(accounts map[string]tradeblocks.NetworkAccountBlock, sequence int) *tradeblocks.NetworkAccountBlock {
	for _, b := range accounts {
		if b.Sequence == sequence {
			return &b
		}
	}
	return nil
}

func swapWithSequence(swaps map[string]tradeblocks.NetworkSwapBlock, sequence int) *tradeblocks.NetworkSwapBlock {
	for _, b := range swaps {
		if b.Sequence == sequence {
			return &b
		}
	}
	return nil
}

func orderWithSequence(orders map[string]tradeblocks.NetworkOrderBlock, sequence int) *tradeblocks.NetworkOrderBlock {
	for _, b := range orders {
		if b.Sequence == sequence {
			return &b
		}
	}
	return nil
}

func (n *Node) bootstrapAccounts(hostURL string, client *web.Client) (map[string]tradeblocks.NetworkAccountBlock, error) {
	// Create get all blocks request and register this server
	r, err := client.NewGetAccountBlocksRequest()
	if err != nil {
		return nil, err
	}
	r.Header.Add("TradeBlocks-Register", hostURL)

	// Execute get all blocks request
	res, err := n.client.Do(r)
	if err != nil {
		return nil, err
	}

	// Decode response
	return client.DecodeGetAccountBlocksResponse(res)
}

func (n *Node) bootstrapSwaps(hostURL string, client *web.Client) (map[string]tradeblocks.NetworkSwapBlock, error) {
	// Create get all blocks request and register this server
	r, err := client.NewGetSwapBlocksRequest()
	if err != nil {
		return nil, err
	}
	r.Header.Add("TradeBlocks-Register", hostURL)

	// Execute get all blocks request
	res, err := n.client.Do(r)
	if err != nil {
		return nil, err
	}

	// Decode response
	return client.DecodeGetSwapBlocksResponse(res)
}

func (n *Node) bootstrapOrders(hostURL string, client *web.Client) (map[string]tradeblocks.NetworkOrderBlock, error) {
	// Create get all blocks request and register this server
	r, err := client.NewGetOrderBlocksRequest()
	if err != nil {
		return nil, err
	}
	r.Header.Add("TradeBlocks-Register", hostURL)

	// Execute get all blocks request
	res, err := n.client.Do(r)
	if err != nil {
		return nil, err
	}

	// Decode response
	return client.DecodeGetOrderBlocksResponse(res)
}

func (n *Node) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if addr := r.Header.Get("TradeBlocks-Register"); addr != "" {
		n.addPeer(addr)
	}
	n.server.ServeHTTP(rw, r)
}

func (n *Node) handleBlock(b app.TypedBlock) {
	// TODO don't broadcast if block already seen

	// Save block
	switch b.T {
	case "account":
		if err := n.storage.SaveAccountBlock(b.AccountBlock); err != nil {
			log.Println(err)
		}
		if err := n.server.BroadcastBlock(b.AccountBlock); err != nil {
			log.Println(err)
		}
	case "swap":
		if err := n.storage.SaveSwapBlock(b.SwapBlock); err != nil {
			log.Println(err)
		}
		if err := n.server.BroadcastBlock(b.SwapBlock); err != nil {
			log.Println(err)
		}
	case "order":
		if err := n.storage.SaveOrderBlock(b.OrderBlock); err != nil {
			log.Println(err)
		}
		if err := n.server.BroadcastBlock(b.OrderBlock); err != nil {
			log.Println(err)
		}
	}

	// Broadcast to peers
	for address := range n.peers {
		c := web.NewClient(address)
		switch b.T {
		case "account":
			req, err := c.NewPostAccountBlockRequest(b.AccountBlock)
			if err != nil {
				log.Println(err)
				return
			}
			res, err := n.client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			var rb tradeblocks.AccountBlock
			if err := c.DecodeAccountBlockResponse(res, &rb); err != nil {
				log.Println(err)
				return
			}
		case "swap":
			req, err := c.NewPostSwapBlockRequest(b.SwapBlock)
			if err != nil {
				log.Println(err)
				return
			}
			res, err := n.client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			var rb tradeblocks.SwapBlock
			if err := c.DecodeSwapBlockResponse(res, &rb); err != nil {
				log.Println(err)
				return
			}
		case "order":
			req, err := c.NewPostOrderBlockRequest(b.OrderBlock)
			if err != nil {
				log.Println(err)
				return
			}
			res, err := n.client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			var rb tradeblocks.OrderBlock
			if err := c.DecodeOrderBlockResponse(res, &rb); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

// // AddBlock adds the specified block to this node
// func (n *Node) AddBlock(b tradeblocks.Block) error {
// 	if b, ok := b.(*tradeblocks.AccountBlock); ok {
// 		_, err := n.store.AddBlock(b)
// 		return err
// 	}
// 	return fmt.Errorf("node: unsupported block type")
// }

// // GetAccountBlock returns the account block with the specified hash; otherwise nil
// func (n *Node) GetAccountBlock(h string) *tradeblocks.AccountBlock {
// 	b, err := n.store.GetBlock(h)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b
// }

func (n *Node) addPeer(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// TODO do connection check before adding peer
	n.peers[address] = struct{}{}
}

// func (n *Node) accountChangeHandler() app.AccountChangeListener {
// 	return func(hash string, b *tradeblocks.AccountBlock) {
// 		n.mu.Lock()
// 		defer n.mu.Unlock()

// 		if err := n.storage.SaveBlock(hash, b); err != nil {
// 			log.Printf("node: couldn't save block '%s': %s", hash, err.Error())
// 		}

// 		// Send block to account listeners
// 		if err := n.server.BroadcastAccountBlock(b); err != nil {
// 			log.Printf("node: couldn't broadcast account block: %s", err.Error())
// 		}

// 		// Broadcast block if not seen before
// 		if _, found := n.seenAccountBlocks[hash]; !found {
// 			n.seenAccountBlocks[hash] = struct{}{}
// 			for address := range n.peers {
// 				if err := n.broadcastAccountBlock(address, b); err != nil {
// 					log.Println(err.Error())
// 				}
// 			}
// 		}
// 	}
// }

// func (n *Node) broadcastAccountBlock(address string, b *tradeblocks.AccountBlock) error {
// 	c := web.NewClient(address)
// 	r, err := c.NewPostAccountRequest(b)
// 	if err != nil {
// 		return err
// 	}
// 	//ss, _ := app.SerializeAccountBlock(b)
// 	//log.Printf("node: sending %s to %s: %s", hash, address, ss)
// 	res, err := n.client.Do(r)
// 	if err != nil {
// 		return err
// 	}
// 	if res.StatusCode != http.StatusOK {
// 		return fmt.Errorf("node: unexpected status code %d", res.StatusCode)
// 	}
// 	return nil
// }

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
