package src

import (
    "bytes"
    "crypto/elliptic"
    "encoding/gob"
    "fmt"
    "io/ioutil"
    "log"
    "os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
    Wallets map[string]*Wallet
}

func NewWallets(nodeID string) (*Wallets, error) {
    wallets := Wallets{}
    wallets.Wallets = make(map[string]*Wallet)

    err := wallets.LoadFromFile(nodeID)

    return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (this *Wallets) CreateWallet() string {
    wallet := NewWallet()
    address := fmt.Sprintf("%s", wallet.GetAddress())

    this.Wallets[address] = wallet

    return address
}

// LoadFromFile loads wallets from the file
func (this *Wallets) LoadFromFile(nodeID string) error {
    walletFile := fmt.Sprintf(walletFile, nodeID)
    if _, err := os.Stat(walletFile); os.IsNotExist(err) {
        return err
    }

    fileContent, err := ioutil.ReadFile(walletFile)
    if err != nil {
        log.Panic(err)
    }

    var wallets Wallets
    gob.Register(elliptic.P256())
    decoder := gob.NewDecoder(bytes.NewReader(fileContent))
    err = decoder.Decode(&wallets)
    if err != nil {
        log.Panic(err)
    }

    this.Wallets = wallets.Wallets

    return nil
}

// SaveToFile saves wallets to a file
func (this Wallets) SaveToFile(nodeID string) {
    var content bytes.Buffer
    walletFile := fmt.Sprintf(walletFile, nodeID)

    gob.Register(elliptic.P256())

    encoder := gob.NewEncoder(&content)
    err := encoder.Encode(this)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
    if err != nil {
        panic(err)
    }
}

// return a Wallet by its address
func (this Wallets) GetWallet(address string) Wallet {
    return *this.Wallets[address]
}

