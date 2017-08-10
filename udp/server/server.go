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
 *     Initial: 2017/08/10        Yusan Kurban
 */

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	check = make(chan int, 1)
)

func main() {
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", ":3017")
	if err != nil {
		fmt.Println("Can't resolve addr:", err.Error())
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("listen error:", err.Error())
		panic(err)
	}
	defer conn.Close()
	go handlerClient(conn)

	tick := time.Tick(2 * time.Minute)
	for {
		select {
		case <-tick:
			os.Exit(1)
		case <-check:
			tick = time.Tick(2 * time.Minute)
		}
	}
}

func handlerClient(conn *net.UDPConn) {

	data := make([]byte, 1024)
	_, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Println("read udp msg failed with:", err.Error())
		return
	}

	check <- 1
	conn.WriteToUDP([]byte("a"), remoteAddr)
}
