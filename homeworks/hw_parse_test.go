// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks_test

import (
	"bytes"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/devices"
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
    type: homeworks-blind
    controller: home
    id: 1
    level: 50
`

type config struct {
	Controllers []devices.ControllerConfig `yaml:"controllers"`
	Devices     []devices.DeviceConfig     `yaml:"devices"`
}

func TestHWParsing(t *testing.T) {
	var cfg config
	if err := yaml.Unmarshal([]byte(spec), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	logRecorder := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(logRecorder, nil))
	opts := []devices.Option{
		devices.WithLogger(logger),
	}

	ctrls, devs, err := devices.BuildDevices(
		cfg.Controllers,
		cfg.Devices,
		homeworks.SupportedControllers(),
		homeworks.SupportedDevices(),
		opts...)
	if err != nil {
		t.Fatalf("failed to build devices: %v", err)
	}

	cCommSpec := ctrls["home"].Config()
	if got, want := cCommSpec, (devices.ControllerConfigCommon{
		Name: "home",
		Type: "homeworks-qs"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	dCommSpec := devs["living room"].Config()
	if got, want := dCommSpec, (devices.DeviceConfigCommon{
		Name:       "living room",
		Controller: "home",
		Type:       "homeworks-blind"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	cSpec := homeworks.GetQSProcessorConfig(ctrls["home"])
	if got, want := cSpec, (homeworks.QSProcessorConfig{
		IPAddress: "192.168.1.50",
		Timeout:   time.Minute,
		KeepAlive: time.Minute,
		KeyID:     "home"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	dSpec := homeworks.GetBlindConfig(devs["living room"])
	if got, want := dSpec, (homeworks.HWBlindConfig{
		ID:    1,
		Level: 50}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

}
