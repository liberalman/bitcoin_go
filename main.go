package bitcoin_go

func test() {
    blockChain := CreateBlockChain("1DAhvHAataamMB7yg2hyFLeA7LE8LAuo88", "1234")

    blockChain.AddBlock("Send 1 BTC to Andy")
    blockChain.AddBlock("Send 2 mort BTC to Andy")

    /*for _, block := range blockChain.blocks {
        fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
        fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow := NewProofOfWork(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()
    }*/

    defer blockChain.db.Close()

    cli := CLI{blockChain}
    cli.Run()
}

func main() {
    // test()
    cli := CLI{}
    cli.Run()
}
