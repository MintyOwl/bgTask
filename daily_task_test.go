package bgTask

import (
	"syscall"
	"testing"
	"time"
)

func TestRegisterDailyTask(t *testing.T) {
	bg := NewBg()
	defer bg.Wait()
	bg.RegisterDailyTask("unik4", "14:46", func() { p("EVERYDAY TASK 2:46PM") })
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
