// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package schedule_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"cloudeng.io/datetime"
	"cloudeng.io/errors"
	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/schedule"
	"gopkg.in/yaml.v3"
)

type controlConfigSpecific struct {
	IP string `yaml:"ip_address"`
}

type controllerConfig struct {
	devices.ControllerConfigCommon
	cfg controlConfigSpecific
}

type test_controller struct {
	controllerConfig
}

func (t *test_controller) SetConfig(c devices.ControllerConfigCommon) {
	t.controllerConfig.ControllerConfigCommon = c
}

func (t *test_controller) Config() devices.ControllerConfigCommon {
	return t.controllerConfig.ControllerConfigCommon
}

func (t *test_controller) Implementation() any {
	return t
}

func (t *test_controller) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&t.controllerConfig.cfg); err != nil {
		return err
	}
	return nil
}

func new_controller(string, devices.Options) (devices.Controller, error) {
	return &test_controller{}, nil
}

type deviceConfigSpecifc struct {
	On      string `yaml:"on"`
	Off     string `yaml:"off"`
	Another string `yaml:"another"`
}

type deviceConfig struct {
	devices.DeviceConfigCommon
	cfg deviceConfigSpecifc
}

type test_device struct {
	deviceConfig
	ctrl devices.Controller
	out  io.Writer
}

func (t *test_device) SetConfig(c devices.DeviceConfigCommon) {
	t.deviceConfig.DeviceConfigCommon = c
}

func (t *test_device) Config() devices.DeviceConfigCommon {
	return t.deviceConfig.DeviceConfigCommon
}

func (t *test_device) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&t.deviceConfig.cfg)
}

func (t *test_device) SetController(c devices.Controller) {
	t.ctrl = c
}

func (t *test_device) ControlledByName() string {
	return t.DeviceConfigCommon.Controller
}

func (t *test_device) ControlledBy() devices.Controller {
	return t.ctrl
}

func (t *test_device) Implementation() any {
	return t
}

func (t *test_device) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"on":      t.On,
		"off":     t.Off,
		"another": t.Another,
	}
}

func (t *test_device) Timeout() time.Duration {
	return time.Second * 20
}

func (t *test_device) On(context.Context) error {
	fmt.Fprintf(t.out, "on\n")
	return nil
}

func (t *test_device) Off(context.Context) error {
	fmt.Fprintf(t.out, "off\n")
	return nil
}

func (t *test_device) Another(context.Context) error {
	fmt.Fprintf(t.out, "another\n")
	return nil
}

func new_device(string, devices.Options) (devices.Device, error) {
	return &test_device{out: os.Stderr}, nil
}

type slow_test_device struct {
	test_device
	timeout time.Duration
	delay   time.Duration
}

func (st *slow_test_device) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"on": st.On,
	}
}

func (st *slow_test_device) Timeout() time.Duration {
	return st.timeout
}

func (st *slow_test_device) On(context.Context) error {
	time.Sleep(st.delay)
	return nil
}

var supportedDevices = devices.SupportedDevices{
	"device": new_device,
}

var supportedControllers = devices.SupportedControllers{
	"controller": new_controller,
}

type timesource struct {
	ch chan time.Time
}

func (t *timesource) NowIn(loc *time.Location) time.Time {
	n := <-t.ch
	return n.In(loc)
}

func (t *timesource) tick(nextTick time.Time) {
	t.ch <- nextTick
}

type testAction struct {
	when   time.Time
	action schedule.Action
}

func allScheduled(s *schedule.Scheduler, yp datetime.YearAndPlace) ([]testAction, []time.Time) {
	actions := []testAction{}
	times := []time.Time{}
	for active := range s.Scheduled(yp) {
		if len(active.Actions) == 0 {
			continue
		}
		times = append(times, datetime.Time(yp, active.Date, active.Actions[0].Due))
		for _, action := range active.Actions {
			actions = append(actions, testAction{
				action: action.Action,
				when:   datetime.Time(yp, active.Date, action.Due),
			})
		}
	}
	return actions, times
}

