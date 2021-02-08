package test

import "testing"

func TestGetBalance(t *testing.T) {
	cli.GetBalance(address, nodeID)
}

func TestReindexUTXO(t *testing.T) {
	cli.ReindexUTXO(nodeID)
}