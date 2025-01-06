package homeworks

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cosnicolaou/automation/devices"
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
	case "shadegrp":
		return &HWShadeGroup{
			hwDeviceBase: hwDeviceBase{
				logger: opts.Logger.With(
					"protocol", "homeworks-qs",
					"device", "shadegrp")}}, nil
	case "shade":
		return &HWShade{hwDeviceBase: hwDeviceBase{
			logger: opts.Logger.With(
				"protocol", "homeworks-qs",
				"device", "shade")}}, nil
	}
	return nil, fmt.Errorf("unsupported lutron device type %s", typ)
}

func SupportedDevices() devices.SupportedDevices {
	return devices.SupportedDevices{
		"shadegrp": NewDevice,
		"shade":    NewDevice,
	}
}

func SupportedControllers() devices.SupportedControllers {
	return devices.SupportedControllers{
		"homeworks-qs": NewController,
	}
}

type hwDeviceBase struct {
	devices.DeviceConfigCommon
	processor *QSProcessor
	logger    *slog.Logger
}

func (d *hwDeviceBase) SetConfig(c devices.DeviceConfigCommon) {
	d.DeviceConfigCommon = c
}

func (d *hwDeviceBase) Config() devices.DeviceConfigCommon {
	return d.DeviceConfigCommon
}

func (d *hwDeviceBase) SetController(c devices.Controller) {
	d.processor = c.Implementation().(*QSProcessor)
}

func (d *hwDeviceBase) ControlledByName() string {
	return d.Controller
}

func (d *hwDeviceBase) ControlledBy() devices.Controller {
	return d.processor
}

func (d *hwDeviceBase) Timeout() time.Duration {
	return time.Minute
}

func (d *hwDeviceBase) Conditions() map[string]devices.Condition {
	return map[string]devices.Condition{}
}

func (d *hwDeviceBase) ConditionsHelp() map[string]string {
	return map[string]string{}
}
