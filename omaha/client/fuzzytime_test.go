// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"testing"
	"time"
)

func TestFuzzyDuration(t *testing.T) {
	const d = time.Minute
	for i := 0; i < 1000; i++ {
		f := FuzzyDuration(d, d)
		if f < d/2 {
			t.Errorf("%d < %d", f, d/2)
		} else if f > d+d/2 {
			t.Errorf("%d > %d", f, d+d/2)
		}
	}
}
