package bgTask

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var p = fmt.Println
var pf = fmt.Printf
var spf = fmt.Sprintf

var signals = make(chan os.Signal)
var debugExit = make(chan struct{})

// Bg is a background scheduler to be configured
type Bg struct {
	durationTasks map[string]*job
	dailyTasks    map[string]*job
	location      *time.Location
	showTime      bool
	errMsg        string
	log           func(string) error
	done          chan struct{}
	wg            *sync.WaitGroup
	Errors        []error
}

type job struct {
	key, duration string
	fn            func()
}

// NewBg creates a new background scheduler, allowing to register new 'fire and forget' periodical jobs. Hence during host process termination scheduler doesnt wait for periodical jobs' goroutines to terminate ('fire and forget')
// Eg: bg := bgTask.NewBg(true) // true prints current time in terminal
// bg := bgTask.NewBg() // No argument in NewBg hence no printing of time in terminal
func NewBg(sT ...bool) *Bg {
	var show bool
	if len(sT) > 0 {
		show = true
	}
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	bg := &Bg{
		durationTasks: make(map[string]*job),
		showTime:      show,
		errMsg:        "PANICED",
		done:          make(chan struct{}),
		dailyTasks:    make(map[string]*job),
		location:      time.FixedZone("IST", 19800),
		Errors:        make([]error, 0),
	}
	return bg
}

// NewBgSync is like NewBg except that during host process termination, the scheduler will wait for periodical jobs' goroutines to be terminated
// Eg: bg := bgTask.NewBgSync(true) // true prints current time in terminal
// bg := bgTask.NewBgSync() // No argument in NewBgSync hence no printing of time in terminal
func NewBgSync(sT ...bool) *Bg {
	var show bool
	if len(sT) > 0 {
		show = true
	}
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	bg := &Bg{
		durationTasks: make(map[string]*job),
		showTime:      show,
		errMsg:        "PANICED",
		wg:            &sync.WaitGroup{},
		dailyTasks:    make(map[string]*job),
		location:      time.FixedZone("IST", 19800),
		Errors:        make([]error, 0),
	}
	return bg
}

func (bg *Bg) setDebugger() *Bg {
	return bg
}

// SetLocation allows to set context to the scheduler with a particular time zone.
// This way the tasks running on remote machine can still sync with preferred time zone
func (bg *Bg) SetLocation(loc *time.Location) *Bg {
	bg.location = loc
	return bg
}

// SetLogger allows client to add optional logging facility in case the job handler panics. It will call client's logginf function during panics instead of just printing on terminal
// Eg: bg.SetLogger(myLogger)
// func myLogger(val string) error {fmt.Printf("\n LOGGING > %v \n", val); return nil}
// myLogger should always return nil
func (bg *Bg) SetLogger(logfn func(string) error) *Bg {
	bg.log = logfn
	return bg
}

// SetErrMsg sets a custom error message to be used when the job handler, provided by the client, panics
// Eg: bg.SetErrMsg("HAS PANICED")
func (bg *Bg) SetErrMsg(msg string) *Bg {
	bg.errMsg = msg
	return bg
}

// RegisterTask adds job handler, that will be called for 'duration' provided.
// Eg: bg.RegisterTask("unique_key123", "5s", somefunc). Here, somefunc is reference to client's job handler function that will get called every 5 seconds and is registered with a unique identifier key as the first argument.
func (bg *Bg) RegisterTask(key, duration string, fn func()) {
	_, err := time.ParseDuration(duration)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	bg.durationTasks[key] = &job{fn: fn, duration: duration}
	if bg.wg == nil {
		go bg.startDurationTasks(key)
		return
	}
	bg.wg.Add(1)
	go bg.startDurationTasks(key)
}

// Wait should be used with NewBg() only. This method must to be called by the client at the end of registration or using defer
// Eg: bg := bgTask.NewBg(); defer bg.Wait()
func (bg *Bg) Wait() {
	<-bg.done
}

// SyncWait should be used with NewBgSync() only. This method must to be called by the client at the end of registration
// Eg: bg := bgTask.NewBg(); defer bg.SyncWait()
func (bg *Bg) SyncWait() {
	bg.wg.Wait()
}

func (bg *Bg) startDurationTasks(key string) {
	job := bg.durationTasks[key]
	dur, _ := time.ParseDuration(job.duration)
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-debugExit:
			if bg.wg == nil {
				bg.done <- struct{}{}
				return
			}
			bg.wg.Done()
		case <-signals:
			if bg.wg == nil {
				bg.done <- struct{}{}
				return
			}
			bg.wg.Done()

		case t := <-ticker.C:
			go func() {
				defer func() {
					if err := recover(); err != nil {
						if bg.log != nil {
							bg.log(spf("%v %v\n", key, bg.errMsg))
						} else {
							pf("%v %v\n", key, bg.errMsg)
						}
					}
				}()
				if bg.showTime {
					p(t)
				}
				job.fn()
			}()

		}
	}
}
