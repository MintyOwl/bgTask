package bgTask

import (
	"bytes"
	"syscall"
	"testing"
	"time"
)

func TestSetLogger(t *testing.T) {
	bg := NewBg()
	bg.SetLogger(func(val string) error { return nil })
	if bg.log == nil {
		t.Fail()
	}
}

func TestRegisterTask(t *testing.T) {
	bg := NewBg()

	var tasks []*Task
	task1 := &Task{Key: "unik1", Duration: "1s", TaskFn: func() {}}
	tasks = append(tasks, task1)
	bg.RegisterTasks(tasks)
	bg.Start()
	if bg.heartBeatTasks["unik1"] == nil {
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
	task1 := &Task{Key: "unik1", Duration: "1s", TaskFn: func() { myHandler() }}
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
	logOutput.buf.Reset()
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
