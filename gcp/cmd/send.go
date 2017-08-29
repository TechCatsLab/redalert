// Copyright © 2017 SmartestEE Co.,Ltd..
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
	"redalert/udp/client"
	tcp "redalert/tcp/client"

	"github.com/spf13/cobra"
)

var (
	host     string
	port     string
	packSize int
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "A send file tool.",
	Long:  `send is a tool for send file, base on udp.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		if protocol == "tcp" {
			conf := &tcp.Conf{
				Address:    host,
				Port:       port,
				PackSize:   packSize,
				FileName:   args[0],
			}

			cli, err := tcp.NewClient(conf)
			if err != nil {
				fmt.Println("Connection failed with error:", err)
				return
			}

			cli.Start()
		}

		conf := &client.Conf{
			RemoteAddress: host,
			RemotePort:    port,
			PacketSize:    packSize,
			FileName:      args[0],
		}

		cli, err := client.NewClient(conf)
		if err != nil {
			fmt.Println("Remote server not receive:", err)
			return
		}

		err = cli.Start()
		if err != nil {
			fmt.Println("Client run error:", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.
	sendCmd.Flags().StringVarP(&protocol, "proto", "o", "udp", "send method")
	sendCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "Target host")
	sendCmd.Flags().StringVarP(&port, "port", "p", "17120", "Target port")
	sendCmd.Flags().IntVarP(&packSize, "packetSize", "s", 1024, "Every packet size")
}
