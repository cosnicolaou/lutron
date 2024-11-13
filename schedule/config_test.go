package schedule_test

import (
	"context"
	"reflect"
	"slices"
	"testing"
	"time"

	"cloudeng.io/datetime"
	"github.com/cosnicolaou/lutron/schedule"
)

const auth_config = `
- key_id: home
  user: home
  token: token
`

const simple_config = `location: Local
controllers:
  - name: home
    type: controller
    ip_address: 172.16.1.50
    key_id: home

devices:
  - name: living room
    type: device
    controller: home
    on: on command
    off: off
    key_id: home

  - name: dining room
    type: device
    controller: home
    on: on command
    off: off
    key_id: home

schedules:
  - name: living room
    for: jan, feb
    actions:
      on: 08:12
      off: 20:01:13
    device: living room
 
  - name: dining room
    weekdays: true
    for: jan,feb
    exclude_dates: jan-02, feb-02
    actions:
      on: 12:00
      off: 16:00
    device: dining room

  - name: dining room 2
    ranges:
       - 01/22:2
       - 11/22:12/28
    actions:
      on: 12:00
      another: 12:00
      off: 16:00
    device: dining room
`

func newHourMin(h int, m int) datetime.TimeOfDay {
	return datetime.NewTimeOfDay(h, m, 0)
}

func newHourMinSec(h int, m int, s int) datetime.TimeOfDay {
	return datetime.NewTimeOfDay(h, m, s)
}

func newDate(month datetime.Month, day int) datetime.Date {
	return datetime.NewDate(month, day)
}

func TestParse(t *testing.T) {
	ctx := context.Background()
	yp := datetime.YearAndPlace{}
	keys, cfg, err := schedule.ParseConfig(ctx, yp, []byte(auth_config), []byte(simple_config))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := cfg.Place.String(), time.Local.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := len(cfg.Controllers), 1; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if got, want := len(cfg.Devices), 2; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if got, want := len(cfg.Schedules), 3; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if got, want := cfg.Schedules[0].Name, "living room"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	dev := cfg.Devices[1]
	if got, want := dev.Type, "device"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	vals := []string{}
	for _, c := range dev.Config.Content {
		vals = append(vals, c.Value)

	}
	if got, want := vals, []string{"name", "dining room", "type", "device", "controller", "home", "on", "on command", "off", "off", "key_id", "home"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := len(keys), 1; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
	if got, want := keys["home"].User, "home"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	sched1 := cfg.Schedules[0]

	ntod := newHourMinSec
	nd := newDate

	if got, want := sched1.Actions[0].Action.ActionName, "on"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched1.Actions[1].Action.ActionName, "off"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched1.Actions[0].Due, ntod(8, 12, 0); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched1.Actions[1].Due, ntod(20, 01, 13); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched1.Dates.For, (datetime.MonthList{1, 2}); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	sched2 := cfg.Schedules[1]
	if got, want := sched2.Dates.For, (datetime.MonthList{1, 2}); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched2.Dates.Constraints.Weekdays, true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := sched2.Dates.Constraints.Custom, (datetime.DateList{nd(1, 2), nd(2, 2)}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	sched3 := cfg.Schedules[2]
	expected := datetime.DateRangeList{
		datetime.NewDateRange(nd(1, 22), nd(2, 0)),
		datetime.NewDateRange(nd(11, 22), nd(12, 28)),
	}
	if got, want := sched3.Dates.Ranges, expected; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

}
