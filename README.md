# bgTask
Simple and idiomatic task scheduler

# Usage
```


import (
	"fmt"
	"github.com/MintyOwl/bgTask"
	// ....
)

var task2Panics = func() { panic(""); p("DOING SOMETHING USEFUL EVERY 3 SECS ") }
var task3 = func() { p("DOING SOMETHING ELSE USEFUL AS WELL EVERY 3 SECS ") }

func zapLogger(val string) error {
	var paths []string
	paths = append(paths, "./schedulerZap.json")
	loggr := logging.NewAppLogger("info", paths)
	loggr.Info("Scheduler Info Logs for pending tasks and panics", zap.String("task", val))
	return nil
}

func startScheduler() {
	bg := bgTask.NewBg()
	defer bg.Wait()
	bg.SetLogger(zapLogger).SetErrMsg(" HAS PANICED")
	job1 := func() { ApiCall() }
	jobIsPanicking := task2Panics
	job3 := task3
	bg.RegisterTask("unikey1", "1s", job1)
	bg.RegisterTask("unikey2", "3s", jobIsPanicking)
	bg.RegisterTask("unikey3", "3s", job3)
	bg.RegisterDailyTask("unikey4", "14:46", func() { p("RUNNING THIS TASK AT 2:46PM EVERYDAY") })
}



```
