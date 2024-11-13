package schedule

import (
	"context"
	"fmt"
	"time"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/cmdutil/keystore"
	"cloudeng.io/datetime"
	"cloudeng.io/datetime/schedule"
	"gopkg.in/yaml.v3"
)

type monthList datetime.MonthList
type timeOfDay datetime.TimeOfDay

func (ml *monthList) UnmarshalYAML(node *yaml.Node) error {
	return (*datetime.MonthList)(ml).Parse(node.Value)
}

func (t *timeOfDay) UnmarshalYAML(node *yaml.Node) error {
	return (*datetime.TimeOfDay)(t).Parse(node.Value)
}

type constraintsConfig struct {
	Weekdays bool   `yaml:"weekdays" cmd:"only on weekdays"`
	Weekends bool   `yaml:"weekends" cmd:"only on weekends"`
	Custom   string `yaml:"exclude_dates" cmd:"exclude the specified dates eg: 01/02,jan-02"`
}

func (cc constraintsConfig) parse() (datetime.Constraints, error) {
	dc := datetime.Constraints{
		Weekdays: cc.Weekdays,
		Weekends: cc.Weekends,
	}
	if err := dc.Custom.Parse(cc.Custom); err != nil {
		return datetime.Constraints{}, err
	}
	return dc, nil
}

type datesConfig struct {
	For          monthList         `yaml:"for" cmd:"for the specified months"`
	MirrorMonths bool              `yaml:"mirror_months" cmd:"include the mirror months, ie. those equidistant from the soltices for the set of 'for' months"`
	Ranges       []string          `yaml:"ranges" cmd:"for the specified date ranges"`
	Constraints  constraintsConfig `yaml:",inline" cmd:"constrain the dates"`
}

func (dc *datesConfig) parse() (schedule.Dates, error) {
	d := schedule.Dates{
		For:          datetime.MonthList(dc.For),
		MirrorMonths: dc.MirrorMonths,
	}
	if err := d.Ranges.Parse(dc.Ranges); err != nil {
		return schedule.Dates{}, err
	}
	cc, err := dc.Constraints.parse()
	if err != nil {
		return schedule.Dates{}, err
	}
	d.Constraints = cc
	return d, nil
}

type actionScheduleConfig struct {
	Name    string               `yaml:"name" cmd:"the name of the schedule"`
	Device  string               `yaml:"device" cmd:"the name of the device that the schedule applies to"`
	Dates   datesConfig          `yaml:",inline" cmd:"the dates that the schedule applies to"`
	Actions map[string]timeOfDay `yaml:"actions" cmd:"the actions to be taken when the schedule is current"`
}

type ControllerConfigCommon struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type ControllerConfig struct {
	ControllerConfigCommon
	Config yaml.Node `yaml:",inline"`
}

func (lp *ControllerConfig) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&lp.ControllerConfigCommon); err != nil {
		return err
	}
	return node.Decode(&lp.Config)
}

type DeviceConfigCommon struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Controller string `yaml:"controller"`
}

type DeviceConfig struct {
	DeviceConfigCommon
	Config yaml.Node `yaml:",inline"`
}

func (lp *DeviceConfig) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&lp.DeviceConfigCommon); err != nil {
		return err
	}
	return node.Decode(&lp.Config)
}

type config struct {
	Location    string                 `yaml:"location" cmd:"the location for which the schedule is being created in time.Location format"`
	Controllers []ControllerConfig     `yaml:"controllers" cmd:"the controllers that are being configured"`
	Devices     []DeviceConfig         `yaml:"devices" cmd:"the devices that are being configured"`
	Schedules   []actionScheduleConfig `yaml:"schedules" cmd:"the schedules"`
}

type Action struct {
	DeviceName string
	Device     Device
	ActionName string
	Action     Operation
}

type Schedules struct {
	datetime.YearAndPlace
	Controllers []ControllerConfig
	Devices     []DeviceConfig
	Schedules   []schedule.Annual[Action]
}

func (s Schedules) Lookup(name string) schedule.Annual[Action] {
	for _, sched := range s.Schedules {
		if sched.Name == name {
			return sched
		}
	}
	return schedule.Annual[Action]{}
}

func ParseConfigFiles(ctx context.Context, yp datetime.YearAndPlace, authFile, cfgFile string, uriHandlers map[string]cmdyaml.URLHandler) (keystore.Keys, Schedules, error) {
	keys, err := keystore.ParseConfigURI(ctx, authFile, uriHandlers)
	if err != nil {
		return nil, Schedules{}, err
	}
	var cfg config
	if err := cmdyaml.ParseConfigFile(ctx, cfgFile, &cfg); err != nil {
		return nil, Schedules{}, err
	}
	pcfg, err := cfg.createSchedules(yp)
	if err != nil {
		return nil, Schedules{}, err
	}
	return keys, pcfg, nil
}

func ParseConfig(ctx context.Context, yp datetime.YearAndPlace, authData, cfgData []byte) (keystore.Keys, Schedules, error) {
	keys, err := keystore.Parse(authData)
	if err != nil {
		return nil, Schedules{}, err
	}
	var cfg config
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, Schedules{}, err
	}
	pcfg, err := cfg.createSchedules(yp)
	if err != nil {
		return nil, Schedules{}, err
	}
	return keys, pcfg, err
}

func (cfg config) createSchedules(yp datetime.YearAndPlace) (Schedules, error) {
	var sched Schedules
	if yp.Place == nil {
		location := time.Now().Location()
		if len(cfg.Location) > 0 {
			var err error
			location, err = time.LoadLocation(cfg.Location)
			if err != nil {
				return Schedules{}, err
			}
		}
		yp.Place = location
	}
	sched.YearAndPlace = yp
	names := map[string]struct{}{}
	for _, csched := range cfg.Schedules {
		if _, ok := names[csched.Name]; ok {
			return Schedules{}, fmt.Errorf("duplicate schedule name: %v", csched.Name)
		}
		names[csched.Name] = struct{}{}
		var annual schedule.Annual[Action]
		annual.Name = csched.Name
		dates, err := csched.Dates.parse()
		if err != nil {
			return Schedules{}, err
		}
		annual.Dates = dates
		for name, when := range csched.Actions {
			annual.Actions = append(annual.Actions, schedule.Action[Action]{
				Due:  datetime.TimeOfDay(when),
				Name: name,
				Action: Action{
					DeviceName: csched.Device,
					ActionName: name,
				}})
		}
		annual.Actions.Sort()
		sched.Schedules = append(sched.Schedules, annual)
	}
	sched.Devices = cfg.Devices
	sched.Controllers = cfg.Controllers
	return sched, nil
}
