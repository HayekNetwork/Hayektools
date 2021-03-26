package clockwork

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sort"
	"sync/atomic"
	"time"
)

// Time location, default set by the time.Local (*time.Location)
var loc = time.Local

// Change the time location
func ChangeLoc(newLocation *time.Location) {
	loc = newLocation
}

// ----------------------------------------------------------------------------
// Job
// ----------------------------------------------------------------------------

type Job struct {
	interval uint64 // pause interval * unit bettween runs
	jobFunc  string // the job jobFunc to run, func[jobFunc]
	unit     string // time units, ,e.g. 'minutes', 'hours'...
	atTime   string // optional time at which this job runs

	lastRun  time.Time     // datetime of last run
	nextRun  time.Time     // datetime of next run
	period   time.Duration // cache the period between last an next run
	startDay time.Weekday  // Specific day of the week to start on

	funcs   map[string]interface{}
	fparams map[string]([]interface{})
	running uint32 // atomic
}

// Create a new job with the time interval.
func NewJob(interval uint64) *Job {
	return &Job{
		interval: interval,
		lastRun:  time.Unix(0, 0),
		nextRun:  time.Unix(0, 0),
		startDay: time.Sunday,
		funcs:    make(map[string]interface{}),
		fparams:  make(map[string]([]interface{})),
	}
}

// True if the job should be run now
func (j *Job) shouldRun() bool {
	return !j.isRunning() && time.Now().After(j.nextRun)
}

func (j *Job) isRunning() bool {
	return atomic.LoadUint32(&j.running) == 1
}

// Run the job and immdiately reschedulei it
func (j *Job) run() (result []reflect.Value, err error) {
	if !atomic.CompareAndSwapUint32(&j.running, 0, 1) {
		time.Sleep(time.Millisecond)
		return
	}
	defer atomic.StoreUint32(&j.running, 0)

	defer func() {
		if r := recover(); r != nil {
			var buf bytes.Buffer
			fmt.Fprintln(&buf, "colock.Job.run: recoverd panic:", r)
			for i := 0; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				fmt.Fprintln(&buf, "\t", file, line)
			}
			log.Println(buf.String())
			err = fmt.Errorf("%v", r)
		}
	}()

	f := reflect.ValueOf(j.funcs[j.jobFunc])
	params := j.fparams[j.jobFunc]
	if len(params) != f.Type().NumIn() {
		err = errors.New("clockwork: The number of param is not adapted.")
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}

	result = f.Call(in)
	j.lastRun = time.Now()
	j.scheduleNextRun()
	return
}

// Specifies the jobFunc that should be called every time the job runs
//
func (j *Job) Do(jobFun interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		panic("clockwork: only function can be schedule into the job queue.")
	}

	fname := getFuncName(jobFun)
	j.funcs[fname] = jobFun
	j.fparams[fname] = params
	j.jobFunc = fname

	// schedule the next run
	j.scheduleNextRun()
}

//	s.Every(1).Day().At("10:30").Do(task)
//	s.Every(1).Monday().At("10:30").Do(task)
//
func (j *Job) At(t string) *Job {
	hour := int((t[0]-'0')*10 + (t[1] - '0'))
	min := int((t[3]-'0')*10 + (t[4] - '0'))
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		panic("clockwork: time format error.")
	}

	// time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mock := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, 0, 0, loc)

	if j.unit == "days" {
		if time.Now().After(mock) {
			j.lastRun = mock
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc)
		}
	} else if j.unit == "weeks" {
		if time.Now().After(mock) {
			i := mock.Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), hour, min, 0, 0, loc)
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, hour, min, 0, 0, loc)
		}
	}
	return j
}

//Compute the instant when this job should run next
func (j *Job) scheduleNextRun() {
	if j.lastRun == time.Unix(0, 0) {
		if j.unit == "weeks" {
			i := time.Now().Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), 0, 0, 0, 0, loc)

		} else {
			j.lastRun = time.Now()
		}
	}

	if j.period != 0 {
		// translate all the units to the Seconds
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	} else {
		switch j.unit {
		case "minutes":
			j.period = time.Duration(j.interval * 60)
			break
		case "hours":
			j.period = time.Duration(j.interval * 60 * 60)
			break
		case "days":
			j.period = time.Duration(j.interval * 60 * 60 * 24)
			break
		case "weeks":
			j.period = time.Duration(j.interval * 60 * 60 * 24 * 7)
			break
		case "seconds":
			j.period = time.Duration(j.interval)
		}
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	}
}

// the follow functions set the job's unit with seconds,minutes,hours...

// Set the unit with second
func (j *Job) Second() (job *Job) {
	j.interval = 1
	job = j.Seconds()
	return
}

// Set the unit with seconds
func (j *Job) Seconds() (job *Job) {
	j.unit = "seconds"
	return j
}

// Set the unit  with minute, which interval is 1
func (j *Job) Minute() (job *Job) {
	j.interval = 1
	job = j.Minutes()
	return
}

//set the unit with minute
func (j *Job) Minutes() (job *Job) {
	j.unit = "minutes"
	return j
}

