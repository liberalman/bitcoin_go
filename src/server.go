package src

import (
	"bytes"
	"encoding/gob"
    "fmt"
    "io/ioutil"
    "net"
)

var (
    nodeAddress string
    miningAddress string
    knownNodes = []string{"localhost:3000"}
)

const (
    commandLength = 12
    protocol = "tcp"
    nodeVersion = 1
)

type addr struct {
    AddrList []string
}

type version struct {
    Version int
    BestHeight int
    AddrFrom string
}

type tx struct {
	AddrFrom string
	Transaction []byte
}

type getblocks struct {
    AddrFrom string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte // 12 字节的缓冲区

    for i, c := range command {
        bytes[i] = byte(c) // 依次用命令名进行填充，剩下的字节置空
    }

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
    var command []byte

    for _, b := range bytes {
        if b != 0x0 {
            command = append(command, b)
        }
    }

    return fmt.Sprintf("%s", command)
}

func sendData(address string, data []byte) {}

func sendTx(address string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(address, request)
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if nil != err {
		panic(err)
	}

	return buff.Bytes()
}

func StartServer(nodeID, minerAddress string) {
    nodeAddress = fmt.Sprintf("localhost:%s", nodeID) // 中心节点地址硬编码
    miningAddress = minerAddress // 接收挖矿奖励地址
    ln, err := net.Listen(protocol, nodeAddress)
    if nil != err {
        panic(err)
    }
    defer ln.Close()

    blockChain := NewBlockChain(nodeID)

    if nodeAddress != knownNodes[0] {
        sendVersion(knownNodes[0], blockChain) // 查询是否自己的区块链已过时
    }

    for {
        conn, err := ln.Accept()
        if nil != err {
            panic(err)
        }
        go handleConnection(conn, blockChain)
    }
}

func handleConnection(conn net.Conn, blockChain *BlockChain) {
    request, err := ioutil.ReadAll(conn)
    if nil != err {
        panic(err)
    }

    command := bytesToCommand(request[:commandLength])
    fmt.Printf("Received %s command\n", command)

    switch command {
    case "addr":
        handleAddr(request)
    case "version":
        handleVersion(request, blockChain)
    default:
        fmt.Println("Unknown command!")
    }

    conn.Close()
}

func snedVersion(address string, blockChain *BlockChain) {
    bestHeight := blockChain.GetBestHeight()
    payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})

    request := append(commandToBytes("version"), payload...)

    sendData(address, request)
}

func handleVersion(request []byte, bc *BlockChain) {
    var (
        buff bytes.Buffer
        payload version
    )

    buff.Write(request[commandLength:])
    dec := gob.NewDecoder(&buff)
    err := dec.Decode(&payload)
    if nil != err {
        panic(err)
    }

    myBestHeight := bc.GetBestHeight()
    foreignerBestHeight := payload.BestHeight

    if myBestHeight < foreignerBestHeight {
        sendGetBlocks(payload.AddrFrom) // 对方的区块链更长，请求下载块
    } else if myBestHeight > foreignerBestHeight {
        sendVersion(payload.AddrFrom, bc) // 自身的区块链更长，回复 version 消息
    }

    if !nodeIsKnown(payload.AddrFrom) {
        knownNodes = append(knownNodes, payload.AddrFrom)
    }
}

func sendVersion(address string, blockChain *BlockChain) {
    bestHeight := blockChain.GetBestHeight()
    payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})

    request := append(commandToBytes("version"), payload...)

    sendData(address, request)
}


func sendGetBlocks(address string) {
    payload := gobEncode(getblocks{nodeAddress})
    request := append(commandToBytes("getblocks"), payload...)

    sendData(address, request)
}

func nodeIsKnown(address string) bool {
    for _, node := range knownNodes {
        if node == address {
            return true
        }
    }

    return false
}

func handleAddr(request []byte) {
    var (
        buff bytes.Buffer
        payload addr
    )

    buff.Write(request[commandLength:])
    dec := gob.NewDecoder(&buff)
    err := dec.Decode(&payload)
    if err != nil {
        panic(err)
    }

    knownNodes = append(knownNodes, payload.AddrList...)
    fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
    requestBlocks()
}

func requestBlocks() {
    for _, node := range knownNodes {
        sendGetBlocks(node)
    }
}

