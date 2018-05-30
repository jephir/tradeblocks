package tradeblocks

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/json"
	"io"
	"io/ioutil"
)

// AccountBlock represents a block in the account blockchain
type AccountBlock struct {
	Action         string
	Account        string
	Token          string
	Previous       string
	Representative string
	Balance        float64
	Link           string
	Signature      string
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

// SignBlock signs the block, returns just the error
func (ab *AccountBlock) SignBlock(privateKey io.Reader) error {
	rng := rand.Reader
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	keyBytes, err := ioutil.ReadAll(privateKey)
	if err != nil {
		return err
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return err
	}

	signature, err := rsa.SignPKCS1v15(rng, rsaPrivateKey, crypto.SHA256, hashedBytes[:])
	if err != nil {
		return err
	}

	ab.Signature = string(signature[:])
	return nil
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
		Signature:      "",
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
		Signature:      "",
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
		Signature:      "",
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
		Signature:      "",
	}
}

// SwapBlock represents a block in the swap blockchain
type SwapBlock struct {
	Action       string
	Account      string
	Token        string
	ID           string
	Previous     string
	Left         string
	Right        string
	RefundLeft   string
	RefundRight  string
	Counterparty string
	Want         string
	Quantity     float64
	executor     string
	Fee          float64
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
