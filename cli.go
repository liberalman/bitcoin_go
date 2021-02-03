// @Title 命令行
// @Description 解析输入参数，完成创建链，添加块，查看块等操作
// @Author shouchao.zheng 2021-01-24
// @Update shouchao.zheng 2021-01-24
package bc

import (
    "flag"
    "fmt"
    "os"
    "strconv"
)

type CLI struct {
    blockChain *BlockChain
}

func (this *CLI) Run() {
    var (
        err error
    )
    this.validateArgs()

    addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

    addBlockData := addBlockCmd.String("data", "", "Block data")

    switch os.Args[1] {
    case "addblock":
        err = addBlockCmd.Parse(os.Args[2:])
    case "printchain":
        err = printChainCmd.Parse(os.Args[2:])
    default:
        this.printUsage()
        os.Exit(1)
    }
    if err != nil {
        panic(err)
    }

    if addBlockCmd.Parsed() {
        if *addBlockData == "" {
            addBlockCmd.Usage()
            os.Exit(1)
        }
        this.addBlock(*addBlockData)
    }

    if printChainCmd.Parsed() {
        this.printChain()
    }
}

func (this *CLI) printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
    fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
    fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
    fmt.Println("  listaddresses - Lists all addresses from the wallet file")
    fmt.Println("  printchain - Print all the blocks of the blockchain")
    fmt.Println("  reindexutxo - Rebuilds the UTXO set")
    fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
    fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (this *CLI) validateArgs() {
    if len(os.Args) < 2 {
        this.printUsage()
        os.Exit(1)
    }
}

func (this *CLI) addBlock(data string) {
    this.blockChain.AddBlock(data)
    fmt.Println("Success!")
}

func (this *CLI) printChain() {
    bci := this.blockChain.Iterator()

    for {
        block := bci.Next()

        fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
        //fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow := NewProofOfWork(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }
}

func (this *CLI) send(from, to string, amount int) {
    blockChain := CreateBlockChain(from)
    defer blockChain.db.Close()

    tx := NewUTXOTranscation(from, to, amount, blockChain)
    blockChain.MineBlock([]*Transaction{tx})
    fmt.Println("Success!")
}

