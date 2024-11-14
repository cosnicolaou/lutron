// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/ziutek/telnet"
)

type conn struct {
	options
	mu sync.Mutex
}

func New(name string, opts ...Option) Conn {
	c := &conn{}
	c.options.setDefaults()
	for _, fn := range opts {
		fn(&c.options)
	}
	if c.session == nil {
		c.session = NewSession()
	}
	c.logger = c.logger.With("name", name)
	return c
}

type Conn interface {
	Dial(ctx context.Context, addr string) Conn
	Run(func(Session) error) Conn
	Busy() bool
	Err() error
	Close() error
}

type errConn struct {
	err error
}

func (ec *errConn) Dial(ctx context.Context, addr string) Conn {
	return ec
}

func (ec *errConn) Run(func(Session) error) Conn {
	return ec
}

func (ec *errConn) Close() error {
	return ec.err
}

func (ec *errConn) Err() error {
	return ec.err
}

func (ec *errConn) Busy() bool {
	return false
}

type options struct {
	timeout   time.Duration
	setup     []string
	logger    *slog.Logger
	session   Session
	transport Transport
}

func (o *options) setDefaults() {
	o.timeout = 5 * time.Second
	o.logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
}

type Option func(o *options)

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

func WithSetup(s ...string) Option {
	return func(o *options) {
		o.setup = s
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

func WithSession(s Session) Option {
	return func(o *options) {
		o.session = s
	}
}

func WithTransport(t Transport) Option {
	return func(o *options) {
		o.transport = t
	}
}

func (c *conn) Dial(ctx context.Context, addr string) Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.options.transport != nil {
		return &errConn{err: fmt.Errorf("Transport already configured")}
	}
	conn, err := telnet.Dial("tcp", addr)
	if err != nil {
		return &errConn{err: err}
	}
	c.logger = c.logger.With("addr", conn.RemoteAddr().String())
	transport := &telnetConn{conn: conn, timeout: c.timeout, logger: c.logger}
	c.session.SetTransport(transport)
	c.options.transport = transport
	return c
}

func (c *conn) Run(apply func(Session) error) Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.options.transport == nil {
		return &errConn{err: fmt.Errorf("No transport was configured, either call Dial or create a connection using the WithTransport option")}
	}
	if err := apply(c.session); err != nil {
		return &errConn{err: err}
	}
	return c
}

func (c *conn) Err() error {
	return nil
}

func (c *conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session.Close()
}

func (c *conn) Busy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return false // must be false if we have the lock!
}
