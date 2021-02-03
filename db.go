package main

import (
    "bytes"
    "encoding/gob"
    "github.com/boltdb/bolt"
)



// 区块链迭代器
type BlockChainIterator struct {
    currentHash []byte
    db          *bolt.DB
}

func (b *Block) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)

    if err := encoder.Encode(b); nil != err {
        panic(err)
    }

    return result.Bytes()
}

func (b *Block) DeSerializeBlock(d []byte) *Block {
    var block Block

    decoder := gob.NewDecoder(bytes.NewReader(d))
    if err := decoder.Decode(&block); nil != err {
        panic(err)
    }

    return &block
}

func (bc *BlockChain) Iterator() *BlockChainIterator {
    bci := &BlockChainIterator{bc.tip, bc.db}

    return bci
}

func (i *BlockChainIterator) Next() *Block {
	var block *Block

	if err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = block.DeSerializeBlock(encodedBlock)

		return nil
	}); nil != err {
		panic(err)
	}

	i.currentHash = block.Hash

	return block
}package bc

import (
    "bytes"
    "encoding/gob"
)

// 序列化
func (this *Block) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)

    if err := encoder.Encode(this); nil != err {
        panic(err)
    }

    return result.Bytes()
}

// 反序列化
func DeSerialize(data []byte) *Block {
    var block Block

    decoder := gob.NewDecoder(bytes.NewReader(data))

    if err := decoder.Decode(&block); nil != err {
        panic(err)
    }

    return &block
}
