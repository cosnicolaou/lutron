// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices_test

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/internal/testutil"
	"gopkg.in/yaml.v3"
)

const controllers_spec = `
  - name: c
    type: controller
    operations:
      enable: [on, command, quoted with space]
      disable: [off, command]
    detail: my-location
    key_id: my-key
`
const devices_spec = `
  - name: d
    controller: c
    type: device
    detail: my-device-d
    operations:
      on: [on, command]

  - name: e
    controller: c
    type: device
    operations:
      off: [off, command]
    detail: my-device-e
`

const simple_spec = `controllers:
` + controllers_spec + `
devices:
` + devices_spec

var supportedControllers = devices.SupportedControllers{
	"controller": func(string, devices.Options) (devices.Controller, error) {
		return &testutil.MockController{}, nil
	},
}

var supportedDevices = devices.SupportedDevices{
	"device": func(string, devices.Options) (devices.Device, error) {
		return &testutil.MockDevice{}, nil
	},
}

func init() {
	devices.AvailableControllers = supportedControllers
	devices.AvailableDevices = supportedDevices
}

func compareOperationMaps(got, want map[string][]string) bool {
	if len(got) != len(want) {
		return false
	}
	for k, v := range got {
		if w, ok := want[k]; !ok || !slices.Equal(v, w) {
			return false
		}
	}
	return true
}

func TestParseConfig(t *testing.T) {
	ctx := context.Background()

	system, err := devices.ParseSystemConfig(ctx, "", []byte(simple_spec))
	if err != nil {
		t.Fatalf("failed to parse system config: %v", err)
	}
	ctrls := system.Controllers
	devs := system.Devices

	ctrl := ctrls["c"]
	dev := devs["d"]

	ccfg := ctrl.Config()
	if got, want := ccfg.Operations, (map[string][]string{
		"enable":  {"on", "command", "quoted with space"},
		"disable": {"off", "command"}}); !compareOperationMaps(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	ccfg.Operations = nil

	if got, want := ccfg, (devices.ControllerConfigCommon{
		Name: "c", Type: "controller"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := ctrl.(*testutil.MockController).CustomConfig().(testutil.ControllerDetail), (testutil.ControllerDetail{
		Detail: "my-location", KeyID: "my-key"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	dcfg := dev.Config()
	if got, want := dcfg.Operations, (map[string][]string{
		"on": {"on", "command"}}); !compareOperationMaps(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	dcfg.Operations = nil
	if got, want := dcfg, (devices.DeviceConfigCommon{
		Name: "d", Controller: "c", Type: "device"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := dev.(*testutil.MockDevice).CustomConfig().(testutil.DeviceDetail), (testutil.DeviceDetail{Detail: "my-device-d"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

}

func TestBuildDevices(t *testing.T) {

	var ctrls []devices.ControllerConfig
	var devs []devices.DeviceConfig

	if err := yaml.Unmarshal([]byte(controllers_spec), &ctrls); err != nil {
		t.Fatalf("failed to unmarshal controllers: %v", err)
	}
	if err := yaml.Unmarshal([]byte(devices_spec), &devs); err != nil {
		t.Fatalf("failed to unmarshal devices: %v", err)
	}

	controllers, devices, err := devices.BuildDevices(ctrls, devs, supportedControllers, supportedDevices)

	if err != nil {
		t.Fatalf("failed to build devices: %v", err)
	}

	if got, want := len(controllers), 1; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
	if got, want := len(devices), 2; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	for _, dev := range devices {
		if got, want := dev.ControlledByName(), "c"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if got, want := dev.ControlledBy(), controllers["c"]; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	if got, want := devices["d"].(*testutil.MockDevice).Detail.Detail, "my-device-d"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := devices["e"].(*testutil.MockDevice).Detail.Detail, "my-device-e"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

}

func TestParseLocation(t *testing.T) {
	ctx := context.Background()
	gl := func(l string) *time.Location {
		loc, err := time.LoadLocation(l)
		if err != nil {
			t.Fatal(err)
		}
		return loc
	}
	for _, tc := range []struct {
		arg      string
		file     string
		expected *time.Location
	}{
		{"", "", gl("Local")},
		{"", "location:", gl("Local")},
		{"", "location: America/New_York", gl("America/New_York")},
		{"America/New_York", "", gl("America/New_York")},
	} {
		spec := tc.file
		system, err := devices.ParseSystemConfig(ctx, tc.arg, []byte(spec))
		if err != nil {
			t.Fatalf("failed to parse system config: %v", err)
		}
		if got, want := system.Location.String(), tc.expected.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestOperations(t *testing.T) {

	ctx := context.Background()
	system, err := devices.ParseSystemConfig(ctx, "", []byte(simple_spec))
	if err != nil {
		t.Fatalf("failed to parse system config: %v", err)
	}

	var out strings.Builder

	dev := system.Devices["d"]
	pars := dev.Config().Operations["on"]
	if err := dev.Operations()["on"](ctx, &out, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}

	if got, want := out.String(), "device[d].On: [2] on--command\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	out.Reset()

	ctrl := system.Controllers["c"]
	pars = ctrl.Config().Operations["enable"]
	if err := ctrl.Operations()["enable"](ctx, &out, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}
	if got, want := out.String(), "controller[c].Enable: [3] on--command--quoted with space\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	out.Reset()

	pars = ctrl.Config().Operations["disable"]
	if err := ctrl.Operations()["disable"](ctx, &out, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}
	if got, want := out.String(), "controller[c].Disable: [2] off--command\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
