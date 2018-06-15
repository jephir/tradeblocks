package node

import (
	"crypto/rand"
	"crypto/rsa"
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

const keySize = 2048

type peerMap map[string]struct{} // "IP:port" address

type blockHashMap map[string]struct{}

// Node represents a node in the TradeBlocks network
type Node struct {
	dir     string
	store   *app.BlockStore
	storage *fs.BlockStorage
	client  *http.Client
	server  *web.Server

	priv    *rsa.PrivateKey
	address string

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

	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return
	}
	address, err := app.PrivateKeyToAddress(priv)
	if err != nil {
		return
	}

	n = &Node{
		dir:               dir,
		store:             store,
		storage:           storage,
		client:            c,
		server:            server,
		priv:              priv,
		address:           address,
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
	if r.URL.Path == "/address" {
		n.handleAddress().ServeHTTP(rw, r)
		return
	}
	n.server.ServeHTTP(rw, r)
}

func (n *Node) handleAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, n.address)
	}
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

	// Check if block matches an open order
	if b.T == "swap" {
		if err := n.handleSwap(b.SwapBlock); err != nil {
			log.Printf("node: swap error: %s", err.Error())
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

func (n *Node) handleSwap(b *tradeblocks.SwapBlock) error {
	if b.Action == "offer" && b.Executor == n.address {
		order := n.store.GetOrderHead(b.Counterparty, b.ID)
		if order == nil {
			return fmt.Errorf("node: no order found for '%s:%s'", b.Counterparty, b.ID)
		}
		if order.Balance < b.Quantity {
			return fmt.Errorf("node: only '%f' balance remaining to fill order of quantity '%f' in '%s:%s'", order.Balance, b.Quantity, b.Counterparty, b.ID)
		}

		link := tradeblocks.SwapAddress(b.Account, b.ID)
		send := tradeblocks.NewAcceptOrderBlock(order, link, order.Balance-b.Quantity)
		if err := send.SignBlock(n.priv); err != nil {
			return err
		}
		if err := n.store.AddOrderBlock(send); err != nil {
			return fmt.Errorf("error adding order send: %s", err.Error())
		}
		n.server.BlockHandler(app.TypedBlock{
			OrderBlock: send,
			T:          "order",
		})

		commit := tradeblocks.NewCommitBlock(b, send)
		if err := commit.SignBlock(n.priv); err != nil {
			return err
		}
		if err := n.store.AddSwapBlock(commit); err != nil {
			return fmt.Errorf("error adding swap: %s", err.Error())
		}
		n.server.BlockHandler(app.TypedBlock{
			SwapBlock: commit,
			T:         "swap",
		})
	}
	return nil
}

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
