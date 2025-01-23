//
// File: timer.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:03:56
// Last Modified By: xiaoma20082008 (mmccxx2519@gmail.com>)
// ------------------------------------------------------------------------
//
// Copyright (C) xiaoma20082008. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package timingwheel

import (
	"fmt"
	"sync"
	"time"
)

type hierarchicalTimer struct {
	Timer
	wheel   timingWheel
	startMs uint64
	fireMs  uint64
	started chan bool
	closed  chan bool
	counter uint64
	queue   *hierarchicalBucketList
	lock    sync.Mutex
}

func (timer *hierarchicalTimer) New(fn TimerTask, delay time.Duration) Timeout {
	timer.startup()
	timeout := newTimeout(timer, fn, uint64(time.Now().UnixMilli()), uint64(delay.Milliseconds()))
	timer.addTimeout(timeout)
	return timeout
}

func (timer *hierarchicalTimer) Shutdown() []Timeout {
	timer.closed <- true
	return nil
}

func (timer *hierarchicalTimer) Size() uint64 {
	return timer.counter
}

func (timer *hierarchicalTimer) Close() error {
	timer.Shutdown()
	return nil
}

func (timer *hierarchicalTimer) addTimeout(timeout *hierarchicalTimeout) {
	if !timer.wheel.add(timeout) && !timeout.IsCancelled() {
		go timeout.task(timeout)
	}
}

func (timer *hierarchicalTimer) inc() {
	timer.lock.Lock()
	defer timer.lock.Unlock()
	timer.counter++
}

func (timer *hierarchicalTimer) dec() {
	timer.lock.Lock()
	defer timer.lock.Unlock()
	timer.counter--
}

func (timer *hierarchicalTimer) startup() {
	go timer.loop()
	<-timer.started
}

func (timer *hierarchicalTimer) loop() {
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
			timer.advance(timer.fireMs)
		}
	}
}

func (timer *hierarchicalTimer) advance(timeoutMs uint64) bool {
	fmt.Println("advance...")
	time.Sleep(time.Duration(timeoutMs) * time.Millisecond)
	bucket := timer.wheel.queue.pop(timeoutMs)
	if bucket != nil {
		for bucket != nil {
			timer.wheel.advance(bucket.expireMs)
			bucket.flush(timer.addTimeout)
			bucket = timer.wheel.queue.pop(timeoutMs)
		}
		return true
	} else {
		return false
	}
}

func newTimer(option *TimerOption) *hierarchicalTimer {
	timer := new(hierarchicalTimer)
	timer.wheel = *newWheel(option.TickMs, int(option.WheelSize), uint64(time.Now().UnixMilli()))
	timer.queue = newHierarchicalBucketList()
	timer.fireMs = option.TimeoutMs
	timer.started = make(chan bool)
	timer.closed = make(chan bool)
	return timer
}
