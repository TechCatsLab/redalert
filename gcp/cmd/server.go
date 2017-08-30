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
	"log"

	"github.com/spf13/cobra"

	tcp "redalert/tcp/server"
	"redalert/udp/server"
)

var (
	protocol        string
	serverAddress   string
	serverPort      string
	serverPackSize  int
	serverCacheSize int
	maxConn         int
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
		if protocol == "tcp" {
			tcpConf := tcp.Conf{
				Addr:    serverAddress,
				Port:    serverPort,
				MaxConn: maxConn,
			}

			tcp.StartServer(&tcpConf)
		} else {
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
		}
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	serverCmd.Flags().StringVarP(&protocol, "protocol", "o", "udp", "select a proto to send file, udp or tcp.")
	serverCmd.Flags().StringVarP(&serverAddress, "addr", "a", "127.0.0.1", "addr of server.")
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "17120", "port of server.")
	serverCmd.Flags().IntVarP(&serverPackSize, "pack", "P", 1024, "size of pack.")
	serverCmd.Flags().IntVarP(&serverCacheSize, "cache", "c", 1024, "size of cache.")
	serverCmd.Flags().IntVarP(&maxConn, "max", "M", 10, "TCP max connection.")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
