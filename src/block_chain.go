package src

import (
    "bytes"
    "crypto/ecdsa"
    "encoding/hex"
    "errors"
    "fmt"
    "github.com/boltdb/bolt"
    "log"
    "os"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinBaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
    tip []byte
    db  *bolt.DB
}

// 区块链迭代器
type BlockChainIterator struct {
    currentHash []byte
    db          *bolt.DB
}

// 入链
func (this *BlockChain) AddBlock(data string) {
    var lastHash []byte

    if err := this.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        lastHash = b.Get([]byte("l"))

        return nil
    }); nil != err {
        panic(err)
    }

    newBlock := NewBlock([]*Transaction{}, lastHash)

    if err := this.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        if err := b.Put(newBlock.Hash, newBlock.Serialize()); nil != err {
            panic(err)
        }
        if err := b.Put([]byte("l"), newBlock.Hash); nil != err {
            panic(err)
        }
        this.tip = newBlock.Hash

        return nil
    }); nil != err {
        panic(err)
    }
}

// creates a new blockchain DB
func CreateBlockChain(address, nodeID string) *BlockChain {
    dbFile := fmt.Sprintf(dbFile, nodeID)
    if dbExists(dbFile) {
        fmt.Println("BlockChain already exists.")
        os.Exit(1)
    }

    var tip []byte
    cbtx := CreateCoinBaseTX(address, genesisCoinBaseData)
    genesis := CreateGenesisBlock(cbtx)

    db, err := bolt.Open(dbFile, 0600, nil)
    if nil != err {
        panic(err)
    }

    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))

        if b == nil {
            b, err := tx.CreateBucket([]byte(blocksBucket))
            if nil != err {
                panic(err)
            }
            err = b.Put(genesis.Hash, genesis.Serialize())
            if nil != err {
                panic(err)
            }
            err = b.Put([]byte("l"), genesis.Hash)
            if nil != err {
                panic(err)
            }

            tip = genesis.Hash
        } else {
            tip = b.Get([]byte("l"))
        }

        return nil
    })

    bc := BlockChain{tip, db}

    return &bc
}

// 创建区块链，即有创世块的链， creates a new Blockchain with genesis Block
func NewBlockChain(nodeID string) *BlockChain {
    dbFile := fmt.Sprintf(dbFile, nodeID)
    if dbExists(dbFile) == false {
        fmt.Println("No existing block chain found. Create one first.")
        os.Exit(1)
    }

    var tip []byte
    db, err := bolt.Open(dbFile, 0600, nil)
    if err != nil {
        panic(err)
    }

    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        tip = b.Get([]byte("l"))

        return nil
    })
    if err != nil {
        panic(err)
    }

    bc := BlockChain{tip, db}

    return &bc
}

// 找到所有的未花费输出
func (this *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
    var (
        accumulated    = 0 // 累加货币值
        unSpentOutputs = make(map[string][]int)
        unSpentTXs     = this.FindUnSpentTransaction(address)
    )

Work:
    for _, tx := range unSpentTXs {
        txID := hex.EncodeToString(tx.ID)

        for outIdx, out := range tx.Vout {
            if out.CanBeUnlockedWith(address) && accumulated < amount {
                accumulated += out.Value
                unSpentOutputs[txID] = append(unSpentOutputs[txID], outIdx)

                if accumulated >= amount {
                    break Work
                }
            }
        }
    }

    return accumulated, unSpentOutputs
}

