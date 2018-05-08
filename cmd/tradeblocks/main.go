package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jephir/dexathon"
)

const keySize = 4096

func register(name string) error {
	fmt.Printf("your input for register is %s \n", name)
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return err
	}
	if err := writePrivateKey(name, key); err != nil {
		return err
	}
	if err := writePublicKey(name, &key.PublicKey); err != nil {
		return err
	}
	return login(name)
}

func writePrivateKey(name string, key *rsa.PrivateKey) error {
	f, err := os.Create(name + ".pem")
	if err != nil {
		return err
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return err
	}
	return f.Close()
}

func writePublicKey(name string, key *rsa.PublicKey) error {
	f, err := os.Create(name + ".pub")
	if err != nil {
		return err
	}
	defer f.Close()
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}
	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	}); err != nil {
		return err
	}
	return f.Close()
}

func login(name string) error {
	fmt.Printf("your input for login is %s \n", name)
	return ioutil.WriteFile("user", []byte(name), 0644)
}

func issue(balance float64) error {
	fmt.Printf("your input for issue is %v \n", balance)
	// creates a new BaseTransaction with action = 'issue'
	issue, err := dexathon.NewIssue(balance)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func send(toAccount, token string, amount float64) error {
	fmt.Printf("your input for send is %s %s %v \n", toAccount, token, amount)
	// creates a new BaseTransaction with action = 'send'
	issue, err := dexathon.NewSend(toAccount, token, amount)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func open(sendTx string) error {
	fmt.Printf("your input for open is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'open'
	issue, err := dexathon.NewOpen(sendTx)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func receive(sendTx string) error {
	fmt.Printf("your input for receive is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'receive'
	issue, err := dexathon.NewReceive(sendTx)
	if err != nil {
		return err
	}
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}

func badInputs(funcName string, additionalInfo string) error {
	fmt.Printf("Error in function %s \n", funcName)
	fmt.Printf("Invalid inputs: %v \n", strings.Join(os.Args[2:], ", "))
	fmt.Printf("Additional information: \n%s", additionalInfo)
	return nil
}

func main() {
	var command = os.Args[1]

	switch command {
	case "register":
		goodInputs, addInfo := dexathon.RegisterInputValidation()
		if goodInputs {
			if err := register(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := dexathon.LoginInputValidation()
		if goodInputs {
			if err := login(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := dexathon.IssueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(os.Args[2], 64)
			if err := issue(balance); err != nil {
				panic(err)
			}
		} else {
			badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := dexathon.SendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(os.Args[4], 64)
			if err := send(os.Args[2], os.Args[3], amount); err != nil {
				panic(err)
			}
		} else {
			badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := dexathon.OpenInputValidation()
		if goodInputs {
			if err := open(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := dexathon.ReceiveInputValidation()
		if goodInputs {
			if err := receive(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("receive", addInfo)
		}
	}
}
