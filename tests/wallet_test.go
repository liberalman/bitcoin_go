package test

import (
	"testing"
)

func TestCreateWallet(t *testing.T) {
	cli.CreateWallet(nodeID)
}

func TestListAddress(t *testing.T) {
	cli.ListAddresses(nodeID)
}