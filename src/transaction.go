package src

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

const subsidy = 10 // 挖出新块的奖励金

// 交易
type Transaction struct {
	ID   []byte
	Vin  []TXInput  // 输入
	Vout []TXOutput // 输出
}

// 创建一个 coinbase 交易，即"发行新币"，也就是给旷工奖励一个新币
func CreateCoinBaseTX(to, data string) *Transaction {
	if "" == data {
		//data = fmt.Sprintf("Reward to '%s'", to)
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if nil != err {
			panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{*txout},
	}
	tx.ID = tx.Hash()

	return &tx
}

// Sign signs each input of a Transaction
func (this *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if this.IsCoinBase() {
		return // coinbase 交易没有实际输入，所以不签名
	}

	for _, vin := range this.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := this.TrimmedCopy() // 对修剪后的交易副本签名，而不是对整个完整的交易签名

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

		this.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// 修剪后的交易副本
// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (this *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range this.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range this.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{this.ID, inputs, outputs}

	return txCopy
}

// Hash returns the hash of the Transaction
func (this *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *this
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Verify verifies signatures of Transaction inputs
func (this *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := this.TrimmedCopy() // 副本
	curve := elliptic.P256()

	for inID, vin := range this.Vin {
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
	}

	return true
}

func (this *Transaction) SetID() {
	this.ID = this.Hash()
}

// IsCoinbase checks whether the transaction is coinbase
func (this *Transaction) IsCoinBase() bool {
	return len(this.Vin) == 1 && len(this.Vin[0].Txid) == 0 && this.Vin[0].Vout == -1
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
