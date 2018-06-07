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
	"strings"

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

	return tradeblocks.NewOpenBlock(address, send, balance), nil
}

// Receive receives tokens from a send transaction
func Receive(publicKey io.Reader, previous *tradeblocks.AccountBlock, send *tradeblocks.AccountBlock, amount float64) (*tradeblocks.AccountBlock, error) {
	return tradeblocks.NewReceiveBlock(previous, send, amount), nil
}

// PublicKeyToAddress returns the string serialization of the specified public key
func PublicKeyToAddress(publicKey io.Reader) (address string, err error) {
	buf, err := ioutil.ReadAll(publicKey)
	if err != nil {
		return "", err
	}
	return addressPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

// PrivateKeyToAddress returns the string serialization of the specified private key
func PrivateKeyToAddress(priv *rsa.PrivateKey) (address string, err error) {
	b, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return
	}
	r := bytes.NewReader(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: b,
	}))
	return PublicKeyToAddress(r)
}

// AddressToPublicKey returns the byte array of the specified address
func AddressToPublicKey(address string) (publicKey []byte, err error) {
	addressNoPrefix := strings.TrimPrefix(address, addressPrefix)
	byteKey, err := base64.RawURLEncoding.DecodeString(addressNoPrefix)
	return byteKey, err
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
