package web

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
	"net/http"
	"sync"
)

// Server implements a TradeBlocks node
type Server struct {
	mux     *http.ServeMux
	service *service
}

// NewServer allocates and returns a new server
func NewServer(blockstore *app.BlockStore) *Server {
	s := &Server{
		service: &service{
			blockstore: blockstore,
		},
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/account", s.handleAccountBlock())
	s.mux.HandleFunc("/blocks", s.handleBlocks())
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
			if err := s.service.addBlock(&b); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := json.NewEncoder(w).Encode(b); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
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

func (s *service) addBlock(block *tradeblocks.AccountBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blockstore.AddBlock(block)
}
