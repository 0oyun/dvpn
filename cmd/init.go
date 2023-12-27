package cmd

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"gopkg.in/yaml.v3"

	"github.com/toy-playground/miniDVPN/config"
)

// 生成p2p2节点的配置文件
func InitRun() {

	configFileName := *interfaceName + ".yaml"

	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		panic(err)
	}
	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		panic(err)
	}

	b, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("peer id %s\n", id.String())

	new := config.Config{
		Interface: config.Interface{
			Name:       *interfaceName,
			ID:         id.String(),
			ListenPort: *port,
			Address:    *address,
			PrivateKey: string(b),
		},
	}

	out, err := yaml.Marshal(&new)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(configFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(out)
	if err != nil {
		panic(err)
	}
	// Print config creation message to user
	fmt.Printf("Initialized new config at %s\n", configFileName)

}
