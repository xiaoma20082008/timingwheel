//
// File: api.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-24 00:04:12
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
	"io"
	"time"
)

type TimerTask func(timeout Timeout)

type Timer interface {
	io.Closer
	New(task TimerTask, delay time.Duration) Timeout
	Shutdown() []Timeout
	Size() uint64
}

type Timeout interface {
	fmt.Stringer
	Timer() Timer
	Task() TimerTask
	IsExpired() bool
	IsCancelled() bool
	Cancel() bool
	Expire() bool
	StartMs() uint64
	DelayMs() uint64
	ExpireMs() uint64
}

type TimerOption struct {
	Name      string
	TickMs    uint64
	WheelSize uint64
	TimeoutMs uint64
	Extended  map[string]any
}

type Options func(*TimerOption)

func NewTimer(options ...Options) Timer {
	cfg := &TimerOption{
		Name:      "TimingWheelTimer",
		TickMs:    1,
		WheelSize: 1000,
		TimeoutMs: 100,
		Extended:  make(map[string]any),
	}
	for _, opt := range options {
		opt(cfg)
	}
	return newTimer(cfg)
}
