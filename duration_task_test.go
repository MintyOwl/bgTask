package bgTask

import (
	"bytes"
	"syscall"
	"testing"
	"time"
)

func TestNewBg(t *testing.T) {
	bg := NewBg()
	if bg.errMsg != "PANICED" {
		t.Fail()
	}
	if bg.durationTasks == nil {
		t.Fail()
	}
	if bg.wg != nil {
		t.Fail()
	}
	if bg.done == nil {
		t.Fail()
	}
	if bg.log != nil {
		t.Fail()
	}

}

func TestNewBgSync(t *testing.T) {
	bg := NewBgSync()
	if bg.errMsg != "PANICED" {
		t.Fail()
	}
	if bg.durationTasks == nil {
		t.Fail()
	}
	if bg.wg == nil {
		t.Fail()
	}
	if bg.done != nil {
		t.Fail()
	}
	if bg.log != nil {
		t.Fail()
	}

}

func TestSetLogger(t *testing.T) {
	bg := NewBg()
	bg.SetLogger(func(val string) error { return nil })
	if bg.log == nil {
		t.Fail()
	}
}

func TestRegisterTask(t *testing.T) {
	bg := NewBg()
	bg.RegisterTask("unik1", "1s", func() {})
	if bg.durationTasks["unik1"] == nil {
		t.Fail()
	}
}

type data struct {
	buf *bytes.Buffer
}

var buffer *bytes.Buffer

func myHandler() {
	buffer = bytes.NewBufferString("")
	res := data{buf: buffer}
	res.buf.WriteString("TASK RAN")
	p(res.buf.String())

}

func TestStart(t *testing.T) {
	bg := NewBg()
	bg.RegisterTask("unik1", "1s", func() { myHandler() })

	go func() {
		select {
		case <-time.After(2 * time.Second):
			//debugExit <- struct{}{}
			//done2 <- struct{}{}
			go func() { signals <- syscall.SIGINT }()
		}
	}()
	bg.Wait()
	p(buffer.String())
	if buffer.String() != "TASK RAN" {
		t.Fail()
	}
}
