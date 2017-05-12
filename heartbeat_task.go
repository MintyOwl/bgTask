package bgTask

import (
	"errors"
	"fmt"
	"time"
)

var p = fmt.Println
var pf = fmt.Printf
var spf = fmt.Sprintf
var pt = fmt.Print
var spt = fmt.Sprint
var spln = fmt.Sprintln

// SetLocation allows to set context to the scheduler with a particular time zone.
// This way the tasks running on remote machine can still sync with preferred time zone
func (bg *Bg) SetLocation(loc *time.Location) *Bg {
	bg.location = loc
	return bg
}

// SetLogger allows client to add optional logging facility in case the Task handler panics. It will call client's logginf function during panics instead of just printing on terminal
// Eg: bg.SetLogger(myLogger)
// func myLogger(val string) error {fmt.Printf("\n LOGGING > %v \n", val); return nil}
// myLogger should always return nil
func (bg *Bg) SetLogger(logfn func(string) error) *Bg {
	bg.log = logfn
	return bg
}

// SetErrMsg sets a custom error message to be used when the Task handler, provided by the client, panics
// Eg: bg.SetErrMsg("HAS PANICED")
func (bg *Bg) SetErrMsg(msg string) *Bg {
	bg.errMsg = msg
	return bg
}

// registerTask adds Task handler, that will be called for 'duration' provided.
// Eg: bg.registerTask("unique_key123", "5s", somefunc). Here, somefunc is reference to client's Task handler function that will get called every 5 seconds and is registered with a unique identifier KEY as the first argument.
func (bg *Bg) registerTask(task *Task) {
	key := task.Key
	_, err := time.ParseDuration(task.Duration)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	task.hbCancel = make(chan bool)
	bg.heartBeatTasks[key] = task
}

// RegisterTasks adds multiple Tasks all at once.
// Eg: TODO
func (bg *Bg) RegisterTasks(tasks []*Task) {
	for _, task := range tasks {
		bg.registerTask(task)
	}

}

// Start must be called after Registering all tasks. It returns errors, delimited by new line, if any
func (bg *Bg) Start() error {
	if len(bg.Errors) > 0 {
		var allErr string
		for _, v := range bg.Errors {
			allErr += spt(v) + "\n"
		}
		return errors.New(allErr)
	}

	if len(bg.heartBeatTasks) > 0 {
		for key := range bg.heartBeatTasks {
			if bg.wg == nil {
				go bg.startHeartBeatTasks(key)
			} else {
				bg.wg.Add(1)
				go bg.startHeartBeatTasks(key)
			}
		}
	}
	if len(bg.dailyTasks) > 0 {
		for key, task := range bg.dailyTasks {
			if bg.wg == nil {
				go bg.startDailyTask(key, task.dur)
			} else {
				bg.wg.Add(1)
				go bg.startDailyTask(key, task.dur)
			}
		}
	}
	return nil
}

// CancelTask will remove task by the 'key' along with the key. To add again please use registerTask api
func (bg *Bg) CancelTask(key string) {
	defer catchPanic(bg)
	bg.heartBeatTasks[key].hbCancel <- true
	delete(bg.heartBeatTasks, key)
}

func (bg *Bg) handleDisplay(val string) {
	if bg.log != nil {
		bg.log(val)
	} else {
		pt(val)
	}
}

func (bg *Bg) startHeartBeatTasks(key string) {
	task := bg.heartBeatTasks[key]
	dur, _ := time.ParseDuration(task.Duration)
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-task.hbCancel:
			return
		case <-bg.signals:
			if bg.wg == nil {
				bg.done <- struct{}{}
				return
			}
			bg.wg.Done()

		case <-ticker.C:
			go func() {
				defer func() {
					if err := recover(); err != nil {
						bg.handleDisplay(spf("%v %v\n", key, bg.errMsg))
					}
				}()

				task.TaskFn()
			}()
		}
	}
}
