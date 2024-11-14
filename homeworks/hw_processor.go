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

type qsProcessorSpec struct {
	IPAddress string        `yaml:"ip_address"`
	KeepAlive time.Duration `yaml:"keep_alive"`
	AuthID    string        `yaml:"key_id"`
}

type qsProcessor struct {
	processorConfig devices.ControllerConfigCommon
	Spec            qsProcessorSpec `yaml:",inline"`
	conn            *conn
}

func newQSProcessor() *qsProcessor {
	return &qsProcessor{
		conn: &conn{},
	}
}

func (p *qsProcessor) SetConfig(c devices.ControllerConfigCommon) {
	p.processorConfig = c
}

func (p *qsProcessor) Config() devices.ControllerConfigCommon {
	return p.processorConfig
}

func (p *qsProcessor) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&p.Spec)
}

func (p *qsProcessor) Implementation() any {
	return p
}

func (p *qsProcessor) connection(ctx context.Context) (protocol.Conn, error) {
	return p.conn.conn(ctx, p.processorConfig, p.Spec)
}

type conn struct {
	sync.Mutex
	keepAlive time.Duration
	pconn     protocol.Conn
}

func (c *conn) conn(ctx context.Context, cfgBase devices.ControllerConfigCommon,
	cfg qsProcessorSpec) (protocol.Conn, error) {
	c.Lock()
	defer c.Unlock()
	if c.pconn != nil {
		return c.pconn, c.pconn.Err()
	}
	conn := protocol.New(cfgBase.Name).Dial(context.Background(), cfg.IPAddress)
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	keepalive := cfg.KeepAlive
	if keepalive == 0 {
		keepalive = time.Minute * 5
	}
	go c.disconnect(ctx, c.keepAlive)
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
