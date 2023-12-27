// Package cmd use cobra provide cmd function
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	debugF   func()
	branch   string
	commitID string
	date     string
)

// InitCmd init cobra command
func InitCmd(debug func()) error {
	debugF = debug

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
	}

	addRun()

	rootCmd.AddCommand(
		version,
		initialize,
		start,
		close,
	)
	return nil
}

// GetRootCmd return the root cobra.Command
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func addRun() {

	version.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("miniDVPN Version:\n%s-%s-%s\n", branch, date, commitID)
	}

	initialize.Run = func(cmd *cobra.Command, args []string) {
		InitRun()
	}

	start.Run = func(cmd *cobra.Command, args []string) {
		StartRun()
	}

	close.Run = func(cmd *cobra.Command, args []string) {
		CloseRun()
	}

}

func initConfig() {
	if *enableDebug {
		fmt.Println("========DEBUG MODE========")
		debugF()
		time.Sleep(5 * time.Second)
	}
}
