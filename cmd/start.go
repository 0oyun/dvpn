package cmd

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/nxadm/tail"

	"github.com/toy-playground/miniDVPN/config"
	"github.com/toy-playground/miniDVPN/p2p"
	"github.com/toy-playground/miniDVPN/tun"
)

var (
	// iface is the tun device used to pass packets between
	// miniDVPN and the user's machine.
	tunDev *tun.TUN
	// RevLookup allow quick lookups of an incoming stream
	// for security before accepting or responding to any data.
	RevLookup map[string]string
	// activeStreams is a map of active streams to a peer
	activeStreams map[string]network.Stream
)

// 生成p2p2节点的配置文件
func StartRun() {
	configFileName := *startInterfaceName + ".yaml"
	cfg, err := config.Read(configFileName)
	if err != nil {
		panic(err)
	}
	if !*foreground {
		if err := createDaemon(cfg); err != nil {
			fmt.Println("[+] Failed to Create miniDVPN Daemon")
			fmt.Println(err)
			return
		} else {
			fmt.Println("[+] Successfully Created miniDVPN Daemon")
			return
		}
	}

	// Setup reverse lookup hash map for authentication.
	RevLookup = make(map[string]string, len(cfg.Peers))
	for ip, id := range cfg.Peers {
		RevLookup[id.ID] = ip
	}

	fmt.Println("[+] Creating TUN Device")

	// macOS只能有一个peer
	// 用linux做测试
	if runtime.GOOS == "darwin" {
		if len(cfg.Peers) > 1 {
			fmt.Printf("darwin only supports one peer")
		}

		// Grab ip address of only peer in config
		var destPeer string
		for ip := range cfg.Peers {
			destPeer = ip
		}

		// Create new TUN device
		tunDev, err = tun.New(
			cfg.Interface.Name,
			tun.Address(cfg.Interface.Address),
			tun.DestAddress(destPeer),
			tun.MTU(1420),
		)
	} else {
		// Create new TUN device
		tunDev, err = tun.New(
			cfg.Interface.Name,
			tun.Address(cfg.Interface.Address),
			tun.MTU(1420),
		)
	}
	if err != nil {
		panic(err)
	}

	// Setup System Context
	ctx := context.Background()

	fmt.Println("[+] Creating LibP2P Node")

	// Check that the listener port is available.
	port, err := verifyPort(cfg.Interface.ListenPort)
	if err != nil {
		panic(err)
	}

	// Create P2P Node
	host, dht, err := p2p.CreateNode(
		ctx,
		cfg.Interface.PrivateKey,
		port,
		streamHandler,
		cfg.Interface.RemoteAddress,
	)
	if err != nil {
		panic(err)
	}

	// Setup Peer Table for Quick Packet --> Dest ID lookup
	peerTable := make(map[string]peer.ID)
	for ip, id := range cfg.Peers {
		peerTable[ip], err = peer.Decode(id.ID)

		if err != nil {
			panic(err)
		}

	}

	fmt.Println("[+] Setting Up Node Discovery via DHT")

	// Setup P2P Discovery
	go p2p.Discover(ctx, host, dht, peerTable)
	go prettyDiscovery(ctx, host, peerTable)

	// Configure path for lock
	lockPath := cfg.Interface.Name + ".lock"

	// Register the application to listen for SIGINT/SIGTERM
	go signalExit(host, lockPath)

	// Write lock to filesystem to indicate an existing running daemon.
	err = os.WriteFile(lockPath, []byte(fmt.Sprint(os.Getpid())), os.ModePerm)

	// Bring Up TUN Device
	err = tunDev.Up()
	if err != nil {
		panic("unable to bring up tun device")
	}

	fmt.Println("[+] Network Setup Complete...Waiting on Node Discovery")

	// + ----------------------------------------+
	// | Listen For New Packets on TUN Interface |
	// + ----------------------------------------+

	// Initialize active streams map and packet byte array.
	activeStreams = make(map[string]network.Stream)
	var packet = make([]byte, 1420)
	for {
		// Read in a packet from the tun device.
		plen, err := tunDev.Iface.Read(packet)
		if err != nil {
			log.Println(err)
			continue
		}

		// Decode the packet's destination address
		dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()

		fmt.Println("dst: ", dst, "plen: ", plen)
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
		if peer, ok := peerTable[dst]; ok {
			stream, err = host.NewStream(ctx, peer, p2p.Protocol)
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

// singalExit registers two syscall handlers on the system  so that if
// an SIGINT or SIGTERM occur on the system miniDVPN can gracefully
// shutdown and remove the filesystem lock file.
func signalExit(host host.Host, lockPath string) {
	// Wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	// Shut the node down
	err := host.Close()
	if err != nil {
		panic(err)
	}

	// Remove daemon lock from file system.
	err = os.Remove(lockPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("Received signal, shutting down...")

	// Exit the application.
	os.Exit(0)
}

// createDaemon handles creating an independent background process for a
// miniDVPN daemon from the original parent process.
func createDaemon(cfg *config.Config) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	// Generate log path
	logPath := cfg.Interface.Name + ".log"

	// Create Pipe to monitor for daemon output.
	f, err := os.Create(logPath)
	if err != nil {
		return err
	}

	// Create Sub Process
	process, err := os.StartProcess(
		path,
		append(os.Args, "-f"),
		&os.ProcAttr{
			Dir:   ".",
			Env:   os.Environ(),
			Files: []*os.File{nil, f, f},
		},
	)
	if err != nil {
		return err
	}

	// Listen to the child process's log output to determine
	// when the daemon is setup and connected to a set of peers.
	count := 0
	deadlineHit := false
	countChan := make(chan int)
	go func(out chan<- int) {
		numConnected := 0
		t, err := tail.TailFile(logPath, tail.Config{Follow: true})
		if err != nil {
			out <- numConnected
			return
		}
		for line := range t.Lines {
			fmt.Println(line.Text)
			if strings.HasPrefix(line.Text, "[+] Connection to") {
				numConnected++
				if numConnected >= len(cfg.Peers) {
					break
				}
			}
		}
		out <- numConnected
	}(countChan)

	// Block until all clients are connected or for a maximum of 30s.
	select {
	case _, deadlineHit = <-time.After(120 * time.Second):
	case count = <-countChan:
	}

	// Release the created daemon
	err = process.Release()
	if err != nil {
		return err
	}

	// Check if the daemon exited prematurely
	if !deadlineHit && count < len(cfg.Peers) {
		return errors.New("failed to create daemon")
	}
	return nil
}

func streamHandler(stream network.Stream) {
	// If the remote node ID isn't in the list of known nodes don't respond.
	if _, ok := RevLookup[stream.Conn().RemotePeer().String()]; !ok {
		stream.Reset()
		return
	}
	var packet = make([]byte, 1420)
	var packetSize = make([]byte, 2)
	for {
		// Read the incoming packet's size as a binary value.
		_, err := stream.Read(packetSize)
		if err != nil {
			stream.Close()
			return
		}

		// Decode the incoming packet's size from binary.
		size := binary.LittleEndian.Uint16(packetSize)

		// Read in the packet until completion.
		var plen uint16 = 0
		for plen < size {
			tmp, err := stream.Read(packet[plen:size])
			plen += uint16(tmp)
			if err != nil {
				stream.Close()
				return
			}
		}
		tunDev.Iface.Write(packet[:size])
	}
}

func prettyDiscovery(ctx context.Context, node host.Host, peerTable map[string]peer.ID) {
	// Build a temporary map of peers to limit querying to only those
	// not connected.
	tempTable := make(map[string]peer.ID, len(peerTable))
	for ip, id := range peerTable {
		tempTable[ip] = id
	}
	for len(tempTable) > 0 {
		for ip, id := range tempTable {
			stream, err := node.NewStream(ctx, id, p2p.Protocol)
			if err != nil && (strings.HasPrefix(err.Error(), "failed to dial") ||
				strings.HasPrefix(err.Error(), "no addresses")) {
				// Attempt to connect to peers slowly when they aren't found.
				time.Sleep(5 * time.Second)
				continue
			}
			if err == nil {
				fmt.Printf("[+] Connection to %s Successful. Network Ready.\n", ip)
				stream.Close()
			}
			delete(tempTable, ip)
		}
	}
}

func verifyPort(port int) (int, error) {
	var ln net.Listener
	var err error

	// If a user manually sets a port don't try to automatically
	// find an open port.
	if port != 8001 {
		ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return port, errors.New("could not create node, listen port already in use by something else")
		}
	} else {
		// Automatically look for an open port when a custom port isn't
		// selected by a user.
		for {
			ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
			if err == nil {
				break
			}
			if port >= 65535 {
				return port, errors.New("failed to find open port")
			}
			port++
		}
	}
	if ln != nil {
		ln.Close()
	}
	return port, nil
}
