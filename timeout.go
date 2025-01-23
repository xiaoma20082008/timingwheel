//
// File: timeout.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:04:00
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
	"strings"
	"sync"
	"time"
)

const (
	stInitialized = 0
	stExpired     = 1
	stCancelled   = 2
)

type hierarchicalTimeout struct {
	Timeout
	timer    *hierarchicalTimer
	task     TimerTask
	startMs  uint64
	expireMs uint64
	state    int
	lock     sync.Mutex
	bucket   *hierarchicalBucket
	prev     *hierarchicalTimeout
	next     *hierarchicalTimeout
}

func newTimeout(timer *hierarchicalTimer, fn TimerTask, startMs, delayMs uint64) *hierarchicalTimeout {
	timeout := new(hierarchicalTimeout)
	timeout.timer = timer
	timeout.startMs = startMs
	timeout.expireMs = startMs + delayMs
	timeout.task = fn
	timeout.state = stInitialized
	return timeout
}

func (timeout *hierarchicalTimeout) Timer() Timer { return timeout.timer }

func (timeout *hierarchicalTimeout) Task() TimerTask { return timeout.task }

func (timeout *hierarchicalTimeout) IsExpired() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	return timeout.state == stExpired
}

func (timeout *hierarchicalTimeout) IsCancelled() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	return timeout.state == stCancelled
}

func (timeout *hierarchicalTimeout) Cancel() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	if timeout.state == stInitialized {
		timeout.state = stCancelled
		return true
	} else {
		return false
	}
}

func (timeout *hierarchicalTimeout) Expire() bool {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	if timeout.state == stInitialized {
		timeout.state = stExpired
		return true
	} else {
		return false
	}
}

func (timeout *hierarchicalTimeout) StartMs() uint64 {
	return timeout.startMs
}

func (timeout *hierarchicalTimeout) DelayMs() uint64 {
	return timeout.ExpireMs() - timeout.StartMs()
}

func (timeout *hierarchicalTimeout) ExpireMs() uint64 {
	return timeout.expireMs
}

func (timeout *hierarchicalTimeout) String() string {
	deadline := timeout.expireMs - uint64(time.Now().UnixMilli())
	b := strings.Builder{}
	b.WriteString("Timeout(deadline: ")
	if deadline < 0 {
		b.WriteString(fmt.Sprintf("%d ms ago", -deadline))
	} else if deadline > 0 {
		b.WriteString(fmt.Sprintf("%d ms later", deadline))
	} else {
		b.WriteString("now")
	}
	if timeout.IsCancelled() {
		b.WriteString(", cancelled")
	}
	b.WriteString(fmt.Sprintf(", task: %p)", timeout.Task()))
	return b.String()
}

func (timeout *hierarchicalTimeout) remove() {
	bucket := timeout.bucket
	for bucket != nil {
		bucket.remove(timeout)
		bucket = timeout.bucket
	}
}
