//
// File: timingwheel.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:03:51
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
	"container/heap"
	"sync"
)

type timingWheel struct {
	lock      sync.Mutex
	tickMs    uint64
	size      int
	currentMs uint64
	interval  uint64
	buckets   hierarchicalBuckets
	overflow  *timingWheel
	queue     hierarchicalBuckets
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
		if bucket.expire(vid * wheel.tickMs) {
			wheel.queue.Push(bucket)
		}
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
	wheel.lock.Lock()
	defer wheel.lock.Unlock()
	wheel.overflow = newWheel(wheel.tickMs, wheel.size, wheel.currentMs)
}

func (wheel *timingWheel) advance(timeoutMs uint64) {
	if timeoutMs >= wheel.currentMs+wheel.tickMs {
		wheel.currentMs = timeoutMs - (timeoutMs % wheel.tickMs)
		if wheel.overflow != nil {
			wheel.overflow.advance(timeoutMs)
		}
	}
}

func newWheel(tickMs uint64, wheelSize int, startMs uint64) *timingWheel {
	wheel := new(timingWheel)
	wheel.tickMs = tickMs
	wheel.size = wheelSize
	wheel.currentMs = startMs - (startMs % tickMs)
	wheel.interval = tickMs * uint64(wheelSize)
	wheel.buckets = make([]*hierarchicalBucket, wheelSize)
	wheel.queue = make([]*hierarchicalBucket, 0)
	heap.Init(wheel.queue)
	for i := 0; i < wheelSize; i++ {
		wheel.buckets[i] = newBucket()
	}
	return wheel
}
