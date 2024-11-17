package homeworks

import (
	"fmt"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

func NewController(typ string, opts devices.Options) (devices.Controller, error) {
	switch typ {
	case "homeworks-qs":
		return NewQSProcessor(opts), nil
	}
	return nil, fmt.Errorf("unsupported lutron controller/processor type %s", typ)
}

func NewDevice(typ string, opts devices.Options) (devices.Device, error) {
	switch typ {
	case "homeworks-blind":
		return &hwBlind{}, nil
	}
	return nil, fmt.Errorf("unsupported lutron device type %s", typ)
}

func SupportedDevices() devices.SupportedDevices {
	return devices.SupportedDevices{
		"homeworks-blind": NewDevice,
	}
}

func SupportedControllers() devices.SupportedControllers {
	return devices.SupportedControllers{
		"homeworks-qs": NewController,
	}
}

type hwDeviceBase struct {
	devices.DeviceConfigCommon
	controller devices.Controller
	processor  *QSProcessor
}

func (d *hwDeviceBase) SetConfig(c devices.DeviceConfigCommon) {
	d.DeviceConfigCommon = c
}

func (d *hwDeviceBase) Config() devices.DeviceConfigCommon {
	return d.DeviceConfigCommon
}

func (d *hwDeviceBase) SetController(c devices.Controller) {
	d.controller = c
	d.processor = c.Implementation().(*QSProcessor)
}

func (d *hwDeviceBase) ControlledByName() string {
	return d.Controller
}

func (d *hwDeviceBase) ControlledBy() devices.Controller {
	return d.controller
}

func (d *hwDeviceBase) Implementation() any {
	return d
}

func (d *hwDeviceBase) Timeout() time.Duration {
	return time.Minute
}

func (d *hwDeviceBase) UnmarshalYAML(node *yaml.Node) error {
	return nil
}
