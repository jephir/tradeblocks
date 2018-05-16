package main

import (
	"encoding/json"
	"github.com/jephir/tradeblocks"
	"log"
	"net/http"

	"github.com/jephir/tradeblocks/app"
)

func main() {
	srv := newServer()
	log.Fatal(http.ListenAndServe(":8080", srv))
}

type server struct {
	blockstore *app.BlockStore
	mux        *http.ServeMux
}

func newServer() *server {
	return &server{
		blockstore: app.NewBlockStore(),
		mux:        http.NewServeMux(),
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *server) routes() {
	s.mux.HandleFunc("/account", s.handleAccountBlock())
}

func (s *server) handleAccountBlock() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			hash := r.FormValue("hash")
			block := s.blockstore.GetBlock(hash)
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
			if err := s.blockstore.AddBlock(&b); err != nil {
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
