//
// File: bucket.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:04:10
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

type hierarchicalBucket struct {
	root     *hierarchicalTimeout
	lock     sync.Mutex
	expireMs uint64
}

type hierarchicalBuckets []*hierarchicalBucket

func (bl hierarchicalBuckets) Len() int {
	return len(bl)
}
func (bl hierarchicalBuckets) Less(i, j int) bool {
	return bl[i].expireMs < bl[j].expireMs
}
func (bl hierarchicalBuckets) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}
func (bl hierarchicalBuckets) Push(x any) {
	heap.Push(bl, x)
}
func (bl hierarchicalBuckets) Pop() any {
	return heap.Pop(bl).(any)
}
func (bl hierarchicalBuckets) pop(timeoutMs uint64) *hierarchicalBucket {
	return nil
}

// hierarchicalBucketList wrap for []*hierarchicalBucket
type hierarchicalBucketList struct {
	expireMs uint64
	lock     sync.Mutex
	cond     *sync.Cond
	list     hierarchicalBuckets
}

func (bl *hierarchicalBucketList) Len() int {
	return len(bl.list)
}
func (bl *hierarchicalBucketList) Less(i, j int) bool {
	return bl.list[i].expireMs < bl.list[j].expireMs
}
func (bl *hierarchicalBucketList) Swap(i, j int) {
	bl.list[i], bl.list[j] = bl.list[j], bl.list[i]
}
func (bl *hierarchicalBucketList) Push(x any) {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	heap.Push(bl.list, x)
	bl.cond.Broadcast()
}
func (bl *hierarchicalBucketList) Pop() any {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	for bl.list.Len() == 0 {
		bl.cond.Wait()
	}
	return bl.pop(0)
}

func (bl *hierarchicalBucketList) pop(timeoutMs uint64) *hierarchicalBucket {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	now := uint64(time.Now().UnixMilli())
	for bl.list.Len() == 0 && (timeoutMs-now) > timeoutMs {
		bl.cond.Wait()
	}
	var bucket *hierarchicalBucket
	if bl.list.Len() == 0 {
		bucket = nil
	} else {
		bucket = heap.Pop(bl).(*hierarchicalBucket)
	}
	return bucket
}

func newHierarchicalBucketList() *hierarchicalBucketList {
	list := new(hierarchicalBucketList)
	list.lock = sync.Mutex{}
	list.cond = sync.NewCond(&list.lock)
	list.expireMs = 0
	list.list = make(hierarchicalBuckets, 0)
	return list
}

func newBucket() *hierarchicalBucket {
	bucket := new(hierarchicalBucket)
	bucket.lock = sync.Mutex{}
	bucket.root = newTimeout(nil, nil, 0, 0)
	bucket.root.prev = bucket.root
	bucket.root.next = bucket.root
	return bucket
}

func (bucket *hierarchicalBucket) expire(timeoutMs uint64) bool {
	return false
}

func (bucket *hierarchicalBucket) flush(f func(timeout *hierarchicalTimeout)) {

}

func (bucket *hierarchicalBucket) add(timeout *hierarchicalTimeout) {
	done := false
	for !done {
		timeout.remove()

		bucket.lock.Lock()
		defer bucket.lock.Unlock()

		if timeout.bucket == nil {
			// put the timeout to the end of the bucket.
			tail := bucket.root.prev

			timeout.prev = tail
			timeout.next = bucket.root
			timeout.bucket = bucket

			tail.next = timeout
			bucket.root.prev = timeout
			timeout.timer.inc()
			done = true
		}
	}
}

func (bucket *hierarchicalBucket) remove(timeout *hierarchicalTimeout) {
	timeout.lock.Lock()
	defer timeout.lock.Unlock()
	if timeout.bucket == bucket {
		timeout.next.prev = timeout.prev
		timeout.prev.next = timeout.next
		timeout.prev = nil
		timeout.next = nil
		timeout.bucket = nil
		timeout.timer.dec()
	}
}
