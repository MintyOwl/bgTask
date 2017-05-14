package bgTask

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Bg is a background scheduler to be configured
type Bg struct {
	heartBeatTasks map[string]*Task
	dailyTasks     map[string]*Task
	location       *time.Location
	devel          bool
	errMsg         string
	log            *Logger
	done           chan struct{}
	wg             *sync.WaitGroup
	Errors         []error
	signals        chan os.Signal
	cancel         chan string
	storage        string
	pTasks         []pendingTask
	mu             *sync.RWMutex
}

type Logger struct {
	Info func(string) error
	Err  func(string) error
}

type pendingTask struct {
	Key           string `json:"key"`
	CorrectedTime string `json:"corrected_time"` // correcedTime must be of the 'bgTaskStdTimeDay' format
}

// Task is task definition which is used to Register itself with registerTask and RegisterDailyTask api methods
// Duration is used with only registerTask, whereas RelativeTime is used with only RegisterDailyTask
// TaskFn is the callback function when triggered appropriately
type Task struct {
	Key, Duration, RelativeTime string
	TaskFn                      func() error
	dur                         time.Duration // dur is for daily task only
	hbCancel                    chan bool
	dailyCancel                 chan bool
}

// NewBg creates a new background scheduler, allowing to register new 'fire and forget' periodical Tasks. Hence during host process termination scheduler doesnt wait for periodical Tasks' goroutines to terminate ('fire and forget')
// Eg: bg := bgTask.NewBg(true) // true prints current time in terminal
// bg := bgTask.NewBg() // No argument in NewBg hence no printing of time in terminal
func NewBg() *Bg {
	bg := &Bg{
		Errors:         make([]error, 0),
		heartBeatTasks: make(map[string]*Task),
		dailyTasks:     make(map[string]*Task),
		devel:          false,
		errMsg:         "PANICED",
		done:           make(chan struct{}),
		location:       time.FixedZone("IST", 19800),
		signals:        make(chan os.Signal),
		cancel:         make(chan string),
		mu:             new(sync.RWMutex),
		pTasks:         make([]pendingTask, 0),
	}
	signal.Notify(bg.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return bg
}

// NewBgSync is like NewBg except that during host process termination, the scheduler will wait for periodical Tasks' goroutines to be terminated
// Eg: bg := bgTask.NewBgSync(true) // true prints current time in terminal
// bg := bgTask.NewBgSync() // No argument in NewBgSync hence no printing of time in terminal
func NewBgSync() *Bg {
	bg := &Bg{
		Errors:         make([]error, 0),
		heartBeatTasks: make(map[string]*Task),
		dailyTasks:     make(map[string]*Task),
		devel:          false,
		errMsg:         "PANICED",
		wg:             &sync.WaitGroup{},
		location:       time.FixedZone("IST", 19800),
		signals:        make(chan os.Signal),
		cancel:         make(chan string),
		mu:             new(sync.RWMutex),
		pTasks:         make([]pendingTask, 0),
	}
	signal.Notify(bg.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return bg
}

// Wait should be used with NewBg() only. This method must to be called by the client at the end of registration or using defer
// Eg: bg := bgTask.NewBg(); defer bg.Wait()
func (bg *Bg) Wait() {
	<-bg.done
}

// SyncWait should be used with NewBgSync() only. This method must to be called by the client at the end of registration
// Eg: bg := bgTask.NewBg(); defer bg.SyncWait()
func (bg *Bg) SyncWait() {
	bg.wg.Wait()
}
