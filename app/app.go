package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"

	"github.com/jephir/tradeblocks"
)

const addressPrefix = "xtb:"

// Register creates a new key pair with the specified local name
func Register(privateKey io.Writer, publicKey io.Writer, name string, keySize int) error {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return err
	}
	if err := pem.Encode(privateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return err
	}
	bytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	if err := pem.Encode(publicKey, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	}); err != nil {
		return err
	}
	return nil
}

// Issue creates a new crypto coin with the specified balance
func Issue(publicKey io.Reader, balance float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewIssueBlock(address, balance), nil
}

// PublicKeyToAddress returns the string serialization of the specified public key
func PublicKeyToAddress(publicKey io.Reader) (address string, err error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, publicKey); err != nil {
		return "", err
	}
	return addressPrefix + base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// Send transfers tokens to the specified account
func Send(publicKey io.Reader, previous *tradeblocks.AccountBlock, to string, amount float64) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewSendBlock(address, previous, to, amount), nil
}

// Open creates a new account blockchain
func Open(publicKey io.Reader, send *tradeblocks.AccountBlock) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewOpenBlock(address, send), nil
}

// Receive receives tokens from a send transaction
func Receive(publicKey io.Reader, previous *tradeblocks.AccountBlock, send *tradeblocks.AccountBlock) (*tradeblocks.AccountBlock, error) {
	address, err := PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewReceiveBlock(address, previous, send), nil
}
