package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func register(name string) {
	fmt.Printf("your input for register is %s \n", name)
}

func login(name string) {
	fmt.Printf("your input for login is %s \n", name)
}

func issue(balance float64) {
	fmt.Printf("your input for issue is %v \n", balance)
	// creates a new BaseTransaction with action = 'issue'
	var issue = NewIssue(balance)
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
}

func send(toAccount, token string, amount float64) {
	fmt.Printf("your input for send is %s %s %v \n", toAccount, token, amount)
	// creates a new BaseTransaction with action = 'send'
	var issue = NewSend(toAccount, token, amount)
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
}

func open(sendTx string) {
	fmt.Printf("your input for open is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'open'
	var issue = NewOpen(sendTx)
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
}

func receive(sendTx string) {
	fmt.Printf("your input for receive is %s \n", sendTx)
	// creates a new BaseTransaction with action = 'receive'
	var issue = NewReceive(sendTx)
	fmt.Printf("the issue at hand is \n")
	spew.Dump(issue)
}

func badInputs(funcName string, additionalInfo string) {
	fmt.Printf("Error in function %s \n", funcName)
	fmt.Printf("Invalid inputs: %v \n", strings.Join(os.Args[2:], ", "))
	fmt.Printf("Additional information: \n%s", additionalInfo)
}

func main() {
	var command = os.Args[1]

	switch command {
	case "register":
		goodInputs, addInfo := registerInputValidation()
		if goodInputs {
			register(os.Args[2])
		} else {
			badInputs("register", addInfo)
		}
	case "login":
		goodInputs, addInfo := loginInputValidation()
		if goodInputs {
			login(os.Args[2])
		} else {
			badInputs("login", addInfo)
		}
	case "issue":
		goodInputs, addInfo := issueInputValidation()
		if goodInputs {
			balance, _ := strconv.ParseFloat(os.Args[2], 64)
			issue(balance)
		} else {
			badInputs("issue", addInfo)
		}
	case "send":
		goodInputs, addInfo := sendInputValidation()
		if goodInputs {
			amount, _ := strconv.ParseFloat(os.Args[4], 64)
			send(os.Args[2], os.Args[3], amount)
		} else {
			badInputs("send", addInfo)
		}
	case "open":
		goodInputs, addInfo := openInputValidation()
		if goodInputs {
			open(os.Args[2])
		} else {
			badInputs("open", addInfo)
		}
	case "receive":
		goodInputs, addInfo := receiveInputValidation()
		if goodInputs {
			receive(os.Args[2])
		} else {
			badInputs("receive", addInfo)
		}
	}
}
