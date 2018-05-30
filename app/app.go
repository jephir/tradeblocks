package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"

	"github.com/jephir/tradeblocks"
)

const addressPrefix = "xtb:"

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

// Open creates a new account blockchain
func Open(publicKey io.Reader, send *tradeblocks.AccountBlock, balance float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}

	// check if the send block is already claimed
	// sendhash := send.Hash()
	// if AlreadyLinked(sendhash) {
	// 	return nil, errors.New("Block on chain already claimed the given send")
	// }

	return tradeblocks.NewOpenBlock(address, send, balance), nil
}

// Receive receives tokens from a send transaction
func Receive(publicKey io.Reader, previous *tradeblocks.AccountBlock, send *tradeblocks.AccountBlock, amount float64) (*tradeblocks.AccountBlock, error) {
	// check if the Link field is already claimed
	sendHash := send.Hash()
	// if AlreadyLinked(sendHash) {
	// 	return nil, errors.New("Block on chain already claimed the given send")
	// }

	return tradeblocks.NewReceiveBlock(previous, sendHash, amount), nil
}

// PublicKeyToAddress returns the string serialization of the specified public key
func PublicKeyToAddress(publicKey io.Reader) (address string, err error) {
	buf, err := ioutil.ReadAll(publicKey)
	if err != nil {
		return "", err
	}
	return addressPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

// AddressToPublicKey returns the string serialization of the specified public key
// do the opposite of what the rawurlencoding.DECODE? on the publicKey
func AddressToPublicKey(address string) (publicKey string, err error) {
	byteKey, err := base64.RawURLEncoding.DecodeString(address)
	publicKey = string(byteKey[:])
	return addressPrefix + publicKey, err
}

// AccountBlockHash returns the hash of the specified account block
func AccountBlockHash(block *tradeblocks.AccountBlock) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil)), nil
}

// SerializeAccountBlock returns the string representation of the specified account block
func SerializeAccountBlock(block *tradeblocks.AccountBlock) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AlreadyLinked checks if there is a block that claims the given hash in Link field
// Works for: Open, Receive
// Does not work for: Send (link is to account)
// For Swaps, if origination swap, check if there's another swap with Left equal to hash
// If counterswap, check if swap with Right equal to hash
// func AlreadyLinked(hash string) bool {
// 	return false
// }
