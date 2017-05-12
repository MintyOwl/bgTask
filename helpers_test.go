package bgTask

import (
	"strings"
	"syscall"
	"testing"
	"time"
)

const bgStaticTime = "15:04"

func TestThisIsInPast(t *testing.T) {
	yesterday := time.Now().In(time.FixedZone("GMT", 0)).Add(-24 * time.Hour)
	dateToBeTested := yesterday.Format(bgTaskStdTimeDay)
	bg := NewBg()
	if bg.thisIsInPast(dateToBeTested) != true {
		t.Fail()
	}
	tomorrow := time.Now().In(time.FixedZone("GMT", 0)).Add(24 * time.Hour)
	dateToBeTested = tomorrow.Format(bgTaskStdTimeDay)
	if bg.thisIsInPast(dateToBeTested) != false {
		t.Fail()
	}
}

func TestCorrectDate(t *testing.T) {
	now := time.Now()
	//nowInFormat := now.Format(bgTaskStdTimeDay)
	In := now.Add(-1 * time.Hour)
	InToBeTested := In.Format(bgStaticTime)
	InWrong := In.Format(bgTaskStdTimeDay)
	InTomorrow := In.Add(24 * time.Hour)
	InExpected := InTomorrow.Format(bgTaskStdTimeDay)
	bg := NewBg()
	got := bg.getCorrectDate(InToBeTested)
	if got != InExpected {
		t.Fail()
	}
	if got == InWrong {
		t.Fail()
	}

	In = now.Add(1 * time.Hour)
	InToBeTested = In.Format(bgStaticTime)
	InExpected = In.Format(bgTaskStdTimeDay)
	//bg := NewBg()
	got = bg.getCorrectDate(InToBeTested)
	if got != InExpected {
		t.Fail()
	}

}

func TestWrongDates(t *testing.T) {
	bg := NewBg().SetErrMsg("HAS PANICKED")
	bg.getCorrectDate("invalid")
	if len(bg.Errors) == 0 {
		t.Fail()
	}
	if len(bg.StringifyErrors()) == 0 {
		t.Fail()
	}
}

func TestCatchPanic(t *testing.T) {
	var panicVar string
	bg := NewBg().SetLogger(func(val string) error { panicVar = val; return nil })
	var dailyTasks []*Task
	task1 := &Task{Key: "unik4", RelativeTime: "04:46", TaskFn: func() { p("EVERYDAY TASK 2:46AM") }}
	dailyTasks = append(dailyTasks, task1)
	bg.RegisterDailyTasks(dailyTasks)
	bg.Start()
	if bg.GetDailyTaskByKey("unik4") == nil {
		t.Fail()
	}
	go func() {
		select {
		case <-time.After(1 * time.Second):
			bg.CancelTask("unik1")
			_, ok := bg.heartBeatTasks["unik1"]
			if ok {
				t.Fail()
			}
			bg.CancelDailyTask("unik2")
			bg.signals <- syscall.SIGINT
		case <-time.After(2 * time.Second):
			bg.signals <- syscall.SIGINT
		}
	}()

	bg.Wait()
	if !strings.Contains(panicVar, "bgTask has panicked") {
		t.Fail()
	}
}

func TestMisc(t *testing.T) {
	//	bg := NewBg().SetDevel()

}
