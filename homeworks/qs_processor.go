// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloudeng.io/cmdutil/keystore"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/streamconn"
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

	mu           sync.Mutex
	idle         *streamconn.IdleTimer
	session      streamconn.Session
	closeContext context.Context
	closeCancel  context.CancelFunc
	closeCh      chan struct{}
}

func NewQSProcessor(opts devices.Options) *QSProcessor {
	return &QSProcessor{
		logger: opts.Logger,
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

func (p *QSProcessor) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, out io.Writer, args ...string) error {
			t, err := p.GetTime(ctx)
			if out != nil && err == nil {
				fmt.Fprintf(out, "gettime: %v\n", t)
			}
			return err
		},
		"getlocation": func(ctx context.Context, out io.Writer, args ...string) error {
			lat, long, err := p.GetLatLong(ctx)
			if out != nil && err == nil {
				fmt.Fprintf(out, "latlong: %vN %vW\n", lat, long)
			}
			return err
		},
		"getsuntimes": func(ctx context.Context, out io.Writer, args ...string) error {
			rise, set, err := p.GetSunriseSunset(ctx)
			if out != nil && err == nil {
				fmt.Fprintf(out, "sunrise: %v, sunset: %v\n",
					rise.Format("15:04:05"), set.Format("15:04:05"))
			}
			return err
		},
		"os_version": func(ctx context.Context, out io.Writer, args ...string) error {
			osv, err := p.Version(ctx)
			if out != nil && err == nil {
				fmt.Fprintf(out, "%v\n", osv)
			}
			return err
		},
	}
}

func (p *QSProcessor) OperationsHelp() map[string]string {
	return map[string]string{
		"gettime":     "get the current time, date and timezone",
		"getlocation": "get the current location in latitude and longitude",
		"getsuntimes": "get the current sunrise and sunset times in local time",
		"os_version":  "get the OS version running on QS processor",
	}
}

func (p *QSProcessor) SystemQuery(ctx context.Context, action protocol.SystemActions) (string, error) {
	s := p.Session(ctx)
	response, err := protocol.System(ctx, s, false, action)
	if err != nil {
		return "", fmt.Errorf("QSProcessor.System: %v: %v", action, err)
	}
	return response, nil
}

func (p *QSProcessor) GetLatLong(ctx context.Context) (float64, float64, error) {
	latlong, err := p.SystemQuery(ctx, protocol.SystemLatLong)
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(latlong, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected response: %v", latlong)
	}
	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %v", err)
	}
	long, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %v", err)
	}
	return lat, long, nil
}

func (p *QSProcessor) GetSunriseSunset(ctx context.Context) (time.Time, time.Time, error) {
	sunrise, err := p.SystemQuery(ctx, protocol.SystemSunrise)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunset, err := p.SystemQuery(ctx, protocol.SystemSunset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	sunriseT, err := time.Parse("15:04:05", sunrise)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunsetT, err := time.Parse("15:04:05", sunset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return sunriseT, sunsetT, nil
}

func (p *QSProcessor) GetTime(ctx context.Context) (time.Time, error) {
	date, err := p.SystemQuery(ctx, protocol.SystemDate)
	if err != nil {
		return time.Time{}, err
	}
	tod, err := p.SystemQuery(ctx, protocol.SystemTime)
	if err != nil {
		return time.Time{}, err
	}
	tz, err := p.SystemQuery(ctx, protocol.SystemTimeZone)
	if err != nil {
		return time.Time{}, err
	}
	tzn := protocol.NormalizeTimeZone(tz)
	sysTime, err := time.Parse("01/02/2006 15:04:05 -07:00", date+" "+tod+" "+tzn)
	if err != nil {
		return time.Time{}, err
	}
	return sysTime, nil
}

func (p *QSProcessor) Version(ctx context.Context) (string, error) {
	data, err := p.SystemQuery(ctx, protocol.SystemOSRev)
	if err != nil {
		return "", err
	}
	return data, nil
}

// Session returns an authenticated session to the QS processor. If
// an error is encountered then an error session is returned.
func (p *QSProcessor) Session(ctx context.Context) streamconn.Session {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.session != nil {
		return p.session
	}
	transport, err := streamconn.DialTelnet(ctx, p.IPAddress, p.Timeout, p.logger)
	if err != nil {
		return streamconn.NewErrorSession(err)
	}
	p.idle = streamconn.NewIdleTimer(p.KeepAlive)
	p.session = streamconn.NewSession(transport, p.idle)

	// Authenticate
	keys := keystore.AuthFromContextForID(ctx, p.KeyID)
	err = protocol.QSLogin(ctx, p.session, keys.User, keys.Token)
	if err != nil {
		p.session.Close(ctx)
		p.session = nil
		p.idle = nil
		return streamconn.NewErrorSession(err)
	}
	p.closeContext, p.closeCancel = context.WithCancel(ctx)
	p.closeCh = make(chan struct{})
	go p.idleClose(p.closeContext, p.idle)
	return p.session
}

func (p *QSProcessor) CloseSession(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.session == nil {
		return nil
	}
	err := p.session.Close(ctx)
	p.session = nil
	p.idle = nil
	p.closeCancel()
	closeTimeout := time.Minute
	select {
	case <-p.closeCh:
	case <-time.After(closeTimeout):
		if err == nil {
			err = fmt.Errorf("failed to close session after %v", closeTimeout)
		}
	}
	return err
}

func (p *QSProcessor) idleClose(ctx context.Context, idle *streamconn.IdleTimer) {
	for {
		select {
		case <-time.After(idle.Remaining()):
			if idle.Expired() {
				p.CloseSession(ctx)
				return
			}
		case <-ctx.Done():
			return
		case <-p.closeCh:
			return
		}
	}
}

func (p *QSProcessor) Close(ctx context.Context) error {
	return p.CloseSession(ctx)
}
