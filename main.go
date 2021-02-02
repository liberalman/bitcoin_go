package main

import (
    "bytes"
    "crypto/sha256"
    "encoding/binary"
    "fmt"
    "github.com/boltdb/bolt"
    "math"
    "math/big"
    "strconv"
    "time"
)

var (
    maxNonce = math.MaxInt64
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

type Block struct {
    Timestamp     int64  // 当前时间戳，也就是区块创建的时间
    Data          []byte // 区块存储的实际有效信息，也就是交易
    PrevBlockHash []byte // 前一个块的哈希，即父哈希
    Hash          []byte // 当前块的哈希
    Nonce         int
}

type BlockChain struct {
    //blocks []*Block

    tip []byte
    db  *bolt.DB
}

func (bc *BlockChain) AddBlock(data string) {
    var lastHash []byte

    if err := bc.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        lastHash = b.Get([]byte("l"))

        return nil
    }); nil != err {
        panic(err)
    }

    newBlock := NewBlock(data, lastHash)

    if err := bc.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        if err := b.Put(newBlock.Hash, newBlock.Serialize()); nil != err {
            panic(err)
        }
        if err := b.Put([]byte("l"), newBlock.Hash); nil != err {
            panic(err)
        }
        bc.tip = newBlock.Hash

        return nil
    }); nil != err {
        panic(err)
    }

    //prevBlock := bc.blocks[len(bc.blocks)-1]
    //newBlock := NewBlock(data, prevBlock.Hash)
    //bc.blocks = append(bc.blocks, newBlock)
}

func NewGenesisBlock() *Block {
    return NewBlock("Genesis Block", []byte{})
}

func NewBlockChain() *BlockChain {
    var tip []byte
    db, err := bolt.Open(dbFile, 0600, nil)
    if nil != err {
        panic(err)
    }

    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))

        if b == nil {
            genesis := NewGenesisBlock()
            b, err := tx.CreateBucket([]byte(blocksBucket))
            if nil != err {
                panic(err)
            }
            err = b.Put(genesis.Hash, genesis.Serialize())
            err = b.Put([]byte("l"), genesis.Hash)
            tip = genesis.Hash
        } else {
            tip = b.Get([]byte("l"))
        }

        return nil
    })

    bc := BlockChain{tip, db}

    //return &BlockChain{[]*Block{NewGenesisBlock()}}
    return &bc
}

func (b *Block) SetHash() {
    timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
    headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
    hash := sha256.Sum256(headers)

    b.Hash = hash[:]
}

func NewBlock(data string, preBlockHash []byte) *Block {
    block := &Block{time.Now().Unix(), []byte(data), preBlockHash, []byte{}, 0}
    //block.SetHash()
    pow := NewProofOfWork(block)
    nonce, hash := pow.Run()

    block.Hash = hash[:]
    block.Nonce = nonce

    return block
}

const targetBits = 24

type ProofOfWork struct {
    block  *Block
    target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits))

    pow := &ProofOfWork{b, target}

    return pow
}

func test() {
    data1 := []byte("I like donuts")
    data2 := []byte("I like donutsca07ca")
    targetBits := 24
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits))
    fmt.Printf("%x\n", sha256.Sum256(data1))
    fmt.Printf("%64x\n", target)
    fmt.Printf("%x\n", sha256.Sum256(data2))
}

// 将一个 int64 转化为一个字节数组(byte array)
func IntToHex(num int64) []byte {
    buff := new(bytes.Buffer)
    err := binary.Write(buff, binary.BigEndian, num)
    if err != nil {
        panic(err)
    }

    return buff.Bytes()
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
    data := bytes.Join(
        [][]byte{
            pow.block.PrevBlockHash,
            pow.block.Data,
            IntToHex(pow.block.Timestamp),
            IntToHex(int64(targetBits)),
            IntToHex(int64(nonce)),
        },
        []byte{})
    return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    fmt.Printf("Mining the block containning \"%s\"\n", pow.block.Data)
    for nonce < maxNonce {
        data := pow.prepareData(nonce)
        hash = sha256.Sum256(data)
        hashInt.SetBytes(hash[:])

        if hashInt.Cmp(pow.target) == -1 {
            fmt.Printf("\r%x", hash)
            break
        } else {
            nonce++
        }
    }
    fmt.Print("\n\n")

    return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
    var hashInt big.Int

    data := pow.prepareData(pow.block.Nonce)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])

    isValid := hashInt.Cmp(pow.target) == -1

    return isValid
}

func main() {
    bc := NewBlockChain()

    bc.AddBlock("Send 1 BTC to Ivan")
    bc.AddBlock("Send 2 more BTC to Ivan")

    /*for _, block := range bc.blocks {
        fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
        fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)

        pow := NewProofOfWork(block)
        fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()
    }*/

    //test()
}
