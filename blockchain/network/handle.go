package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/toy-playground/miniDVPN/blockchain/block"
)

func handleStream(stream network.Stream) {
	data, err := io.ReadAll(stream)
	if err != nil {
		panic(err)
	}
	cmd, content := splitMessage(data)
	fmt.Printf("received CMD: %s\n", cmd)
	switch command(cmd) {
	case cVersion:
		go handleVersion(content)
	case cGetHash:
		go handleGetHash(content)
	case cHashMap:
		go handleHashMap(content)
	case cGetBlock:
		go handleGetBlock(content)
	case cBlock:
		go handleBlock(content)
	case cTransaction:
		go handleTransaction(content)
	case cExit:
		go handleExit(content)
	default:
		fmt.Printf("received unknown CMD: %s\n", data)
		var packet = make([]byte, 1420)
		var packetSize = make([]byte, 2)
		for {
			// reading packetSize from data
			// unshift the packetSize from data
			packetSize, data = data[:2], data[2:]

			// Decode the incoming packet's size from binary.
			size := binary.LittleEndian.Uint16(packetSize)
			packet, data = data[:size], data[size:]
			if tunDev != nil {
				tunDev.Iface.Write(packet[:size])
			}
		}
	}
}

func vpnStreamHandler(stream network.Stream) {
	var packet = make([]byte, 1420)
	fmt.Printf("vpn stream handler : %s\n", stream.Conn().RemotePeer().String())
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

func handleExit(content []byte) {
	e := Exit{}
	e.deserialize(content)
	fmt.Printf("receive exit message from %s, ready to delete this node from peerPool\n", e.Addrfrom)
	peer := buildPeerInfoByAddr(e.Addrfrom)
	delete(peerPool, fmt.Sprint(peer.ID))
}

func handleTransaction(content []byte) {
	t := Transactions{}
	t.Deserialize(content)
	if len(t.Txs) == 0 {
		return
	}
	for _, tx := range t.Txs {
		if string(tx.TxType) == "addClient" {
			fmt.Printf("received addClient transaction,")
			a := string(tx.Data)[:strings.Index(string(tx.Data), "-,-")]
			ipInfo := buildPeerInfoByAddr(a)
			// peerPool[fmt.Sprint(ipInfo.ID)] = ipInfo
			b := string(tx.Data)[strings.Index(string(tx.Data), "-,-")+3:]
			TunPeerPool[b] = ipInfo.ID
			fmt.Printf("client info: %s, %s\n", a, b)

		}
	}
	mineBlock(t)
}

var lock = sync.Mutex{}

func mineBlock(t Transactions) {
	lock.Lock()
	defer lock.Unlock()

	bc := block.NewBlockchain()
	for {
		currentHeight := bc.GetLastBlockHeight()
		if currentHeight >= block.NewestBlockHeight {
			break
		}
		time.Sleep(time.Second * 1)
	}
	nTs := make([]block.Transaction, len(t.Txs))
	for i := range t.Txs {
		nTs[i].TxType = block.TransactionType(t.Txs[i].TxType)
		nTs[i].Data = t.Txs[i].Data
	}
	bc.SendTx(nTs, send)
}

func handleBlock(content []byte) {
	b := &block.Block{}
	b.Deserialize(content)
	fmt.Printf("current node has received block from other node, the block hash is: %x\n", b.Hash)
	bc := block.NewBlockchain()

	if b.Verify() {
		fmt.Printf("block hash verify success, ready to add this block to local database\n")

		currentHash := bc.GetBlockHashByHeight(b.Height)
		if b.Height == 1 && currentHash == nil {
			bc.AddBlock(b)
			fmt.Printf("genesis block added to local database\n")
		}
		lastBlockHash := bc.GetBlockHashByHeight(b.Height - 1)
		if lastBlockHash == nil {
			for {
				time.Sleep(time.Second)
				lastBlockHash = bc.GetBlockHashByHeight(b.Height - 1)
				if lastBlockHash != nil {
					fmt.Printf("Block height %d is not synchronized, waiting for synchronization...\n", b.Height-1)
					break
				}
			}
		}
		if bytes.Equal(lastBlockHash, b.PreHash) {
			bc.AddBlock(b)
			fmt.Printf(" block added to local database\n")
		} else {
			fmt.Println("prev hash verify failed")
		}
	} else {
		fmt.Println("block hash verify failed")
	}
}

func handleGetBlock(content []byte) {
	g := getBlock{}
	g.deserialize(content)
	bc := block.NewBlockchain()
	blockBytes := bc.GetBlockByHash(g.BlockHash)
	data := jointMessage(cBlock, blockBytes)
	fmt.Println("ready to send block to other node, the block hash is: ", g.BlockHash)
	send.SendMessage(buildPeerInfoByAddr(g.AddrFrom), data)
}

func handleHashMap(content []byte) {
	h := hash{}
	h.deserialize(content)
	hm := h.HashMap
	bc := block.NewBlockchain()
	lastHeight := bc.GetLastBlockHeight()
	targetHeight := lastHeight + 1
	for {
		hash := hm[targetHeight]
		if hash == nil {
			break
		}
		g := getBlock{hash, localAddr}
		data := jointMessage(cGetBlock, g.serialize())
		send.SendMessage(buildPeerInfoByAddr(h.AddrFrom), data)
		targetHeight++
	}
}

func handleGetHash(content []byte) {
	g := getHash{}
	g.deserialize(content)
	bc := block.NewBlockchain()
	lastHeight := bc.GetLastBlockHeight()
	hm := hashMap{}
	for i := g.Height + 1; i <= lastHeight; i++ {
		hm[i] = bc.GetBlockHashByHeight(i)
	}
	h := hash{hm, localAddr}
	data := jointMessage(cHashMap, h.serialize())
	send.SendMessage(buildPeerInfoByAddr(g.AddrFrom), data)
}

func handleVersion(content []byte) {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	v := version{}
	v.deserialize(content)
	dstAddr := v.TunAddrFrom[:strings.Index(v.TunAddrFrom, "/")]
	if _, ok := TunPeerPool[dstAddr]; !ok {
		TunPeerPool[dstAddr] = buildPeerInfoByAddr(v.AddrFrom).ID
	}
	bc := block.NewBlockchain()
	fmt.Printf("received Version: %d, current Height: %d\n", v.Height, bc.GetLastBlockHeight())
	if block.NewestBlockHeight > v.Height {
		for {
			currentHeight := bc.GetLastBlockHeight()
			if currentHeight < block.NewestBlockHeight {
				time.Sleep(time.Second)
			} else {
				newV := version{versionInfo, currentHeight, localAddr, InterfaceAddress}
				data := jointMessage(cVersion, newV.serialize())
				send.SendMessage(buildPeerInfoByAddr(v.AddrFrom), data)
				break
			}
		}
	} else if block.NewestBlockHeight < v.Height {
		gh := getHash{block.NewestBlockHeight, localAddr}
		block.NewestBlockHeight = v.Height
		data := jointMessage(cGetHash, gh.serialize())
		send.SendMessage(buildPeerInfoByAddr(v.AddrFrom), data)
	} else {
		fmt.Printf("block height are equal \n")
	}
}
