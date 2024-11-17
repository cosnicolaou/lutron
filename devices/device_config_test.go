// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices_test

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type detail struct {
	Detail string `yaml:"detail"`
	KeyID  string `yaml:"key_id"`
}

type controller struct {
	devices.ControllerConfigCommon
	out    *strings.Builder
	Detail detail `yaml:",inline"`
}

func (c *controller) SetConfig(cfg devices.ControllerConfigCommon) {
	c.ControllerConfigCommon = cfg
}

func (c controller) Config() devices.ControllerConfigCommon {
	return c.ControllerConfigCommon
}

func (c *controller) CustomConfig() any {
	return c.Detail
}

func (c *controller) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&c.Detail)
}

func (c *controller) On(ctx context.Context, args ...string) error {
	if c.out != nil {
		fmt.Fprintf(c.out, "controller[%s].On: [%d] %v\n", c.Name, len(args), strings.Join(args, "--"))
	}
	return nil
}

func (c *controller) Off(ctx context.Context, args ...string) error {
	if c.out != nil {
		fmt.Fprintf(c.out, "controller[%s].Off: [%d] %v\n", c.Name, len(args), strings.Join(args, "--"))
	}
	return nil
}

func (c *controller) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{"on": c.On, "off": c.Off}
}

func (c *controller) Implementation() any {
	return c
}

type deviceConfig struct {
	Detail detail `yaml:",inline"`
}

type device struct {
	devices.DeviceConfigCommon
	controller devices.Controller
	cfg        deviceConfig
	out        *strings.Builder
}

func (d *device) SetConfig(cfg devices.DeviceConfigCommon) {
	d.DeviceConfigCommon = cfg
}

func (d device) Config() devices.DeviceConfigCommon {
	return d.DeviceConfigCommon
}

func (d *device) CustomConfig() any {
	return d.cfg
}

func (d *device) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&d.cfg)
}

func (d *device) Implementation() any {
	return d
}

func (d *device) SetController(c devices.Controller) {
	d.controller = c
}

func (d *device) ControlledByName() string {
	return d.Controller
}

func (d *device) ControlledBy() devices.Controller {
	return d.controller
}

func (d *device) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{"on": d.On}
}

func (d *device) Timeout() time.Duration {
	return time.Second
}

func (d *device) On(ctx context.Context, args ...string) error {
	if d.out != nil {
		fmt.Fprintf(d.out, "device[%s][%d] %v\n", d.Name, len(args), strings.Join(args, "--"))
	}
	return nil
}

const controllers_spec = `
  - name: c
    type: controller
    operations:
      on: [on, command, quoted with space]
      off: [off, command]
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
		return &controller{}, nil
	},
}

var supportedDevices = devices.SupportedDevices{
	"device": func(string, devices.Options) (devices.Device, error) {
		return &device{}, nil
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
		"on":  {"on", "command", "quoted with space"},
		"off": {"off", "command"}}); !compareOperationMaps(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	ccfg.Operations = nil

	if got, want := ccfg, (devices.ControllerConfigCommon{
		Name: "c", Type: "controller"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := ctrl.(*controller).Detail, (detail{
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

	if got, want := dev.(*device).cfg.Detail, (detail{Detail: "my-device-d"}); !reflect.DeepEqual(got, want) {
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

	if got, want := devices["d"].(*device).cfg.Detail.Detail, "my-device-d"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := devices["e"].(*device).cfg.Detail.Detail, "my-device-e"; got != want {
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
	oc, od := devices.AvailableControllers, devices.AvailableDevices
	defer func() {
		devices.AvailableControllers, devices.AvailableDevices = oc, od
	}()

	var out strings.Builder

	devices.AvailableControllers = devices.SupportedControllers{
		"controller": func(string, devices.Options) (devices.Controller, error) {
			return &controller{out: &out}, nil
		},
	}

	devices.AvailableDevices = devices.SupportedDevices{
		"device": func(string, devices.Options) (devices.Device, error) {
			return &device{out: &out}, nil
		},
	}

	ctx := context.Background()
	system, err := devices.ParseSystemConfig(ctx, "", []byte(simple_spec))
	if err != nil {
		t.Fatalf("failed to parse system config: %v", err)
	}

	dev := system.Devices["d"]
	pars := dev.Config().Operations["on"]
	if err := dev.Operations()["on"](ctx, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}

	if got, want := out.String(), "device[d][2] on--command\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	out.Reset()

	ctrl := system.Controllers["c"]
	pars = ctrl.Config().Operations["on"]
	if err := ctrl.Operations()["on"](ctx, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}
	if got, want := out.String(), "controller[c].On: [3] on--command--quoted with space\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	out.Reset()

	pars = ctrl.Config().Operations["off"]
	if err := ctrl.Operations()["off"](ctx, pars...); err != nil {
		t.Errorf("failed to perform operation: %v", err)
	}
	if got, want := out.String(), "controller[c].Off: [2] off--command\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
