package network

import "github.com/libp2p/go-libp2p/core/host"

var (
	RendezvousString = "miniDVPNRendezous"
	ProtocolID       = "/miniDVPN/1.0.0"
	ListenHost       = "0.0.0.0"
	ListenPort       = "4001"
	localHost        host.Host
	localAddr        string
)

const versionInfo = byte(0x00)

const prefixCMDLength = 12

type command string

const (
	cVersion     command = "version"
	cGetHash     command = "getHash"
	cHashMap     command = "hashMap"
	cGetBlock    command = "getBlock"
	cBlock       command = "block"
	cTransaction command = "transaction"
	cExit        command = "exit"
)
