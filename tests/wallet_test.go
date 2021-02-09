package test

import (
    . "bitcoin_go/src"
    "encoding/hex"
    "fmt"
	"testing"

    // third package
    "github.com/stretchr/testify/assert"
)

func TestCreateWallet(t *testing.T) {
	cli.CreateWallet(nodeID)
}

func TestGetAddress(t *testing.T) {
    wallet := Wallet{}
    addr := wallet.GetAddress()
    assert.Equal(t, "1HT7xU2Ngenf7D4yocz2SAcnNLW7rK8d4E", string(addr))
    fmt.Println("address 1: ", string(addr))

    wallet2 := NewWallet()
    addr = wallet2.GetAddress()
    fmt.Println("address 2: ", string(addr))
    fmt.Println("PublicKey 2: ", hex.EncodeToString(wallet2.PublicKey))
    fmt.Println("PrivateKey 2: ", wallet2.PrivateKey)
}

func TestListAddress(t *testing.T) {
	cli.ListAddresses(nodeID)
}
