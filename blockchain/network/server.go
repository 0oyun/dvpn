package network

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"github.com/toy-playground/miniDVPN/blockchain/block"
	"github.com/toy-playground/miniDVPN/blockchain/tun"
)

// 在P2P网络中已发现的节点池
// key:节点ID  value:节点详细信息
var peerPool = make(map[string]peer.AddrInfo)
var ctx = context.Background()
var send = Send{}
var PeerChan chan peer.AddrInfo
var (
	// iface is the tun device used to pass packets between
	// miniDVPN and the user's machine.
	tunDev      *tun.TUN
	TunPeerPool = make(map[string]peer.ID)
	// activeStreams is a map of active streams to a peer
	activeStreams map[string]network.Stream

	h p2phost.Host
)

type CLI interface {
	ReceiveCMD()
}

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

// 启动本地节点
func StartNode(cli CLI) {
	//先获取本地区块最新高度
	bc := block.NewBlockchain()
	block.NewestBlockHeight = bc.GetLastBlockHeight()
	fmt.Printf("[*] listen IP Address: %s Port: %s\n", ListenHost, ListenPort)
	r := rand.Reader
	// 为本地节点创建RSA密钥对
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		fmt.Printf("generate keypair failed: %s\n", err)
		panic(err)
	}

	// ListenHost = GetMyIP()
	// 创建本地节点地址信息
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ListenHost, ListenPort))
	//传入地址信息，RSA密钥对信息，生成libp2p本地host信息
	h, err = libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.NoSecurity,
		libp2p.NATPortMap(),
		libp2p.DefaultMuxers,
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.FallbackDefaults,
	)
	if err != nil {
		panic(err)
	}

	// macOS只能有一个peer
	// 用linux做测试
	if runtime.GOOS == "darwin" {
		// Grab ip address of only peer in config
	} else {
		// Create new TUN device
		tunDev, err = tun.New(
			InterfaceName,
			tun.Address(InterfaceAddress),
			tun.MTU(1420),
		)
		if err != nil {
			panic(err)
		}
		err = tunDev.Up()
		if err != nil {
			panic("unable to bring up tun device")
		}
	}

	//写入全局变量本地主机信息
	localHost = h
	//写入全局变量本地P2P节点地址详细信息
	localAddr = fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", ListenHost, ListenPort, h.ID().String())
	fmt.Printf("[*] your p2p address: %s\n", localAddr)
	//启动监听本地端口，并且传入一个处理流的函数，当本地节点接收到流的时候回调处理流的函数
	h.SetStreamHandler(protocol.ID(ProtocolID), handleStream)
	h.SetStreamHandler(protocol.ID(VPNID), vpnStreamHandler)
	//寻找p2p网络并加入到节点池里
	go findP2PPeer()
	//监测节点池,如果发现网络当中节点有变动则打印到屏幕
	go monitorP2PNodes()
	// //启一个go程去向其他p2p节点发送高度信息，来进行更新区块数据
	// go sendVersionToPeers()
	//启动程序的命令行输入环境
	go cli.ReceiveCMD()

	go streams()
	signalHandle()
}

