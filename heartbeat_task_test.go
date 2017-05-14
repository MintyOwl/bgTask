package bgTask

import (
	"bytes"
	"syscall"
	"testing"
	"time"
)

func TestSetLogger(t *testing.T) {
	logger := &Logger{
		Info: func(val string) error { return nil },
		Err:  func(val string) error { return nil },
	}
	bg := NewBg()
	bg.SetLogger(logger)
	if bg.log == nil {
		t.Fail()
	}
}

func TestRegisterBadTask(t *testing.T) {
	bg := NewBg()

	var tasks []*Task
	task1 := &Task{Key: "unik1", Duration: "1", TaskFn: func() error { return nil }}
	tasks = append(tasks, task1)
	bg.RegisterTasks(tasks)
	if bg.Start() == nil {
		t.Fail()
	}

}
func TestRegisterTask(t *testing.T) {
	bg := NewBg()

	var tasks []*Task
	task1 := &Task{Key: "unik1", Duration: "1s", TaskFn: func() error { return nil }}
	tasks = append(tasks, task1)
	bg.RegisterTasks(tasks)
	bg.Start()
	if bg.GetTaskByKey("unik1") == nil {
		t.Fail()
	}
	bg.CancelTask("unik1")
	if _, ok := bg.heartBeatTasks["unik1"]; ok {
		t.Fail()
	}
}

type data struct {
	buf *bytes.Buffer
}

var output data

func myHandler() {
	buffer := bytes.NewBufferString("")
	output = data{buf: buffer}
	output.buf.WriteString("TASK RAN")
}

func TestBg(t *testing.T) {
	bg := NewBg()
	setup(bg, t)
}

func setup(bg *Bg, t *testing.T) {
	var tasks []*Task
	task1 := &Task{Key: "unik1", Duration: "1s", TaskFn: func() error { myHandler(); return nil }}
	tasks = append(tasks, task1)
	bg.RegisterTasks(tasks)

	bg.Start()
	go func() {
		select {
		case <-time.After(1100 * time.Millisecond):
			go func() { bg.signals <- syscall.SIGINT }()
		}
	}()
	if bg.wg != nil {
		bg.SyncWait()
	} else {
		bg.Wait()
	}
	if output.buf.String() != "TASK RAN" {
		t.Fail()
	}

	output.buf.Reset()
	if bg.wg == nil {
		p("Heartbeat Test")
	} else {
		p("Heartbeat with Sync Test")
	}

}

func TestBgSync(t *testing.T) {
	bg := NewBgSync()
	setup(bg, t)

}
