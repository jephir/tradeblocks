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
	"fmt"
	"io"
	"strings"
)

// Block represents any block type
type Block interface {
	Hash() string
	SignBlock(*rsa.PrivateKey) error
}

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

// Normalize trims all whitespace in the block
func (ab *AccountBlock) Normalize() {
	ab.Action = strings.TrimSpace(ab.Action)
	ab.Account = strings.TrimSpace(ab.Account)
	ab.Token = strings.TrimSpace(ab.Token)
	ab.Previous = strings.TrimSpace(ab.Previous)
	ab.Representative = strings.TrimSpace(ab.Representative)
	ab.Link = strings.TrimSpace(ab.Link)
	ab.Signature = strings.TrimSpace(ab.Signature)
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

// Equals returns an error if this block doesn't equal the specified block
func (ab *AccountBlock) Equals(o *AccountBlock) error {
	if o.Action != ab.Action {
		return fmt.Errorf("blockgraph: action '%s' doesn't equal '%s'", o.Action, ab.Action)
	}
	if o.Account != ab.Account {
		return fmt.Errorf("blockgraph: account '%s' doesn't equal '%s'", o.Account, ab.Account)
	}
	if o.Token != ab.Token {
		return fmt.Errorf("blockgraph: token '%s' doesn't equal '%s'", o.Token, ab.Token)
	}
	if o.Previous != ab.Previous {
		return fmt.Errorf("blockgraph: previous '%s' doesn't equal '%s'", o.Previous, ab.Previous)
	}
	if o.Representative != ab.Representative {
		return fmt.Errorf("blockgraph: representative '%s' doesn't equal '%s'", o.Representative, ab.Representative)
	}
	if o.Balance != ab.Balance {
		return fmt.Errorf("blockgraph: balance '%f' doesn't equal '%f'", o.Balance, ab.Balance)
	}
	if o.Link != ab.Link {
		return fmt.Errorf("blockgraph: link '%s' doesn't equal '%s'", o.Link, ab.Link)
	}
	if o.Signature != ab.Signature {
		return fmt.Errorf("blockgraph: signature '%s' doesn't equal '%s'", o.Signature, ab.Signature)
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
func NewOpenBlock(account string, send interface{}, balance float64) (openBlock *AccountBlock) {
	switch sendTyped := send.(type) {
	case AccountBlock:
		openBlock = &AccountBlock{
			Action:         "open",
			Account:        account,
			Token:          sendTyped.Token,
			Previous:       "",
			Representative: "",
			Balance:        balance,
			Link:           sendTyped.Hash(),
			Signature:      "",
		}
	case SwapBlock:
		if sendTyped.Action == "commit" {
			openBlock = &AccountBlock{
				Action:         "open",
				Account:        account,
				Token:          sendTyped.Token,
				Previous:       "",
				Representative: "",
				Balance:        balance,
				Link:           sendTyped.Hash(),
				Signature:      "",
			}
		} else {
			openBlock = nil
		}

	default:
		openBlock = nil
	}
	return
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
func NewReceiveBlock(previous *AccountBlock, send interface{}, amount float64) (receiveBlock *AccountBlock) {
	switch sendTyped := send.(type) {
	case AccountBlock:
		receiveBlock = &AccountBlock{
			Action:         "receive",
			Account:        previous.Account,
			Token:          previous.Token,
			Previous:       previous.Hash(),
			Representative: previous.Representative,
			Balance:        previous.Balance + amount,
			Link:           sendTyped.Hash(),
			Signature:      "",
		}
	case SwapBlock:
		if sendTyped.Action == "commit" {
			receiveBlock = &AccountBlock{
				Action:         "receive",
				Account:        previous.Account,
				Token:          previous.Token,
				Previous:       previous.Hash(),
				Representative: previous.Representative,
				Balance:        previous.Balance + amount,
				Link:           sendTyped.Hash(),
				Signature:      "",
			}
		} else {
			receiveBlock = nil
		}

	default:
		receiveBlock = nil
	}
	return
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

// NewOfferBlock is the originating swap
func NewOfferBlock(account string, send *AccountBlock, ID string, counterparty string, want string, quantity float64, executor string, fee float64) *SwapBlock {
	return &SwapBlock{
		Action:       "offer",
		Account:      account,
		Token:        send.Token,
		ID:           ID,
		Previous:     "",
		Left:         send.Hash(),
		Right:        "",
		RefundLeft:   "",
		RefundRight:  "",
		Counterparty: counterparty,
		Want:         want,
		Quantity:     quantity,
		Executor:     executor,
		Fee:          fee,
		Signature:    "",
	}
}

// NewCommitBlock is the committing swap
func NewCommitBlock(send *AccountBlock, previous *SwapBlock) *SwapBlock {
	return &SwapBlock{
		Action:       "commit",
		Account:      previous.Account,
		Token:        previous.Token,
		ID:           previous.ID,
		Previous:     previous.Hash(),
		Left:         previous.Left,
		Right:        send.Hash(),
		RefundLeft:   "",
		RefundRight:  "",
		Counterparty: previous.Counterparty,
		Want:         previous.Want,
		Quantity:     previous.Quantity,
		Executor:     previous.Executor,
		Fee:          previous.Fee,
		Signature:    "",
	}
}

// NewRefundLeftBlock refunds the left. Previous should be the offer block
func NewRefundLeftBlock(previous *SwapBlock, refundTo string) *SwapBlock {
	return &SwapBlock{
		Action:       "refund-left",
		Account:      previous.Account,
		Token:        previous.Token,
		ID:           previous.ID,
		Previous:     previous.Hash(),
		Left:         previous.Left,
		Right:        "",
		RefundLeft:   refundTo,
		RefundRight:  previous.RefundRight,
		Counterparty: previous.Counterparty,
		Want:         previous.Want,
		Quantity:     previous.Quantity,
		Executor:     previous.Executor,
		Fee:          previous.Fee,
		Signature:    "",
	}
}

// NewRefundRightBlock refunds the left. Previous should be the commit block
func NewRefundRightBlock(refundLeft *SwapBlock, counterSend *AccountBlock, refundTo string) *SwapBlock {
	return &SwapBlock{
		Action:       "refund-right",
		Account:      refundLeft.Account,
		Token:        refundLeft.Token,
		ID:           refundLeft.ID,
		Previous:     refundLeft.Hash(),
		Left:         refundLeft.Left,
		Right:        counterSend.Hash(),
		RefundLeft:   refundLeft.Account,
		RefundRight:  refundTo,
		Counterparty: refundLeft.Counterparty,
		Want:         refundLeft.Want,
		Quantity:     refundLeft.Quantity,
		Executor:     refundLeft.Executor,
		Fee:          refundLeft.Fee,
		Signature:    "",
	}
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

// SignedAccountBlock returns a signed version of the specified block with the specified private key
func SignedAccountBlock(b *AccountBlock, priv *rsa.PrivateKey) (*AccountBlock, error) {
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	return b, nil
}

// SignedSwapBlock returns a signed version of the specified block with the specified private key
func SignedSwapBlock(b *SwapBlock, priv *rsa.PrivateKey) (*SwapBlock, error) {
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	return b, nil
}

// SignedOrderBlock returns a signed version of the specified block with the specified private key
func SignedOrderBlock(b *OrderBlock, priv *rsa.PrivateKey) (*OrderBlock, error) {
	if err := b.SignBlock(priv); err != nil {
		return nil, err
	}
	return b, nil
}
