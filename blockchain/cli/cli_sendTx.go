package cli

import (
	"fmt"

	"github.com/toy-playground/miniDVPN/blockchain/block"
	"github.com/toy-playground/miniDVPN/blockchain/network"
)

func (cli Cli) sendTx(txtype, data string) {
	blc := block.NewBlockchain()
	blc.CreateTransaction(txtype, data, network.Send{})
	fmt.Println("transaction finish!")
}