// 找到包含未花费输出的交易
func (this *BlockChain) FindUnSpentTransaction(address string) []Transaction {
    var (
        unSpentTXs []Transaction
        spentTXOs  = make(map[string][]int)
        bci        = this.Iterator()
    )

    for {
        block := bci.Next()

        for _, tx := range block.Transcations {
            txID := hex.EncodeToString(tx.ID)

        Outputs:
            for outIdx, out := range tx.Vout {
                // Was the output spent?
                if spentTXOs[txID] != nil {
                    for _, spentOut := range spentTXOs[txID] {
                        if spentOut == outIdx {
                            continue Outputs
                        }
                    }
                }

                if out.CanBeUnlockedWith(address) {
                    unSpentTXs = append(unSpentTXs, *tx)
                }
            }

            if tx.IsCoinBase() == false {
                for _, in := range tx.Vin {
                    if in.CanUnlockOutputWith(address) {
                        inTxID := hex.EncodeToString(in.Txid)
                        spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                    }
                }
            }

            if len(block.PrevBlockHash) == 0 { // 到链尾了，轮询结束
                break
            }
        }
    }

    return unSpentTXs
}

func (this *BlockChain) Iterator() *BlockChainIterator {
    return &BlockChainIterator{
        currentHash: this.tip,
        db:          this.db,
    }
}

func (this *BlockChainIterator) Next() *Block {
    var block *Block

    if err := this.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        encodedBlock := b.Get(this.currentHash)
        block = DeSerialize(encodedBlock)

        return nil
    }); nil != err {
        panic(err)
    }

    this.currentHash = block.PrevBlockHash

    return block
}

// MineBlock mines a new block with the provided transactions
func (this *BlockChain) MineBlock(transactions []*Transaction) *Block {
    var (
    	lastHash []byte
    )

    for _, tx := range transactions {
        // TODO: ignore transaction if it's not valid
        if this.VerifyTransaction(tx) != true {
            panic("ERROR: Invalid transaction")
        }
    }

    err := this.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        lastHash = b.Get([]byte("l"))
        return nil
    })
    if err != nil {
        panic(err)
    }

    newBlock := NewBlock(transactions, lastHash)

    err = this.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        err := b.Put(newBlock.Hash, newBlock.Serialize())
        if err != nil {
            panic(err)
        }

        err = b.Put([]byte("l"), newBlock.Hash)
        if err != nil {
            panic(err)
        }

        this.tip = newBlock.Hash

        return nil
    })
    if err != nil {
        panic(err)
    }

    return newBlock

    /*
    var lastHash []byte
    var lastHeight int

    for _, tx := range transactions {
        // TODO: ignore transaction if it's not valid
        if this.VerifyTransaction(tx) != true {
            panic("ERROR: Invalid transaction")
        }
    }

    err := this.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        lastHash = b.Get([]byte("l"))

        blockData := b.Get(lastHash)
        block := DeSerialize(blockData)

        lastHeight = block.Height

        return nil
    })
    if err != nil {
        panic(err)
    }

    newBlock := NewBlock(transactions, lastHash, lastHeight+1)

    err = this.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        err := b.Put(newBlock.Hash, newBlock.Serialize())
        if err != nil {
            panic(err)
        }

        err = b.Put([]byte("l"), newBlock.Hash)
        if err != nil {
            panic(err)
        }

        this.tip = newBlock.Hash

        return nil
    })
    if err != nil {
        panic(err)
    }

    return newBlock*/
}

// VerifyTransaction verifies transaction input signatures
func (this *BlockChain) VerifyTransaction(tx *Transaction) bool {
    if tx.IsCoinBase() {
        return true
    }

    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
        prevTX, err := this.FindTransaction(vin.Txid)
        if nil != err {
            panic(err)
        }
        prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }

    return tx.Verify(prevTXs)
}

func (this *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
    bci := this.Iterator()

    for {
        block := bci.Next()

        for _, tx := range block.Transcations {
            if bytes.Compare(tx.ID, ID) == 0 {
                return *tx, nil
            }
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }

    return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction signs inputs of a Transaction
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
        prevTX, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            log.Panic(err)
        }
        prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }

    tx.Sign(privKey, prevTXs)
}

func dbExists(dbFile string) bool {
    if _, err := os.Stat(dbFile); os.IsNotExist(err) {
        return false
    }

    return true
}

func (this *BlockChain) FindUTXO() map[string]TXOutputs {
    utxo := make(map[string]TXOutputs)
    return utxo
}
