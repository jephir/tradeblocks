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
	mux *http.ServeMux

	mu         sync.Mutex
	blockstore *app.BlockStore
}

// NewServer allocates and returns a new server
func NewServer(blockstore *app.BlockStore) *Server {
	s := &Server{
		blockstore: blockstore,
		mux:        http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/account", s.handleAccountBlock())
}

func (s *Server) handleAccountBlock() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			hash := r.FormValue("hash")
			block, err := s.getBlock(hash)
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
			if err := s.addBlock(&b); err != nil {
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

func (s *Server) getBlock(hash string) (*tradeblocks.AccountBlock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blockstore.GetBlock(hash)
}

func (s *Server) addBlock(block *tradeblocks.AccountBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blockstore.AddBlock(block)
}
