//
// File: timingwheel_test.go
// Project: timingwheel
// File Created: 2025-01-23
// Author: xiaoma20082008 (mmccxx2519@gmail.com)
//
// ------------------------------------------------------------------------
// Last Modified At: 2025-01-23 20:24:23
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
	"testing"
)

func TestTimer(t *testing.T) {
	timer := NewTimer()
	defer timer.Shutdown()
	done := make(chan bool)
	timeout1 := timer.New(func(timeout Timeout) {
		fmt.Println("running after 500ms")
		done <- true
	}, 500)
	fmt.Println(timeout1)
	<-done

	//tests := []struct {
	//	name string
	//	want Timer
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		if got := NewTimer(); !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("NewTimer() = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}
