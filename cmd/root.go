package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

var (
	HTTPAddr *net.IP
	HTTPPort *int

	ESAddr *[]net.IP
	Debug  *bool
)

var rootCmd = &cobra.Command{
	Use:   "metrics-meta",
	Short: "A tool that can be used as a component for a metric searching system based on Elasticsearch",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	HTTPAddr = rootCmd.PersistentFlags().IPP("http.addr", "l", net.ParseIP("0.0.0.0"), "Address to bind to.")
	HTTPPort = rootCmd.PersistentFlags().IntP("http.port", "p", 8080, "Port to bind to")
	ESAddr = rootCmd.PersistentFlags().IPSliceP("es.addr", "H", []net.IP{}, "ElasticSearch targets")
	Debug = rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable verbose logging")
}