//set the unit with hour, which interval is 1
func (j *Job) Hour() (job *Job) {
	j.interval = 1
	job = j.Hours()
	return
}

// Set the unit with hours
func (j *Job) Hours() (job *Job) {
	j.unit = "hours"
	return j
}

// Set the job's unit with day, which interval is 1
func (j *Job) Day() (job *Job) {
	j.interval = 1
	job = j.Days()
	return
}

// Set the job's unit with days
func (j *Job) Days() *Job {
	j.unit = "days"
	return j
}

// s.Every(1).Monday().Do(task)
// Set the start day with Monday
func (j *Job) Monday() (job *Job) {
	j.interval = 1
	j.startDay = 1
	job = j.Weeks()
	return
}

// Set the start day with Tuesday
func (j *Job) Tuesday() (job *Job) {
	j.interval = 1
	j.startDay = 2
	job = j.Weeks()
	return
}

// Set the start day woth Wednesday
func (j *Job) Wednesday() (job *Job) {
	j.interval = 1
	j.startDay = 3
	job = j.Weeks()
	return
}

// Set the start day with thursday
func (j *Job) Thursday() (job *Job) {
	j.interval = 1
	j.startDay = 4
	job = j.Weeks()
	return
}

// Set the start day with friday
func (j *Job) Friday() (job *Job) {
	j.interval = 1
	j.startDay = 5
	job = j.Weeks()
	return
}

// Set the start day with saturday
func (j *Job) Saturday() (job *Job) {
	j.interval = 1
	j.startDay = 6
	job = j.Weeks()
	return
}

// Set the start day with sunday
func (j *Job) Sunday() (job *Job) {
	j.interval = 1
	j.startDay = 0
	job = j.Weeks()
	return
}

// Set the units as weeks
func (j *Job) Weeks() *Job {
	j.unit = "weeks"
	return j
}

// ----------------------------------------------------------------------------
// Scheduler
// ----------------------------------------------------------------------------

// Class Scheduler, the only data member is the list of jobs.
type Scheduler struct {
	jobs []*Job
}

// Create a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Datetime when the next job should run.
func (s *Scheduler) NextRun() (*Job, time.Time) {
	if len(s.jobs) == 0 {
		return nil, time.Now()
	}
	sort.Slice(s.jobs, func(i, j int) bool {
		return s.jobs[j].nextRun.After(s.jobs[i].nextRun)
	})
	return s.jobs[0], s.jobs[0].nextRun
}

// Schedule a new periodic job
func (s *Scheduler) Every(interval uint64) *Job {
	job := NewJob(interval)
	s.jobs = append(s.jobs, job)
	return job
}

// Run all the jobs that are scheduled to run.
func (s *Scheduler) RunPending() {
	sort.Slice(s.jobs, func(i, j int) bool {
		return s.jobs[j].nextRun.After(s.jobs[i].nextRun)
	})

	for _, job := range s.jobs {
		if job.shouldRun() {
			go job.run()
		}
	}
}

// Run all jobs regardless if they are scheduled to run or not
func (s *Scheduler) RunAll() {
	for _, job := range s.jobs {
		go job.run()
	}
}

// Remove specific job j
func (s *Scheduler) Remove(fn interface{}) {
	fname := getFuncName(fn)
	for i := 0; i < len(s.jobs); i++ {
		if s.jobs[i].jobFunc == fname {
			s.jobs[i] = s.jobs[len(s.jobs)-1]
			s.jobs = s.jobs[:len(s.jobs)-1]
			i--
		}
	}
}

// Delete all scheduled jobs
func (s *Scheduler) Clear() {
	s.jobs = s.jobs[:0]
}

// Start all the pending jobs
// Add seconds ticker
func (s *Scheduler) Start() chan bool {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunPending()
			case <-stopped:
				return
			}
		}
	}()

	return stopped
}

// ----------------------------------------------------------------------------
// Default Scheduler
// ----------------------------------------------------------------------------

// The following methods are shortcuts for not having to
// create a Schduler instance

var defaultScheduler = NewScheduler()

// Schedule a new periodic job
func Every(interval uint64) *Job {
	return defaultScheduler.Every(interval)
}

// Run all jobs that are scheduled to run
//
// Please note that it is *intended behavior that run_pending()
// does not run missed jobs*. For example, if you've registered a job
// that should run every minute and you only call run_pending()
// in one hour increments then your job won't be run 60 times in
// between but only once.
func RunPending() {
	defaultScheduler.RunPending()
}

// Run all jobs regardless if they are scheduled to run or not.
func RunAll() {
	defaultScheduler.RunAll()
}

// Run all jobs that are scheduled to run
func Start() chan bool {
	return defaultScheduler.Start()
}

// Clear
func Clear() {
	defaultScheduler.Clear()
}

// Remove
func Remove(j interface{}) {
	defaultScheduler.Remove(j)
}

// NextRun gets the next running time
func NextRun() (job *Job, time time.Time) {
	return defaultScheduler.NextRun()
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

// for given function fn , get the name of funciton.
func getFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf((fn)).Pointer()).Name()
}

// ----------------------------------------------------------------------------
// END
// ----------------------------------------------------------------------------
