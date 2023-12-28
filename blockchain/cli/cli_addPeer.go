package cli

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/toy-playground/miniDVPN/blockchain/network"
)

func (cli *Cli) addPeer(addr string) {
	peerInfo, err := peer.AddrInfoFromString(addr)
	if err != nil {
		fmt.Printf("addPeer error: %v\n", err)
		return
	}
	network.PeerChan <- *peerInfo
}
