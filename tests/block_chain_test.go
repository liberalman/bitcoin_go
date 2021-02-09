package test

import "testing"

var address = "18vhdHeZ2XJLSSd861p4XxFVYwaLeNcGP2" // 通过 ListAddress() 查出来地址列表，赋值到这里。

func TestCreateBlockChain(t *testing.T) {
	cli.CreateBlockChain(address, nodeID)
}

func TestPrintChain(t *testing.T) {
	cli.PrintChain(nodeID)
}

