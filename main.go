package main

import (
	"fmt"

	"github.com/toy-playground/miniDVPN/cmd"
)

func main() {

	err := cmd.InitCmd(debug)
	if err != nil {
		fmt.Println("cmd init fail: ", err)
		return
	}

	err = cmd.GetRootCmd().Execute()
	if err != nil {
		fmt.Println("cmd execute fail: ", err)
	}
}
