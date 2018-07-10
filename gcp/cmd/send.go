/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co.,Ltd..
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2017/08/30        Liu JiaChang
 */

package cmd

import (
	"fmt"

	tcp "github.com/TechCatsLab/redalert/tcp/client"
	"github.com/TechCatsLab/redalert/udp/client"
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
				Address:  host,
				Port:     port,
				PackSize: packSize,
				FileName: args[0],
			}

			cli := tcp.NewClient(conf)
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
