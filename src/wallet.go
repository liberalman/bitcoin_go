package src

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

const (
	version1           = byte(0x00)
	addressCheckSumLen = 4
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey // 私钥
	PublicKey  []byte           // 公钥
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}

	return &wallet
}

// 生成密钥对
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader) // 私钥
	if nil != err {
		panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) // 公钥

	return *private, pubKey
}

// 生成钱包地址
func (this *Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(this.PublicKey)

	versionedPayload := append([]byte{version1}, pubKeyHash...) // (version + pubKeyHash)
	checksum := checkSum(versionedPayload)                      // 计算出 (版本+公钥) 的校验和

	fullPayload := append(versionedPayload, checksum...) // 把校验和追加到尾部 (version + pubKeyHash + checksum)
	address := Base58Encode(fullPayload)                 // 上面三个组合，经过base58编码之后，就生成了地址

	return address
}

// 计算公钥的哈希值
func HashPubKey(pubKey []byte) []byte {
	// 使用 RIPEMD160(SHA256(PubKey)) 方法，进行两次哈希。一次是SHA256，一次是RIPEMD160。
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if nil != err {
		panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// check if address is valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualCheckSum := pubKeyHash[len(pubKeyHash)-addressCheckSumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressCheckSumLen]
	targetCheckSum := checkSum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualCheckSum, targetCheckSum) == 0
}

// 计算校验和
func checkSum(payload []byte) []byte {
	// 先进行两次哈希
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressCheckSumLen] // 取两次哈希结果的前4个字节作为 校验和
}
