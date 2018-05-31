package tradeblocks

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"io"
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
	withoutSig := &AccountBlock{
		Action:         ab.Action,
		Account:        ab.Account,
		Token:          ab.Token,
		Previous:       ab.Previous,
		Representative: ab.Representative,
		Balance:        ab.Balance,
		Link:           ab.Link,
		Signature:      "",
	}
	b, err := json.Marshal(withoutSig)
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
func (ab *AccountBlock) SignBlock(priv *rsa.PrivateKey) error {
	rng := rand.Reader
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	signature, err := rsa.SignPKCS1v15(rng, priv, crypto.SHA256, hashedBytes[:])
	if err != nil {
		return err
	}

	ab.Signature = base64.StdEncoding.EncodeToString(signature)
	return nil
}

// VerifyBlock signs the block, returns just the error
func (ab *AccountBlock) VerifyBlock(pubKey *rsa.PublicKey) error {
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	decodedSig, err := base64.StdEncoding.DecodeString(ab.Signature)
	if err != nil {
		return err
	}

	errVerify := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashedBytes[:], decodedSig)
	if errVerify != nil {
		return errVerify
	}
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
	Executor     string
	Fee          float64
	Signature    string
}

// Hash returns the hash of this block
func (ab *SwapBlock) Hash() string {
	withoutSig := &SwapBlock{
		Action:       ab.Action,
		Account:      ab.Account,
		Token:        ab.Token,
		ID:           ab.ID,
		Previous:     ab.Previous,
		Left:         ab.Left,
		Right:        ab.Right,
		RefundLeft:   ab.RefundLeft,
		RefundRight:  ab.RefundRight,
		Counterparty: ab.Counterparty,
		Want:         ab.Want,
		Quantity:     ab.Quantity,
		Executor:     ab.Executor,
		Fee:          ab.Fee,
		Signature:    "",
	}
	b, err := json.Marshal(withoutSig)
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
func (ab *SwapBlock) SignBlock(priv *rsa.PrivateKey) error {
	rng := rand.Reader
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	signature, err := rsa.SignPKCS1v15(rng, priv, crypto.SHA256, hashedBytes[:])
	if err != nil {
		return err
	}

	ab.Signature = base64.StdEncoding.EncodeToString(signature)
	return nil
}

// VerifyBlock signs the block, returns just the error
func (ab *SwapBlock) VerifyBlock(pubKey *rsa.PublicKey) error {
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	decodedSig, err := base64.StdEncoding.DecodeString(ab.Signature)
	if err != nil {
		return err
	}

	errVerify := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, decoded[:], decodedSig)
	if errVerify != nil {
		return errVerify
	}
	return nil
}

// OrderBlock represents a block in the order blockchain
type OrderBlock struct {
	Action    string
	Account   string
	Token     string
	ID        string
	Previous  string
	Balance   float64
	Quote     string
	Price     float64
	Link      string
	Partial   bool
	Executor  string
	Signature string
}

// Hash returns the hash of this block
func (ab *OrderBlock) Hash() string {
	withoutSig := &OrderBlock{
		Action:    ab.Action,
		Account:   ab.Account,
		Token:     ab.Token,
		ID:        ab.ID,
		Previous:  ab.Previous,
		Balance:   ab.Balance,
		Quote:     ab.Quote,
		Price:     ab.Price,
		Link:      ab.Link,
		Partial:   ab.Partial,
		Executor:  ab.Executor,
		Signature: "",
	}
	b, err := json.Marshal(withoutSig)
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
func (ab *OrderBlock) SignBlock(priv *rsa.PrivateKey) error {
	rng := rand.Reader
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	signature, err := rsa.SignPKCS1v15(rng, priv, crypto.SHA256, hashedBytes[:])
	if err != nil {
		return err
	}

	ab.Signature = base64.StdEncoding.EncodeToString(signature)
	return nil
}

// VerifyBlock signs the block, returns just the error
func (ab *OrderBlock) VerifyBlock(pubKey *rsa.PublicKey) error {
	hashed := ab.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hashed)
	if err != nil {
		return err
	}

	hashedBytes := []byte(decoded)

	decodedSig, err := base64.StdEncoding.DecodeString(ab.Signature)
	if err != nil {
		return err
	}

	errVerify := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashedBytes[:], decodedSig)
	if errVerify != nil {
		return errVerify
	}
	return nil
}