type recorder struct {
	sync.Mutex
	out *bytes.Buffer
}

func (r *recorder) Write(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	return r.out.Write(p)
}

func (r *recorder) Lines() []string {
	lines := []string{}
	for _, l := range bytes.Split(r.out.Bytes(), []byte("\n")) {
		if len(l) == 0 {
			continue
		}
		lines = append(lines, string(l))
	}
	return lines
}

type logEntry struct {
	Sched      string    `json:"sched"`
	Msg        string    `json:"msg"`
	Op         string    `json:"op"`
	Due        time.Time `json:"due"`
	NumActions int       `json:"#actions"`
	Error      string    `json:"err"`
}

func (r *recorder) Logs(t *testing.T) []logEntry {
	entries := []logEntry{}
	for _, l := range bytes.Split(r.out.Bytes(), []byte("\n")) {
		if len(l) == 0 {
			continue
		}
		var e logEntry
		if err := json.Unmarshal(l, &e); err != nil {
			t.Errorf("failed to unmarshal: %v: %v", string(l), err)
			return nil
		}
		if e.NumActions != 0 || e.Msg == "late" {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

func containsError(logs []logEntry) error {
	for _, l := range logs {
		if l.Error != "" {
			return errors.New(l.Error)
		}
	}
	return nil
}

func newRecorder() *recorder {
	return &recorder{out: bytes.NewBuffer(nil)}
}

func setupSchedules(t *testing.T, yp datetime.YearAndPlace, config string) (spec schedule.Schedules, devs map[string]devices.Device, deviceRecorder, logRecorder *recorder, logger *slog.Logger) {
	_, spec, err := schedule.ParseConfig(context.Background(), yp, []byte(auth_config), []byte(config))
	if err != nil {
		t.Fatal(err)
	}

	deviceRecorder = newRecorder()
	logRecorder = newRecorder()
	logger = slog.New(slog.NewJSONHandler(logRecorder, nil))

	supportedDevices := devices.SupportedDevices{
		"device": func(string, devices.Options) (devices.Device, error) {
			return &test_device{out: deviceRecorder}, nil
		},
		"slow_device": func(string, devices.Options) (devices.Device, error) {
			return &slow_test_device{
				test_device: test_device{out: deviceRecorder},
				timeout:     time.Millisecond * 10,
				delay:       time.Minute,
			}, nil
		},
		"hanging_device": func(string, devices.Options) (devices.Device, error) {
			return &slow_test_device{
				test_device: test_device{out: deviceRecorder},
				timeout:     time.Hour,
				delay:       time.Hour,
			}, nil
		},
	}

	_, devs, err = devices.BuildDevices(spec.Controllers, spec.Devices, supportedControllers, supportedDevices)
	if err != nil {
		t.Fatal(err)
	}

	return spec, devs, deviceRecorder, logRecorder, logger
}

func TestScheduler(t *testing.T) {
	ctx := context.Background()
	yp := datetime.YearAndPlace{Year: 2021}

	spec, devices, deviceRecorder, logRecorder, logger := setupSchedules(t, yp, simple_config)
	yp = spec.YearAndPlace

	ts := &timesource{ch: make(chan time.Time, 1)}
	opts := []schedule.Option{
		schedule.WithTimeSource(ts),
		schedule.WithLogger(logger),
	}

	// dining room schedule
	diningRoom := spec.Lookup("dining room 2")
	scheduler, err := schedule.NewScheduler(diningRoom, yp.Place, devices, opts...)
	if err != nil {
		t.Fatal(err)
	}

	all, times := allScheduled(scheduler, yp)
	var errs errors.M
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errs.Append(scheduler.RunYear(ctx, yp))
		wg.Done()
	}()
	for _, t := range times {
		ts.tick(t)
		time.Sleep(time.Millisecond * 2)
	}
	wg.Wait()
	if err := errs.Err(); err != nil {
		t.Fatal(err)
	}

	logs := logRecorder.Logs(t)
	if err := containsError(logs); err != nil {
		t.Fatal(err)
	}

	// 01/22:2, 11/22:12/28 translates to:
	// 10+28+9+28 days
	days := 10 + 28 + 9 + 28
	if got, want := len(all), days*3; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	if got, want := len(logs), days*3; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	timeIterator := 0
	for i := range len(logs) / 3 {
		lg1, lg2, lg3 := logs[i*3], logs[i*3+1], logs[i*3+2]
		// expect to see on, another, off, or another, on, off
		// since on and another are co-scheduled.
		if got, want := lg1.Due, times[timeIterator]; !got.Equal(want) {
			t.Errorf("%#v: got %v, want %v", lg1, got, want)
		}
		if got, want := lg2.Due, times[timeIterator]; !got.Equal(want) {
			t.Errorf("%#v: got %v, want %v", lg2, got, want)
		}
		timeIterator++
		if got, want := lg3.Due, times[timeIterator]; !got.Equal(want) {
			t.Errorf("%#v: got %v, want %v", lg3, got, want)
		}
		timeIterator++
		want1 := "on"
		want2 := "another"
		if lg1.Op == "another" {
			want2, want1 = want1, want2
		}
		if got, want := lg1.Op, want1; got != want {
			t.Errorf("%#v: got %v, want %v", lg1, got, want)
		}
		if got, want := lg2.Op, want2; got != want {
			t.Errorf("%#v: got %v, want %v", lg2, got, want)
		}
		if got, want := lg3.Op, "off"; got != want {
			t.Errorf("%#v: got %v, want %v", lg3, got, want)
		}
	}

	lines := deviceRecorder.Lines()
	if got, want := len(lines), days*3; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
	for i := range len(lines) / 3 {
		l1, l2, l3 := lines[i*3], lines[i*3+1], lines[i*3+2]
		// expect to see on, another, off, or another, on, off
		// since on and another are co-scheduled.
		want1 := "on"
		want2 := "another"
		if l1 == "another" {
			want2, want1 = want1, want2
		}
		if got, want := l1, want1; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := l2, want2; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := l3, "off"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

const scheduletest_config = `
location: Local
controllers:
  - name: magnolia
    type: controller
devices:
  - name: device
    type: device
  - name: slow
    type: slow_device
  - name: hanging
    type: hanging_device
schedules:
  - name: simple
    device: device
    actions:
      on: 00:00:00
      off: 00:00:01
  - name: slow
    device: slow
    actions:
      on: 00:00:00
  - name: hanging
    device: hanging
    actions:
      on: 00:00:00
 `

func TestScheduleRealTime(t *testing.T) {
	ctx := context.Background()
	yp := datetime.YearAndPlace{Year: time.Now().Year()}

	spec, devices, deviceRecorder, logRecorder, logger := setupSchedules(t, yp, scheduletest_config)
	yp = spec.YearAndPlace

	now := time.Now().In(yp.Place)
	today := datetime.DateFromTime(now)
	sched := spec.Lookup("simple")
	sched.Dates.Ranges = []datetime.DateRange{datetime.NewDateRange(today, today)}

	sched.Actions[0].Due = datetime.TimeOfDayFromTime(now.Add(time.Second))
	sched.Actions[1].Due = datetime.TimeOfDayFromTime(now.Add(2 * time.Second))

	opts := []schedule.Option{schedule.WithLogger(logger)}
	scheduler, err := schedule.NewScheduler(sched, yp.Place, devices, opts...)
	if err != nil {
		t.Fatal(err)
	}

	all, _ := allScheduled(scheduler, yp)

	var errs errors.M
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errs.Append(scheduler.RunYear(ctx, yp))
		wg.Done()
	}()
	wg.Wait()
	if err := errs.Err(); err != nil {
		t.Fatal(err)
	}

	logs := logRecorder.Logs(t)
	if err := containsError(logs); err != nil {
		t.Fatal(err)
	}
	lines := deviceRecorder.Lines()

	if got, want := len(logs), len(all); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
	if got, want := len(lines), len(all); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
	for i := range all {
		if got, want := logs[i].Due, all[i].when; !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := logs[i].Op, all[i].action.ActionName; got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		if got, want := lines[i], all[i].action.ActionName; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestTimeout(t *testing.T) {
	ctx := context.Background()
	yp := datetime.YearAndPlace{Year: time.Now().Year()}

	spec, devices, _, logRecorder, logger := setupSchedules(t, yp, scheduletest_config)
	yp = spec.YearAndPlace

	for _, tc := range []struct {
		sched  string
		cancel bool
		errmsg string
	}{
		{"slow", false, "context deadline exceeded"},   // timeout
		{"hanging", true, "context deadline exceeded"}, // hanging, must be canceled
	} {
		ctx, cancel := context.WithCancel(ctx)

		now := time.Now().In(yp.Place)
		today := datetime.DateFromTime(now)
		sched := spec.Lookup(tc.sched) // slow device schedule
		sched.Dates.Ranges = []datetime.DateRange{datetime.NewDateRange(today, today)}

		sched.Actions[0].Due = datetime.TimeOfDayFromTime(now.Add(time.Second))

		opts := []schedule.Option{schedule.WithLogger(logger)}
		scheduler, err := schedule.NewScheduler(sched, yp.Place, devices, opts...)
		if err != nil {
			t.Fatal(err)
		}
		if tc.cancel {
			go func() {
				time.Sleep(time.Second)
				cancel()
			}()
		}
		if err := scheduler.RunYear(ctx, yp); err != nil {
			t.Fatal(err)
		}

		logs := logRecorder.Logs(t)
		if err := containsError(logs); err == nil || err.Error() != tc.errmsg {
			t.Errorf("unexpected or missing error: %v", err)
		}

		cancel()
	}
}

const multiyear_config = `
location: Local
controllers:
  - name: home
    type: controller
devices:
  - name: device
    type: device
schedules:
  - name: simple
    device: device
    actions:
      on: 00:00:00
    ranges:
      - 02/25:02
 `

func TestMultiYear(t *testing.T) {
	ctx := context.Background()
	yp := datetime.YearAndPlace{Year: 2023}

	spec, devices, deviceRecorder, logRecorder, logger := setupSchedules(t, yp, multiyear_config)
	yp = spec.YearAndPlace

	ts := &timesource{ch: make(chan time.Time, 1)}
	opts := []schedule.Option{
		schedule.WithTimeSource(ts),
		schedule.WithLogger(logger),
	}

	scheduler, err := schedule.NewScheduler(spec.Lookup("simple"), yp.Place, devices, opts...)
	if err != nil {
		t.Fatal(err)
	}

	all2023, times2023 := allScheduled(scheduler, yp)
	all2024, times2024 := allScheduled(scheduler, datetime.YearAndPlace{Year: 2024, Place: yp.Place})
	times := append(times2023, times2024...)
	all := append(all2023, all2024...)
	if len(times) != len(all) {
		t.Fatalf("mismatch: %v %v", len(times), len(all))
	}

	var errs errors.M
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errs.Append(scheduler.RunYears(ctx, yp, 2))
		wg.Done()
	}()
	for _, t := range times {
		ts.tick(t)
		time.Sleep(time.Millisecond * 2)
	}
	wg.Wait()
	if err := errs.Err(); err != nil {
		t.Fatal(err)
	}

	logs := logRecorder.Logs(t)
	if err := containsError(logs); err != nil {
		t.Fatal(err)
	}

	for i, l := range logs {
		if got, want := l.Due, times[i]; !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	lines := deviceRecorder.Lines()
	if got, want := len(lines), 9; got != want {
		t.Errorf("got %d, want %d", got, want)
	}

}
