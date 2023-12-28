package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Cli struct {
}

func printUsage() {
	fmt.Println("--------------------------------------------- ")
	fmt.Println("Usage:")
	fmt.Println("\thelp                                             ")
	fmt.Println("\tgenesis                                          ")
	fmt.Println("\tsendTx -t TYPE -d DATA                          ")
	fmt.Println("\tprintAll                                  ")
	fmt.Println("--------------------------------------------- ")
}

func New() *Cli {
	return &Cli{}
}

func (cli *Cli) Run() {
	printUsage()
	go cli.startNode()
	cli.ReceiveCMD()
}

func (cli Cli) ReceiveCMD() {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
		cli.userCmdHandle(sendData)
	}
}

func (cli Cli) userCmdHandle(data string) {
	data = strings.TrimSpace(data)
	var cmd string
	var context string
	if strings.Contains(data, " ") {
		cmd = data[:strings.Index(data, " ")]
		context = data[strings.Index(data, " ")+1:]
	} else {
		cmd = data
	}
	switch cmd {
	case "help":
		printUsage()
	case "genesis":
		cli.genesis()
	case "printAll":
		cli.printAllBlocks()
	case "addClient":
		peer := strings.TrimSpace(context[strings.Index(context, "-p")+len("-p") : strings.Index(context, "-v")])
		virtualAddress := strings.TrimSpace(context[strings.Index(context, "-v")+len("-v"):])
		cli.sendTx("addClient", peer+"-,-"+virtualAddress)
	case "sendTx":
		txType := strings.TrimSpace(context[strings.Index(context, "-t")+len("-t") : strings.Index(context, "-d")])
		data := strings.TrimSpace(context[strings.Index(context, "-d")+len("-d"):])
		cli.sendTx(txType, data)
	default:
		printUsage()
	}
}
