package cli

import "github.com/toy-playground/miniDVPN/blockchain/block"

func (cli *Cli) printAllBlocks() {
	bc := block.NewBlockchain()
	bc.PrintAllBlockInfo()
}
