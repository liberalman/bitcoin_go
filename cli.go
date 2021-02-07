// @Title 命令行
// @Description 解析输入参数，完成创建链，添加块，查看块等操作
// @Author shouchao.zheng 2021-01-24
// @Update shouchao.zheng 2021-01-24
package main

import (
	"flag"
	"fmt"
	"log"
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

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}

	//getBalanceCmd := flag.NewFlagSet("get_balance", flag.ExitOnError)
	//createBlockChainCmd := flag.NewFlagSet("create_block_chain", flag.ExitOnError)
	addBlockCmd := flag.NewFlagSet("add_block", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print_chain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("create_block_chain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("create_wallet", flag.ExitOnError)

	//getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	addBlockData := addBlockCmd.String("data", "", "Block data")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	createBlockChainAddress := createBlockChainCmd.String("address", "",
		"The address to send genesis block reward to")

	switch os.Args[1] {
	case "create_block_chain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if nil != err {
			panic(err)
		}
	case "create_wallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "add_block":
		err = addBlockCmd.Parse(os.Args[2:])
	case "print_chain":
		err = printChainCmd.Parse(os.Args[2:])
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
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

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		this.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddress == "" {
			createBlockChainCmd.Usage()
			os.Exit(1)
		}
		this.createBlockChain(*createBlockChainAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		this.createWallet(nodeID)
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

func (this *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		panic("ERROR: Recipient address is not valid")
	}

	blockChain := CreateBlockChain(from, nodeID)
	set := UTXOSet{blockChain}
	defer blockChain.db.Close()

	wallets, err := NewWallets(nodeID)
	if nil != err {
		panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &set)

	if mineNow {
		cbTX := CreateCoinBaseTX(from, "")
		txs := []*Transaction{cbTX, tx}

		newBlock := blockChain.MineBlock(txs)
		set.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}

func (this *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (this *CLI) createBlockChain(address, nodeID string) {
	if !ValidateAddress(address) {
		panic("ERROR: Address is not valid")
	}

	blockChain := CreateBlockChain(address, nodeID)
	defer blockChain.db.Close()

	set := UTXOSet{blockChain}
	set.ReIndex()

	fmt.Println("Done!")
}
