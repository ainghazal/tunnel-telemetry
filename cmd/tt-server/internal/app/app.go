/*
Copyright © 2024 ain ghazal <ain@openobservatory.org>
*/
package app

import (
	"fmt"
	"os"

	"github.com/ainghazal/tunnel-telemetry/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	defaultConfigFile = "/etc/tunneltelemetry/config.yaml"
	defaultCacheDir   = "/var/www/.cache"
	defaultHTTPAddr   = ":8080"
	defaultHTTPSAddr  = ":443"
)

type flag int

const (
	flagAllowPublicEndpoint flag = iota
	flagAutoTLS
	flagAutoTLSCacheDir
	flagCollectorID
	flagDebug
	flagDebugGeolocation
	flagHostname
	flagListenAddr
)

var allFlags = map[flag]string{
	flagAllowPublicEndpoint: "allow-public-endpoint",
	flagAutoTLS:             "autotls",
	flagAutoTLSCacheDir:     "autotls-cache-dir",
	flagCollectorID:         "collector-id",
	flagDebug:               "debug",
	flagDebugGeolocation:    "debug-geolocation",
	flagHostname:            "hostname",
	flagListenAddr:          "listen",
}

func (f flag) String() string {
	if name, ok := allFlags[f]; ok {
		return name
	}
	return ""
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tt-server",
	Short: "Run a tunnel-telemetry collector server",
	Long: `Tunnel-telemetry collector server.

A collector server receives reports from tunnel clients,
and optionally stores them and/or relays them to an upstream collector.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &config.Config{
			AllowPublicEndpoint: viper.GetBool(flagAllowPublicEndpoint.String()),
			AutoTLS:             viper.GetBool(flagAutoTLS.String()),
			AutoTLSCacheDir:     viper.GetString(flagAutoTLSCacheDir.String()),
			CollectorID:         viper.GetString(flagCollectorID.String()),
			Debug:               viper.GetBool(flagDebug.String()),
			DebugGeolocation:    viper.GetBool(flagDebugGeolocation.String()),
			Hostname:            viper.GetString(flagHostname.String()),
			ListenAddr:          viper.GetString(flagListenAddr.String()),
		}

		if cfg.AutoTLS && cfg.Hostname == "" {
			fmt.Println("ERROR: empty --hostname")
			os.Exit(1)
		}

		if cfg.ListenAddr == "" {
			if cfg.AutoTLS {
				cfg.ListenAddr = defaultHTTPSAddr
			} else {
				cfg.ListenAddr = defaultHTTPAddr
			}
		}

		startEchoServer(cfg)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile, "config file")

	rootCmd.Flags().BoolP(flagAllowPublicEndpoint.String(), "", false, "allow publishing of the endpoints IP")
	rootCmd.Flags().BoolP(flagAutoTLS.String(), "", false, "use autotls to manage LetsEncrypt Certificates (default: false)")
	rootCmd.Flags().StringP(flagAutoTLSCacheDir.String(), "", defaultCacheDir, "dir to cache autotls material")
	rootCmd.Flags().StringP(flagCollectorID.String(), "", "", "collector ID to add to enrich reports with")
	rootCmd.Flags().BoolP(flagDebug.String(), "d", false, "set debug level in logs")
	rootCmd.Flags().BoolP(flagDebugGeolocation.String(), "", false, "get real IP from headers (potentially insecure!)")
	rootCmd.Flags().StringP(flagHostname.String(), "", "", "hostname (for autotls certs)")
	rootCmd.Flags().StringP(flagListenAddr.String(), "", "", "address to listen on (:8080 or :443 if autotls is set)")
}

// initConfig reads config file and any relevant ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(defaultConfigFile)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv() // read any environment variables that match

	for _, flg := range allFlags {
		viper.BindPFlag(flg, rootCmd.Flags().Lookup(flg))
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
