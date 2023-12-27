package config

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for a libp2p node.

type Config struct {
	Interface Interface       `yaml:"interface"`
	Peers     map[string]Peer `yaml:"peers"`
}

type Interface struct {
	Name       string `yaml:"name"`
	ID         string `yaml:"id"`
	ListenPort int    `yaml:"listen_port"`
	Address    string `yaml:"address"`
	PrivateKey string `yaml:"private_key"`
}

type Peer struct {
	ID string `yaml:"id"`
}

func Read(filename string) (*Config, error) {
	in, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := Config{
		Interface: Interface{
			Name:       "peer0",
			ID:         "",
			ListenPort: 8080,
			Address:    "10.1.1.1/24",
			PrivateKey: "",
		},
	}

	err = yaml.Unmarshal(in, &result)
	if err != nil {
		return nil, err
	}

	for ip := range result.Peers {
		if net.ParseIP(ip) == nil {
			return nil, fmt.Errorf("%s is not a valid ip address", ip)
		}
	}
	return &result, nil
}
