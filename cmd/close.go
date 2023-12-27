package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/toy-playground/miniDVPN/tun"
)

// 结束vpn守护进程
func CloseRun() {
	lockFileName := *closeInterfaceName + ".lock"

	out, err := os.ReadFile(lockFileName)
	if err != nil {
		panic(err)
	}

	pid, err := strconv.Atoi(string(out))
	if err != nil {
		panic(err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		panic(err)
	}

	err0 := process.Signal(os.Interrupt)

	var err1 error
	if runtime.GOOS != "darwin" {
		err1 = tun.Delete(*closeInterfaceName)
	}

	if err0 != nil || err1 != nil {
		fmt.Println("[+] failed to close miniVPN " + *closeInterfaceName + " daemon\n")
		if err0 != nil {
			fmt.Println(err0)
		}
		if err1 != nil {
			fmt.Println(err1)
		}

	}

	fmt.Println("[+] deleted miniVPN " + *closeInterfaceName + " daemon")

}
