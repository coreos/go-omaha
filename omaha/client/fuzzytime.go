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
	"math/rand"
	"time"
)

func init() {
	// Ensure seeding the prng is never forgotten, that would defeat
	// the whole point of using fuzzy timers to guard against a DoS.
	rand.Seed(time.Now().UnixNano())
}

// FuzzyDuration randomizes the duration d within the range specified
// by fuzz. Specifically the value range is: [d-(fuzz/2), d+(fuzz/2)]
// The result will never be negative.
func FuzzyDuration(d, fuzz time.Duration) time.Duration {
	if fuzz < 0 {
		return d
	}
	// apply range [-fuzz/2, fuzz/2]
	d += time.Duration(rand.Int63n(int64(fuzz)+1) - (int64(fuzz) / 2))
	if d < 0 {
		return 0
	}
	return d
}

// FuzzyAfter waits for the fuzzy duration to elapse and then sends the
// current time on the returned channel. See FuzzyDuration.
func FuzzyAfter(d, fuzz time.Duration) <-chan time.Time {
	return time.After(FuzzyDuration(d, fuzz))
}

// FuzzySleep pauses the current goroutine for the fuzzy duration d.
// See FuzzyDuration.
func FuzzySleep(d, fuzz time.Duration) {
	time.Sleep(FuzzyDuration(d, fuzz))
}
