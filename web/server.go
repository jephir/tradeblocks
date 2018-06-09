package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

// Server implements a TradeBlocks node
type Server struct {
	mux         *http.ServeMux
	blockStream *sse
	store       *app.BlockStore2
}

// NewServer allocates and returns a new server
func NewServer(blockstore *app.BlockStore2) *Server {
	s := &Server{
		mux:         http.NewServeMux(),
		blockStream: newSSE(),
		store:       blockstore,
	}
	s.routes()
	s.blockStream.connectHandler = s.blockEvents
	return s
}

func (s *Server) blockEvents() []event {
	var result []event
	s.store.Blocks(func(sequence int, b tradeblocks.Block) bool {
		hash := b.Hash()
		ss, err := blockEvent(hash, b)
		if err != nil {
			log.Println(err)
		}
		result = append(result, ss)
		return true
	})
	return result
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/block", s.handleBlock())
	s.mux.HandleFunc("/blocks", s.handleBlocks())
	s.mux.HandleFunc("/head", s.handleHead())
}

func (s *Server) handleBlock() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			hash := r.FormValue("hash")
			block := s.store.Block(hash)
			if block == nil {
				serverError(w, "no block found with hash '"+hash+"'", http.StatusBadRequest)
				return
			}
			if err := json.NewEncoder(w).Encode(block); err != nil {
				serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "POST":
			t := r.FormValue("type")
			if t == "" {
				serverError(w, "missing query param 'type'", http.StatusBadRequest)
				return
			}
			switch t {
			case "account":
				var b tradeblocks.AccountBlock
				if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
					serverError(w, "error decoding block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := s.store.AddAccountBlock(&b); err != nil {
					serverError(w, "can't add block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := json.NewEncoder(w).Encode(b); err != nil {
					serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
					return
				}
			case "swap":
				var b tradeblocks.SwapBlock
				if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
					serverError(w, "error decoding block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := s.store.AddSwapBlock(&b); err != nil {
					serverError(w, "can't add block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := json.NewEncoder(w).Encode(b); err != nil {
					serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
					return
				}
			case "order":
				var b tradeblocks.OrderBlock
				if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
					serverError(w, "error decoding block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := s.store.AddOrderBlock(&b); err != nil {
					serverError(w, "can't add block: "+err.Error(), http.StatusBadRequest)
					return
				}
				if err := json.NewEncoder(w).Encode(b); err != nil {
					serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
					return
				}
			default:
				serverError(w, "invalid query type '"+t+"'", http.StatusBadRequest)
			}
		default:
			serverError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) handleHead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			serverError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		t := r.FormValue("type")
		if t == "" {
			serverError(w, "missing query param 'type'", http.StatusBadRequest)
			return
		}
		switch t {
		case "account":
			account := r.FormValue("account")
			token := r.FormValue("token")
			block := s.store.GetAccountHead(account, token)
			if block == nil {
				serverError(w, "no head found for account '"+account+"' and token '"+token+"'", http.StatusBadRequest)
			}
			if err := json.NewEncoder(w).Encode(block); err != nil {
				serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "swap":
			account := r.FormValue("account")
			id := r.FormValue("id")
			block := s.store.GetSwapHead(account, id)
			if block == nil {
				serverError(w, "no head found for account '"+account+"' and id '"+id+"'", http.StatusBadRequest)
			}
			if err := json.NewEncoder(w).Encode(block); err != nil {
				serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "order":
			account := r.FormValue("account")
			id := r.FormValue("id")
			block := s.store.GetOrderHead(account, id)
			if block == nil {
				serverError(w, "no head found for account '"+account+"' and id '"+id+"'", http.StatusBadRequest)
			}
			if err := json.NewEncoder(w).Encode(block); err != nil {
				serverError(w, "error encoding block: "+err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			serverError(w, "invalid query type '"+t+"'", http.StatusBadRequest)
		}
	}
}

func (s *Server) handleBlocks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("stream") != "" {
			s.blockStream.ServeHTTP(w, r)
			return
		}
		switch r.Method {
		case "GET":
			t := r.FormValue("type")
			switch t {
			// case "":
			// 	var blocks []tradeblocks.NetworkBlock
			// 	s.store.AccountBlocks(func(hash string, b *tradeblocks.AccountBlock) bool {
			// 		block := tradeblocks.NetworkBlock{
			// 			Block: b,
			// 			Type:  "account",
			// 		}
			// 		blocks = append(blocks, block)
			// 		return true
			// 	})
			// 	s.store.SwapBlocks(func(hash string, b *tradeblocks.SwapBlock) bool {
			// 		block := tradeblocks.NetworkBlock{
			// 			Block: b,
			// 			Type:  "swap",
			// 		}
			// 		blocks = append(blocks, block)
			// 		return true
			// 	})
			// 	s.store.OrderBlocks(func(hash string, b *tradeblocks.OrderBlock) bool {
			// 		block := tradeblocks.NetworkBlock{
			// 			Block: b,
			// 			Type:  "order",
			// 		}
			// 		blocks = append(blocks, block)
			// 		return true
			// 	})
			// 	sort.Slice(blocks, func(i, j int) bool {
			// 		a := blocks[i].Hash()
			// 		b := blocks[j].Hash()
			// 		return s.store.SequenceLess(a, b)
			// 	})
			// 	if err := json.NewEncoder(w).Encode(blocks); err != nil {
			// 		serverError(w, "error encoding blocks: "+err.Error(), http.StatusInternalServerError)
			// 		return
			// 	}
			case "account":
				result := make(map[string]tradeblocks.NetworkAccountBlock)
				s.store.AccountBlocks(func(sequence int, b *tradeblocks.AccountBlock) bool {
					hash := b.Hash()
					result[hash] = tradeblocks.NetworkAccountBlock{
						AccountBlock: b,
						Sequence:     sequence,
					}
					return true
				})
				if err := json.NewEncoder(w).Encode(result); err != nil {
					serverError(w, "error encoding blocks: "+err.Error(), http.StatusInternalServerError)
					return
				}
			case "swap":
				result := make(map[string]tradeblocks.NetworkSwapBlock)
				s.store.SwapBlocks(func(sequence int, b *tradeblocks.SwapBlock) bool {
					hash := b.Hash()
					result[hash] = tradeblocks.NetworkSwapBlock{
						SwapBlock: b,
						Sequence:  sequence,
					}
					return true
				})
				if err := json.NewEncoder(w).Encode(result); err != nil {
					serverError(w, "error encoding blocks: "+err.Error(), http.StatusInternalServerError)
					return
				}
			case "order":
				result := make(map[string]tradeblocks.NetworkOrderBlock)
				s.store.OrderBlocks(func(sequence int, b *tradeblocks.OrderBlock) bool {
					hash := b.Hash()
					result[hash] = tradeblocks.NetworkOrderBlock{
						OrderBlock: b,
						Sequence:   sequence,
					}
					return true
				})
				if err := json.NewEncoder(w).Encode(result); err != nil {
					serverError(w, "error encoding blocks: "+err.Error(), http.StatusInternalServerError)
					return
				}
			default:
				serverError(w, "invalid query type '"+t+"'", http.StatusBadRequest)
			}

		default:
			serverError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

// BroadcastBlock broadcasts the specified block to all event listeners
func (s *Server) BroadcastBlock(b tradeblocks.Block) error {
	e, err := blockEvent(b.Hash(), b)
	if err != nil {
		return err
	}
	s.blockStream.broadcast <- e
	return nil
}

func blockEvent(hash string, b tradeblocks.Block) (event, error) {
	var res struct {
		tradeblocks.Block
		Hash string
	}
	res.Block = b
	res.Hash = hash
	return json.Marshal(res)
}

func serverError(w http.ResponseWriter, error string, code int) {
	log.Printf("web: %s", error)
	http.Error(w, error, code)
}
