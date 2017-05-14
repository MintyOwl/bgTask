package bgTask

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var logOutput bugLogger

type bugLogger struct {
	info *bytes.Buffer
}

func loggerSetup() {
	buffer := bytes.NewBufferString("")
	logOutput = bugLogger{info: buffer}
}
func myLogger(val string) error {
	logOutput.info.WriteString(val)
	return nil
}

func TestRegisterDailyTask(t *testing.T) {
	loggerSetup()
	loggr := &Logger{
		Info: myLogger,
	}
	bg := NewBg().SetLogger(loggr).SetDevel()
	defer bg.Wait()
	var dailyTasks []*Task
	task1 := &Task{Key: "unik4", RelativeTime: "04:46", TaskFn: func() error { p("This is unik4 task being run as SetDevel is true"); return nil }}
	dailyTasks = append(dailyTasks, task1)
	bg.RegisterDailyTasks(dailyTasks)
	bg.Start()
	<-time.After(50 * time.Millisecond)
	if !strings.Contains(logOutput.info.String(), "will start after") {
		t.Fail()
	}
	if bg.dailyTasks["unik4"] == nil {
		t.Fail()
	}

	bg.signals <- syscall.SIGINT
	if len(bg.Errors) > 0 {
		t.Fail()
	}

	if bg.location.String() != "IST" {
		t.Fail()
	}
	bg.SetLocation(time.FixedZone("GMT", 0))
	if bg.location.String() != "GMT" {
		t.Fail()
	}
	loggr = &Logger{
		Info: myLogger,
	}
	bg = NewBg().SetLogger(loggr)
	task44 := &Task{Key: "unik44", RelativeTime: "26:12", TaskFn: func() error { p("This is unik44 task being run"); return nil }}
	dailyTasks = append(dailyTasks, task44)
	bg.RegisterDailyTasks(dailyTasks)

	if bg.Start() == nil {
		t.Fail()
	}
}

func TestPersistence(t *testing.T) {
	bg := NewBg()
	p("Persistence Test")
	bg.Persistence("testStore")

	dir := filepath.Join("testStore", "bgTasks")
	fullPath := filepath.Join(dir, storeFile)
	f, err := os.Open(fullPath)
	if err != nil {
		f.Close()
		t.Fail()
	}

	var dailyTask2 = func() error {
		p("RUNNING TASK unikey5")
		return nil
	}
	var bgDailyTask = &Task{Key: "unikey5", RelativeTime: "19:21", TaskFn: dailyTask2}
	var dailyTasks []*Task
	dailyTasks = append(dailyTasks, bgDailyTask)
	bg.RegisterDailyTasks(dailyTasks)
	if bg.GetDailyTaskByKey("unikey5") == nil {
		t.Fail()
	}

}
