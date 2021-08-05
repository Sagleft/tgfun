package tgfun

import "time"

type cronObject struct {
	TimerTime time.Duration
	Callback  cronCallback
}

func newCronHandler(callback cronCallback, timerTime time.Duration) cronObject {
	return cronObject{
		TimerTime: timerTime,
		Callback:  callback,
	}
}

type cronCallback func()

func (c *cronObject) run() {
	for true {
		c.Callback()
		time.Sleep(c.TimerTime)
	}
}
