package network

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

func buildPeerInfoByAddr(addrs string) peer.AddrInfo {
	pi, err := peer.AddrInfoFromString(addrs)
	if err != nil {
		fmt.Printf("buildPeerInfoByAddr err:%s\n", err)
	}
	return *pi
}

func jointMessage(cmd command, content []byte) []byte {
	b := make([]byte, prefixCMDLength)
	for i, v := range []byte(cmd) {
		b[i] = v
	}
	joint := make([]byte, 0)
	joint = append(b, content...)
	return joint
}

func splitMessage(message []byte) (cmd string, content []byte) {
	cmdBytes := message[:prefixCMDLength]
	newCMDBytes := make([]byte, 0)
	for _, v := range cmdBytes {
		if v != byte(0) {
			newCMDBytes = append(newCMDBytes, v)
		}
	}
	cmd = string(newCMDBytes)
	content = message[prefixCMDLength:]
	return
}
