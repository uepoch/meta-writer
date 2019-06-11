// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"bytes"
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/uepoch/metrics-meta/filter"

	"github.com/spf13/cobra"
)

var (
	TcpAddr     *net.IP
	TcpPort     *uint
	BloomN      *uint
	BloomP      *float64
	BloomShards *uint
	BloomFlush  *time.Duration
)

// writerCmd represents the writer command
var writerCmd = &cobra.Command{
	Use:   "writer",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := filter.NewShardedBFilter(*BloomN, *BloomP, *BloomShards, *BloomFlush)
		if err != nil {
			zap.L().Fatal("error during filter initialization", zap.Error(err))
			return
		}
		addr, err := net.ResolveTCPAddr("tcp", TcpAddr.String()+":"+strconv.Itoa(int(*TcpPort)))
		if err != nil {
			zap.L().Fatal("error during listener parsing", zap.Error(err))
			return
		}
		err = RunWriter(addr, f)
		if err != nil {
			zap.L().Fatal("error during writer execution", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(writerCmd)

	BloomN = writerCmd.Flags().Uint("bloom.N", 100000, "The target number of elements to be present in the bloom filter(s). Must not be used with bloom.Size")
	BloomShards = writerCmd.Flags().Uint("bloom.shards", 5, "How many different filters to keep")
	BloomP = writerCmd.Flags().Float64("bloom.P", 10e-3, "Target precision (chances of false positive, lower is better at the cost of memory). Must not be used with bloom.Size")
	BloomFlush = writerCmd.Flags().Duration("bloom.flush.interval", 10*time.Minute, "Time between the clear of one of the filters")
	TcpAddr = writerCmd.Flags().IP("tcp.addr", net.ParseIP("0.0.0.0"), "Address to bind the carbon listener.")
	TcpPort = writerCmd.Flags().Uint("tcp.port", 3343, "Port to bind the carbon listener")
}

type Writer struct {
	connPool chan *net.TCPConn
	listener *net.TCPListener
	filter   filter.Filter
}

func (w *Writer) AcceptAll() {
	l := w.listener
	acceptor := w.connPool
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			zap.L().Error("error while accepting conn", zap.Error(err))
		}
		zap.L().Debug("accepted connection", zap.String("addr", c.RemoteAddr().String()))
		acceptor <- c
	}
}

func (w *Writer) HandleConns() {
	for {
		select {
		case c, ok := <-w.connPool:
			if ok {
				w.handleConn(c)
			} else {
				break
			}
		}
	}
}

func (w *Writer) handleConn(c *net.TCPConn) error {
	defer c.Close()
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		metric := scanner.Bytes()
		l := zap.L().With(zap.ByteString("key", metric))
		l.Debug("received metric")
		metric = metric[:bytes.IndexByte(metric, ' ')]
		l = zap.L().With(zap.ByteString("key", metric))
		if w.filter.ContainsOrUpdate(metric) {
			l.Info("filter hit")
		} else {
			l.Info("filter miss")
		}
	}
	err := scanner.Err()
	if err != nil {
		zap.L().Error("error reading on conn", zap.Error(err))
	}
	return err
}

func (w *Writer) Run() error {
	go w.AcceptAll()
	w.HandleConns()
	return nil
}

func RunWriter(addr *net.TCPAddr, f filter.Filter) error {
	l, err := net.ListenTCP("tcp", addr)
	zap.L().Info("listening", zap.Stringer("addr", addr))
	defer l.Close()
	if err != nil {
		return err
	}

	ch := make(chan *net.TCPConn, 10)
	w := &Writer{
		ch,
		l,
		f,
	}
	return w.Run()
}
