// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package schedule

import (
	"context"
	"fmt"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/datetime"
	"cloudeng.io/datetime/schedule"
	"github.com/cosnicolaou/lutron/devices"
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

type schedulesConfig struct {
	Schedules []actionScheduleConfig `yaml:"schedules" cmd:"the schedules"`
}

type Schedules struct {
	Schedules []schedule.Annual[devices.Action]
}

func (s Schedules) Lookup(name string) schedule.Annual[devices.Action] {
	for _, sched := range s.Schedules {
		if sched.Name == name {
			return sched
		}
	}
	return schedule.Annual[devices.Action]{}
}

func ParseScheduleConfigFile(ctx context.Context, cfgFile string, devs map[string]devices.Device) (Schedules, error) {
	var cfg schedulesConfig
	if err := cmdyaml.ParseConfigFile(ctx, cfgFile, &cfg); err != nil {
		return Schedules{}, err
	}
	pcfg, err := cfg.createSchedules(devs)
	if err != nil {
		return Schedules{}, err
	}
	return pcfg, nil
}

func ParseScheduleConfig(ctx context.Context, cfgData []byte, devs map[string]devices.Device) (Schedules, error) {
	var cfg schedulesConfig
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return Schedules{}, err
	}
	pcfg, err := cfg.createSchedules(devs)
	if err != nil {
		return Schedules{}, err
	}
	return pcfg, err
}

func (cfg schedulesConfig) createSchedules(devs map[string]devices.Device) (Schedules, error) {
	var sched Schedules
	names := map[string]struct{}{}
	for _, csched := range cfg.Schedules {
		if _, ok := names[csched.Name]; ok {
			return Schedules{}, fmt.Errorf("duplicate schedule name: %v", csched.Name)
		}
		names[csched.Name] = struct{}{}
		var annual schedule.Annual[devices.Action]
		annual.Name = csched.Name
		dates, err := csched.Dates.parse()
		if err != nil {
			return Schedules{}, err
		}
		annual.Dates = dates
		for name, when := range csched.Actions {
			if _, ok := devs[csched.Device]; !ok {
				return Schedules{}, fmt.Errorf("unknown device: %s", csched.Device)
			}
			annual.Actions = append(annual.Actions, schedule.Action[devices.Action]{
				Due:  datetime.TimeOfDay(when),
				Name: name,
				Action: devices.Action{
					DeviceName: csched.Device,
					ActionName: name,
				}})
		}
		annual.Actions.Sort()
		sched.Schedules = append(sched.Schedules, annual)
	}
	return sched, nil
}