func streams() {
	activeStreams = make(map[string]network.Stream)
	var packet = make([]byte, 1420)
	for {
		// Read in a packet from the tun device.
		if tunDev == nil {
			if len(TunPeerPool) != 0 {
				if runtime.GOOS == "darwin" {
					// Grab ip address of only peer in config
					// var destPeer string
					// for _, ip := range peerPool {
					// 	// /ip4/10.181.212.66/tcp/9001 extract ip address
					// 	ipString := ip.Addrs[0].String()
					// 	destPeer = ipString[strings.Index(ipString, "ip4/")+len("ip4/") : strings.Index(ipString, "/tcp")]
					// 	fmt.Printf("destPeer: %s\n", destPeer)
					// }
					var dstAddress string
					for dst := range TunPeerPool {
						dstAddress = dst
					}
					var err error
					// Create new TUN device
					tunDev, err = tun.New(
						InterfaceName,
						tun.Address(InterfaceAddress),
						tun.DestAddress(dstAddress),
						tun.MTU(1420),
					)
					if err != nil {
						fmt.Printf("unable to create tun device: %s\n", err)
					}
					err = tunDev.Up()
					if err != nil {
						fmt.Printf("unable to bring up tun device: %s\n", err)
					}
				}
			}
			time.Sleep(time.Second)
			continue
		}
		plen, err := tunDev.Iface.Read(packet)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Decode the packet's destination address
		dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()
		fmt.Printf("dst: %s, :plen: %d\n", dst, plen)
		// Check if we already have an open connection to the destination peer.
		stream, ok := activeStreams[dst]
		if ok {
			// Write out the packet's length to the libp2p stream to ensure
			// we know the full size of the packet at the other end.
			err = binary.Write(stream, binary.LittleEndian, uint16(plen))
			if err == nil {
				// Write the packet out to the libp2p stream.
				// If everyting succeeds continue on to the next packet.
				_, err = stream.Write(packet[:plen])
				if err == nil {
					continue
				}
			}
			// If we encounter an error when writing to a stream we should
			// close that stream and delete it from the active stream map.
			stream.Close()
			delete(activeStreams, dst)
		}

		// Check if the destination of the packet is a known peer to
		// the interface.
		if peer, ok := TunPeerPool[dst]; ok {
			stream, err = h.NewStream(ctx, peer, protocol.ID(VPNID))
			// stream, err = libhost.NewStream(ctx, peer, p2p.Protocol)
			if err != nil {
				continue
			}
			// Write packet length
			err = binary.Write(stream, binary.LittleEndian, uint16(plen))
			if err != nil {
				stream.Close()
				continue
			}
			// Write the packet
			_, err = stream.Write(packet[:plen])
			if err != nil {
				stream.Close()
				continue
			}

			// If all succeeds when writing the packet to the stream
			// we should reuse this stream by adding it active streams map.
			activeStreams[dst] = stream
		}
	}
}

// 启动mdns寻找p2p网络 并等节点连接
func findP2PPeer() {
	PeerChan = initMDNS(ctx, localHost, RendezvousString)
	for {
		peer := <-PeerChan // will block untill we discover a peer
		fmt.Print("Found peer:", peer, "\n")
		//将发现的节点加入节点池
		peerPool[fmt.Sprint(peer.ID)] = peer
	}
}

// 一个监测程序,监测当前网络中已发现的节点
func monitorP2PNodes() {
	currentPeerPoolNum := 0
	for {
		peerPoolNum := len(peerPool)

		if peerPoolNum != currentPeerPoolNum && peerPoolNum != 0 {
			fmt.Printf("-------- P2P node pool has changed, current node pool has %d nodes --------\n", peerPoolNum)
			for _, v := range peerPool {
				fmt.Println("|   ", v.ID.String(), "   |")
			}
			fmt.Printf("----------------------------------\n")
			currentPeerPoolNum = peerPoolNum
			send.SendVersionToPeers(block.NewestBlockHeight)
		} else if peerPoolNum != currentPeerPoolNum && peerPoolNum == 0 {
			fmt.Println("-------- P2P node pool has changed, current node pool has no nodes --------")

			currentPeerPoolNum = peerPoolNum
			fmt.Printf("----------------------------------\n")
		}
		time.Sleep(time.Second)
	}
}

// // 向其他p2p节点发送高度信息，来进行更新区块数据
// func sendVersionToPeers() {
// 	//如果节点池中还未存在节点的话,一直循环 直到发现已连接节点
// 	for {
// 		if len(peerPool) == 0 {
// 			time.Sleep(time.Second)
// 			continue
// 		} else {
// 			break
// 		}
// 	}
// 	send.SendVersionToPeers(block.NewestBlockHeight)
// }

// 节点退出信号处理
func signalHandle() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	send.SendSignOutToPeers()
	fmt.Println("local node exit!")
	time.Sleep(time.Second)
	os.Exit(0)
}
