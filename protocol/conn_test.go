// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/protocol"
	"github.com/reiver/go-telnet"
)

func runServer(t *testing.T, handler telnet.Handler, wg *sync.WaitGroup) net.Listener {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &telnet.Server{
		Handler: handler,
	}
	go func() {
		server.Serve(listener)
		wg.Done()
	}()
	return listener
}

func TestClient(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	server := runServer(t, telnet.EchoHandler, &wg)
	defer func() {
		server.Close()
		wg.Wait()
	}()

	logRecorder := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(logRecorder, nil))
	addr := server.Addr().String()

	transport, err := protocol.DialTelnet(addr, time.Minute, logger)
	if err != nil {
		t.Fatal(err)
	}

	session := protocol.NewSession(transport)

	conn := protocol.New("test-client",

		protocol.WithSession(session),
		protocol.WithLogger(logger),
	).Dial(context.Background(), addr)

	err := conn.
		Run(func(s protocol.Session) error {
			s.Send("hello")
			s.Send("world")
			read := s.ReadUntil("world\r\n")
			s.Append(read)
			return s.Err()
		}).
		Err()
	if err != nil {
		t.Fatal(err)
	}

	err = conn.Run(func(s protocol.Session) error {
		s.Send("and")
		s.Send("again")
		read := s.ReadUntil("again\r\n")
		s.Append(read)
		return s.Err()
	}).Close()

	if err != nil {
		t.Fatal(err)
	}

	entries := []string{}
	for e := range session.Entries() {
		entries = append(entries, e)
	}
	expected := []string{"hello\r\nworld\r\n", "and\r\nagain\r\n"}
	if got, want := entries, expected; !slices.Equal(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
