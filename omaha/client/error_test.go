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

func init() {
	// use quicker backoff for testing
	backoffStart = time.Millisecond
	backoffTries = 3
}

type tmpErr struct{}

func (e tmpErr) Error() string   { return "fake temporary error" }
func (e tmpErr) Temporary() bool { return true }
func (e tmpErr) Timeout() bool   { return false }

func TestExpNetBackoff(t *testing.T) {
	tries := 0
	err := expNetBackoff(func() error {
		tries++
		if tries < 2 {
			return tmpErr{}
		}
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tries != 2 {
		t.Errorf("unexpected # of tries: %d", tries)
	}
}
