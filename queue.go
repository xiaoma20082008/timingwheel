//
// File: queue.go
// Project: timingwheel
// File Created: 2025-01-24
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:13:06
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
	"time"
)

type delayQueue struct {
	lock sync.RWMutex
	cond *sync.Cond
	data hierarchicalBuckets
}

func (q *delayQueue) offer(bucket *hierarchicalBucket) {
	q.lock.Lock()
	defer q.lock.Unlock()
	heap.Push(&q.data, bucket)
	q.cond.Broadcast()
}

func (q *delayQueue) poll(duration time.Duration) *hierarchicalBucket {
	q.lock.Lock()
	defer q.lock.Unlock()
	for q.data.Len() == 0 && duration > 0 {
		q.cond.Wait()
	}
	if q.data.Len() == 0 {
		return nil
	} else {
		return heap.Pop(q.data).(*hierarchicalBucket)
	}
}

func newDelayQueue() *delayQueue {
	dq := new(delayQueue)
	dq.lock = sync.RWMutex{}
	dq.cond = sync.NewCond(&dq.lock)
	dq.data = make(hierarchicalBuckets, 16)
	return dq
}
