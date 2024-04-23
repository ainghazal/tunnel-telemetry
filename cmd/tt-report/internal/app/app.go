/*
Copyright © 2024 ain ghazal <ain@openobservatory.org>
*/
package app

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// TODO: pass report from json file in disk.
	reportFile string
)

type flag int

const (
	flagDebug flag = iota
	flagSkipGeolocation
)

var allFlags = map[flag]string{
	flagDebug:           "debug",
	flagSkipGeolocation: "skip-geolocation",
}

func (f flag) String() string {
	if name, ok := allFlags[f]; ok {
		return name
	}
	return ""
}

type config struct {
	Debug           bool
	SkipGeolocation bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tt-server",
	Short: "Run a tunnel-telemetry collector server",
	Long: `Tunnel-telemetry collector server.

A collector server receives reports from tunnel clients,
and optionally stores them and/or relays them to an upstream collector.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &config{
			Debug:           viper.GetBool(flagDebug.String()),
			SkipGeolocation: viper.GetBool(flagSkipGeolocation.String()),
		}
		processAndSubmitReport(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolP(flagDebug.String(), "d", false, "set debug level in logs")
	rootCmd.Flags().BoolP(flagSkipGeolocation.String(), "", false, "skip geolocation using stun/https apis")
}

// initConfig reads config file and any relevant ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read any environment variables that match

	for _, flg := range allFlags {
		viper.BindPFlag(flg, rootCmd.Flags().Lookup(flg))
	}

	/*
		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	*/
}
