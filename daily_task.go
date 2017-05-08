package bgTask

import (
	"os"
	"time"
)

const (
	bgTaskStdTimeDay = "15:04 2006-01-02"
	bgTaskStdDay     = "2006-01-02"
)

func getCorrectDate(in string, loc *time.Location) (string, error) {
	today := time.Now().Format(bgTaskStdDay)
	today = in + " " + today
	tIn, err := time.ParseInLocation(bgTaskStdTimeDay, today, loc)
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	if tIn.Sub(now) > 0 {
		today = time.Now().Format(bgTaskStdDay)
		correctDate := in + " " + today
		// write correctDate to file for persistense
		return correctDate, nil
	}
	tom := now.Add(24 * time.Hour)
	tomS := tom.Format(bgTaskStdDay)
	correctDate := in + " " + tomS
	return correctDate, nil
}

// RegisterDailyTask is used for tasks that are run only once a day.
// It accepts the unique key as the first param, a time in the format of "13:45", meaning
// at 1:45 PM everyday, 'fn' callback function will be called
// bg.RegisterDailyTask("uniqueKey123", "13:45", func() { fmt.Println("CALL BACK at 1:45PM everyday") })
func (bg *Bg) RegisterDailyTask(key, relativeTime string, fn func()) {
	CorrectDate, err := getCorrectDate(relativeTime, bg.location)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	t1, err := time.ParseInLocation(bgTaskStdTimeDay, CorrectDate, bg.location)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	bg.dailyTasks[key] = &job{fn: fn}
	dur, err := time.ParseDuration(spf("%v", t1.Sub(time.Now())))
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	if len(bg.Errors) > 0 {
		p(bg.Errors)
		os.Exit(2)
	}
	if bg.wg == nil {
		go bg.startDailyTask(key, dur)
		return
	}
	bg.wg.Add(1)
	go bg.startDailyTask(key, dur)
}

func (bg *Bg) startDailyTask(key string, dur time.Duration) {
	dur1 := dur
	job := bg.dailyTasks[key]
	if bg.log != nil {
		bg.log(spf("Task for %v will start after %v\n", key, dur))
	} else {
		pf("\nTask for %v will start after %v\n", key, dur)
	}

	go func() {
		for {
			select {
			case <-bg.signals:
				if bg.wg == nil {
					bg.done <- struct{}{}
					return
				}
				bg.wg.Done()
			case t := <-time.After(dur1):
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

					dur1 = 24 * time.Hour
					job.fn()
					// remove from persistence
				}()
			}
		}

	}()
	//<-done

}
