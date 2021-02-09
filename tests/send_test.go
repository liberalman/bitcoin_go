package test

import (
	"testing"
)

var (
	from = "18vhdHeZ2XJLSSd861p4XxFVYwaLeNcGP2"
	to = "1LKMabNYff5xKot4FRnmMnxSG6C1SHjN96"
	amount = 1
	mineNow = true
)

func TestSend(t *testing.T) {
	cli.Send(from, to, amount, nodeID, mineNow)
}