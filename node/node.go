package node

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"github.com/jephir/tradeblocks/db"
	"github.com/jephir/tradeblocks/web"
)

const keySize = 2048

type peerMap map[string]struct{} // "IP:port" address

type blockHashMap map[string]struct{}

// Node represents a node in the TradeBlocks network
type Node struct {
	store  *app.BlockStore
	client *http.Client
	server *web.Server

	priv    *rsa.PrivateKey
	address string

	mu                sync.Mutex
	peers             peerMap
	seenAccountBlocks blockHashMap
}

// NewNode creates a new node or returns an error if it fails.
func NewNode(dir string) (n *Node, err error) {
	f := filepath.Join(dir, "tradeblocks.db")
	store, err := app.NewPersistBlockStore(f)
	if err != nil {
		return nil, err
	}
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
		store:             store,
		client:            c,
		server:            server,
		priv:              priv,
		address:           address,
		peers:             make(peerMap),
		seenAccountBlocks: make(blockHashMap),
	}
	if err != nil {
		return
	}
	server.BlockHandler = n.handleBlock
	return
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
	return nil
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
	defer res.Body.Close()

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
	defer res.Body.Close()

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
	defer res.Body.Close()

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
		if err := n.server.BroadcastBlock(b.AccountBlock); err != nil {
			log.Println(err)
		}
		if err := n.confirmBlock(b.AccountBlock); err != nil {
			log.Printf("node: confirm error: %s", err.Error())
		}
	case "swap":
		if err := n.server.BroadcastBlock(b.SwapBlock); err != nil {
			log.Println(err)
		}
		if err := n.confirmBlock(b.SwapBlock); err != nil {
			log.Printf("node: confirm error: %s", err.Error())
		}
	case "order":
		if err := n.server.BroadcastBlock(b.OrderBlock); err != nil {
			log.Println(err)
		}
		if err := n.confirmBlock(b.OrderBlock); err != nil {
			log.Printf("node: confirm error: %s", err.Error())
		}
		// TODO save confirm blocks
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
			defer res.Body.Close()
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
			defer res.Body.Close()
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
			defer res.Body.Close()
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
		order, err := n.store.GetOrderHead(b.Counterparty, b.ID)
		if err != nil {
			return err
		}
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

func (n *Node) confirmBlock(b tradeblocks.Block) error {
	address := b.Address()
	previous, err := n.store.GetConfirmHead(n.address, address)
	if err != nil {
		if err == db.ErrNotFound {
			previous = nil
		} else {
			return err
		}
	}
	hash := b.Hash()
	cb := tradeblocks.NewConfirmBlock(previous, n.address, address, hash)
	if err := cb.SignBlock(n.priv); err != nil {
		return err
	}
	if err := n.store.AddConfirmBlock(cb); err != nil {
		return fmt.Errorf("error adding confirm: %s", err.Error())
	}
	n.server.BlockHandler(app.TypedBlock{
		ConfirmBlock: cb,
		T:            "confirm",
	})
	return nil
}

func (n *Node) addPeer(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// TODO do connection check before adding peer
	n.peers[address] = struct{}{}
}

// Sync flushes all unbroadcasted blocks to known peers
func (n *Node) Sync() error {
	return nil
}

func blocksDir(dir string) string {
	return filepath.Join(dir, "blocks")
}
