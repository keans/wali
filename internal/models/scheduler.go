package models

import (
	"time"
)

type OnTickCallback = func(t time.Time)

type Scheduler struct {
	Interval int64 // in ms
	ticker   *time.Ticker
	OnTick   OnTickCallback
}

func NewScheduler(interval int64, onTick OnTickCallback,
	autoStart bool) *Scheduler {

	scheduler := &Scheduler{
		Interval: interval,
		OnTick:   onTick,
	}

	if autoStart {
		scheduler.Start()
	}

	return scheduler
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(time.Duration(s.Interval) * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-s.ticker.C:
				s.OnTick(t)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
}
