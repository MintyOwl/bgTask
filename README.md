# bgTask
Simple and persistent task scheduler

# Usage

```
	bg := bgTask.NewBg()
	// block the goroutine for ever
	defer  bg.Wait()

	// prepare and register heartbeat tasks
	var allBgTasks []*bgTask.Task
	allBgTasks = append(allBgTasks, sometask, othertask)
	bg.RegisterTasks(allBgTasks)

	// prepare and register daily tasks
	var dailyTasks []*bgTask.Task
	dailyTasks = append(dailyTasks, somedailytask, someotherdailytask)
	bg.RegisterDailyTasks(dailyTasks)

	// After successful registrations call Start()
	bg.Start()
```

# Full Example
```
package main

package main

import (
	"bytes"
	"fmt"
	golog "log"
	"time"

	"github.com/MintyOwl/bgTask"
)

var p = fmt.Println
var buf bytes.Buffer

func init() {
	logger = golog.New(&buf, "bgTask_", golog.Lshortfile)
}

var logger *golog.Logger

// loggers akways return nil
func goLogger(val string) error {
	logger.Print(val)
	return nil
}

func Ping() {
	p("Pinging every sec")
}

var task1 = func() { Ping() }
var taskIsPanicking = func() { panic(""); p("DOING SOMETHING USEFUL EVERY 3 SECS ") }
var task3 = func() { p("DOING SOMETHING ELSE USEFUL AS WELL EVERY 3 SECS ") }

var bgTask1 = &bgTask.Task{Key: "unikey1", Duration: "1s", TaskFn: task1}
var bgTask2 = &bgTask.Task{Key: "unikey2", Duration: "3s", TaskFn: taskIsPanicking}
var bgTask3 = &bgTask.Task{Key: "unikey3", Duration: "3s", TaskFn: task3}

var dailyTask1 = func() { p("RUNNING unik4 TASK AT 1 16 PM EVERYDAY") }
var bgDailyTask1 = &bgTask.Task{Key: "unikey4", RelativeTime: "13:16", TaskFn: dailyTask1}

var dailyTask2 = func() {
	p("RUNNING unik5 TASK AT 7 46  PM EVERYDAY")
}
var bgDailyTask2 = &bgTask.Task{Key: "unikey5", RelativeTime: "19:46", TaskFn: dailyTask2}

func startScheduler() {
	bg := bgTask.NewBg().SetLocation(time.FixedZone("GMT", 0))
	defer func() { p(buf.String()); bg.Wait() }()
	bg.SetLogger(goLogger).SetErrMsg(" HAS PANICED")

	var allBgTasks []*bgTask.Task
	allBgTasks = append(allBgTasks, bgTask1, bgTask3)
	bg.RegisterTasks(allBgTasks)

	var dailyTasks []*bgTask.Task
	dailyTasks = append(dailyTasks, bgDailyTask1, bgDailyTask2)
	bg.Persistence() // pending daily tasks are persisted, so that when server crashes before executing the appropriate task, it will rerun again during server restart
	bg.RegisterDailyTasks(dailyTasks)
	p(bg.Errors) // check for errors if any. Proceed ahead only when there are none

	// After registering both heartbeats and daily tasks, call Start
	bg.Start()

	go func() {
		select {
		case <-time.After(10 * time.Millisecond):
			p("Get Task By Key")
			p(bg.GetDailyTaskByKey("unikey5"))

		}
	}()
	go func() {
		select {
		case <-time.After(100 * time.Millisecond):
			bg.CancelDailyTask("unikey5")
			//bg.CancelTask("unikey3")
			<-time.After(100 * time.Millisecond) // wait until go's internal scheduler finishes
			p(bg.GetDailyTaskByKey("unikey5"))
		}
	}()
}

func main() {
	startScheduler()
}



```