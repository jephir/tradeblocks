package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jephir/tradeblocks"
	"github.com/jephir/tradeblocks/app"
)

func main() {
	var command = os.Args[1]

	switch command {
	case "register":
		goodInputs, addInfo := tradeblocks.RegisterInputValidation()
		if goodInputs {
			if err := register(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := tradeblocks.LoginInputValidation()
		if goodInputs {
			if err := login(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := tradeblocks.IssueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(os.Args[2], 64)
			if err := app.Issue(balance); err != nil {
				panic(err)
			}
		} else {
			badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := tradeblocks.SendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(os.Args[4], 64)
			if err := app.Send(os.Args[2], os.Args[3], amount); err != nil {
				panic(err)
			}
		} else {
			badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := tradeblocks.OpenInputValidation()
		if goodInputs {
			if err := app.Open(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := tradeblocks.ReceiveInputValidation()
		if goodInputs {
			if err := app.Receive(os.Args[2]); err != nil {
				panic(err)
			}
		} else {
			badInputs("receive", addInfo)
		}
	}
}

func badInputs(funcName string, additionalInfo string) error {
	fmt.Printf("Error in function %s \n", funcName)
	fmt.Printf("Invalid inputs: %v \n", strings.Join(os.Args[2:], ", "))
	fmt.Printf("Additional information: \n%s", additionalInfo)
	return nil
}

func register(name string) error {
	fmt.Printf("your input for register is %s \n", name)
	privateKeyFile, err := os.Create(name + ".pem")
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()
	publicKeyFile, err := os.Create(name + ".pub")
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()
	if err := app.Register(privateKeyFile, publicKeyFile, name); err != nil {
		return err
	}
	if err := publicKeyFile.Close(); err != nil {
		return err
	}
	if err := privateKeyFile.Close(); err != nil {
		return err
	}
	return nil
}

func login(name string) error {
	fmt.Printf("your input for login is %s \n", name)
	return ioutil.WriteFile("user", []byte(name), 0644)
}

func issue(balance float64) error {
	fmt.Printf("your input for issue is %v \n", balance)
	// creates a new BaseTransaction with action = 'issue'
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
	return nil
}