// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cloudeng.io/cmdutil/keystore"
	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/protocol"

	"gopkg.in/yaml.v3"
)

type QSProcessorConfig struct {
	IPAddress string        `yaml:"ip_address"`
	Timeout   time.Duration `yaml:"timeout"`
	KeepAlive time.Duration `yaml:"keep_alive"`
	KeyID     string        `yaml:"key_id"`
	Verbose   bool          `yaml:"verbose"`
}

type QSProcessor struct {
	devices.ControllerConfigCommon
	QSProcessorConfig `yaml:",inline"`
	logger            *slog.Logger
	conn              *conn
}

func NewQSProcessor(opts devices.Options) *QSProcessor {
	return &QSProcessor{
		logger: opts.Logger,
		conn:   &conn{opts: opts},
	}
}

func (p *QSProcessor) SetConfig(c devices.ControllerConfigCommon) {
	p.ControllerConfigCommon = c
}

func (p *QSProcessor) Config() devices.ControllerConfigCommon {
	return p.ControllerConfigCommon
}

func (p *QSProcessor) CustomConfig() any {
	return p.QSProcessorConfig
}

func (p *QSProcessor) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&p.QSProcessorConfig); err != nil {
		return err
	}
	return nil
}

func (p *QSProcessor) Implementation() any {
	return p
}

func (p *QSProcessor) SystemQuery(ctx context.Context, action protocol.SystemActions) (string, error) {
	c, err := p.connection(ctx)
	if err != nil {
		return "", err
	}
	var response string
	err = c.Run(func(s protocol.Session) error {
		var err error
		response, err = protocol.System(s, false, action)
		return err
	}).Err()
	if err != nil {
		return "", fmt.Errorf("QSProcessor.System: %v: %v", action, err)
	}
	return response, nil
}

func (p *QSProcessor) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args ...string) error {
			t, err := p.GetTime(ctx)
			fmt.Printf("TIME: %v\n", t)
			return err
		},
	}
}

func (p *QSProcessor) GetTime(ctx context.Context) (time.Time, error) {
	tod, err := p.SystemQuery(ctx, protocol.SystemTime)
	if err != nil {
		return time.Time{}, err
	}
	date, err := p.SystemQuery(ctx, protocol.SystemDate)
	if err != nil {
		return time.Time{}, err
	}
	tz, err := p.SystemQuery(ctx, protocol.SystemTimeZone)
	if err != nil {
		return time.Time{}, err
	}
	todT, err := time.Parse("15:04:05", tod)
	if err != nil {
		return time.Time{}, err
	}
	dateT, err := time.Parse("01/02/06", date)
	if err != nil {
		return time.Time{}, err
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(dateT.Year(), dateT.Month(), dateT.Day(), todT.Hour(), todT.Minute(), todT.Second(), 0, loc), nil
}

func (p *QSProcessor) connection(ctx context.Context) (protocol.Conn, error) {
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
		// use lslog here.
		fmt.Printf("reusing connection to %s: (err %v)\n", cfg.IPAddress, c.pconn.Err())
		return c.pconn, c.pconn.Err()
	}
	keys := keystore.AuthFromContextForID(ctx, cfg.KeyID)

	fmt.Printf("connecting to %s: %v\n", cfg.IPAddress, keys.User)

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

	err := conn.Run(func(s protocol.Session) error {
		return protocol.QSLogin(s, keys.User, keys.Token)
	}).Err()
	if err != nil {
		return nil, err
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
