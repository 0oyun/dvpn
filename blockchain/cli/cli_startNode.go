package cli

import (
	"github.com/toy-playground/miniDVPN/blockchain/network"
)

func (cli Cli) startNode() {
	network.StartNode(cli)
}
