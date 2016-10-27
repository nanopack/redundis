// Package config is the central location for all configuration.
package config

import (
	"path/filepath"
	"time"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ListenAddress    = "127.0.0.1:6379"  // Redundis listen address
	SentinelAddress  = "127.0.0.1:26379" // Address of sentinel node
	SentinelPassword = ""                // Sentinel password
	MonitorName      = "test"            // Name of sentinel monitor

	masterWait   = 30
	notReady     = 30
	sentinelPoll = 10

	TimeoutMasterWait   = time.Duration(masterWait) * time.Second   // Time to wait for node to transition to master (seconds)
	TimeoutNotReady     = time.Duration(notReady) * time.Second     // Time to wait to connect to redis|sentinel (seconds)
	TimeoutSentinelPoll = time.Duration(sentinelPoll) * time.Second // Time to wait for sentinel to respond (seconds)

	LogLevel = "info" // Log level [fatal, error, info, debug, trace]
	Log      lumber.Logger
)

func init() {
	Log = lumber.NewConsoleLogger(lumber.LvlInt("info"))
}

// AddFlags lets cobra know what flags to parse.
func AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&ListenAddress, "listen-address", "l", ListenAddress, "Redundis listen address")
	cmd.Flags().StringVarP(&SentinelAddress, "sentinel-address", "s", SentinelAddress, "Address of sentinel node")
	cmd.Flags().StringVarP(&SentinelPassword, "sentinel-password", "p", SentinelPassword, "Sentinel password")
	cmd.Flags().StringVarP(&MonitorName, "monitor-name", "m", MonitorName, "Name of sentinel monitor")

	// timeouts
	cmd.Flags().IntVarP(&masterWait, "master-wait", "t", masterWait, "Time to wait for node to transition to master (seconds)")
	cmd.Flags().IntVarP(&notReady, "ready-wait", "r", notReady, "Time to wait to connect to redis|sentinel (seconds)")
	cmd.Flags().IntVarP(&sentinelPoll, "sentinel-wait", "w", sentinelPoll, "Time to wait for sentinel to respond (seconds)")

	cmd.Flags().StringVarP(&LogLevel, "log-level", "L", LogLevel, "Log level [fatal, error, info, debug, trace]")

	TimeoutMasterWait = time.Duration(masterWait) * time.Second
	TimeoutNotReady = time.Duration(notReady) * time.Second
	TimeoutSentinelPoll = time.Duration(sentinelPoll) * time.Second
}

// ReadConfigFile reads a specified config file, if set, overriding flag settings.
func ReadConfigFile(configFile string) error {
	if configFile == "" {
		return nil
	}

	// Set defaults to whatever might be there already
	viper.SetDefault("listen-address", ListenAddress)
	viper.SetDefault("sentinel-address", SentinelAddress)
	viper.SetDefault("sentinel-password", SentinelPassword)
	viper.SetDefault("monior-name", MonitorName)
	viper.SetDefault("master-wait", masterWait)
	viper.SetDefault("ready-wait", notReady)
	viper.SetDefault("sentinel-wait", sentinelPoll)
	viper.SetDefault("log-level", LogLevel)

	filename := filepath.Base(configFile)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(configFile))

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Set values. Config file will override commandline
	ListenAddress = viper.GetString("listen-address")
	SentinelAddress = viper.GetString("sentinel-address")
	SentinelPassword = viper.GetString("sentinel-password")
	MonitorName = viper.GetString("monior-name")
	TimeoutMasterWait = time.Duration(viper.GetInt("master-wait")) * time.Second
	TimeoutNotReady = time.Duration(viper.GetInt("ready-wait")) * time.Second
	TimeoutSentinelPoll = time.Duration(viper.GetInt("sentinel-wait")) * time.Second
	LogLevel = viper.GetString("log-level")

	return nil
}
