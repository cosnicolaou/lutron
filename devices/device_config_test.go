// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type detail struct {
	Detail string `yaml:"detail"`
}

type controller struct {
	devices.ControllerConfigCommon
	Detail detail `yaml:",inline"`
}

func (c *controller) SetConfig(cfg devices.ControllerConfigCommon) {
	c.ControllerConfigCommon = cfg
}

func (c controller) Config() devices.ControllerConfigCommon {
	return c.ControllerConfigCommon
}

func (c *controller) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&c.Detail)
}

func (c *controller) Implementation() any {
	return c
}

type device struct {
	devices.DeviceConfigCommon
	controller devices.Controller
	Detail     detail `yaml:",inline"`
}

func (d *device) SetConfig(cfg devices.DeviceConfigCommon) {
	d.DeviceConfigCommon = cfg
}

func (d device) Config() devices.DeviceConfigCommon {
	return d.DeviceConfigCommon
}

func (d *device) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&d.Detail)
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
	return map[string]devices.Operation{"foo": nil}
}

func (d *device) Timeout() time.Duration {
	return time.Second
}

const controller_spec = `name: c
type: controller
detail: my-location
`

const device_spec = `name: d
controller: c
type: device
detail: my-device
`

var supportedControllers = devices.SupportedControllers{
	"controller": func(string, ...devices.Option) (devices.Controller, error) {
		return &controller{}, nil
	},
}

var supportedDevices = devices.SupportedDevices{
	"device": func(string, ...devices.Option) (devices.Device, error) {
		return &device{}, nil
	},
}

func TestParseConfig(t *testing.T) {
	var c devices.ControllerConfig
	if err := yaml.Unmarshal([]byte(controller_spec), &c); err != nil {
		t.Fatalf("failed to unmarshal controller: %v", err)
	}
	var d devices.DeviceConfig
	if err := yaml.Unmarshal([]byte(device_spec), &d); err != nil {
		t.Fatalf("failed to unmarshal device: %v", err)
	}

	ctrls, devs, err := devices.BuildDevices(
		[]devices.ControllerConfig{c},
		[]devices.DeviceConfig{d},
		supportedControllers,
		supportedDevices)
	if err != nil {
		t.Fatalf("failed to build devices: %v", err)
	}

	ctrl := ctrls["c"]
	dev := devs["d"]

	if got, want := ctrl.Config(), c.ControllerConfigCommon; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := ctrl.(*controller).Detail, (detail{Detail: "my-location"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := dev.Config(), d.DeviceConfigCommon; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	if got, want := dev.(*device).Detail, (detail{Detail: "my-device"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

}
