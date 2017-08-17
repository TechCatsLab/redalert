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
 *     Initial: 2017/08/11        Liu jiachang
 */

package client




//import (
//	"sync"
//	"testing"
//)
//
//var (
//	Connection []*TcpClient
//	remoteAddr = "127.0.0.1:9999"
//	writeMsg   = "you are the pretty sunshine of my eye\n"
//)
//
//func BenchmarkTcpClient(b *testing.B) {
//	group := &sync.WaitGroup{}
//	for i := 0; i < 50; i++ {
//		cli, err := NewTcpClient(remoteAddr)
//		if err != nil {
//			b.Error("create client error:", err)
//		}
//
//		err = cli.DialTcp()
//		if err == nil {
//			Connection = append(Connection, cli)
//		} else {
//			b.Error("dial tcp error:", err)
//		}
//	}
//
//	group.Add(50)
//	b.ResetTimer()
//	for j := 0; j < 50; j++ {
//		temp := j
//		go func() {
//			for i := 0; i < b.N; i++ {
//				Connection[temp].ReadMessage()
//				Connection[temp].WriteMessage(writeMsg)
//			}
//			group.Done()
//		}()
//	}
//
//	group.Wait()
//}
