package tradeblocks

import "encoding/json"
import "crypto/sha256"
import "io"
import "bytes"
import "encoding/base32"

// AccountBlock represents a block in the account blockchain
type AccountBlock struct {
	Action         string
	Account        string
	Token          string
	Previous       string
	Representative string
	Balance        float64
	Link           string
}

// Hash returns the hash of this block
func (ab *AccountBlock) Hash() string {
	b, err := json.Marshal(ab)
	if err != nil {
		panic(err)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		panic(err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil))

}

// NewIssueBlock initializes a new crypto token
func NewIssueBlock(account string, balance float64) *AccountBlock {
	return &AccountBlock{
		Action:         "issue",
		Account:        account,
		Token:          account,
		Previous:       "",
		Representative: "",
		Balance:        balance,
		Link:           "",
	}
}

// NewOpenBlock initializes the start of an account blockchain
func NewOpenBlock(account string, send *AccountBlock, balance float64) *AccountBlock {
	return &AccountBlock{
		Action:         "open",
		Account:        account,
		Token:          send.Token,
		Previous:       "",
		Representative: "",
		Balance:        balance,
		Link:           send.Hash(),
	}
}

// NewSendBlock initializes a send to the specified address
func NewSendBlock(previous *AccountBlock, to string, amount float64) *AccountBlock {
	return &AccountBlock{
		Action:         "send",
		Account:        previous.Account,
		Token:          previous.Token,
		Previous:       previous.Hash(),
		Representative: previous.Representative,
		Balance:        previous.Balance - amount,
		Link:           to,
	}
}

// NewReceiveBlock initializes a receive of tokens
func NewReceiveBlock(previous *AccountBlock, send *AccountBlock, amount float64) *AccountBlock {
	return &AccountBlock{
		Action:         "receive",
		Account:        previous.Account,
		Token:          previous.Token,
		Previous:       previous.Hash(),
		Representative: previous.Representative,
		Balance:        previous.Balance + amount,
		Link:           send.Hash(),
	}
}

// SwapBlock represents a block in the swap blockchain
type SwapBlock struct {
	Action       string
	Account      string
	ID           string
	Previous     string
	Left         string
	Right        string
	Counterparty string
	Want         string
	Quantity     float64
}

// OrderBlock represents a block in the order blockchain
type OrderBlock struct {
	Action   string
	Account  string
	Token    string
	ID       string
	Previous string
	Balance  float64
	Quote    string
	Price    float64
	Link     string
	Partial  bool
}
