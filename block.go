package main

import (
    "bytes"
    "crypto/sha256"
    "time"
)

type Block struct {
    Timestamp     int64          // 当前时间戳，也就是区块创建的时间
    Transcations  []*Transaction // 区块存储的实际有效信息，也就是交易
    PrevBlockHash []byte         // 前一个块的哈希，即父哈希，256-bit，32 bytes
    Hash          []byte         // 当前块的哈希，256-bit，32 bytes
    Nonce         int
}

// 生成一个新的块
func NewBlock(transactions []*Transaction, preBlockHash []byte) *Block {
    block := &Block{
        Transcations:  transactions,
        PrevBlockHash: preBlockHash,
        Timestamp:     time.Now().Unix(),
        Nonce:         0,
    }
    pow := NewProofOfWork(block)
    nonce, hash := pow.Run()

    block.Hash = hash
    block.Nonce = nonce

    return block
}

/*// 计算块的Hash值
func (this *Block) SetHash() {
    timestamp := fmt.Sprintf("%d", this.Timestamp)
    // Join函数的功能是用字节切片sep把s中的每个字节切片连成一个字节切片并返回.
    headers := bytes.Join([][]byte{this.PrevBlockHash, this.Data, []byte(timestamp)}, []byte{})

    hash := sha256.Sum256(headers)

    this.Hash = hash[:]
}*/

// 创建 创世块（genesis block）
func CreateGenesisBlock(coinBase *Transaction) *Block {
    return NewBlock([]*Transaction{coinBase}, []byte{})
}

func (this *Block) HashTranscations() []byte {
    var (
        txHashes [][]byte
        txHash [32]byte
    )

    for _, tx := range this.Transcations {
        txHashes = append(txHashes, tx.ID) // 得每笔交易的哈希
    }
    txHash = sha256.Sum256(bytes.Join(txHashes, []byte{})) // 将它们组合之后计算哈希

    return txHash[:]
}


