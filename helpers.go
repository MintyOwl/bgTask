package bgTask

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// StringifyErrors will stringify all bg.Errors delimited by new line
func (bg *Bg) StringifyErrors() string {
	var allErrs string
	for _, err := range bg.Errors {
		allErrs += spln(err)
	}
	return allErrs
}

func catchPanic(bg *Bg, args ...string) {
	if err := recover(); err != nil {
		bgPanicMsg := spf("\nbgTask has panicked\n")
		var e string
		if len(args) > 0 {
			e = args[0]
		}
		otherErrs := bg.StringifyErrors()
		panicErr := spf("PANIC ERROR > %v ", err)
		bgPanicMsg += e + panicErr + otherErrs
		if bg.log != nil {
			bg.log.Err(bgPanicMsg)
		} else {
			p(bgPanicMsg)
		}
	}
}

// flushPTask flushed current pending tasks to storage
func (bg *Bg) flushPTask() {
	bg.mu.Lock()
	defer bg.mu.Unlock()
	b, err := json.Marshal(bg.deDup())
	if err != nil {
		p(err)
	}

	ioutil.WriteFile(bg.storage, b, 0644)
}

// thisIsInPast is used to see whether correctedDate is a time in past from now
func (bg *Bg) thisIsInPast(correctedDate string) bool {
	corrected, err := time.ParseInLocation(bgTaskStdTimeDay, correctedDate, bg.location)
	if err != nil {
		// there shouldn't be any errors here since this was pulled from the storage, but just in case
		bg.Errors = append(bg.Errors, err)
		return false
	}
	if bg.now().Sub(corrected) > 0 {
		return true
	}
	return false
}

func (bg *Bg) handleDisplay(val string, infoType bool) {
	if bg.log != nil && infoType {
		bg.log.Info(val)
	} else if infoType == false {
		bg.log.Err(val)
		return
	} else {
		pt(val)
	}
}

// if the 'in' in the format "15:04" time has already passed, update it the value with tomorrow's timestamp.
func (bg *Bg) getCorrectDate(in string) string {
	today := bg.now().Format(bgTaskStdDay)
	today = in + " " + today
	tIn, err := time.ParseInLocation(bgTaskStdTimeDay, today, bg.location)
	if err != nil {
		bg.Errors = append(bg.Errors, err)
	}
	now := bg.now()
	if tIn.Sub(now) > 0 {
		today = bg.now().Format(bgTaskStdDay)
		correctDate := in + " " + today
		return correctDate
	}
	tom := now.Add(24 * time.Hour)
	tomS := tom.Format(bgTaskStdDay)
	correctDate := in + " " + tomS
	return correctDate
}

func (bg *Bg) now() time.Time {
	return time.Now().In(bg.location)
}

func taskExistInSlice(key, correctedTime string, pTasks []pendingTask) bool {
	for _, v := range pTasks {
		if v.Key == key && v.CorrectedTime == correctedTime {
			return true
		}
	}
	return false
}

func (bg *Bg) deDup() []pendingTask {
	var newpTasks []pendingTask
	for _, pTask := range bg.pTasks {
		if !taskExistInSlice(pTask.Key, pTask.CorrectedTime, newpTasks) {
			newpTasks = append(newpTasks, pTask)
		}
	}
	return newpTasks
}

// GetStorage gives you storage location for pending tasks
func (bg *Bg) GetStorage() string {
	return bg.storage
}
