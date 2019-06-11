// Copyright © 2019 m.conraux@criteo.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metrics-meta",
	Short: "A tool that can be used as a component for a metric searching system based on Elasticsearch",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().IPP("http.addr", "l", net.ParseIP("0.0.0.0"), "Address to bind to.")
	rootCmd.PersistentFlags().IntP("http.port", "p", 3343, "Port to bind to")
	rootCmd.PersistentFlags().IPSliceP("es.addr", "H", []net.IP{}, "ElasticSearch targets")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable verbose logging")

}
