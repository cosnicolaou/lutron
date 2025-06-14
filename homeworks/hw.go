package homeworks

import (
	"fmt"

	"github.com/cosnicolaou/automation/devices"
)

func NewController(typ string, opts devices.Options) (devices.Controller, error) {
	if typ == "homeworks-qs" {
		return NewQSProcessor(opts), nil
	}
	return nil, fmt.Errorf("unsupported lutron controller/processor type %s", typ)
}

func NewDevice(typ string, _ devices.Options) (devices.Device, error) {
	switch typ {
	case "shadegrp":
		return &HWShadeGroup{hwShadeBase: hwShadeBase{}}, nil
	case "shade":
		return &HWShade{hwShadeBase: hwShadeBase{}}, nil
	case "contact-closure-open-close":
		return &ContactClosureOpenClose{}, nil

	}
	return nil, fmt.Errorf("unsupported lutron device type %s", typ)
}

func SupportedDevices() devices.SupportedDevices {
	return devices.SupportedDevices{
		"shadegrp":                   NewDevice,
		"shade":                      NewDevice,
		"contact-closure":            NewDevice,
		"contact-closure-open-close": NewDevice,
	}
}

func SupportedControllers() devices.SupportedControllers {
	return devices.SupportedControllers{
		"homeworks-qs": NewController,
	}
}
