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

	mgr      *streamconn.SessionManager
	ondemand *netutil.OnDemandConnection[streamconn.Transport, *QSProcessor]
}

func NewQSProcessor(_ devices.Options) *QSProcessor {
	p := &QSProcessor{
		mgr: &streamconn.SessionManager{},
	}
	p.ondemand = netutil.NewOnDemandConnection(p)
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

func (p *QSProcessor) runOperation(ctx context.Context, op func(context.Context, *streamconn.Session, devices.OperationArgs) (any, error), args devices.OperationArgs) (any, error) {
	ctx, sess, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.Release()
	return op(ctx, sess, args)
}

func (p *QSProcessor) contactClosurePulse(ctx context.Context, id []byte, pulse time.Duration, l0, l1 byte) (any, error) {
	ctx, sess, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.Release()
	pars := make([]byte, 0, 32)
	pars = append(pars, id...)
	pars = append(pars, ',', '1', ',', l0)
	// Ignore any response since the response may refer
	// to integration IDs that don't match the request.
	// This happens when the contact closure is activated
	// via a visor control for example where the request is
	// sent to the visor control, but the system issues
	// monitoring commands that refer to the integration IDs
	// of the devices connected to the visor control.
	err = protocol.NewCommand(protocol.OutputCommands, true, pars).Invoke(ctx, sess)
	if err != nil {
		return nil, err
	}
	time.Sleep(pulse)
	pars[len(pars)-1] = l1
	err = protocol.NewCommand(protocol.OutputCommands, true, pars).Invoke(ctx, sess)
	return nil, err
}

func (p *QSProcessor) getTime(ctx context.Context, sess *streamconn.Session, args devices.OperationArgs) (any, error) {
	t, err := protocol.GetTime(ctx, sess)
	if err == nil {
		fmt.Fprintf(args.Writer, "gettime: %v\n", t)
	}
	return struct {
		Time string `json:"time"`
	}{Time: t.String()}, err
}

func (p *QSProcessor) getLocation(ctx context.Context, sess *streamconn.Session, args devices.OperationArgs) (any, error) {
	lat, long, err := protocol.GetLatLong(ctx, sess)
	if err == nil {
		fmt.Fprintf(args.Writer, "latlong: %vN %vW\n", lat, long)
	}
	return struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{Latitude: lat, Longitude: long}, err
}

func (p *QSProcessor) getSuntimes(ctx context.Context, sess *streamconn.Session, args devices.OperationArgs) (any, error) {
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
}

func (p *QSProcessor) getOSVersion(ctx context.Context, sess *streamconn.Session, args devices.OperationArgs) (any, error) {
	osv, err := protocol.GetVersion(ctx, sess)
	if err == nil {
		fmt.Fprintf(args.Writer, "%v\n", osv)
	}
	return struct {
		OSVersion string `json:"os_version"`
	}{OSVersion: osv}, err
}

func (p *QSProcessor) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return p.runOperation(ctx, p.getTime, args)
		},
		"getlocation": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return p.runOperation(ctx, p.getLocation, args)
		},
		"getsuntimes": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return p.runOperation(ctx, p.getSuntimes, args)
		},
		"os_version": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return p.runOperation(ctx, p.getOSVersion, args)
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

func (p *QSProcessor) Connect(ctx context.Context, idle netutil.IdleReset) (streamconn.Transport, error) {
	conn, err := telnet.Dial(ctx, p.ControllerConfigCustom.IPAddress, p.Timeout)
	if err != nil {
		return nil, err
	}
	session := p.mgr.New(conn, idle)
	defer session.Release()

	// Authenticate
	keys := keystore.AuthFromContextForID(ctx, p.ControllerConfigCustom.KeyID)
	if err := protocol.QSLogin(ctx, session, keys.User, keys.Token); err != nil {
		conn.Close(ctx)
		return nil, err
	}
	return conn, nil
}

func (p *QSProcessor) Disconnect(ctx context.Context, conn streamconn.Transport) error {
	return conn.Close(ctx)
}

func (p *QSProcessor) loggingContext(ctx context.Context) context.Context {
	return ctxlog.WithAttributes(ctx, "protocol", "homeworks-qs")
}

// Session returns an authenticated session to the QS processor. If
// an error is encountered then an error session is returned.
// It also adds the protocol name to the context for logging purposes.
// The session must be released when the operation is complete.
func (p *QSProcessor) session(ctx context.Context) (context.Context, *streamconn.Session, error) {
	ctx = ctxlog.WithAttributes(ctx, "protocol", "homeworks-qs")
	conn, idle, err := p.ondemand.Connection(ctx)
	if err != nil {
		return ctx, nil, err
	}
	session := p.mgr.New(conn, idle)
	return ctx, session, nil
}

func (p *QSProcessor) Close(ctx context.Context) error {
	ctx = p.loggingContext(ctx)
	return p.ondemand.Close(ctx)
}
