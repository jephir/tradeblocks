package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/jephir/tradeblocks"
)

const addressPrefix = "xtb:"

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// Register creates a new key pair with the specified local name
func Register(privateKey io.Writer, publicKey io.Writer, name string, keySize int) (address string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", err
	}
	if err := pem.Encode(privateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return "", err
	}
	b, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", err
	}
	p := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: b,
	})
	address, err = PublicKeyToAddress(bytes.NewReader(p))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(publicKey, bytes.NewReader(p)); err != nil {
		return "", err
	}
	return
}

// Issue creates a new crypto coin with the specified balance
func Issue(publicKey io.Reader, balance float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewIssueBlock(address, balance), nil
}

// Send transfers tokens to the specified account
func Send(publicKey io.Reader, previous *tradeblocks.AccountBlock, to string, amount float64) (*tradeblocks.AccountBlock, error) {
	return tradeblocks.NewSendBlock(previous, to, amount), nil
}

// OpenFromSend creates a new account blockchain from a send
func OpenFromSend(publicKey io.Reader, send *tradeblocks.AccountBlock, balance float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}

	return tradeblocks.NewOpenBlockFromSend(address, send, balance), nil
}

// OpenFromSwap creates a new account blockchain from a swap
func OpenFromSwap(publicKey io.Reader, token string, swap *tradeblocks.SwapBlock, balance float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}

	return tradeblocks.NewOpenBlockFromSwap(address, token, swap, balance), nil
}

// Receive receives tokens from a send transaction
func Receive(publicKey io.Reader, previous *tradeblocks.AccountBlock, send *tradeblocks.AccountBlock, amount float64) (*tradeblocks.AccountBlock, error) {
	return tradeblocks.NewReceiveBlockFromSend(previous, send, amount), nil
}

//Offer creates an offer for a swap
func Offer(publicKey io.Reader, send *tradeblocks.AccountBlock, ID string, counterparty string, want string, quantity float64, executor string, fee float64) (*tradeblocks.SwapBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}

	return tradeblocks.NewOfferBlock(address, send, ID, counterparty, want, quantity, executor, fee), nil
}

//Commit creates a commit for a swap
func Commit(publicKey io.Reader, offer *tradeblocks.SwapBlock, right *tradeblocks.AccountBlock) (*tradeblocks.SwapBlock, error) {
	return tradeblocks.NewCommitBlock(offer, right), nil
}

//RefundLeft creates a refund-left for a swap
func RefundLeft(publicKey io.Reader, offer *tradeblocks.SwapBlock, refundTo string) (*tradeblocks.SwapBlock, error) {
	return tradeblocks.NewRefundLeftBlock(offer, refundTo), nil
}

//RefundRight creates a refund-left for a swap
func RefundRight(publicKey io.Reader, refundLeft *tradeblocks.SwapBlock, counterSend *tradeblocks.AccountBlock, refundTo string) (*tradeblocks.SwapBlock, error) {
	return tradeblocks.NewRefundRightBlock(refundLeft, counterSend, refundTo), nil
}

//CreateOrder creates an order
func CreateOrder(publicKey io.Reader, send *tradeblocks.AccountBlock, balance float64, ID string, partial bool, quote string, price float64, executor string, fee float64) (*tradeblocks.OrderBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}

	return tradeblocks.NewCreateOrderBlock(address, send, balance, ID, partial, quote, price, executor, fee), nil
}

//AcceptOrder creates an accept order for an order
func AcceptOrder(publicKey io.Reader, previous *tradeblocks.OrderBlock, link string, balance float64) (*tradeblocks.OrderBlock, error) {
	return tradeblocks.NewAcceptOrderBlock(previous, link, balance), nil
}

//RefundOrder creates an accept order for an order
func RefundOrder(publicKey io.Reader, previous *tradeblocks.OrderBlock, refundTo string) (*tradeblocks.OrderBlock, error) {
	return tradeblocks.NewRefundOrderBlock(previous, refundTo), nil
}

// PublicKeyToAddress returns the string serialization of the specified public key
func PublicKeyToAddress(publicKey io.Reader) (address string, err error) {
	buf, err := ioutil.ReadAll(publicKey)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(buf)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return PublicKeyRSAToAddress(pub)
	default:
		break // fall through
	}
	return "", fmt.Errorf("app: key type is not RSA")
}

// ErrInvalidAddress is returned when the address is not in the correct format
var ErrInvalidAddress = errors.New("app: invalid address")

// AddressToPublicKey decodes the specified address into a public key
func AddressToPublicKey(address string) (*rsa.PublicKey, error) {
	addr := strings.TrimPrefix(address, addressPrefix)
	b, err := encoding.DecodeString(addr)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	buf := bytes.NewBuffer(b)
	pub := new(rsa.PublicKey)
	var e int32
	if err := binary.Read(buf, binary.BigEndian, &e); err != nil {
		return nil, ErrInvalidAddress
	}
	pub.E = int(e)
	pub.N = big.NewInt(0)
	pub.N.SetBytes(buf.Bytes())
	return pub, nil
}

// PrivateKeyToAddress returns the string serialization of the specified private key
func PrivateKeyToAddress(priv *rsa.PrivateKey) (string, error) {
	return PublicKeyRSAToAddress(&priv.PublicKey)
}

// PublicKeyRSAToAddress returns the string serialization of the specified public key
func PublicKeyRSAToAddress(pub *rsa.PublicKey) (string, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int32(pub.E)); err != nil {
		return "", err
	}
	if _, err := buf.Write(pub.N.Bytes()); err != nil {
		return "", err
	}
	return addressPrefix + encoding.EncodeToString(buf.Bytes()), nil
}

// SerializeAccountBlock returns the string representation of the specified account block
func SerializeAccountBlock(block *tradeblocks.AccountBlock) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UniqueID returns a unique ID for creating a swap or order blockchain
func UniqueID() string {
	// https://stackoverflow.com/a/25736155
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
