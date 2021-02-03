package main

import (
    "encoding/hex"
    "fmt"
)

const subsidy = 10

// 交易输出
type TXOutput struct {
    Value        int    // 币值
    ScriptPubKey string // 锁定脚本
}

type TXInput struct {
    Txid      []byte // 之前交易的 ID
    Vout      int
    ScriptSig string
}

// 交易
type Transaction struct {
    ID   []byte
    Vin  []TXInput  // 输入
    Vout []TXOutput // 输出
}

func NewUTXOTranscation(from, to string, amount int, blockChain *BlockChain) *Transaction {
    var (
        inputs  []TXInput
        outputs []TXOutput
    )

    acc, validOutputs := blockChain.FindSpendableOutputs(from, amount) // 找到所有的未花费输出

    if acc < amount {
        panic("ERROR: Not enough funds")
    }

    // Build a list of inputs from validOutputs
    for txid, outs := range validOutputs {
        txID, _ := hex.DecodeString(txid)

        for _, out := range outs {
            input := TXInput{txID, out, from}
            inputs = append(inputs, input)
        }
    }

    // Build a list of outputs
    outputs = append(outputs, TXOutput{amount, to}) // 接收者地址锁定
    if acc > amount {
        outputs = append(outputs, TXOutput{acc - amount, from}) // a change（找零）, 发送者地址锁定
    }

    tx := Transaction{nil, inputs, outputs}
    tx.SetID()

    return &tx
}

// 创建一个 coinbase 交易，即"发行新币"，也就是给旷工奖励一个新币
func CreateCoinBaseTX(to, data string) *Transaction {
    if "" == data {
        data = fmt.Sprintf("Reward to '%s'", to)
    }

    txin := TXInput{[]byte{}, -1, data}
    txout := TXOutput{subsidy, to}
    tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
    tx.SetID()

    return &tx
}

// 输入解锁
func (this *TXInput) CanUnlockOutputWith(unlockingData string) bool {
    return this.ScriptSig == unlockingData
}

// 输出解锁
func (this *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
    return this.ScriptPubKey == unlockingData
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

/*
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
*/
func (this *Transaction) SetID() {}