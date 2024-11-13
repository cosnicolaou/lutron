package lutron

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ControllerSpec struct {
	Name      string `yaml:"name"`
	IPAddress string `yaml:"ip_address"`
	AuthID    string `yaml:"auth_id"`
}

type DeviceSpec struct {
	On_Action  string `yaml:"on_script"`
	Off_action string `yaml:"off_script"`
}

func New(typ string) (yaml.Unmarshaler, error) {
	switch typ {
	case "lutron-controller":
		return &Controller{}, nil
	case "lutron-device":
		return &Device{}, nil
	}
	return &Device{}, fmt.Errorf("unsupported device type %s", typ)
}

var (
	controllers = map[string]*Controller{}
	devices     = map[string]*Device{}
)

func (d *Device) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&d.spec); err != nil {
		return err
	}
	return nil
}

func (d *Controller) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&d.spec); err != nil {
		return err
	}
	// Validate connection to controller.
	return nil
}

type Controller struct {
	spec ControllerSpec
}

type Device struct {
	spec DeviceSpec
}

/*

func (lp *LutronDeviceSpec) UnmarshalYAML(value *yaml.Node) error {
	var spec LutronDeviceSpec
	if err := value.Decode(&spec); err != nil {
		return err
	}
	return nil
}

func newLutronDevice(typ string) yaml.Unmarshaler {
	return &LutronDevice{}
}

func (lp *LutronDeviceSpec) UnmarshalYAML(value *yaml.Node) error {
	var spec LutronDeviceSpec
	if err := value.Decode(&spec); err != nil {
		return err
	}
	return nil
}



type LutronDevice struct {
	spec LutronDeviceSpec
}

func (ld *LutronDevice) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&ld.spec); err != nil {
		return err
	}
	return nil
}

func (ld *LutronDevice) On(ctx context.Context) error {
	return nil
}

func (ld *LutronDevice) Off(ctx context.Context) error {
	return nil
}
*/
