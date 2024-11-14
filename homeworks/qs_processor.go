// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"sync"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/protocol"

	"gopkg.in/yaml.v3"
)

type QSProcessorConfig struct {
	IPAddress string        `yaml:"ip_address"`
	Timeout   time.Duration `yaml:"timeout"`
	KeepAlive time.Duration `yaml:"keep_alive"`
	KeyID     string        `yaml:"key_id"`
}

type qsProcessor struct {
	devices.ControllerConfigCommon
	QSProcessorConfig `yaml:",inline"`

	conn *conn
}

func newQSProcessor(opts devices.Options) *qsProcessor {
	return &qsProcessor{
		conn: &conn{opts: opts},
	}
}

func (p *qsProcessor) SetConfig(c devices.ControllerConfigCommon) {
	p.ControllerConfigCommon = c
}

func (p *qsProcessor) Config() devices.ControllerConfigCommon {
	return p.ControllerConfigCommon
}

func (p *qsProcessor) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&p.QSProcessorConfig)
}

func (p *qsProcessor) Implementation() any {
	return p
}

func (p *qsProcessor) connection(ctx context.Context) (protocol.Conn, error) {
	return p.conn.conn(ctx, p.ControllerConfigCommon, p.QSProcessorConfig)
}

type conn struct {
	sync.Mutex
	opts  devices.Options
	pconn protocol.Conn
}

func (c *conn) conn(ctx context.Context, cfgBase devices.ControllerConfigCommon,
	cfg QSProcessorConfig) (protocol.Conn, error) {
	c.Lock()
	defer c.Unlock()
	if c.pconn != nil {
		return c.pconn, c.pconn.Err()
	}

	conn := c.opts.ProtocolConnection
	if conn == nil {
		opts := []protocol.Option{protocol.WithTimeout(cfg.Timeout)}
		if c.opts.Logger != nil {
			opts = append(opts, protocol.WithLogger(c.opts.Logger))
		}
		conn = protocol.New(cfgBase.Name, opts...).
			Dial(context.Background(), cfg.IPAddress)
	}
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	c.pconn = conn

	keepalive := cfg.KeepAlive
	if keepalive == 0 {
		keepalive = time.Minute * 5
	}
	go c.disconnect(ctx, cfg.KeepAlive)
	return conn, nil
}

func (c *conn) disconnect(ctx context.Context, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout):
			c.Lock()
			defer c.Unlock()
			if c.pconn.Busy() {
				continue
			}
			err := c.pconn.Close()
			c.pconn = nil
			return err
		}
	}
}
