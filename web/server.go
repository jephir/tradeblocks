package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

// Server implements a TradeBlocks node
type Server struct {
	mux           *http.ServeMux
	accountStream *sse
	service       *service
}

// NewServer allocates and returns a new server
func NewServer(blockstore *app.BlockStore) *Server {
	s := &Server{
		mux:           http.NewServeMux(),
		accountStream: newSSE(),
		service: &service{
			blockstore: blockstore,
		},
	}
	s.routes()
	s.accountStream.connectHandler = func() []event {
		var result []event

		for r := range s.service.getBlocks() {
			ss, err := accountBlockEvent(r.block)
			if err != nil {
				log.Println(err)
			}
			result = append(result, ss)
		}
		return result
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.Handle("/accounts", s.accountStream)
	s.mux.HandleFunc("/account", s.handleAccountBlock())
	s.mux.HandleFunc("/head", s.handleAccountHead())
	s.mux.HandleFunc("/blocks", s.handleBlocks())
	s.mux.Handle("/ui", http.StripPrefix("/ui/", http.FileServer(http.Dir("web/public"))))
}

func (s *Server) handleAccountBlock() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			hash := r.FormValue("hash")
			block, err := s.service.getBlock(hash)
			if err != nil {
				http.Error(w, "Couldn't get block.", http.StatusInternalServerError)
				return
			}
			if block == nil {
				http.Error(w, "No block found.", http.StatusBadRequest)
				return
			}
			if err := json.NewEncoder(w).Encode(block); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "POST":
			var b tradeblocks.AccountBlock
			if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if _, err := s.service.addBlock(&b); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := json.NewEncoder(w).Encode(b); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// ss, _ := app.SerializeAccountBlock(&b)
			//log.Printf("web: added block %s: %s", b.Hash(), ss)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) handleAccountHead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account := r.FormValue("account")
		token := r.FormValue("token")
		block, err := s.service.getHeadBlock(account, token)
		if err != nil {
			http.Error(w, "Couldn't get block.", http.StatusInternalServerError)
			return
		}
		if block == nil {
			http.Error(w, "No block found.", http.StatusBadRequest)
		}
		if err := json.NewEncoder(w).Encode(block); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleBlocks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			result := make(app.AccountBlocksMap)
			for r := range s.service.getBlocks() {
				result[r.hash] = r.block
			}
			if err := json.NewEncoder(w).Encode(result); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

// BroadcastAccountBlock sends the specified account block to all event listeners
func (s *Server) BroadcastAccountBlock(b *tradeblocks.AccountBlock) error {
	e, err := accountBlockEvent(b)
	if err != nil {
		return err
	}
	s.accountStream.broadcast <- e
	return nil
}

func accountBlockEvent(b *tradeblocks.AccountBlock) (event, error) {
	var res struct {
		*tradeblocks.AccountBlock
		Hash string
	}
	res.AccountBlock = b
	res.Hash = b.Hash()
	return json.Marshal(res)
}

// service represents concurrency-safe resources that the HTTP handlers can access
type service struct {
	mu         sync.RWMutex
	blockstore *app.BlockStore
}

type hashAccountBlock struct {
	hash  string
	block *tradeblocks.AccountBlock
}

func (s *service) getBlocks() <-chan hashAccountBlock {
	// Use goroutine to hold open read mutex while returning blocks
	ch := make(chan hashAccountBlock)
	go func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		for hash, block := range s.blockstore.AccountBlocks {
			ch <- hashAccountBlock{
				hash:  hash,
				block: block,
			}
		}
		close(ch)
	}()
	return ch
}

func (s *service) getBlock(hash string) (*tradeblocks.AccountBlock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockstore.GetBlock(hash)
}

func (s *service) getHeadBlock(account, token string) (*tradeblocks.AccountBlock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blockstore.GetHeadBlock(account, token)
}

func (s *service) addBlock(block *tradeblocks.AccountBlock) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, err := s.blockstore.AddBlock(block)
	if _, ok := err.(*app.BlockConflictError); ok {
		var highest int
		var hash string
		for _, vote := range s.blockstore.VoteBlocks {
			if vote.Order > highest {
				highest = vote.Order
				hash = vote.Link
			}
		}
		if hash != block.Hash() {
			return "", fmt.Errorf("server: specified block '%s' is purged", block.Hash())
		}
	}
	return hash, err
}
