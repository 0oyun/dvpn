package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "miniDVPN",
		Example: "miniDVPN",
	}

	version = &cobra.Command{
		Use:     "version",
		Short:   "get current version",
		Example: "miniDVPN version",
	}

	initialize = &cobra.Command{
		Use:     "init",
		Short:   "initialize the miniDVPN Config",
		Example: "miniDVPN init -i peer1 -p 8080 -a 10.1.1.2/24",
	}

	start = &cobra.Command{
		Use:     "start",
		Short:   "start a miniDVPN daemon",
		Example: "miniDVPN start",
	}

	close = &cobra.Command{
		Use:     "close",
		Short:   "close a miniDVPN daemon",
		Example: "miniDVPN close -i peer1",
	}
	// --p
	enableDebug = rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")

	// init flags
	port          = initialize.PersistentFlags().IntP("port", "p", 8080, "listen port of server")
	interfaceName = initialize.PersistentFlags().StringP("interface", "i", "peer0", "interface name of miniDVPN")
	address       = initialize.PersistentFlags().StringP("address", "a", "10.1.1.1/24", "inner address of miniDVPN")

	// start flags
	startInterfaceName = start.PersistentFlags().StringP("interface", "i", "peer0", "interface name of miniDVPN")
	foreground         = start.PersistentFlags().BoolP("foreground", "f", false, "run in foreground")

	// close flags
	closeInterfaceName = close.PersistentFlags().StringP("interface", "i", "peer0", "interface name of miniDVPN")
)
