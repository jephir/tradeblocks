package tradeblocks

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"encoding/base64"
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

	errSign := issueBlock.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	decodedSig, err := base64.StdEncoding.DecodeString(issueBlock.Signature)
	if err != nil {
		t.Fatal(err)
	}

	errVerify := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hashedBytes[:], decodedSig)
	if errVerify != nil {
		t.Fatalf("verify failed with: %v", errVerify)
	}

}

func TestVerifyBlock(t *testing.T) {
	//make a block
	issueBlock := NewIssueBlock("xtb:test", 100)
	if issueBlock.Signature != "" {
		t.Fatal("Signature was not empty string on new block")
	}

	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err)
	}

	// sign it
	errSign := issueBlock.SignBlock(key)
	if errSign != nil {
		t.Fatal(errSign)
	}

	errVerify := issueBlock.VerifyBlock(&key.PublicKey)
	if errVerify != nil {
		t.Fatal(errVerify)
	}
}
