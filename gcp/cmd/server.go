// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"log"

	"github.com/spf13/cobra"

	"redalert/udp/server"
)

var (
	serverAddress   string
	serverPort      string
	serverPackSize  int
	serverCacheSize int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := server.Conf{
			Address:    serverAddress,
			Port:       serverPort,
			PacketSize: serverPackSize,
			CacheCount: serverCacheSize,
		}

		h := server.Provider{}

		_, err := server.NewServer(&conf, &h)
		if err != nil {
			log.Print(err)
			return
		}
		select {}
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	serverCmd.Flags().StringVarP(&serverAddress, "addr", "a", "127.0.0.1", "addr of server.")
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "17120", "port of server.")
	serverCmd.Flags().IntVarP(&serverPackSize, "pack", "P", 1024, "size of pack.")
	serverCmd.Flags().IntVarP(&serverCacheSize, "cache", "c", 1024, "size of cache.")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
