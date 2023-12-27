package network

import (
	"bufio"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/toy-playground/miniDVPN/blockchain/block"
)

type Send struct {
}

func (s Send) SendSignOutToPeers() {
	m := Exit{localAddr}
	data := jointMessage(cExit, m.serialize())
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
}

// 向网络中其他节点发送高度信息
func (s Send) SendVersionToPeers(lastHeight int) {
	newV := version{versionInfo, lastHeight, localAddr}
	data := jointMessage(cVersion, newV.serialize())
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
	fmt.Printf("send height %d to other node\n", lastHeight)
}

// 向网络中其他节点发送交易信息
func (s Send) SendTransToPeers(ts []block.Transaction) {
	nts := make([]Transaction, len(ts))
	for i := range ts {
		nts[i].AddrFrom = localAddr
		nts[i].Data = ts[i].Data
		nts[i].TxType = TransactionType(ts[i].TxType)
	}
	tss := Transactions{nts}
	go handleTransaction(tss.Serialize())
	data := jointMessage(cTransaction, tss.Serialize())
	fmt.Printf("ready to send %d tx to other node\n", len(tss.Txs))
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
}

// 基础发送信息方法
func (Send) SendMessage(peer peer.AddrInfo, data []byte) {
	//连接传入的对等节点
	if err := localHost.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:\n", err.Error())
	}
	stream, err := localHost.NewStream(ctx, peer.ID, protocol.ID(ProtocolID))
	if err != nil {
		fmt.Println("Stream open failed\n", err.Error())
	} else {
		// cmd, _ := splitMessage(data)
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		_, err := rw.Write(data)
		if err != nil {
			panic(err)
		}
		//向流中写入所有缓冲数据
		err = rw.Flush()
		if err != nil {
			panic(err)
		}
		//关闭流，完成一次信息的发送
		err = stream.Close()
		if err != nil {
			panic(err)
		}
		// fmt.Printf("send cmd:%s to peer:%v\n", cmd, peer)
	}
}
