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

var logOutput logger

type logger struct {
	buf *bytes.Buffer
}

func loggerSetup() {
	buffer := bytes.NewBufferString("")
	logOutput = logger{buf: buffer}
}
func myLogger(val string) error {
	logOutput.buf.WriteString(val)
	return nil
}
func TestRegisterDailyTask(t *testing.T) {
	loggerSetup()
	bg := NewBg().SetLogger(myLogger).SetDevel()
	defer bg.Wait()
	var dailyTasks []*Task
	task1 := &Task{Key: "unik4", RelativeTime: "04:46", TaskFn: func() { p("This is unik4 task being run as SetDevel is true") }}
	dailyTasks = append(dailyTasks, task1)
	bg.RegisterDailyTasks(dailyTasks)
	bg.Start()
	<-time.After(50 * time.Millisecond)
	if !strings.Contains(logOutput.buf.String(), "will start after") {
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
}

func TestPersistence(t *testing.T) {
	bg := NewBg()
	p("Persistence Test")
	bg.Persistence("testStore")

	dir := filepath.Join("testStore", "bgTasks")
	fullPath := filepath.Join(dir, storeFile)
	_, err := os.Open(fullPath)
	if err != nil {
		t.Fail()
	}

	var dailyTask2 = func() {
		p("RUNNING TASK unikey5")
	}
	var bgDailyTask = &Task{Key: "unikey5", RelativeTime: "19:21", TaskFn: dailyTask2}
	var dailyTasks []*Task
	dailyTasks = append(dailyTasks, bgDailyTask)
	bg.RegisterDailyTasks(dailyTasks)
	if bg.GetDailyTaskByKey("unikey5") == nil {
		t.Fail()
	}

}
