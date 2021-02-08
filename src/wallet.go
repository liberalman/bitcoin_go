package src

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey []byte
}

func NewWallet() *Wallet {
    private, public := newKeyPair()
    wallet := Wallet{private, public}

    return &wallet
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

// check if address is valid
func ValidateAddress(address string) bool {
    return true
    //pubKeyHash := Base58Decode([]byte(address))
    //actualCheckSum := pubKeyHash[len(pubKeyHash) - addressChecksumLen : ]
    //version := pubKeyHash[0]
    //pubKeyHash =
}
