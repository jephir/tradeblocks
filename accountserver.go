package dexathon

import (
	"encoding/json"
	"net/http"
)

// AccountServer stores and handles updates for account blockchains.
type AccountServer struct {
	accounts map[string][]*AccountBlock // Keyed by Account field
}

// NewAccountServer initializes a new account server.
func NewAccountServer() *AccountServer {
	return &AccountServer{
		accounts: make(map[string][]*AccountBlock),
	}
}

func (s *AccountServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id := r.FormValue("account")
		account, ok := s.accounts[id]
		if !ok {
			http.Error(w, "No account found.", http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(account); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "POST":
		var block AccountBlock
		if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.accounts[block.Account] = append(s.accounts[block.Account], &block)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
