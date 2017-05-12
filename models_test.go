package bgTask

import (
	"testing"
)

func TestNewBg(t *testing.T) {
	bg := NewBg()
	if bg.errMsg != "PANICED" {
		t.Fail()
	}
	if bg.heartBeatTasks == nil {
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
	if bg.heartBeatTasks == nil {
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
