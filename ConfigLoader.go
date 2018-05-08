package dexathon

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
)

// LoadAccount : loads the hash used for account in transactions
// TODO
func LoadAccount() (string, error) {
	return readAccount()
}

// LoadBalance : loads the account balance from local, currently stub
// TODO
func LoadBalance() (balance float64) {
	balance = -1.0
	return
}

// LoadPrevious : loads the previous transaction hash, currently stub
// TODO
func LoadPrevious() (previous string) {
	previous = "previous placeholder"
	return
}

// LoadRepresentative : loads the current representative
// TODO
func LoadRepresentative() (representative string) {
	representative = "representative placeholder"
	return
}

// LoadToken : loads the token? Talk to julian
// TODO
func LoadToken() (token string) {
	token = "token placeholder"
	return
}

func readAccount() (string, error) {
	user, err := ioutil.ReadFile("user")
	if err != nil {
		return "", err
	}
	f, err := os.Open(string(user) + ".pub")
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func readPrivateKey() (*rsa.PrivateKey, error) {
	user, err := ioutil.ReadFile("user")
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile(string(user) + ".pem")
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(bytes)
}
