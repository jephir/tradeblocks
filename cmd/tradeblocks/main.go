package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

const keySize = 4096

func main() {
	var command = os.Args[1]
	var block *tradeblocks.AccountBlock
	var err error

	switch command {
	case "register":
		goodInputs, addInfo := registerInputValidation()
		if goodInputs {
			address, err := register(os.Args[2])
			if err != nil {
				panic(err)
			}
			fmt.Println(address)
		} else {
			badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := loginInputValidation()
		if goodInputs {
			address, err := login(os.Args[2])
			if err != nil {
				panic(err)
			}
			fmt.Println(address)
		} else {
			badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := issueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(os.Args[2], 64)
			block, err = issue(balance)
			if err != nil {
				panic(err)
			}
		} else {
			badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := sendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(os.Args[4], 64)
			block, err = send(os.Args[2], os.Args[3], amount)
			if err != nil {
				panic(err)
			}
		} else {
			badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := openInputValidation()
		if goodInputs {
			block, err = open(os.Args[2])
			if err != nil {
				panic(err)
			}
		} else {
			badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := receiveInputValidation()
		if goodInputs {
			block, err = receive(os.Args[2])
			if err != nil {
				panic(err)
			}
		} else {
			badInputs("receive", addInfo)
		}
	}
	if block != nil {
		h, err := app.AccountBlockHash(block)
		if err != nil {
			panic(err)
		}
		fmt.Println(h)
	}
}

func badInputs(funcName string, additionalInfo string) error {
	fmt.Printf("Error in function %s \n", funcName)
	fmt.Printf("Invalid inputs: %v \n", strings.Join(os.Args[2:], ", "))
	fmt.Printf("Additional information: \n%s", additionalInfo)
	os.Exit(1)
	return nil
}

func register(name string) (address string, err error) {
	privateKeyFile, err := os.Create(name + ".pem")
	if err != nil {
		return
	}
	defer privateKeyFile.Close()
	publicKeyFile, err := os.Create(name + ".pub")
	if err != nil {
		return
	}
	defer publicKeyFile.Close()
	address, err = app.Register(privateKeyFile, publicKeyFile, name, keySize)
	if err != nil {
		return
	}
	if err := publicKeyFile.Close(); err != nil {
		return "", err
	}
	if err := privateKeyFile.Close(); err != nil {
		return "", err
	}
	return
}

func login(name string) (address string, err error) {
	if err := ioutil.WriteFile("user", []byte(name), 0644); err != nil {
		return "", err
	}
	f, err := openPublicKey()
	if err != nil {
		return
	}
	defer f.Close()
	address, err = app.PublicKeyToAddress(f)
	return
}

func issue(balance float64) (*tradeblocks.AccountBlock, error) {
	key, err := openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	return app.Issue(key, balance)
}

func send(to string, token string, amount float64) (*tradeblocks.AccountBlock, error) {
	key, err := openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	previous, err := getHeadBlock(key, token)
	if err != nil {
		return nil, err
	}
	return app.Send(key, previous, to, amount)
}

func open(link string) (*tradeblocks.AccountBlock, error) {
	key, err := openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	send, err := getBlock(link)
	if err != nil {
		return nil, err
	}
	return app.Open(key, send)
}

func receive(link string) (*tradeblocks.AccountBlock, error) {
	key, err := openPublicKey()
	if err != nil {
		return nil, err
	}
	defer key.Close()
	send, err := getBlock(link)
	if err != nil {
		return nil, err
	}
	previous, err := getHeadBlock(key, send.Token)
	if err != nil {
		return nil, err
	}
	return app.Receive(key, previous, send)
}

func openPublicKey() (*os.File, error) {
	user, err := ioutil.ReadFile("user")
	if err != nil {
		return nil, err
	}
	return os.Open(string(user) + ".pub")
}

func getHeadBlock(publicKey io.Reader, token string) (*tradeblocks.AccountBlock, error) {
	address, err := app.PublicKeyToAddress(publicKey)
	if err != nil {
		return nil, err
	}
	// TODO Get the block from the server
	return &tradeblocks.AccountBlock{
		Action:         "open",
		Account:        address,
		Token:          token,
		Previous:       "",
		Representative: "",
		Balance:        100,
		Link:           "",
	}, nil
}

func getBlock(hash string) (*tradeblocks.AccountBlock, error) {
	// TODO Get the block from the server
	return &tradeblocks.AccountBlock{
		Action:         "open",
		Account:        "xtb:test",
		Token:          "xtb:testtoken",
		Previous:       "",
		Representative: "",
		Balance:        100,
		Link:           "",
	}, nil
}
