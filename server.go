package main

import (
	"bytes"
	"encoding/gob"
)

var nodeAddress string
var knownNodes = []string{"localhost:3000"}

type tx struct {
	AddFrom string
	Transaction []byte
}

func commandToBytes(command string) []byte {
	var bytes []byte
	return bytes
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
