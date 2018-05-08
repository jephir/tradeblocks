package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"
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
	address, err := publicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	return tradeblocks.NewIssueBlock(address, balance), nil
}

func publicKeyToAddress(publicKey io.Reader) (address string, err error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, publicKey); err != nil {
		return "", err
	}
	return addressPrefix + base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil

}

func Send(toAccount, token string, amount float64) error {
	fmt.Printf("your input for send is %s %s %v \n", toAccount, token, amount)
	// creates a new BaseTransaction with action = 'send'
	issue, err := tradeblocks.NewSend(toAccount, token, amount)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func Open(sendTx string) error {
	fmt.Printf("your input for open is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'open'
	issue, err := tradeblocks.NewOpen(sendTx)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func Receive(sendTx string) error {
	fmt.Printf("your input for receive is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'receive'
	issue, err := tradeblocks.NewReceive(sendTx)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}
