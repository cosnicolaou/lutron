// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks_test

import (
	"fmt"
	"testing"

	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/homeworks"
	"gopkg.in/yaml.v3"
)

const spec = `
controllers:
  - name: home
    type: homeworks-qs
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

func TestHWBlindParsing(t *testing.T) {
	var cfg config
	if err := yaml.Unmarshal([]byte(spec), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	ctrls, devs, err := devices.BuildDevices(
		cfg.Controllers,
		cfg.Devices,
		homeworks.SupportedControllers(),
		homeworks.SupportedDevices())
	if err != nil {
		t.Fatalf("failed to build devices: %v", err)
	}

	fmt.Printf("ctrls: %v\n", ctrls)
	fmt.Printf("devs: %v\n", devs)
	t.Fail()
}
