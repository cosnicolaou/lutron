// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloudeng.io/cmdutil/keystore"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/netutil"
	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/automation/net/streamconn/telnet"
	"github.com/cosnicolaou/lutron/protocol"

	"gopkg.in/yaml.v3"
)

type QSProcessorConfig struct {
	IPAddress string        `yaml:"ip_address"`
	KeepAlive time.Duration `yaml:"keep_alive"`
	KeyID     string        `yaml:"key_id"`
	Verbose   bool          `yaml:"verbose"`
}

type QSProcessor struct {
	devices.ControllerBase[QSProcessorConfig]
	logger *slog.Logger

	ondemand *netutil.OnDemandConnection[streamconn.Session, *QSProcessor]
}

func NewQSProcessor(opts devices.Options) *QSProcessor {
	p := &QSProcessor{
		logger: opts.Logger.With("protocol", "homeworks-qs"),
	}
	p.ondemand = netutil.NewOnDemandConnection(p, streamconn.NewErrorSession)
	return p
}

func (p *QSProcessor) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&p.ControllerConfigCustom); err != nil {
		return err
	}
	if p.ControllerConfigCustom.KeepAlive == 0 {
		return fmt.Errorf("keep_alive must be specified")
	}
	p.ondemand.SetKeepAlive(p.ControllerConfigCustom.KeepAlive)
	return nil
}

func (p *QSProcessor) Implementation() any {
	return p
}

func (p *QSProcessor) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args devices.OperationArgs) error {
			t, err := protocol.GetTime(ctx, p.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "gettime: %v\n", t)
			}
			return err
		},
		"getlocation": func(ctx context.Context, args devices.OperationArgs) error {
			lat, long, err := protocol.GetLatLong(ctx, p.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "latlong: %vN %vW\n", lat, long)
			}
			return err
		},
		"getsuntimes": func(ctx context.Context, args devices.OperationArgs) error {
			rise, set, err := protocol.GetSunriseSunset(ctx, p.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "sunrise: %v, sunset: %v\n",
					rise.Format("15:04:05"), set.Format("15:04:05"))
			}
			return err
		},
		"os_version": func(ctx context.Context, args devices.OperationArgs) error {
			osv, err := protocol.GetVersion(ctx, p.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "%v\n", osv)
			}
			return err
		},
	}
}

func (*QSProcessor) OperationsHelp() map[string]string {
	return map[string]string{
		"gettime":     "get the current time, date and timezone",
		"getlocation": "get the current location in latitude and longitude",
		"getsuntimes": "get the current sunrise and sunset times in local time",
		"os_version":  "get the OS version running on QS processor",
	}
}

func (p *QSProcessor) Connect(ctx context.Context, idle netutil.IdleReset) (streamconn.Session, error) {
	transport, err := telnet.Dial(ctx, p.ControllerConfigCustom.IPAddress, p.Timeout, p.logger)
	if err != nil {
		return nil, err
	}
	session := streamconn.NewSession(transport, idle)

	// Authenticate
	keys := keystore.AuthFromContextForID(ctx, p.ControllerConfigCustom.KeyID)
	if err := protocol.QSLogin(ctx, session, keys.User, keys.Token); err != nil {
		session.Close(ctx)
		return nil, err
	}
	return session, nil
}

func (p *QSProcessor) Disconnect(ctx context.Context, sess streamconn.Session) error {
	return sess.Close(ctx)
}

// Session returns an authenticated session to the QS processor. If
// an error is encountered then an error session is returned.
func (p *QSProcessor) Session(ctx context.Context) streamconn.Session {
	return p.ondemand.Connection(ctx)
}

func (p *QSProcessor) Close(ctx context.Context) error {
	return p.ondemand.Close(ctx)
}
