package tradeblocks

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"testing"
)

func TestHash(t *testing.T) {
	expect := "VXF6FV3YI4MDD6T7AYHSD3JAUQ4GUXRTPVESHTDXUCFESS4BZGIA"
	b := NewIssueBlock("xtb:test", 100)
	h := b.Hash()
	if h != expect {
		t.Fatalf("Hash was incorrect, got: %s, want: %s", h, expect)
	}
}

func TestSignBlock(t *testing.T) {
	issueBlock := NewIssueBlock("xtb:test", 100)
	if issueBlock.Signature != "" {
		t.Fatal("Signature was not empty string on new block")
	}

	// save the hash before the block changes
	issueHash := issueBlock.Hash()
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(issueHash)
	hashedBytes := []byte(decoded)

	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	r := bytes.NewReader(keyBytes)

	errSign := issueBlock.SignBlock(r)
	if errSign != nil {
		t.Fatal(errSign)
	}

	if len(issueBlock.Signature) != 64 {
		t.Fatalf("Hash length was incorrect, got: %v, want: %v", len(issueBlock.Signature), 64)
	}

	decodedSig := []byte(issueBlock.Signature)

	errVerify := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hashedBytes[:], decodedSig)
	if errVerify != nil {
		t.Fatalf("verify failed with: %v", errVerify)
	}
}
