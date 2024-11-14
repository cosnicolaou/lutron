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

type Transport interface {
	Send(text string) error
	ReadUntil(text string) (string, error)
	Expect(text string) error
	Close() error
}

type telnetConn struct {
	conn    *telnet.Conn
	timeout time.Duration
	logger  *slog.Logger
}

func (tc *telnetConn) Send(text string) error {
	if err := tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout)); err != nil {
		return err
	}
	buf := slices.Clone([]byte(text))
	buf = append(buf, '\r', '\n')
	_, err := tc.conn.Write(buf)
	tc.logger.Info("sent", "text", text, "err", err)
	return err
}

func (tc *telnetConn) ReadUntil(text string) (string, error) {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		return "", err
	}
	buf, err := tc.conn.ReadUntil(text)
	out := string(buf)
	tc.logger.Info("readUntil", "text", text, "err", err)
	return out, err
}

func (tc *telnetConn) Expect(text string) error {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		return err
	}
	err := tc.conn.SkipUntil(text)
	tc.logger.Info("expect", "text", text, "err", err)
	return err
}

func (tc *telnetConn) Close() error {
	return tc.conn.Close()
}
