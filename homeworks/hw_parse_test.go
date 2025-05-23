// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks_test

import (
	"bytes"
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/lutron/homeworks"
	"gopkg.in/yaml.v3"
)

const spec = `
controllers:
  - name: home
    type: homeworks-qs
    ip_address: 192.168.1.50
    timeout: 1m
    keep_alive: 1m
    key_id: home

devices:
  - name: living room
    type: shadegrp
    controller: home
    id: 1
    level: 50
`

type config struct {
	Controllers []devices.ControllerConfig `yaml:"controllers"`
	Devices     []devices.DeviceConfig     `yaml:"devices"`
}

func TestHWParsing(t *testing.T) {
	ctx := context.Background()
	var cfg config
	if err := yaml.Unmarshal([]byte(spec), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	logRecorder := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(logRecorder, nil))
	ctx = ctxlog.WithLogger(ctx, logger)
	opts := []devices.Option{
		devices.WithDevices(homeworks.SupportedDevices()),
		devices.WithControllers(homeworks.SupportedControllers()),
	}

	ctrls, devs, err := devices.CreateSystem(ctx,
		cfg.Controllers,
		cfg.Devices,
		opts...)
	if err != nil {
		t.Fatalf("failed to build devices: %v", err)
	}

	cCommSpec := ctrls["home"].Config()
	if got, want := cCommSpec, (devices.ControllerConfigCommon{
		Name:        "home",
		Type:        "homeworks-qs",
		RetryConfig: devices.RetryConfig{Timeout: time.Minute},
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	dCommSpec := devs["living room"].Config()
	if got, want := dCommSpec, (devices.DeviceConfigCommon{
		Name:           "living room",
		ControllerName: "home",
		RetryConfig:    devices.RetryConfig{Timeout: time.Minute},
		Type:           "shadegrp"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	cSpec := ctrls["home"].CustomConfig().(homeworks.QSProcessorConfig)
	if got, want := cSpec, (homeworks.QSProcessorConfig{
		IPAddress: "192.168.1.50",
		KeepAlive: time.Minute,
		KeyID:     "home"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	dSpec := devs["living room"].CustomConfig().(homeworks.HWShadeConfig)
	if got, want := dSpec, (homeworks.HWShadeConfig{ID: 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

}
