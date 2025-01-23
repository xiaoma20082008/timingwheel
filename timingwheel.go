package github.com/xiaoma20082008/timingwheel

import (
	"sync"
	"time"
)

//==================================== apis ====================================

type TimerTask func(timeout Timeout)

type Timer interface {
	New(task TimerTask, delayMs uint64) Timeout
	Shutdown() []Timeout
}

type Timeout interface {
	Timer() Timer
	Task() TimerTask
	IsExpired() bool
	IsCancelled() bool
	Cancel() bool
	Expire() bool
}

func NewTimer() Timer {
	timer := new(hierarchicalTimer)
	timer.wheel = *newWheel(1, 1000, uint64(time.Now().UnixMilli()))
	timer.fireMs = 100
	timer.started = make(chan bool)
	timer.closed = make(chan bool)
	return timer
}

//================================= implements =================================

var counter int
var lock sync.Mutex

func inc() {
	lock.Lock()
	defer lock.Unlock()
	counter++
}

func dec() {
	lock.Lock()
	defer lock.Unlock()
	counter--
}

type timingWheel struct {
	lock      sync.Mutex
	tickMs    uint64
	size      int
	currentMs uint64
	interval  uint64
	buckets   buckets
	overflow  *timingWheel
}

func (wheel *timingWheel) add(timeout *hierarchicalTimeout) bool {
	expiration := timeout.expireMs
	if timeout.IsCancelled() {
		// Cancelled
		return false
	} else if expiration < wheel.currentMs+wheel.tickMs {
		// Expired
		return false
	} else if expiration < wheel.currentMs+wheel.interval {
		// Put it into it's bucket
		vid := expiration / wheel.tickMs
		bid := vid % uint64(wheel.size)
		bucket := wheel.buckets[bid]
		bucket.add(timeout)
		return true
	} else {
		// Out of the interval. Put it into the parent timer
		if wheel.overflow == nil {
			wheel.addWheel()
		}
		return wheel.overflow.add(timeout)
	}
}

func (wheel *timingWheel) addWheel() {
	lock.Lock()
	defer lock.Unlock()
	wheel.overflow = newWheel(wheel.tickMs, wheel.size, wheel.currentMs)
}

func newWheel(tickMs uint64, wheelSize int, startMs uint64) *timingWheel {
	wheel := new(timingWheel)
	wheel.tickMs = tickMs
	wheel.size = wheelSize
	wheel.currentMs = startMs - (startMs % tickMs)
	wheel.interval = tickMs * uint64(wheelSize)
	wheel.buckets = make(buckets, wheelSize)
	for i := 0; i < wheelSize; i++ {
		wheel.buckets[i] = newBucket()
	}
	return wheel
}

type hierarchicalTimer struct {
	Timer
	wheel   timingWheel
	startMs uint64
	fireMs  uint64
	started chan bool
	closed  chan bool
}

func (timer *hierarchicalTimer) New(fn TimerTask, delayMs uint64) Timeout {
	timer.Startup()
	timeout := newTimeout(timer, fn, uint64(time.Now().UnixMilli())+delayMs)
	if !timer.wheel.add(timeout) && !timeout.IsCancelled() {
		go fn(timeout)
	}
	return timeout
}

func (timer *hierarchicalTimer) Shutdown() []Timeout {
	timer.closed <- true
	return nil
}

func (timer *hierarchicalTimer) Startup() {
	go timer.Loop()
	<-timer.started
}

func (timer *hierarchicalTimer) Loop() {
	timer.startMs = uint64(time.Now().UnixMilli())
	timer.started <- true
LOOP:
	for {
		select {
		case v := <-timer.closed:
			if v {
				break LOOP
			}
		default:
			timer.Advance(timer.fireMs)
		}
	}
}

func (timer *hierarchicalTimer) Advance(timeoutMs uint64) {
}

type buckets []*hierarchicalTimeoutBucket

type hierarchicalTimeoutBucket struct {
	root *hierarchicalTimeoutBucket
}

func newBucket() *hierarchicalTimeoutBucket {
	return nil
}

func (bucket *hierarchicalTimeoutBucket) add(timeout *hierarchicalTimeout) {}

func (bucket *hierarchicalTimeoutBucket) del(timeout *hierarchicalTimeout) {}

const (
	st_INIT      = 0
	st_EXPIRED   = 1
	st_CANCELLED = 2
)

type hierarchicalTimeout struct {
	Timeout
	timer    *hierarchicalTimer
	task     TimerTask
	expireMs uint64
	state    int
	lock     sync.Mutex
}

func newTimeout(timer *hierarchicalTimer, fn TimerTask, expireMs uint64) *hierarchicalTimeout {
	timeout := new(hierarchicalTimeout)
	timeout.timer = timer
	timeout.expireMs = expireMs
	timeout.task = fn
	timeout.state = st_INIT
	return timeout
}

func (timeout *hierarchicalTimeout) Timer() Timer { return timeout.timer }

func (timeout *hierarchicalTimeout) Task() TimerTask { return timeout.task }

func (timeout *hierarchicalTimeout) IsExpired() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	return timeout.state == st_EXPIRED
}

func (timeout *hierarchicalTimeout) IsCancelled() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	return timeout.state == st_CANCELLED
}

func (timeout *hierarchicalTimeout) Cancel() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	timeout.state = st_CANCELLED
	return true
}

func (timeout *hierarchicalTimeout) Expire() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	timeout.state = st_EXPIRED
	timeout.task(timeout)
	return true
}
