// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"log/slog"
	"time"

	"github.com/ziutek/telnet"
)

type telnetConn struct {
	conn    *telnet.Conn
	timeout time.Duration
	logger  *slog.Logger
}

func DialTelnet(ctx context.Context, addr string, timeout time.Duration, logger *slog.Logger) (Transport, error) {
	logger.Log(ctx, slog.LevelInfo, "dialing", "addr", addr)
	conn, err := telnet.Dial("tcp", addr)
	if err != nil {
		logger.Log(ctx, slog.LevelWarn, "dial failed", "addr", addr, "err", err)
		return nil, err
	}
	logger = logger.With("addr", conn.RemoteAddr().String())
	return &telnetConn{conn: conn, timeout: timeout, logger: logger}, nil
}

func (tc *telnetConn) Send(ctx context.Context, buf []byte) (int, error) {
	if err := tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout)); err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "send failed to set read deadline", "err", err)
		return -1, err
	}
	n, err := tc.conn.Write(buf)
	tc.logger.Log(ctx, slog.LevelInfo, "sent", "text", string(buf), "err", err)
	return n, err
}

func (tc *telnetConn) ReadUntil(ctx context.Context, expected []string) ([]byte, error) {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "readUntil failed to set read deadline", "err", err)
		return nil, err
	}
	buf, err := tc.conn.ReadUntil(expected...)
	if err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "readUntil failed", "text", expected, "err", err)
		return nil, err
	}
	tc.logger.Log(ctx, slog.LevelInfo, "readUntil", "text", expected)
	return buf, err
}

func (tc *telnetConn) Close(ctx context.Context) error {
	if err := tc.conn.Close(); err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "close failed", "err", err)
	}
	tc.logger.Log(ctx, slog.LevelInfo, "close")
	return nil
}
