// Redundis is a smart, sentinel aware proxy for redis that connects redis
// clients to the correct master node.
//
// Usage
//
// To run as a server, using the defaults, starting redundis is as simple as
//  redundis
// For more specific usage information, refer to the help doc (redundis -h):
//
//  Usage:
//    redundis [flags]
//
//  Flags:
//    -c, --config-file="": Config file location for redundis
//    -l, --listen-address="127.0.0.1:6379": Redundis listen address
//    -L, --log-level="info": Log level [fatal, error, info, debug, trace]
//    -t, --master-wait=30: Time to wait for node to transition to master (seconds)
//    -m, --monitor-name="test": Name of sentinel monitor
//    -r, --ready-wait=30: Time to wait to connect to redis|sentinel (seconds)
//    -s, --sentinel-address="127.0.0.1:26379": Address of sentinel node
//    -p, --sentinel-password="": Sentinel password
//    -w, --sentinel-wait=10: Time to wait for sentinel to respond (seconds)
//
package main

import (
	"os"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/redundis/config"
	"github.com/nanopack/redundis/core"
)

var (
	configFile string // Config file location for redundis

	// redundis command
	Redundis = &cobra.Command{
		Use:   "redundis",
		Short: "redundis redis-proxy",
		Long:  ``,

		Run: startRedundis,
	}
)

func main() {
	Redundis.Flags().StringVarP(&configFile, "config-file", "c", "", "Config file location for redundis")
	config.AddFlags(Redundis)

	Redundis.Execute()
}

// startRedundis reads a specified config file, initializes the logger, and starts
// redundis
func startRedundis(ccmd *cobra.Command, args []string) {
	if err := config.ReadConfigFile(configFile); err != nil {
		config.Log.Fatal("Failed to read config - %v", err)
		os.Exit(1)
	}

	// initialize logger
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	config.Log.Prefix("[redundis]")

	// initialize redundis
	err := redundis.Start()
	if err != nil {
		config.Log.Fatal("Failed to listen - %v", err)
		os.Exit(1)
	}
}
