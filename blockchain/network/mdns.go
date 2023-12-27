package network

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func initMDNS(ctx context.Context, peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	notifee := discoveryNotifee{
		PeerChan: make(chan peer.AddrInfo),
	}
	disc := mdns.NewMdnsService(peerhost, rendezvous, &notifee)

	err := disc.Start()
	if err != nil {
		fmt.Println("init MDNS err:", err)
		panic(err)
	}
	return notifee.PeerChan
}
