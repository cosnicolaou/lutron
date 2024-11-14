package homeworks

import (
	"fmt"

	"github.com/cosnicolaou/lutron/devices"
)

func NewController(typ string) (devices.Controller, error) {
	switch typ {
	case "homeworks-qs":
		return newQSProcessor(), nil
	}
	return nil, fmt.Errorf("unsupported lutron controller/processor type %s", typ)
}

func NewDevice(typ string) (devices.Device, error) {
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
