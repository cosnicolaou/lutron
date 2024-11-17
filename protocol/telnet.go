// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"log/slog"
	"slices"
	"time"

	"github.com/ziutek/telnet"
)

type telnetConn struct {
	conn    *telnet.Conn
	timeout time.Duration
	logger  *slog.Logger
}

func DialTelnet(addr string, timeout time.Duration, logger *slog.Logger) (Transport, error) {
	conn, err := telnet.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	logger = logger.With("addr", conn.RemoteAddr().String())
	return &telnetConn{conn: conn, timeout: timeout, logger: logger}, nil
}

func (tc *telnetConn) Send(text string) error {
	if err := tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout)); err != nil {
		return err
	}
	buf := slices.Clone([]byte(text))
	_, err := tc.conn.Write(buf)
	tc.logger.Info("sent", "text", text, "err", err)
	return err
}

func (tc *telnetConn) ReadUntil(text ...string) (string, error) {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		return "", err
	}
	buf, err := tc.conn.ReadUntil(text...)
	out := string(buf)
	tc.logger.Info("readUntil", "text", text, "err", err)
	return out, err
}

func (tc *telnetConn) Expect(text ...string) error {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		return err
	}
	err := tc.conn.SkipUntil(text...)
	tc.logger.Info("expect", "text", text, "err", err)
	return err
}

func (tc *telnetConn) Close() error {
	return tc.conn.Close()
}
