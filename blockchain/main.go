package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/toy-playground/miniDVPN/blockchain/block"
	"github.com/toy-playground/miniDVPN/blockchain/cli"
	"github.com/toy-playground/miniDVPN/blockchain/database"
	"github.com/toy-playground/miniDVPN/blockchain/network"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	listenHost := viper.GetString("network.listen_host")
	listenPort := viper.GetString("network.listen_port")
	rendezvousString := viper.GetString("network.rendezvous_string")
	protocolID := viper.GetString("network.protocol_id")
	virtualHost := viper.GetString("interface.address")
	network.ListenHost = listenHost
	network.RendezvousString = rendezvousString
	network.ProtocolID = protocolID
	network.ListenPort = listenPort
	database.ListenPort = listenPort
	network.VirtualHost = virtualHost
	block.ListenPort = listenPort
}

func main() {
	c := cli.New()
	c.Run()
}
