// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"time"

	"cloudeng.io/cmdutil/keystore"
	"cloudeng.io/logging/ctxlog"
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

	ondemand *netutil.OnDemandConnection[streamconn.Session, *QSProcessor]
}

func NewQSProcessor(_ devices.Options) *QSProcessor {
	p := &QSProcessor{}
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
		"gettime": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := p.Session(ctx)
			t, err := protocol.GetTime(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "gettime: %v\n", t)
			}
			return struct {
				Time string `json:"time"`
			}{Time: t.String()}, err
		},
		"getlocation": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := p.Session(ctx)
			lat, long, err := protocol.GetLatLong(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "latlong: %vN %vW\n", lat, long)
			}
			return struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			}{Latitude: lat, Longitude: long}, err
		},
		"getsuntimes": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := p.Session(ctx)
			rise, set, err := protocol.GetSunriseSunset(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "sunrise: %v, sunset: %v\n",
					rise.Format("15:04:05"), set.Format("15:04:05"))
			}
			return struct {
				SunRise string `json:"sunrise"`
				SunSet  string `json:"sunset"`
			}{
				SunRise: rise.Format("15:04:05"),
				SunSet:  set.Format("15:04:05"),
			}, err
		},
		"os_version": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := p.Session(ctx)
			osv, err := protocol.GetVersion(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "%v\n", osv)
			}
			return struct {
				OSVersion string `json:"os_version"`
			}{OSVersion: osv}, err
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
	transport, err := telnet.Dial(ctx, p.ControllerConfigCustom.IPAddress, p.Timeout)
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

func (p *QSProcessor) loggingContext(ctx context.Context) context.Context {
	return ctxlog.WithAttributes(ctx, "protocol", "homeworks-qs")
}

// Session returns an authenticated session to the QS processor. If
// an error is encountered then an error session is returned.
func (p *QSProcessor) Session(ctx context.Context) (context.Context, streamconn.Session) {
	ctx = p.loggingContext(ctx)
	return ctx, p.ondemand.Connection(ctx)
}

func (p *QSProcessor) Close(ctx context.Context) error {
	ctx = p.loggingContext(ctx)
	return p.ondemand.Close(ctx)
}
