// Copyright © 2017 jsharkc
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
	"os"

	"github.com/spf13/cobra"

	"redalert/udp/client"
)

var (
	host     string
	port     string
	packSize int
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sendcli",
	Short: "A send file tool.",
	Long: `
sendcli is a tool for send file, base on udp.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		conf := &client.Conf{
			RemoteAddress: host,
			RemotePort:    port,
			PacketSize:    packSize,
		}

		cli, err := client.NewClient(conf, nil)
		if err != nil {
			fmt.Println("Remote server not receive:", err)
			return
		}

		err = cli.PrepareFile(args[0])
		if err != nil {
			fmt.Println("Prepare client error:", err)
			return
		}

		err = cli.StartRun()
		if err != nil {
			fmt.Println("Client run error:", err)
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sendcli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "Target host")
	RootCmd.Flags().StringVarP(&port, "port", "p", "17120", "Target port")
	RootCmd.Flags().IntVarP(&packSize, "packetSize", "s", 1024, "Every packet size")
}
