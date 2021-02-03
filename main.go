package main

func main() {
    blockChain := CreateBlockChain("")

    /*bc.AddBlock("Send 1 BTC to Andy")
      bc.AddBlock("Send 2 mort BTC to Andy")

      for _, block := range bc.blocks {
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

// go env -w GO111MODULE=on
// go env -w GOPROXY=https://goproxy.io,direct
