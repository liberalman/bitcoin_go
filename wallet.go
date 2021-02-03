package main

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/gob"
    "fmt"
    "golang.org/x/crypto/ripemd160"
    "io/ioutil"
    "log"
    "os"
)

const walletFile = "wallet_%s.dat"
const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey []byte
}

type Wallets struct {
    Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
    private, public := newKeyPair()
    wallet := Wallet{private, public}

    return &wallet
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

func newKeyPair() (ecdsa.PrivateKey, []byte) {
    curve := elliptic.P256()
    private, err := ecdsa.GenerateKey(curve, rand.Reader)
    if nil != err {
        panic(err)
    }
    pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

    return *private, pubKey
}

func (this *Wallet) GetAddress() []byte {
    pubKeyHash := HashPubKey(this.PublicKey)

    versionedPayload := append([]byte{version}, pubKeyHash...)
    checksum := checkSum(versionedPayload)

    fullPayload := append(versionedPayload, checksum...)
    address := Base58Encode(fullPayload)

    return address
}

func HashPubKey(pubKey []byte) []byte {
    publicSHA256 := sha256.Sum256(pubKey)

    RIPEMD160Hasher := ripemd160.New()
    _, err := RIPEMD160Hasher.Write(publicSHA256[:])
    if nil != err {
        panic(err)
    }
    publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

    return publicRIPEMD160
}

func checkSum(payload []byte) []byte {
    firstSHA := sha256.Sum256(payload)
    secondSHA := sha256.Sum256(firstSHA[:])

    return secondSHA[:addressChecksumLen]
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
