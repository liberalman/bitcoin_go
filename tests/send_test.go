package test

import (
	"testing"
)

var (
	from = "13ETDxHQCzmgc7TfpC76eKrKNY9SmacSsj"
	to = "1wjuf2kvUSd6mraLcHa6s5BnS5VvJDYDh"
	amount = 1
	mineNow = true
)

func TestSend(t *testing.T) {
	cli.Send(from, to, amount, nodeID, mineNow)
}