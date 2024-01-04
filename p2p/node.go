package p2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
)

// Protocol is a descriptor for the miniDVPN P2P Protocol.
const Protocol = "/miniDVPN/vpn/1.0.0"

func GetMyIP() string {
	var MyIP string

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Printf("get my ip failed: %s\n", err)
	} else {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		MyIP = localAddr.IP.String()
	}
	fmt.Printf("[*] your ip address: %s\n", MyIP)
	return MyIP
}

// CreateNode creates an internal Libp2p nodes and returns it and it's DHT Discovery service.
func CreateNode(ctx context.Context, inputKey string, port int, handler network.StreamHandler, remoteIP string) (node host.Host, dhtOut *dht.IpfsDHT, err error) {
	// Unmarshal Private Key
	privateKey, err := crypto.UnmarshalPrivateKey([]byte(inputKey))
	if err != nil {
		return
	}

	ip := GetMyIP()
	ip6quic := fmt.Sprintf("/ip6/::/udp/%d/quic", port)
	ip4quic := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)

	ip6tcp := fmt.Sprintf("/ip6/::/tcp/%d", port)
	ip4tcp := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	ip4tcp2 := fmt.Sprintf("/ip4/%s/tcp/%d", ip, port)
	fmt.Printf("ip4tcp: %s\n", ip4tcp2)

	// Create libp2p node
	node, err = libp2p.New(
		libp2p.ListenAddrStrings(ip6quic, ip4quic, ip6tcp, ip4tcp, ip4tcp2),
		libp2p.Identity(privateKey),
		libp2p.NoSecurity,
		libp2p.NATPortMap(),
		libp2p.DefaultMuxers,
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.FallbackDefaults,
	)
	if err != nil {
		return
	}

	// Setup Hyprspace Stream Handler
	node.SetStreamHandler(Protocol, handler)

	// Create DHT Subsystem
	dhtOut = dht.NewDHTClient(ctx, node, datastore.NewMapDatastore())

	tmp3 := fmt.Sprintf("/ip4/%s/tcp/9003/p2p/QmabJcpncmRMH8YM6AZKtx4TRa6Q1fkB1GTp1JuftiHwao", remoteIP)
	tmp2 := fmt.Sprintf("/ip4/%s/tcp/9002/p2p/QmNWYjnJqUHFMuLRH7KxZfgZBjQ5YfHBA9VHGHHaibQehp", remoteIP)
	tmp := fmt.Sprintf("/ip4/%s/tcp/9001/p2p/Qmcvb91UBUtgFHdEQDTNVW1tT9B8ebagojJNaGnhJmW5iY", remoteIP)
	// Define Bootstrap Nodes.
	peers := []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	}
	peers = append(peers, tmp)
	peers = append(peers, tmp2)
	peers = append(peers, tmp3)

	// Convert Bootstap Nodes into usable addresses.
	BootstrapPeers := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return node, dhtOut, err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return node, dhtOut, err
		}
		pi, ok := BootstrapPeers[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			BootstrapPeers[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	lock := sync.Mutex{}
	count := 0
	wg.Add(len(BootstrapPeers))
	for _, peerInfo := range BootstrapPeers {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := node.Connect(ctx, *peerInfo)
			if err == nil {
				lock.Lock()
				count++
				lock.Unlock()

			}
		}(peerInfo)
	}
	wg.Wait()

	if count < 1 {
		return node, dhtOut, errors.New("unable to bootstrap libp2p node")
	}

	return node, dhtOut, nil
}
