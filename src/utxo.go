package src

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
)

const subsidy = 10
const utxoBucket = "chainstate"

// 交易输出
type TXOutput struct {
	Value int // 币值
	//ScriptPubKey string // 锁定脚本
	PubKeyHash []byte // 公钥的哈希值
}

// 交易输入
type TXInput struct {
	Txid      []byte // 之前交易的 ID
	Vout      int
	Signature []byte
	PubKey    []byte
}

// 交易
type Transaction struct {
	ID   []byte
	Vin  []TXInput  // 输入
	Vout []TXOutput // 输出
}

type UTXOSet struct {
	BlockChain *BlockChain
}

func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var (
		inputs  []TXInput
		outputs []TXOutput
	)

	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount) // 找到所有的未花费输出

	if acc < amount {
		panic("ERROR: Not enough funds")
	}

	// Build a list of inputs from validOutputs
	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to)) // 接收者地址锁定
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change（找零）, 发送者地址锁定
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	UTXOSet.BlockChain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

// 创建一个 coinbase 交易，即"发行新币"，也就是给旷工奖励一个新币
func CreateCoinBaseTX(to, data string) *Transaction {
	if "" == data {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		Signature: []byte{},
		PubKey:    []byte{},
	}
	txout := TXOutput{
		Value:      subsidy,
		PubKeyHash: []byte{}}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// 输入解锁
func (this *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	//return this.ScriptSig == unlockingData
	return false
}

// 输出解锁
func (this *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	//return this.PubKeyHash == unlockingData
	return false
}

// IsCoinbase checks whether the transaction is coinbase
func (this *Transaction) IsCoinBase() bool {
	return len(this.Vin) == 1 && len(this.Vin[0].Txid) == 0 && this.Vin[0].Vout == -1
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			panic("ERROR: Previous transaction is not correct")
		}
	}

	/*txCopy := tx.TrimmedCopy()
	  curve := elliptic.P256()

	  for inID, vin := range tx.Vin {
	      prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
	      txCopy.Vin[inID].Signature = nil
	      txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

	      r := big.Int{}
	      s := big.Int{}
	      sigLen := len(vin.Signature)
	      r.SetBytes(vin.Signature[:(sigLen / 2)])
	      s.SetBytes(vin.Signature[(sigLen / 2):])

	      x := big.Int{}
	      y := big.Int{}
	      keyLen := len(vin.PubKey)
	      x.SetBytes(vin.PubKey[:(keyLen / 2)])
	      y.SetBytes(vin.PubKey[(keyLen / 2):])

	      dataToVerify := fmt.Sprintf("%x\n", txCopy)

	      rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	      if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
	          return false
	      }
	      txCopy.Vin[inID].PubKey = nil
	  }*/

	return true
}

func (this *Transaction) SetID() {
	this.ID = this.Hash()
}

func (this *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(this.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (this *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address) // 先解码
	// 去掉第1个字节（版本号）和最后4个字节（校验值），取中间的即公钥的哈希值
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressCheckSumLen]
	this.PubKeyHash = pubKeyHash
}

func (this *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(this.PubKeyHash, pubKeyHash) == 0
}

// Hash returns the hash of the Transaction
func (this *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *this
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Serialize returns a serialized Transaction
func (this *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(this)
	if err != nil {
		panic(err)
	}

	return encoded.Bytes()
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.BlockChain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return accumulated, unspentOutputs
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// 修剪后的交易副本
// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

type TXOutputs struct {
	Outputs []TXOutput
}

// serializes TXOutputs
func (this *TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(this)
	if nil != err {
		panic(err)
	}

	return buff.Bytes()
}

// deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		panic(err)
	}

	return outputs
}

// rebuilds the UTXO set
func (this *UTXOSet) ReIndex() {
	db := this.BlockChain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketName); nil != err && err != bolt.ErrBucketNotFound {
			panic(err)
		}

		if _, err := tx.CreateBucket(bucketName); nil != err {
			panic(err)
		}

		return nil
	})
	if nil != err {
		panic(err)
	}

	utxo := this.BlockChain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range utxo {
			key, err := hex.DecodeString(txID)
			if nil != err {
				panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if nil != err {
				panic(err)
			}
		}

		return nil
	})
}

// finds UTXO for a public key hash
func (this *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var utxos []TXOutput
	db := this.BlockChain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					utxos = append(utxos, out)
				}
			}
		}

		return nil
	})
	if nil != err {
		panic(err)
	}

	return utxos
}

// updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (this *UTXOSet) Update(block *Block) {
	db := this.BlockChain.db

	err := db.Update(func(tx *bolt.Tx) error {
		return nil
	})
	if nil != err {
		panic(err)
	}
}

// CountTransactions returns the number of transactions in the UTXO set
func (this *UTXOSet) CountTransactions() int {
	db := this.BlockChain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return counter
}

func (this *Transaction) PrintTransaction() {
	fmt.Printf("|----- transaction %v -----|\n", hex.EncodeToString(this.ID))
	for _, in := range this.Vin {
		fmt.Printf("|Vin | PubKey: %s, Signature: %s, Txid: %s|\n", hex.EncodeToString(in.PubKey),
			hex.EncodeToString(in.Signature), hex.EncodeToString(in.Txid))
	}
	for _, out := range this.Vout {
		fmt.Printf("|Vout| Value: %d, PubKeyHash: %s|\n", out.Value, hex.EncodeToString(out.PubKeyHash))
	}
}
