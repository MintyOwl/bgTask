package bgTask

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	bgTaskStdTimeDay = "15:04 2006-01-02"
	bgTaskStdDay     = "2006-01-02"
	storeFile        = "bgTasks_.json"
)

func getCorrectDate(in string, loc *time.Location) (string, error) {
	today := time.Now().In(loc).Format(bgTaskStdDay)
	today = in + " " + today
	tIn, err := time.ParseInLocation(bgTaskStdTimeDay, today, loc)
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	if tIn.Sub(now) > 0 {
		today = time.Now().Format(bgTaskStdDay)
		correctDate := in + " " + today
		return correctDate, nil
	}
	tom := now.Add(24 * time.Hour)
	tomS := tom.Format(bgTaskStdDay)
	correctDate := in + " " + tomS
	return correctDate, nil
}

// SetDevel must not be called on production
func (bg *Bg) SetDevel() *Bg {
	bg.devel = true
	return bg
}

// Persistence accepts directory where the scheduler will put all pending daily tasks in a file named bgTask_.json
// By default the scheduler will use operating system specific temp directory if possible
// Persistence is only meant to be used with daily tasks
func (bg *Bg) Persistence(directory ...string) *Bg {
	var dir = ""
	if len(directory) > 0 {
		dir = directory[0]
	}
	var absDir, absDirWithFile string
	var subDir = "bgTasks"
	var err error
	dir = filepath.Join(dir, subDir)
	absDir, err = filepath.Abs(dir)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	err = os.MkdirAll(absDir, 0644)
	absDirWithFile = filepath.Join(absDir, storeFile)
	f, err := os.OpenFile(absDirWithFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
		f.Close()
	}
	defer f.Close()
	bg.storage = absDirWithFile
	return bg
}

// GetDailyTaskByKey will get you daily Task by the unique key provided during registration
func (bg *Bg) GetDailyTaskByKey(key string) *Task {
	task, ok := bg.dailyTasks[key]
	if !ok {
		return nil
	}
	return task
}

// registerDailyTask is used for tasks that are run only once a day.
func (bg *Bg) registerDailyTask(task *Task) {
	key := task.Key
	var err error
	for _, pTask := range bg.pTasks {
		if pTask.Key == task.Key {
			if bg.thisIsInPast(pTask.CorrectedTime) {
				newTask := task
				newTask.Key = key + "_repeated_" + strconv.Itoa(int(time.Now().Unix()))
				go bg.startDailyTask(newTask.Key, 500*time.Millisecond, newTask)
				<-time.After(1 * time.Second)
				bg.removeTaskByDate(pTask.CorrectedTime)
			}

		}
	}

	correctDate, err := getCorrectDate(task.RelativeTime, bg.location)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	t1, err := time.ParseInLocation(bgTaskStdTimeDay, correctDate, bg.location)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}

	bg.pTasks = append(bg.pTasks, pendingTask{Key: key, CorrectedTime: correctDate})
	task.dailyCancel = make(chan bool)
	bg.dailyTasks[key] = task
	if bg.devel {
		task.dur = 50 * time.Millisecond
	} else {
		task.dur, err = time.ParseDuration(spf("%v", t1.Sub(time.Now())))
	}

	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
}

// RegisterDailyTasks is used for tasks that are run only once a day.
func (bg *Bg) RegisterDailyTasks(tasks []*Task) {
	var pendingTasks []pendingTask
	if bg.storage != "" {
		bg.mu.RLock()
		b, _ := ioutil.ReadFile(bg.storage)
		bg.mu.RUnlock()
		json.Unmarshal(b, &pendingTasks)
	}
	if len(pendingTasks) > 0 {
		bg.pTasks = pendingTasks
	}
	for _, task := range tasks {
		bg.registerDailyTask(task)
	}
	bg.flushPTask()
}

// CancelDailyTask will remove task by the 'key' along with the key. To add again use RegisterTask
func (bg *Bg) CancelDailyTask(key string) {
	defer catchPanic(bg)
	bg.dailyTasks[key].dailyCancel <- true

}

func (bg *Bg) removeTaskByDate(correctedDate string) {
	var newpTasks = bg.pTasks
	for _, pTask := range newpTasks {
		bg.pTasks = make([]pendingTask, 0)
		if pTask.CorrectedTime != correctedDate && !bg.thisIsInPast(correctedDate) {
			bg.pTasks = append(bg.pTasks, pTask)
		}
	}
	bg.flushPTask()
}

// removeTask must remove the task by the 'key' from storage. Personally, I think this should be synchronous operation. Hence everytime, we remove a single task, its going to block the storage writer
func (bg *Bg) removeTask(key string) error {
	var newpTasks = bg.pTasks
	bg.pTasks = make([]pendingTask, 0)
	for _, pTask := range newpTasks {
		if pTask.Key != key {
			bg.pTasks = append(bg.pTasks, pTask)
		}
	}
	bg.flushPTask()
	return nil
}

func (bg *Bg) startDailyTask(key string, dur time.Duration, taskToBeRunImmediately ...*Task) {
	var task *Task
	if len(taskToBeRunImmediately) > 0 {
		task = taskToBeRunImmediately[0]
	} else {
		task = bg.dailyTasks[key]
	}

	bg.handleDisplay(spf("Task for %v will start after %v\n", key, dur), true)
	go func() {
		for {
			select {
			case <-task.dailyCancel:
				if bg.storage != "" {
					bg.removeTask(key)
				}

				delete(bg.dailyTasks, key)
				return
			case <-bg.signals:
				if bg.wg == nil {
					bg.done <- struct{}{}
					return
				}
				bg.wg.Done()
			case <-time.After(task.dur):
				if bg.devel {
					task.dur = 5 * time.Second
					p("startDailyTask 5 Sec instead of 24Hours")
				} else {
					task.dur = 24 * time.Hour

				}

				go func() {
					defer catchPanic(bg, spf("%v %v\n", key, bg.errMsg))
					err := task.TaskFn()
					if err != nil {
						bg.handleDisplay(err.Error(), false)
					}
					if bg.storage != "" && len(taskToBeRunImmediately) == 0 {
						bg.removeTask(key)
					}

				}()
			}
		}
	}()
}
