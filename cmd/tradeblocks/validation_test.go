package main

import (
	"fmt"
	"testing"
)

func TestValidation(t *testing.T) {
	ok, _ := registerInputValidation([]string{"tradeblocks", "register", "test"})
	if !ok {
		t.Fatalf("register failed; expected ok")
	}
}

func TestOfferValidation(t *testing.T) {
	ok, _ := offerInputValidation([]string{"tradeblocks", "offer", "left"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want", "badquantity"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want", "10.0"})
	if !ok {
		t.Fatalf("offer failed; expected ok")
	}

	// tests for optionals
	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want", "10.0", "only-executor"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want", "10.0", "executor", "bad-fee"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = offerInputValidation([]string{"tradeblocks", "offer", "left", "ID", "counter", "want", "10.0", "executor", "10.0"})
	if !ok {
		t.Fatalf("offer failed; expected ok")
	}
}

func TestCommitValidation(t *testing.T) {
	ok, _ := commitInputValidation([]string{"tradeblocks", "commit", "offer", "right"})
	if !ok {
		t.Fatalf("offer failed; expected ok")
	}

	ok, _ = commitInputValidation([]string{"tradeblocks", "commit", "only offer"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = commitInputValidation([]string{"tradeblocks", "commit", "offer", "right", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}
}

func TestRefundLeftValidation(t *testing.T) {
	ok, _ := refundLeftInputValidation([]string{"tradeblocks", "refund-left", "offer"})
	if !ok {
		t.Fatalf("offer failed; expected ok")
	}

	ok, _ = refundLeftInputValidation([]string{"tradeblocks", "refund-left"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = refundLeftInputValidation([]string{"tradeblocks", "refund-left", "offer", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}
}

func TestRefundRightValidation(t *testing.T) {
	ok, _ := refundRightInputValidation([]string{"tradeblocks", "refund-right", "refundLeft"})
	if !ok {
		t.Fatalf("offer failed; expected ok")
	}

	ok, _ = refundRightInputValidation([]string{"tradeblocks", "refund-right"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = refundRightInputValidation([]string{"tradeblocks", "refund-right", "refundLeft", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}
}

func TestCreateOrderValidation(t *testing.T) {
	ok, _ := createOrderInputValidation([]string{"tradeblocks", "create-order", "send", "id", "True", "quote"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = createOrderInputValidation([]string{"tradeblocks", "create-order", "send", "id", "True", "quote", "10.0", "executor", "not-float"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = createOrderInputValidation([]string{"tradeblocks", "create-order", "send", "id", "not bool", "quote", "10.0", "executor", "10.0", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = createOrderInputValidation([]string{"tradeblocks", "create-order", "send", "id", "True", "quote", "not float", "executor", "10.0"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, err := createOrderInputValidation([]string{"tradeblocks", "create-order", "send", "id", "True", "quote", "10.0", "executor", "10.0"})
	if !ok {
		fmt.Printf("err %v \n", err)
		t.Fatalf("offer failed; expected ok")
	}
}

func TestAcceptOrderValidation(t *testing.T) {
	ok, _ := acceptOrderInputValidation([]string{"tradeblocks", "accept-order", "order"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = acceptOrderInputValidation([]string{"tradeblocks", "accept-order", "order", "link", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, err := acceptOrderInputValidation([]string{"tradeblocks", "accept-order", "order", "link"})
	if !ok {
		fmt.Printf("err %v \n", err)
		t.Fatalf("offer failed; expected ok")
	}
}

func TestRefundOrderValidation(t *testing.T) {
	ok, _ := refundOrderInputValidation([]string{"tradeblocks", "refund-order"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, _ = refundOrderInputValidation([]string{"tradeblocks", "refund-order", "order", "extra"})
	if ok {
		t.Fatalf("offer failed; expected error")
	}

	ok, err := refundOrderInputValidation([]string{"tradeblocks", "refund-order", "order"})
	if !ok {
		fmt.Printf("err %v \n", err)
		t.Fatalf("offer failed; expected ok")
	}
}
