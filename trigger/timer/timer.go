package timer

import (
	"context"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/carlescere/scheduler"
	"strconv"
)

var log = logger.GetLogger("trigger-flogo-timer")

type TimerTrigger struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	timers   []*scheduler.Job

	handlers []*trigger.Handler
}

//NewFactory create a new Trigger factory
func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &TimerFactory{metadata: md}
}

// TimerFactory Timer Trigger factory
type TimerFactory struct {
	metadata *trigger.Metadata
}

//New Creates a new trigger instance for a given id
func (t *TimerFactory) New(config *trigger.Config) trigger.Trigger {
	return &TimerTrigger{metadata: t.metadata, config: config}
}

// Metadata implements trigger.Trigger.Metadata
func (t *TimerTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

// Init implements trigger.Init
func (t *TimerTrigger) Initialize(ctx trigger.InitContext) error {
	t.handlers = ctx.GetHandlers()
	return nil
}

// Start implements ext.Trigger.Start
func (t *TimerTrigger) Start() error {
	log.Debug("Start")
	handlers := t.handlers

	log.Debug("Processing handlers")
	for _, handler := range handlers {
		t.scheduleRepeating(handler)
	}

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *TimerTrigger) Stop() error {

	log.Debug("Stopping endpoints")
	for _, timer := range t.timers {

		if timer.IsRunning() {
			//log.Debug("Stopping timer for : ", k)
			timer.Quit <- true
		}
	}

	t.timers = nil

	return nil
}

func (t *TimerTrigger) scheduleRepeating(endpoint *trigger.Handler) {
	log.Info("Scheduling a repeating job")

	fn2 := func() {
		log.Debug("-- Starting \"Repeating\" (repeat) timer action")

		_, err := endpoint.Handle(context.Background(), nil)
		if err != nil {
			log.Error("Error running handler: ", err.Error())
		}
	}

	var interval int = 0
	if seconds := endpoint.GetStringSetting("seconds"); seconds != "" {
		seconds, _ := strconv.Atoi(seconds)
		interval = interval + seconds
	}
	if minutes := endpoint.GetStringSetting("minutes"); minutes != "" {
		minutes, _ := strconv.Atoi(minutes)
		interval = interval + minutes*60
	}
	if hours := endpoint.GetStringSetting("hours"); hours != "" {
		hours, _ := strconv.Atoi(hours)
		interval = interval + hours*3600
	}

	timerJob := scheduler.Every(interval).Seconds()

	if endpoint.GetStringSetting("notImmediate") == "true" {
		timerJob.NotImmediately()
	}

	_, err := timerJob.Run(fn2)
	if err != nil {
		log.Error("Error scheduleRepeating (first) flo err: ", err.Error())
	}

	t.timers = append(t.timers, timerJob)
}

type PrintJob struct {
	Msg string
}

func (j *PrintJob) Run() error {
	log.Debug(j.Msg)
	return nil
}
