package cli

import (
	"github.com/toy-playground/miniDVPN/blockchain/block"
	"github.com/toy-playground/miniDVPN/blockchain/network"
)

func (cli *Cli) genesis() {
	bc := block.NewBlockchain()
	bc.CreateGenesisBlock(network.Send{})
}
